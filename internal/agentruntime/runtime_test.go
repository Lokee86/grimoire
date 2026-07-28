package agentruntime

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/Lokee86/grimoire/internal/agentquery"
	"github.com/Lokee86/grimoire/internal/knowledge"
	"github.com/Lokee86/grimoire/internal/repostate"
)

func TestExecuteReturnsFullEvidenceWithoutSession(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "docs/design.md", "# Graph ownership\nArcana owns repository graph traversal.\n")
	response, err := Execute(context.Background(), Request{
		Request: agentquery.Request{Mode: "search", Root: root, Query: "graph ownership", Limit: 4},
	}, testOptions(root))
	if err != nil {
		t.Fatal(err)
	}
	if len(response.SourceMatches) != 1 {
		t.Fatalf("source response = %#v", response.SourceMatches)
	}
	if len(response.DocumentMatches) == 0 || response.DocumentMatches[0].Path != "docs/design.md" {
		t.Fatalf("document response = %#v", response.DocumentMatches)
	}
	if response.Delta != nil {
		t.Fatalf("unexpected session delta: %#v", response.Delta)
	}
}

func TestDocumentMatchesRemainSeparateAndInspectable(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "internal/graph.go", "package internal\n\nfunc TraverseGraph() {}\n")
	writeTestFile(t, root, "docs/design.md", "# Graph ownership\nArcana owns repository graph traversal.\n")
	response, err := Execute(context.Background(), Request{
		Request: agentquery.Request{Mode: "search", Root: root, Query: "graph ownership", Limit: 4},
	}, testOptions(root))
	if err != nil {
		t.Fatal(err)
	}
	if len(response.SourceMatches) != 1 || len(response.DocumentMatches) != 1 {
		t.Fatalf("discovery lanes were not separate: source=%#v documents=%#v", response.SourceMatches, response.DocumentMatches)
	}
	if response.DocumentMatches[0].Handle == "" || response.DocumentMatches[0].StartLine <= 0 {
		t.Fatalf("document match has no inspectable handle and range: %#v", response.DocumentMatches[0])
	}

	inspection, err := Execute(context.Background(), Request{
		Request: agentquery.Request{
			Mode: "inspect", Root: root, Anchor: response.DocumentMatches[0].Handle, Limit: 4,
		},
	}, testOptions(root))
	if err != nil {
		t.Fatal(err)
	}
	if len(inspection.DocumentMatches) != 1 || inspection.DocumentMatches[0].Handle != response.DocumentMatches[0].Handle {
		t.Fatalf("document handle did not inspect exact evidence: %#v", inspection.DocumentMatches)
	}
	if len(inspection.SourceMatches) != 0 || len(inspection.SymbolMatches) != 0 {
		t.Fatalf("document inspection leaked into code lanes: %#v", inspection)
	}
}

func TestExecuteResolvesSessionNodeHandlesBeforeQuery(t *testing.T) {
	root := t.TempDir()
	originalHandle := "grimoire:v1:stable-node-handle"
	var requests []agentquery.Request
	options := testOptions(root)
	options.ExecuteQuery = func(_ context.Context, request agentquery.Request) (agentquery.Response, error) {
		requests = append(requests, request)
		response := agentquery.Response{
			Schema:   agentquery.SchemaVersion,
			Mode:     request.Mode,
			Snapshot: agentquery.Snapshot{Source: "source-1"},
		}
		if len(requests) == 1 {
			span := agentquery.Range{
				Path: "internal/target.go", StartLine: 10, EndLine: 12,
				Handle: agentquery.NewSourceHandle("source-1", "internal/target.go", 10, 12),
			}
			response.SourceMatches = []agentquery.Result{{
				Rank: 1, Provider: "lexicon", Kind: "function",
				Node: agentquery.Node{
					Handle: agentquery.Handle{Value: originalHandle, Provider: "lexicon", Snapshot: "lexicon-1", NodeIdentity: "node-one"},
					Kind:   "function", Name: "Target", Path: "internal/target.go", Span: &span,
				},
			}}
		}
		return response, nil
	}

	first, err := Execute(context.Background(), Request{
		Request: agentquery.Request{Mode: "search", Root: root, Query: "target", Limit: 4},
		Session: "stable-handles",
	}, options)
	if err != nil {
		t.Fatal(err)
	}
	if first.Delta == nil || len(first.Delta.NewNodes) != 1 || len(first.Delta.NewSourceRanges) != 1 {
		t.Fatalf("search delta = %#v", first.Delta)
	}
	opaque := first.Delta.NewNodes[0].Handle.String()
	if !strings.HasPrefix(opaque, "g1_") {
		t.Fatalf("session handle = %q", opaque)
	}

	_, err = Execute(context.Background(), Request{
		Request: agentquery.Request{Mode: "inspect", Root: root, Handles: []string{opaque}, Limit: 4},
		Session: "stable-handles",
	}, options)
	if err != nil {
		t.Fatal(err)
	}
	if len(requests) < 2 || len(requests[1].Handles) != 1 || requests[1].Handles[0] != originalHandle {
		t.Fatalf("inspect request did not restore original handle: %#v", requests)
	}

	_, err = Execute(context.Background(), Request{
		Request: agentquery.Request{Mode: "trace", Root: root, Anchor: opaque, Limit: 4},
		Session: "stable-handles",
	}, options)
	if err != nil {
		t.Fatal(err)
	}
	if len(requests) < 3 || requests[2].Anchor != originalHandle {
		t.Fatalf("trace request did not restore original handle: %#v", requests)
	}

	sourceOpaque := first.Delta.NewSourceRanges[0].Handle.String()
	_, err = Execute(context.Background(), Request{
		Request: agentquery.Request{Mode: "inspect", Root: root, Handles: []string{sourceOpaque}, Limit: 4},
		Session: "stable-handles",
	}, options)
	if err != nil {
		t.Fatal(err)
	}
	expectedSource := agentquery.NewSourceHandle("source-1", "internal/target.go", 10, 12).Value
	if len(requests) < 4 || len(requests[3].Handles) != 1 || requests[3].Handles[0] != expectedSource {
		t.Fatalf("source-range handle did not restore exact source handle: %#v", requests)
	}
}

func TestDiscoveryDefaultsUseIndependentLaneLimit(t *testing.T) {
	search := normalizeRequest(Request{Request: agentquery.Request{Mode: "search"}}, Options{})
	if search.Limit != 12 {
		t.Fatalf("search default limit = %d, want 12", search.Limit)
	}
	trace := normalizeRequest(Request{Request: agentquery.Request{Mode: "trace"}}, Options{})
	if trace.Limit != 8 {
		t.Fatalf("trace default limit = %d, want 8", trace.Limit)
	}
	enabled := true
	if includeDocuments(Request{Request: agentquery.Request{CodeOnly: true}, IncludeDocuments: &enabled}) {
		t.Fatal("code_only must suppress documents even when include_documents is true")
	}
}

func TestDocumentVectorsAreExplicitOptIn(t *testing.T) {
	if useDocumentVectors(Request{}) {
		t.Fatal("document vectors should be disabled by default")
	}
	enabled := true
	if !useDocumentVectors(Request{UseDocumentVectors: &enabled}) {
		t.Fatal("explicit document vector opt-in was ignored")
	}
	disabled := false
	if useDocumentVectors(Request{UseDocumentVectors: &disabled}) {
		t.Fatal("explicit document vector disable was ignored")
	}
}

func TestCompactKnowledgeResultsBoundsAgentPayload(t *testing.T) {
	links := make([]knowledge.CodeLink, 12)
	for index := range links {
		links[index] = knowledge.CodeLink{Kind: "symbol", Value: "Symbol", SourcePath: "pkg/file.go"}
	}
	result := knowledge.Result{
		Text:      strings.Repeat("évidence ", 300),
		Reasons:   []string{"one", "two", "three", "four", "five"},
		CodeLinks: links,
	}

	compacted := compactKnowledgeResults([]knowledge.Result{result})
	if len(compacted) != 1 {
		t.Fatalf("result count = %d", len(compacted))
	}
	if len(compacted[0].Text) > 1203 || !utf8.ValidString(compacted[0].Text) || !strings.HasSuffix(compacted[0].Text, "…") {
		t.Fatalf("text was not compacted safely: bytes=%d valid=%v suffix=%q", len(compacted[0].Text), utf8.ValidString(compacted[0].Text), compacted[0].Text[len(compacted[0].Text)-3:])
	}
	if len(compacted[0].CodeLinks) != 8 {
		t.Fatalf("code links = %d, want 8", len(compacted[0].CodeLinks))
	}
	if len(compacted[0].Reasons) != 4 {
		t.Fatalf("reasons = %d, want 4", len(compacted[0].Reasons))
	}
	if len(result.CodeLinks) != 12 || len(result.Reasons) != 5 {
		t.Fatal("compaction mutated caller-owned result")
	}
}

func TestInvestigationResponseRecordsDirectRelationships(t *testing.T) {
	subject := agentquery.Node{Handle: agentquery.Handle{Value: "subject", Provider: "lexicon"}, Kind: "function", Name: "Caller"}
	object := agentquery.Node{Handle: agentquery.Handle{Value: "object", Provider: "lexicon"}, Kind: "function", Name: "Callee"}
	converted := investigationResponse(agentquery.Response{
		Snapshot: agentquery.Snapshot{Source: "source-1"},
		RelationshipMatches: []agentquery.RelationshipMatch{{
			Provider: "lexicon", Subject: subject, Direction: "outgoing", Relation: "calls", Object: object,
		}},
	}, nil)
	if len(converted.GraphPaths) != 1 || len(converted.GraphPaths[0].Edges) != 1 || converted.GraphPaths[0].Edges[0] != "calls" {
		t.Fatalf("direct relationship was not recorded as graph evidence: %#v", converted.GraphPaths)
	}
}

func TestExecuteSessionReturnsOnlyNewEvidenceThenPriorHandles(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "docs/design.md", "# Graph ownership\nArcana owns repository graph traversal.\n")
	request := Request{
		Request: agentquery.Request{Mode: "search", Root: root, Query: "graph ownership", Limit: 4},
		Session: "agent-session",
	}
	first, err := Execute(context.Background(), request, testOptions(root))
	if err != nil {
		t.Fatal(err)
	}
	if len(first.SourceMatches) != 0 || len(first.DocumentMatches) != 0 || first.Delta == nil {
		t.Fatalf("session response leaked full evidence: %#v", first)
	}
	if len(first.Delta.NewNodes) == 0 || len(first.Delta.NewDocuments) == 0 {
		t.Fatalf("first delta = %#v", first.Delta)
	}
	second, err := Execute(context.Background(), request, testOptions(root))
	if err != nil {
		t.Fatal(err)
	}
	if second.Delta == nil || len(second.Delta.NewNodes) != 0 || len(second.Delta.NewDocuments) != 0 {
		t.Fatalf("second delta replayed evidence: %#v", second.Delta)
	}
	if len(second.Delta.PriorNodeHandles) == 0 || len(second.Delta.PriorDocuments) == 0 {
		t.Fatalf("second delta lacks prior handles: %#v", second.Delta)
	}
}

func testOptions(root string) Options {
	return Options{
		DefaultMode: repostate.CurrentOnly,
		EnsureRepository: func(context.Context, repostate.Options) (repostate.Status, error) {
			return repostate.Status{
				Repository:              repostate.RepositoryStatus{Root: root},
				Grimoire:                repostate.ComponentStatus{Status: "current", Prepared: true},
				DeterministicQueryReady: true,
			}, nil
		},
		ExecuteQuery: func(context.Context, agentquery.Request) (agentquery.Response, error) {
			handle := agentquery.Handle{Value: "grimoire://source/source-1/node/caller", Provider: "source", Snapshot: "source-1", NodeIdentity: "caller"}
			return agentquery.Response{
				Schema:        agentquery.SchemaVersion,
				Mode:          "search",
				Snapshot:      agentquery.Snapshot{Source: "source-1"},
				SourceMatches: []agentquery.Result{{Rank: 1, Provider: "source", Kind: "function", Node: agentquery.Node{Handle: handle, Kind: "function", Name: "caller", Path: filepath.ToSlash(filepath.Join("internal", "caller.go"))}, Reasons: []string{"exact symbol"}}},
			}, nil
		},
	}
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
