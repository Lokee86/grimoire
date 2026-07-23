package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Lokee86/grimoire/internal/compiler"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/retrieve"
	"github.com/Lokee86/grimoire/internal/selection"
)

type retrievalQualityCase struct {
	Name          string   `json:"name"`
	Query         string   `json:"query"`
	SemanticPaths []string `json:"semantic_paths"`
	MustInclude   []string `json:"must_include"`
	MustSources   []string `json:"must_sources"`
	Budget        int      `json:"budget"`
}

func TestRetrievalQualityFixtures(t *testing.T) {
	root := filepath.Join("testdata", "retrieval-quality")
	state := filepath.Join(t.TempDir(), "state")
	if err := Run([]string{"index", "--root", root, "--state", state}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	snapshot, err := index.Load(state)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cases []retrievalQualityCase
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}

	for _, fixture := range cases {
		t.Run(fixture.Name, func(t *testing.T) {
			semantic := fixtureSemanticCandidates(snapshot, fixture.Query, fixture.SemanticPaths)
			exact := retrieve.Exact(snapshot, fixture.Query, 20)
			merged := mergeFixtureCandidates(exact, semantic)
			curated := selection.Curate(snapshot, merged)
			pkg, err := compiler.Compile(
				fixture.Query, fixture.Budget, snapshot.Version, snapshot.Tokenizer,
				fixtureSources(curated), curated,
			)
			if err != nil {
				t.Fatal(err)
			}
			assertFixtureCoverage(t, fixture, pkg)
			first, err := compiler.Marshal(pkg)
			if err != nil {
				t.Fatal(err)
			}
			repeated := selection.Curate(snapshot, mergeFixtureCandidates(exact, semantic))
			secondPkg, err := compiler.Compile(
				fixture.Query, fixture.Budget, snapshot.Version, snapshot.Tokenizer,
				fixtureSources(repeated), repeated,
			)
			if err != nil {
				t.Fatal(err)
			}
			second, err := compiler.Marshal(secondPkg)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(first, second) {
				t.Fatal("fixture package is not deterministic")
			}
		})
	}
}

func TestRetrievalQualityPromotesRetrievedAdjacentEvidence(t *testing.T) {
	chunk := func(id, path, text string, start, end int) index.Chunk {
		return index.Chunk{ID: id, Path: path, StartLine: start, EndLine: end, TokenCount: 24, Text: text}
	}
	primary := chunk("alpha-main", "internal/alpha.go", "package internal\nfunc AlphaOwner() { AlphaCore() }", 1, 48)
	required := chunk("alpha-required", "internal/alpha.go", "package internal\nfunc ValidateAlphaSnapshot() { CheckAlphaIdentity(); CheckAlphaDimensions() }", 49, 96)
	snapshot := index.Snapshot{
		Version:   index.FormatVersion,
		Tokenizer: "o200k_base",
		Files: []index.FileRecord{
			{Path: "internal/alpha.go", Chunks: []index.Chunk{primary, required}},
			{Path: "internal/beta.go", Chunks: []index.Chunk{chunk("beta", "internal/beta.go", "package internal\nfunc BetaOwner() {}", 1, 48)}},
			{Path: "internal/gamma.go", Chunks: []index.Chunk{chunk("gamma", "internal/gamma.go", "package internal\nfunc GammaOwner() {}", 1, 48)}},
			{Path: "internal/delta.go", Chunks: []index.Chunk{chunk("delta", "internal/delta.go", "package internal\nfunc DeltaOwner() {}", 1, 48)}},
			{Path: "internal/epsilon.go", Chunks: []index.Chunk{chunk("epsilon", "internal/epsilon.go", "package internal\nfunc EpsilonOwner() {}", 1, 48)}},
			{Path: "internal/zeta.go", Chunks: []index.Chunk{chunk("zeta", "internal/zeta.go", "package internal\nfunc ZetaOwner() {}", 1, 48)}},
		},
	}
	candidates := []retrieve.Candidate{
		{Chunk: primary, Source: "vector", Rank: 1},
		{Chunk: snapshot.Files[1].Chunks[0], Source: "vector", Rank: 2},
		{Chunk: snapshot.Files[2].Chunks[0], Source: "vector", Rank: 3},
		{Chunk: snapshot.Files[3].Chunks[0], Source: "vector", Rank: 4},
		{Chunk: snapshot.Files[4].Chunks[0], Source: "vector", Rank: 5},
		{Chunk: snapshot.Files[5].Chunks[0], Source: "vector", Rank: 6},
		{Chunk: required, Source: "vector", Rank: 7},
	}
	curated := selection.Curate(snapshot, candidates)
	requiredIndex, epsilonIndex := -1, -1
	for index, candidate := range curated {
		switch candidate.Chunk.ID {
		case required.ID:
			requiredIndex = index
		case "epsilon":
			epsilonIndex = index
		}
	}
	if requiredIndex < 0 || epsilonIndex < 0 || requiredIndex > epsilonIndex {
		t.Fatalf("retrieved adjacent evidence was not promoted: %+v", curated)
	}
}

func fixtureSemanticCandidates(snapshot index.Snapshot, query string, paths []string) []retrieve.Candidate {
	ranked := retrieve.Search(snapshot, query, 0)
	byPath := make(map[string]index.Chunk)
	for _, candidate := range ranked {
		if _, exists := byPath[candidate.Chunk.Path]; !exists {
			byPath[candidate.Chunk.Path] = candidate.Chunk
		}
	}
	allByPath := make(map[string][]index.Chunk)
	for _, chunk := range snapshot.AllChunks() {
		allByPath[chunk.Path] = append(allByPath[chunk.Path], chunk)
	}
	result := make([]retrieve.Candidate, 0, len(paths))
	for rank, path := range paths {
		chunk, exists := byPath[path]
		if !exists && len(allByPath[path]) > 0 {
			chunk, exists = allByPath[path][0], true
		}
		if exists {
			result = append(result, retrieve.Candidate{
				Chunk: chunk, Score: float64(len(paths) - rank), Source: "vector", Rank: rank + 1,
				Reasons: []string{"retrieval-quality fixture semantic rank"},
			})
		}
	}
	return result
}

func mergeFixtureCandidates(groups ...[]retrieve.Candidate) []retrieve.Candidate {
	seen := make(map[string]struct{})
	var merged []retrieve.Candidate
	for _, group := range groups {
		for _, candidate := range group {
			if _, exists := seen[candidate.Chunk.ID]; exists {
				continue
			}
			seen[candidate.Chunk.ID] = struct{}{}
			merged = append(merged, candidate)
		}
	}
	return merged
}

func fixtureSources(candidates []retrieve.Candidate) []string {
	seen := make(map[string]struct{})
	var sources []string
	for _, candidate := range candidates {
		if _, exists := seen[candidate.Source]; exists {
			continue
		}
		seen[candidate.Source] = struct{}{}
		sources = append(sources, candidate.Source)
	}
	return sources
}

func assertFixtureCoverage(t *testing.T, fixture retrievalQualityCase, pkg compiler.Package) {
	t.Helper()
	paths := make(map[string]struct{})
	seenRanges := make(map[string]struct{})
	for _, selected := range pkg.Selections {
		paths[selected.Path] = struct{}{}
		key := fmt.Sprintf("%s:%d:%d", selected.Path, selected.StartLine, selected.EndLine)
		if _, exists := seenRanges[key]; exists {
			t.Fatalf("duplicate selected range %s", key)
		}
		seenRanges[key] = struct{}{}
	}
	for _, path := range fixture.MustInclude {
		if _, exists := paths[path]; !exists {
			t.Fatalf("required path %s missing from %+v", path, pkg.Selections)
		}
	}
	sources := make(map[string]struct{})
	for _, source := range pkg.RetrievalSources {
		sources[source] = struct{}{}
	}
	for _, source := range fixture.MustSources {
		if _, exists := sources[source]; !exists {
			t.Fatalf("required source %s missing from %+v", source, pkg.RetrievalSources)
		}
	}
}
