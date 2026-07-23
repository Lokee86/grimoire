package lexiconfacts

import (
	"os"
	"path/filepath"
	"testing"

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
			{ID: "manifest-1", Path: "internal/manifest.go", StartLine: 1, EndLine: 48},
			{ID: "manifest-2", Path: "internal/manifest.go", StartLine: 49, EndLine: 96},
		}},
		{Path: "internal/engine.go", Chunks: []index.Chunk{
			{ID: "engine-1", Path: "internal/engine.go", StartLine: 1, EndLine: 48},
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
	if len(result.Evidence) == 0 || result.Evidence[0].Node == nil {
		t.Fatalf("expected first-class Lexicon evidence, got %+v", result.Evidence)
	}
	if result.Evidence[0].Node.Name != "ValidateSnapshot" || result.Evidence[0].Provider != source {
		t.Fatalf("unexpected Lexicon symbol evidence: %+v", result.Evidence[0])
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
		}
	}
	if !foundRelated {
		t.Fatalf("call relationship did not recover helper chunk: %+v", candidates)
	}
}

func TestSearchRequiresExport(t *testing.T) {
	_, err := Search(index.Snapshot{}, "query", t.TempDir(), 10)
	if err == nil {
		t.Fatal("expected missing export error")
	}
}
