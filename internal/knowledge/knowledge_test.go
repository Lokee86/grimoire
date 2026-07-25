package knowledge

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractSectionsPreservesExactSpansAndStableIDs(t *testing.T) {
	content := "# Decision\nWhy this exists.\n\n## Tradeoffs\nUse BM25.\n## Follow-up\nLater.\n"
	sections := extractSections("docs/design.md", content)
	if len(sections) != 3 {
		t.Fatalf("sections = %d, want 3", len(sections))
	}
	for _, section := range sections {
		if got := content[section.StartByte:section.EndByte]; got != section.Text {
			t.Fatalf("section %q span mismatch", section.Heading)
		}
		if section.StartLine > section.EndLine {
			t.Fatalf("invalid line span: %+v", section)
		}
	}
	if sections[1].HeadingPath[0] != "Decision" || sections[1].HeadingPath[1] != "Tradeoffs" {
		t.Fatalf("heading path = %#v", sections[1].HeadingPath)
	}
	changed := extractSections("docs/design.md", strings.Replace(content, "Use BM25.", "Use deterministic BM25.", 1))
	if sections[0].ID != changed[0].ID || sections[1].ID != changed[1].ID {
		t.Fatalf("heading-derived section IDs changed: %q %q", sections[0].ID, changed[0].ID)
	}
}

func TestBuildDiscoversKnowledgeAndReusesUnchangedDocuments(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "docs/architecture.md", "# Architecture\nThe graph owns discovery.\n")
	writeTestFile(t, root, "notes.txt", "A plain design note about rationale.\n")
	writeTestFile(t, root, "internal/implementation.go", "package internal\nfunc ResolveDamage() {}\n")
	writeTestFile(t, root, ".gitignore", "ignored.md\n")
	writeTestFile(t, root, "ignored.md", "# ignored\n")

	first, firstStats, err := Build(root, nil, BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Documents) != 2 || firstStats.Updated != 2 {
		t.Fatalf("first build documents/stats = %d/%+v", len(first.Documents), firstStats)
	}
	second, secondStats, err := Build(root, &first, BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if secondStats.Reused != 2 || secondStats.Updated != 0 || len(second.Documents) != 2 {
		t.Fatalf("reuse stats = %+v", secondStats)
	}
	if !bytes.Equal(mustJSON(t, first.Documents), mustJSON(t, second.Documents)) {
		t.Fatalf("reused documents changed")
	}
	writeTestFile(t, root, "docs/architecture.md", "# Architecture\nThe graph owns discovery and rationale.\n")
	third, thirdStats, err := Build(root, &second, BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if thirdStats.Reused != 1 || thirdStats.Updated != 1 || third.Documents[0].Hash == second.Documents[0].Hash {
		t.Fatalf("incremental update stats/documents = %+v/%+v", thirdStats, third.Documents)
	}
}

func TestPersistentIndexIsStableAndRoundTrips(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "docs/rationale.md", "# Rationale\nPersistent knowledge state is reusable.\n")
	index, _, err := Build(root, nil, BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	state := filepath.Join(root, ".state")
	if err := Save(state, index); err != nil {
		t.Fatal(err)
	}
	firstBytes, err := os.ReadFile(filepath.Join(state, "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(state)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(mustJSON(t, index), mustJSON(t, loaded)) {
		t.Fatalf("round trip changed index")
	}
	if err := Save(state, loaded); err != nil {
		t.Fatal(err)
	}
	secondBytes, err := os.ReadFile(filepath.Join(state, "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstBytes, secondBytes) {
		t.Fatalf("persistent serialization is not deterministic")
	}
}

func TestSearchReturnsExactCitationsFiltersAndDeterministicResults(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "docs/architecture/graph.md", "# Graph ownership\nThe graph owns code discovery.\n\n## Rationale\nDesign rationale keeps semantic retrieval supplemental.\n")
	writeTestFile(t, root, "docs/planning/roadmap.md", "# Roadmap\nThe graph-first migration is planned.\n")
	index, _, err := Build(root, nil, BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	response, err := Search(context.Background(), index, "graph rationale", SearchOptions{TopK: 5, Kind: KindArchitecture, Heading: "Rationale"})
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Results) != 1 {
		t.Fatalf("results = %+v", response.Results)
	}
	result := response.Results[0]
	if result.Text != "## Rationale\nDesign rationale keeps semantic retrieval supplemental.\n" || result.StartByte >= result.EndByte {
		t.Fatalf("citation = %+v", result)
	}
	if !strings.HasPrefix(result.Handle, "knowledge://docs/architecture/graph.md#") || len(result.Reasons) == 0 {
		t.Fatalf("citation handle/reasons = %+v", result)
	}
	again, err := Search(context.Background(), index, "graph rationale", SearchOptions{TopK: 5, Kind: KindArchitecture, Heading: "Rationale"})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(mustJSON(t, response), mustJSON(t, again)) {
		t.Fatalf("search result is not deterministic")
	}
	if _, err := Inspect(index, "", result.Handle); err != nil {
		t.Fatalf("inspect citation: %v", err)
	}
}

func TestCodeLinkHintsRequireExactRepositoryEvidence(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "internal/api.go", "package api\nfunc ResolveDamage() {}\nconst EventName = \"DamageResolved\"\nconst Route = \"/api/damage\"\n")
	writeTestFile(t, root, "config/runtime.yaml", "max_retries: 3\n")
	writeTestFile(t, root, "docs/design.md", "# Contract\nResolveDamage handles DamageResolved at /api/damage using max_retries.\nInventedName is not linked.\n")
	index, _, err := Build(root, nil, BuildOptions{IncludeConfig: true})
	if err != nil {
		t.Fatal(err)
	}
	var section Section
	for _, document := range index.Documents {
		if document.Path == "docs/design.md" {
			section = document.Sections[0]
		}
	}
	if section.ID == "" {
		t.Fatalf("design document not discovered: %+v", index.Documents)
	}
	links := section.CodeLinks
	found := map[string]bool{}
	for _, link := range links {
		found[link.Kind+":"+link.Value] = true
		if link.SourcePath == "" || link.Evidence == "" {
			t.Fatalf("incomplete code link: %+v", link)
		}
	}
	for _, expected := range []string{"symbol:ResolveDamage", "contract:DamageResolved", "endpoint:/api/damage", "config-contract:max_retries"} {
		if !found[expected] {
			t.Fatalf("missing %q in %#v", expected, links)
		}
	}
	for _, link := range links {
		if link.Value == "InventedName" {
			t.Fatalf("invented semantic link: %+v", link)
		}
	}
}

func TestSearchWorksWithoutVectorAndReportsVectorFailure(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "docs/notes.md", "# Lexical\nDeterministic retrieval works without a model.\n")
	index, _, err := Build(root, nil, BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	plain, err := Search(context.Background(), index, "deterministic retrieval", SearchOptions{TopK: 1})
	if err != nil || plain.VectorUsed || len(plain.Results) != 1 {
		t.Fatalf("plain search = %+v, err=%v", plain, err)
	}
	failed, err := Search(context.Background(), index, "deterministic retrieval", SearchOptions{TopK: 1, Vector: failingRanker{}})
	if err != nil || failed.VectorUsed || failed.VectorError == "" || len(failed.Results) != 1 {
		t.Fatalf("fallback search = %+v, err=%v", failed, err)
	}
}

type failingRanker struct{}

func (failingRanker) Rank(context.Context, string, []Section) (map[string]float64, error) {
	return nil, errors.New("vector state unavailable")
}

func writeTestFile(t *testing.T, root, relative, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
