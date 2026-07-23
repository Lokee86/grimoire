package app

import (
	"testing"

	"github.com/Lokee86/grimoire/internal/structure"
)

func TestInterleaveStructuralEvidencePreservesProviderRanks(t *testing.T) {
	lexicon := []structure.Evidence{
		{Provider: "lexicon", Rank: 1},
		{Provider: "lexicon", Rank: 2},
		{Provider: "lexicon", Rank: 3},
	}
	arcana := []structure.Evidence{
		{Provider: "arcana", Rank: 1},
		{Provider: "arcana", Rank: 2},
	}
	result := interleaveStructuralEvidence(lexicon, arcana)
	wantProviders := []string{"lexicon", "arcana", "lexicon", "arcana", "lexicon"}
	wantRanks := []int{1, 1, 2, 2, 3}
	if len(result) != len(wantProviders) {
		t.Fatalf("interleaved %d items, want %d", len(result), len(wantProviders))
	}
	for index := range result {
		if result[index].Provider != wantProviders[index] || result[index].Rank != wantRanks[index] {
			t.Fatalf("unexpected item %d: %+v", index, result[index])
		}
	}
}
