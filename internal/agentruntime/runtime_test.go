package agentruntime

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Lokee86/grimoire/internal/agentquery"
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
	if response.Query == nil || len(response.Query.Results) != 1 {
		t.Fatalf("query response = %#v", response.Query)
	}
	if len(response.Knowledge) == 0 || response.Knowledge[0].Path != "docs/design.md" {
		t.Fatalf("knowledge response = %#v", response.Knowledge)
	}
	if response.Delta != nil {
		t.Fatalf("unexpected session delta: %#v", response.Delta)
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
	if first.Query != nil || len(first.Knowledge) != 0 || first.Delta == nil {
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
				Schema:   agentquery.SchemaVersion,
				Mode:     "search",
				Snapshot: agentquery.Snapshot{Source: "source-1"},
				Results:  []agentquery.Result{{Rank: 1, Provider: "source", Kind: "function", Node: agentquery.Node{Handle: handle, Kind: "function", Name: "caller", Path: filepath.ToSlash(filepath.Join("internal", "caller.go"))}, Reasons: []string{"exact symbol"}}},
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
