package lexiconfacts

import (
	"testing"

	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/retrieve"
)

func TestSearchDetailedScopedCannotSelectOutsideLexicalRanges(t *testing.T) {
	corpus := &Corpus{facts: library{
		nodes: map[string]Node{
			"owner": {
				ID: "owner", Kind: "function", Name: "RecoverChannel", Path: "internal/recovery.go",
				Span: &Span{Path: "internal/recovery.go", StartLine: 10, EndLine: 20},
			},
			"distractor": {
				ID: "distractor", Kind: "function", Name: "ReconnectRecovery", Path: "internal/distractor.go",
				Span: &Span{Path: "internal/distractor.go", StartLine: 1, EndLine: 8},
			},
		},
	}}
	snapshot := index.Snapshot{Files: []index.FileRecord{
		{Path: "internal/recovery.go", Chunks: []index.Chunk{{
			ID: "recovery", Path: "internal/recovery.go", StartLine: 1, EndLine: 30, Text: "recover channel",
		}}},
		{Path: "internal/distractor.go", Chunks: []index.Chunk{{
			ID: "distractor", Path: "internal/distractor.go", StartLine: 1, EndLine: 30, Text: "reconnect recovery",
		}}},
	}}
	scopes := []retrieve.Candidate{{
		Chunk: snapshot.Files[0].Chunks[0], Source: "lexical-file", Rank: 1,
	}}

	result := corpus.SearchDetailedScoped(snapshot, "reconnect recovery", scopes, 8)
	if len(result.Seeds) != 1 || result.Seeds[0].Identity != "owner" {
		t.Fatalf("scoped seeds escaped lexical range: %+v", result.Seeds)
	}
	if len(result.Candidates) != 1 || result.Candidates[0].Chunk.Path != "internal/recovery.go" {
		t.Fatalf("scoped candidates escaped lexical range: %+v", result.Candidates)
	}
}
