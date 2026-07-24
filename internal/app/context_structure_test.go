package app

import (
	"testing"

	"github.com/Lokee86/grimoire/internal/structure"
)

func TestParseContextStructuralProvidersAllowsArcanaWithoutEmittingLexicon(t *testing.T) {
	emitLexicon, arcanaEnabled, err := parseContextStructuralProviders("arcana")
	if err != nil {
		t.Fatal(err)
	}
	if emitLexicon || !arcanaEnabled {
		t.Fatalf("emitLexicon=%v arcanaEnabled=%v", emitLexicon, arcanaEnabled)
	}
}

func TestParseContextStructuralProvidersRejectsUnknownProvider(t *testing.T) {
	if _, _, err := parseContextStructuralProviders("lexicon,unknown"); err == nil {
		t.Fatal("expected unsupported provider error")
	}
}

func TestMergeArcanaSeedsBalancesSemanticAndLexiconRecall(t *testing.T) {
	lexicon := []structure.Node{
		{Name: "LexiconOne", Path: "lexicon/one.go"},
		{Name: "Shared", Path: "shared.go"},
		{Name: "LexiconThree", Path: "lexicon/three.go"},
		{Name: "LexiconFour", Path: "lexicon/four.go"},
	}
	semantic := []structure.Node{
		{Name: "SemanticOne", Path: "semantic/one.go"},
		{Name: "Shared", Path: "shared.go"},
		{Name: "SemanticThree", Path: "semantic/three.go"},
		{Name: "SemanticFour", Path: "semantic/four.go"},
	}

	result := mergeArcanaSeeds(lexicon, semantic, 6)
	want := []string{
		"SemanticOne", "LexiconOne", "Shared",
		"SemanticThree", "LexiconThree", "SemanticFour",
	}
	if len(result) != len(want) {
		t.Fatalf("merged %d seeds, want %d: %+v", len(result), len(want), result)
	}
	for index := range want {
		if result[index].Name != want[index] {
			t.Fatalf("seed %d=%q, want %q: %+v", index, result[index].Name, want[index], result)
		}
	}
}

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
