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

	result, err := SearchDetailed(snapshot, "Where does ValidateSnapshot check dimensions?", directory, 10)
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
		directContext.EstimatedTokens != 17 || directContext.RedundancyKey == "" {
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
				candidate.Context.EstimatedTokens != 13 || candidate.Context.RedundancyKey == "" {
				t.Fatalf("relationship candidate descriptor missing supporting metadata: %+v", candidate.Context)
			}
		}
	}
	if !foundRelated {
		t.Fatalf("call relationship did not recover helper chunk: %+v", candidates)
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
