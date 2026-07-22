package retrieve

import (
	"testing"

	"github.com/Lokee86/grimoire/internal/index"
)

func TestSearchRanksFilenameAndContentDeterministically(t *testing.T) {
	snapshot := index.Snapshot{
		Version: index.FormatVersion,
		Files: []index.FileRecord{
			{Path: "internal/damage/resolver.go", Chunks: []index.Chunk{{
				ID: "a", Path: "internal/damage/resolver.go", StartLine: 1, EndLine: 2,
				TokenCount: 5, Text: "package damage\nfunc ResolveDamage() {}",
			}}},
			{Path: "internal/player/state.go", Chunks: []index.Chunk{{
				ID: "b", Path: "internal/player/state.go", StartLine: 1, EndLine: 2,
				TokenCount: 5, Text: "package player\n// damage is recorded here",
			}}},
		},
	}

	results := Search(snapshot, "damage resolver", 10)
	if len(results) != 2 {
		t.Fatalf("expected two results, got %d", len(results))
	}
	if results[0].Chunk.Path != "internal/damage/resolver.go" {
		t.Fatalf("unexpected first result: %+v", results[0])
	}
	if results[0].Score <= results[1].Score {
		t.Fatalf("expected strict score ordering: %+v", results)
	}
}

func TestSearchUsesStablePathTieBreak(t *testing.T) {
	snapshot := index.Snapshot{Version: index.FormatVersion, Files: []index.FileRecord{
		{Path: "b.go", Chunks: []index.Chunk{{Path: "b.go", StartLine: 1, Text: "needle", TokenCount: 1}}},
		{Path: "a.go", Chunks: []index.Chunk{{Path: "a.go", StartLine: 1, Text: "needle", TokenCount: 1}}},
	}}
	results := Search(snapshot, "needle", 10)
	if results[0].Chunk.Path != "a.go" {
		t.Fatalf("tie break was not stable: %+v", results)
	}
}
