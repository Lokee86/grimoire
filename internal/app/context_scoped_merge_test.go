package app

import (
	"testing"

	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/retrieve"
)

func TestMergeScopedContextProvidersDoesNotInjectOrReorderStructuralCandidates(t *testing.T) {
	base := []retrieve.Candidate{
		{Chunk: index.Chunk{ID: "first"}, Source: "lexical", Rank: 1},
		{Chunk: index.Chunk{ID: "second"}, Source: "lexical-file", Rank: 2},
	}
	lexicon := []retrieve.Candidate{
		{Chunk: index.Chunk{ID: "second"}, Source: "lexicon", Rank: 1, Reasons: []string{"resolved declaration"}},
		{Chunk: index.Chunk{ID: "outside"}, Source: "lexicon", Rank: 2},
	}
	arcana := []retrieve.Candidate{
		{Chunk: index.Chunk{ID: "outside-graph"}, Source: "arcana", Rank: 1},
	}

	merged := mergeScopedContextProviders(10, nil, base, lexicon, arcana)
	if len(merged) != 2 || merged[0].Chunk.ID != "first" || merged[1].Chunk.ID != "second" {
		t.Fatalf("structural candidates changed lexical authority: %+v", merged)
	}
	if merged[1].Source != "lexical-file" {
		t.Fatalf("structural inspection replaced lexical provenance: %+v", merged[1])
	}
	found := false
	for _, reason := range merged[1].Reasons {
		if reason == "inspected by lexicon rank 1" {
			found = true
		}
	}
	if !found {
		t.Fatalf("structural inspection was not attached to discovered range: %+v", merged[1])
	}
}
