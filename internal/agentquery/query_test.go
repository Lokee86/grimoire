package agentquery

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/Lokee86/grimoire/internal/index"
)

func TestSearchHandleInspectsExactPreparedSource(t *testing.T) {
	root, facts := queryFixture(t)
	search, err := Execute(context.Background(), Request{
		Schema: SchemaVersion, Mode: "search", Root: root,
		Query: "SubmitLogin", LexiconFacts: facts, Limit: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	var source Handle
	for _, lane := range [][]Result{search.ExactMatches, search.SourceMatches} {
		for _, result := range lane {
			if result.Node.Handle.Provider == "source" && result.Node.Path == "client/login.gd" {
				source = result.Node.Handle
				break
			}
		}
	}
	if source.Value == "" || source.Snapshot == "" {
		t.Fatalf("search did not return a snapshot-qualified source handle: exact=%+v source=%+v", search.ExactMatches, search.SourceMatches)
	}

	inspection, err := Execute(context.Background(), Request{
		Schema: SchemaVersion, Mode: "inspect", Root: root,
		Handles: []string{source.Value}, LexiconFacts: facts, Adjacent: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(inspection.Inspections) != 1 ||
		!strings.Contains(inspection.Inspections[0].Source, "func SubmitLogin") {
		t.Fatalf("handle did not inspect exact prepared source: %+v", inspection.Inspections)
	}
	if inspection.Inspections[0].ContainingSpan.Handle.Provider != "source" {
		t.Fatalf("inspection range has no source handle: %+v", inspection.Inspections[0])
	}
}

func TestSearchKeepsDocumentationOutOfSourceLanes(t *testing.T) {
	root, facts := queryFixture(t)
	response, err := Execute(context.Background(), Request{
		Schema: SchemaVersion, Mode: "search", Root: root,
		Query: "POST session start", LexiconFacts: facts, Limit: 10, CodeOnly: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(response.ExactMatches)+len(response.SourceMatches) == 0 {
		t.Fatal("search returned no source evidence")
	}
	for _, lane := range [][]Result{response.ExactMatches, response.SourceMatches, response.SymbolMatches} {
		for _, result := range lane {
			if strings.HasPrefix(result.Node.Path, "docs/") || strings.HasSuffix(result.Node.Path, ".md") {
				t.Fatalf("source or symbol lane returned documentation: %+v", result)
			}
		}
	}
}

func TestSearchAppliesLimitPerEvidenceLane(t *testing.T) {
	root, facts := queryFixture(t)
	response, err := Execute(context.Background(), Request{
		Schema: SchemaVersion, Mode: "search", Root: root,
		Query: "SubmitLogin session start", LexiconFacts: facts, Limit: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(response.ExactMatches) != 1 || len(response.SourceMatches) != 1 ||
		len(response.SymbolMatches) != 1 || len(response.RelationshipMatches) != 1 {
		t.Fatalf("lane limits were not independent: exact=%d source=%d symbols=%d relationships=%d",
			len(response.ExactMatches), len(response.SourceMatches),
			len(response.SymbolMatches), len(response.RelationshipMatches))
	}
	if response.ExactMatches[0].Excerpt == "" || response.SourceMatches[0].Excerpt == "" || response.SymbolMatches[0].Excerpt == "" {
		t.Fatalf("discovery results lack bounded source excerpts: exact=%+v source=%+v symbol=%+v",
			response.ExactMatches[0], response.SourceMatches[0], response.SymbolMatches[0])
	}
	if response.RelationshipMatches[0].Relation != "calls-endpoint" {
		t.Fatalf("unexpected relationship match: %+v", response.RelationshipMatches[0])
	}
}

func TestTraceFollowsInterstackEndpointFromReturnedHandle(t *testing.T) {
	root, facts := queryFixture(t)
	search, err := Execute(context.Background(), Request{
		Schema: SchemaVersion, Mode: "search", Root: root,
		Query: "SubmitLogin", LexiconFacts: facts, Limit: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	var anchor string
	for _, result := range search.SymbolMatches {
		if result.Provider == "lexicon" && result.Node.Name == "SubmitLogin" {
			anchor = result.Node.Handle.Value
		}
	}
	if anchor == "" {
		t.Fatalf("missing Lexicon anchor: %+v", search.SymbolMatches)
	}

	trace, err := Execute(context.Background(), Request{
		Schema: SchemaVersion, Mode: "trace", Root: root,
		Anchor: anchor, LexiconFacts: facts, Depth: 3, Limit: 10, Detail: "full",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range trace.Paths {
		relations := make([]string, len(path.Steps))
		for index, step := range path.Steps {
			relations[index] = step.Relation
			if step.From.Value == "" || step.To.Value == "" {
				t.Fatalf("trace step lost handles: %+v", step)
			}
		}
		if slices.Equal(relations, []string{"calls-endpoint", "handled-by"}) &&
			path.Nodes[len(path.Nodes)-1].Name == "create" {
			return
		}
	}
	t.Fatalf("interstack endpoint path was not returned: %+v", trace.Paths)
}

func TestOrientIsCompactAndSuggestsProgressiveExpansion(t *testing.T) {
	root, facts := queryFixture(t)
	response, err := Execute(context.Background(), Request{
		Schema: SchemaVersion, Mode: "orient", Root: root,
		LexiconFacts: facts, Limit: 8,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(response.SourceMatches)+len(response.SymbolMatches) == 0 || len(response.Suggestions) == 0 {
		t.Fatalf("orient did not return anchors and expansions: %+v", response)
	}
	for _, lane := range [][]Result{response.SourceMatches, response.SymbolMatches} {
		for _, result := range lane {
			if result.Node.Handle.Value == "" {
				t.Fatalf("orient result has no stable handle: %+v", result)
			}
		}
	}
}

func TestImpactHonorsIncomingRelationFilter(t *testing.T) {
	root, facts := queryFixture(t)
	search, err := Execute(context.Background(), Request{
		Schema: SchemaVersion, Mode: "search", Root: root,
		Query: "SessionsController#create", LexiconFacts: facts, Limit: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	var anchor string
	for _, result := range search.SymbolMatches {
		if result.Provider == "lexicon" && result.Node.Name == "create" {
			anchor = result.Node.Handle.Value
		}
	}
	response, err := Execute(context.Background(), Request{
		Schema: SchemaVersion, Mode: "impact", Root: root,
		Anchor: anchor, LexiconFacts: facts, Direction: "incoming",
		Relations: []string{"handled-by"}, Depth: 3, Limit: 8,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Dependents) != 1 ||
		response.Dependents[0].Relation != "handled-by" ||
		response.Dependents[0].Node.Kind != "http-endpoint" {
		t.Fatalf("impact did not honor relation filter: %+v", response.Dependents)
	}
}

func queryFixture(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	client := "package client\n\nfunc SubmitLogin() { post(\"/session/start\") }\n"
	server := "package server\n\nfunc create() { startSession() }\n"
	writeFixture(t, filepath.Join(root, "client", "login.gd"), client)
	writeFixture(t, filepath.Join(root, "server", "sessions_controller.rb"), server)
	writeFixture(t, filepath.Join(root, "docs", "contracts.md"), "# Contracts\nPOST /session/start\n")

	snapshot, _, err := index.Build(root, nil, index.BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err := index.Save(filepath.Join(root, ".grimoire"), snapshot); err != nil {
		t.Fatal(err)
	}

	facts := filepath.Join(root, "facts")
	writeFixture(t, filepath.Join(facts, "client.jsonl"), strings.Join([]string{
		`{"record":"lexicon","language":"gdscript","repository":"fixture"}`,
		`{"record":"node","id":"sender","kind":"method","name":"SubmitLogin","path":"client/login.gd","qualified_name":"client/login.gd::SubmitLogin","span":{"path":"client/login.gd","start_line":3,"end_line":3}}`,
	}, "\n")+"\n")
	writeFixture(t, filepath.Join(facts, "server.jsonl"), strings.Join([]string{
		`{"record":"lexicon","language":"ruby","repository":"fixture"}`,
		`{"record":"node","id":"handler","kind":"method","name":"create","path":"server/sessions_controller.rb","qualified_name":"SessionsController#create","span":{"path":"server/sessions_controller.rb","start_line":3,"end_line":3}}`,
	}, "\n")+"\n")
	writeFixture(t, filepath.Join(facts, "interstack.jsonl"), strings.Join([]string{
		`{"record":"lexicon","language":"interstack","repository":"fixture"}`,
		`{"record":"node","id":"endpoint","kind":"http-endpoint","name":"POST /session/start","path":"@interstack/http/session","qualified_name":"http:server:POST /session/start"}`,
		`{"record":"edge","source":"sender","target":"endpoint","relation":"calls-endpoint","span":{"path":"client/login.gd","start_line":3,"end_line":3}}`,
		`{"record":"edge","source":"endpoint","target":"handler","relation":"handled-by","span":{"path":"server/sessions_controller.rb","start_line":3,"end_line":3}}`,
	}, "\n")+"\n")
	return root, facts
}

func writeFixture(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
