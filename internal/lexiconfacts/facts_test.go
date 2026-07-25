package lexiconfacts

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/Lokee86/grimoire/internal/evidence"
	"github.com/Lokee86/grimoire/internal/index"
)

func TestSearchMapsMatchedAndRelatedNodesToPreparedChunks(t *testing.T) {
	directory := t.TempDir()
	data := "" +
		`{"record":"lexicon","language":"go","repository":"example"}` + "\n" +
		`{"record":"node","id":"owner","kind":"function","name":"ValidateSnapshot","path":"internal/manifest.go","qualified_name":"internal/manifest.go::ValidateSnapshot","span":{"path":"internal/manifest.go","start_line":50,"end_line":80}}` + "\n" +
		`{"record":"node","id":"helper","kind":"function","name":"CheckDimensions","path":"internal/engine.go","qualified_name":"internal/engine.go::CheckDimensions","span":{"path":"internal/engine.go","start_line":1,"end_line":20}}` + "\n" +
		`{"record":"node","id":"external","kind":"function","name":"Error","path":"@stdlib/errors","qualified_name":"@stdlib/errors::Error"}` + "\n" +
		`{"record":"edge","source":"owner","target":"helper","relation":"calls"}` + "\n"
	if err := os.WriteFile(filepath.Join(directory, "go.jsonl"), []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	snapshot := index.Snapshot{Files: []index.FileRecord{
		{Path: "internal/manifest.go", Chunks: []index.Chunk{
			{ID: "manifest-1", Path: "internal/manifest.go", StartLine: 1, EndLine: 48, TokenCount: 11},
			{ID: "manifest-2", Path: "internal/manifest.go", StartLine: 49, EndLine: 96, TokenCount: 17},
		}},
		{Path: "internal/engine.go", Chunks: []index.Chunk{
			{ID: "engine-1", Path: "internal/engine.go", StartLine: 1, EndLine: 48, TokenCount: 13},
		}},
	}}

	result, err := SearchDetailed(snapshot, "Where is ValidateSnapshot implemented?", directory, 10)
	if err != nil {
		t.Fatal(err)
	}
	candidates := result.Candidates
	if len(candidates) < 2 {
		t.Fatalf("expected direct and relationship candidates, got %+v", candidates)
	}
	if candidates[0].Chunk.ID != "manifest-2" || candidates[0].Source != source {
		t.Fatalf("unexpected direct candidate: %+v", candidates[0])
	}
	directContext := candidates[0].Context
	if directContext == nil || directContext.Identity != evidence.RangeIdentity("internal/manifest.go", 50, 80) ||
		!slices.Contains(directContext.Roles, evidence.RolePrimary) ||
		!slices.Contains(directContext.GroupIDs, nodeGroupID(Node{ID: "owner", Span: &Span{Path: "internal/manifest.go", StartLine: 50, EndLine: 80}})) ||
		directContext.EstimatedTokens != 17 || directContext.RedundancyKey == "" ||
		directContext.Graph == nil || directContext.Graph.Distance != 0 ||
		directContext.Graph.SymbolRole != "function" || directContext.Graph.ModuleProximity != 1 ||
		directContext.Graph.Centrality <= 0 {
		t.Fatalf("direct candidate descriptor missing structural metadata: %+v", directContext)
	}
	if len(result.Evidence) == 0 || result.Evidence[0].Node == nil {
		t.Fatalf("expected first-class Lexicon evidence, got %+v", result.Evidence)
	}
	if result.Evidence[0].Node.Name != "ValidateSnapshot" || result.Evidence[0].Provider != source {
		t.Fatalf("unexpected Lexicon symbol evidence: %+v", result.Evidence[0])
	}
	if result.Evidence[0].Context == nil || !slices.Contains(result.Evidence[0].Context.Roles, evidence.RoleStructural) ||
		result.Evidence[0].Context.GroupIDs[0] != directContext.GroupIDs[0] ||
		len(result.Evidence[0].Context.Links) != 1 || result.Evidence[0].Context.Links[0].Identity != directContext.Identity {
		t.Fatalf("Lexicon structural descriptor did not link its source candidate: %+v", result.Evidence[0].Context)
	}
	if len(result.Evidence[0].Relationships) != 1 || result.Evidence[0].Relationships[0].Node.Name != "CheckDimensions" {
		t.Fatalf("Lexicon relationship was not preserved: %+v", result.Evidence[0])
	}
	if len(result.Seeds) == 0 || result.Seeds[0].Identity != "owner" {
		t.Fatalf("Arcana seed identity missing: %+v", result.Seeds)
	}
	foundRelated := false
	for _, candidate := range candidates {
		if candidate.Chunk.ID == "engine-1" {
			foundRelated = true
			if candidate.Source != source {
				t.Fatalf("unexpected relationship source: %+v", candidate)
			}
			if candidate.Context == nil || !slices.Contains(candidate.Context.Roles, evidence.RoleSupporting) ||
				candidate.Context.EstimatedTokens != 13 || candidate.Context.RedundancyKey == "" ||
				candidate.Context.Graph == nil || candidate.Context.Graph.Distance != 1 ||
				!slices.Contains(candidate.Context.Graph.Relations, "outgoing:calls") ||
				candidate.Context.Graph.SymbolRole != "function" ||
				candidate.Context.Graph.ModuleProximity != 1 || candidate.Context.Graph.Centrality <= 0 {
				t.Fatalf("relationship candidate descriptor missing supporting metadata: %+v", candidate.Context)
			}
		}
	}
	if !foundRelated {
		t.Fatalf("call relationship did not recover helper chunk: %+v", candidates)
	}
}

func TestSearchTraversesInterstackContractNodesToTheOppositeStack(t *testing.T) {
	directory := t.TempDir()
	files := map[string]string{
		"gdscript.jsonl": "" +
			`{"record":"lexicon","language":"gdscript","repository":"example"}` + "\n" +
			`{"record":"node","id":"sender","kind":"method","name":"SubmitLogin","path":"client/login.gd","qualified_name":"client/login.gd::SubmitLogin","span":{"path":"client/login.gd","start_line":10,"end_line":20}}` + "\n",
		"ruby.jsonl": "" +
			`{"record":"lexicon","language":"ruby","repository":"example"}` + "\n" +
			`{"record":"node","id":"handler","kind":"method","name":"create","path":"server/sessions_controller.rb","qualified_name":"SessionsController#create","span":{"path":"server/sessions_controller.rb","start_line":5,"end_line":12}}` + "\n",
		"interstack.jsonl": "" +
			`{"record":"lexicon","language":"interstack","repository":"example"}` + "\n" +
			`{"record":"node","id":"endpoint","kind":"http-endpoint","name":"POST /session/start","path":"@interstack/http/session","qualified_name":"http:server:POST /session/start"}` + "\n" +
			`{"record":"edge","source":"sender","target":"endpoint","relation":"calls-endpoint"}` + "\n" +
			`{"record":"edge","source":"endpoint","target":"handler","relation":"handled-by"}` + "\n",
	}
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	snapshot := index.Snapshot{Files: []index.FileRecord{
		{Path: "client/login.gd", Chunks: []index.Chunk{{ID: "client", Path: "client/login.gd", StartLine: 1, EndLine: 30, TokenCount: 20}}},
		{Path: "server/sessions_controller.rb", Chunks: []index.Chunk{{ID: "server", Path: "server/sessions_controller.rb", StartLine: 1, EndLine: 20, TokenCount: 18}}},
	}}

	result, err := SearchDetailed(snapshot, "Trace SubmitLogin", directory, 10)
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range result.Candidates {
		if candidate.Chunk.ID != "server" {
			continue
		}
		if candidate.Context == nil || candidate.Context.Graph == nil || candidate.Context.Graph.Distance != 2 ||
			!slices.Contains(candidate.Context.Graph.Relations, "outgoing:calls-endpoint") ||
			!slices.Contains(candidate.Context.Graph.Relations, "outgoing:handled-by") {
			t.Fatalf("opposite-stack candidate lost its two-hop graph path: %+v", candidate)
		}
		return
	}
	t.Fatalf("interstack path did not recover the server handler: %+v", result.Candidates)
}

func TestRelationshipsAggregateOccurrencesAndPreserveSemanticSites(t *testing.T) {
	directory := t.TempDir()
	data := "" +
		`{"record":"lexicon","language":"c-family","repository":"example"}` + "\n" +
		`{"record":"node","id":"run","kind":"function","name":"run","path":"main.c","qualified_name":"run"}` + "\n" +
		`{"record":"node","id":"sink","kind":"function","name":"sink","path":"main.c","qualified_name":"sink"}` + "\n" +
		`{"record":"node","id":"macro","kind":"symbol","name":"FORWARD","path":"main.c","qualified_name":"FORWARD"}` + "\n" +
		`{"record":"node","id":"input","kind":"parameter","name":"input","path":"main.c","qualified_name":"run::input"}` + "\n" +
		`{"record":"node","id":"target","kind":"parameter","name":"target","path":"main.c","qualified_name":"sink::target"}` + "\n" +
		`{"record":"edge","source":"run","target":"sink","relation":"calls","span":{"path":"main.c","start_line":3,"end_line":3},"attributes":{"resolution":"definite","candidate_count":1,"evidence":["macro-body","argument-substitution"],"indirect":"macro","via":["macro"],"macro_body_callee":"sink","macro_definition_span":{"path":"main.c","start_line":2,"end_line":2},"expansion_depth":0,"macro_call_index":0,"substitutions":{"value":"input"},"substituted_arguments":["(input)"]}}` + "\n" +
		`{"record":"edge","source":"run","target":"sink","relation":"calls","span":{"path":"main.c","start_line":4,"end_line":4},"attributes":{"resolution":"definite","candidate_count":1,"evidence":["macro-body"],"indirect":"macro","via":["macro"],"macro_body_callee":"sink","expansion_depth":0}}` + "\n" +
		`{"record":"edge","source":"input","target":"target","relation":"passes-to","span":{"path":"main.c","start_line":3,"end_line":3},"attributes":{"argument_index":0,"expression":"input","via_call":"sink"}}` + "\n"
	if err := os.WriteFile(filepath.Join(directory, "c-family.jsonl"), []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	facts, err := loadDirectory(directory)
	if err != nil {
		t.Fatal(err)
	}

	calls := relationshipsForSeed("run", facts, 12)
	if len(calls) != 1 || calls[0].Relation != "calls" || calls[0].Occurrences != 2 || len(calls[0].Sites) != 2 {
		t.Fatalf("macro call occurrences were not aggregated: %+v", calls)
	}
	first := calls[0].Sites[0]
	if first.ExpansionDepth == nil || *first.ExpansionDepth != 0 || first.DefinitionSpan == nil ||
		first.DefinitionSpan.StartLine != 2 || len(first.Via) != 1 || first.Via[0].Name != "FORWARD" ||
		first.Substitutions["value"] != "input" || first.Arguments[0] != "(input)" {
		t.Fatalf("macro provenance was not preserved: %+v", first)
	}

	flow := relationshipsForSeed("input", facts, 12)
	if len(flow) != 1 || flow[0].Relation != "passes-to" || flow[0].Occurrences != 1 || len(flow[0].Sites) != 1 {
		t.Fatalf("argument flow relationship was not preserved: %+v", flow)
	}
	flowSite := flow[0].Sites[0]
	if flowSite.ArgumentIndex == nil || *flowSite.ArgumentIndex != 0 || flowSite.Expression != "input" ||
		len(flowSite.Via) != 1 || flowSite.Via[0].Name != "sink" {
		t.Fatalf("argument flow provenance was not preserved: %+v", flowSite)
	}
}

func TestDescriptorsWithoutSourceSpansKeepFallbackRangeAndOmitLinks(t *testing.T) {
	chunk := index.Chunk{Path: "internal/fallback.go", StartLine: 1, EndLine: 12, TokenCount: 9}
	node := Node{ID: "fallback", Name: "Fallback", Path: chunk.Path}
	candidates := chunksForNodes(
		index.Snapshot{Files: []index.FileRecord{{Path: chunk.Path, Chunks: []index.Chunk{chunk}}}},
		map[string]scoredNode{"fallback": {node: node, score: 1, primary: true}}, 1,
	)
	if len(candidates) != 1 || candidates[0].Context == nil ||
		candidates[0].Context.Identity != evidence.RangeIdentity(chunk.Path, chunk.StartLine, chunk.EndLine) {
		t.Fatalf("fallback candidate lost its prepared range identity: %+v", candidates)
	}
	structural := evidenceForSeeds([]scoredNode{{node: node, score: 1}}, library{nodes: map[string]Node{}}, 1)
	if len(structural) != 1 || structural[0].Context == nil || len(structural[0].Context.Links) != 0 {
		t.Fatalf("fallback structural evidence unexpectedly linked a source span: %+v", structural)
	}
	data, err := json.Marshal(structural[0])
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 || bytes.Contains(data, []byte(`"links"`)) {
		t.Fatalf("fallback structural serialization did not preserve omitted source links: %s", data)
	}
}

func TestSearchRequiresExport(t *testing.T) {
	_, err := Search(index.Snapshot{}, "query", t.TempDir(), 10)
	if err == nil {
		t.Fatal("expected missing export error")
	}
}
