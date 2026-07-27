package app

import (
	"testing"

	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/queryshape"
	"github.com/Lokee86/grimoire/internal/retrieve"
)

func TestMergeLexicalLanesPreservesBothFronts(t *testing.T) {
	chunks := []retrieve.Candidate{
		{Chunk: index.Chunk{ID: "chunk-a"}, Source: "lexical", Rank: 1},
		{Chunk: index.Chunk{ID: "chunk-b"}, Source: "lexical", Rank: 2},
	}
	files := []retrieve.Candidate{
		{Chunk: index.Chunk{ID: "file-a"}, Source: "lexical-file", Rank: 1},
		{Chunk: index.Chunk{ID: "file-b"}, Source: "lexical-file", Rank: 2},
	}
	merged := mergeLexicalLanes(4, chunks, files)
	want := []string{"chunk-a", "file-a", "chunk-b", "file-b"}
	for index, id := range want {
		if merged[index].Chunk.ID != id {
			t.Fatalf("candidate %d=%q, want %q: %+v", index, merged[index].Chunk.ID, id, merged)
		}
	}
}

func TestIntentFileScopePathsExcludesSupportingFiles(t *testing.T) {
	snapshot := index.Snapshot{Files: []index.FileRecord{
		{Path: "internal/recovery.go", Chunks: []index.Chunk{{
			ID: "source", Path: "internal/recovery.go", Text: "realtime reconnect recovery channel",
		}}},
		{Path: "internal/recovery_test.go", Chunks: []index.Chunk{{
			ID: "test", Path: "internal/recovery_test.go", Text: "realtime reconnect recovery channel realtime reconnect recovery channel",
		}}},
	}}
	intents := []queryshape.RetrievalIntent{{Query: "realtime reconnect recovery", Weight: 1}}
	paths := intentFileScopePaths(snapshot, intents, 4, retrieve.DefaultConfig())
	if len(paths) != 1 || paths[0] != "internal/recovery.go" {
		t.Fatalf("unexpected production scopes: %+v", paths)
	}
}
