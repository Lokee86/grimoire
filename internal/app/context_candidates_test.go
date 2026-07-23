package app

import (
	"slices"
	"testing"

	"github.com/Lokee86/grimoire/internal/evidence"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/retrieve"
)

func TestMergeContextCandidatesPrefersExactAndKeepsProviderEvidence(t *testing.T) {
	chunk := index.Chunk{ID: "alpha", Path: "alpha.go", StartLine: 1, EndLine: 3, Text: "func ResolveAlpha() {}"}
	exact := retrieve.Candidate{
		Chunk: chunk, Score: 100, Source: "exact", Rank: 1,
		Reasons: []string{"identifier ResolveAlpha in content"},
	}
	vector := retrieve.Candidate{
		Chunk: chunk, Score: 0.9, Source: "vector", Rank: 4,
		Reasons: []string{"semantic vector similarity"},
	}

	merged := mergeContextCandidates(10, []retrieve.Candidate{exact}, []retrieve.Candidate{vector})
	if len(merged) != 1 || merged[0].Source != "exact" {
		t.Fatalf("unexpected merge: %+v", merged)
	}
	if !slices.Contains(merged[0].Reasons, "also retrieved by vector rank 4") {
		t.Fatalf("missing provider evidence: %+v", merged[0].Reasons)
	}
}

func TestMergeContextCandidatesMergesSharedEvidenceMetadata(t *testing.T) {
	chunk := index.Chunk{ID: "alpha", Path: "alpha.go", StartLine: 1, EndLine: 3}
	primary := retrieve.Candidate{
		Chunk: chunk, Source: "vector", Rank: 1,
		Context: &evidence.Descriptor{
			Identity: "range:alpha.go:1:3",
			Intents:  []evidence.Intent{evidence.IntentMechanism},
		},
	}
	structural := retrieve.Candidate{
		Chunk: chunk, Source: "lexicon", Rank: 2,
		Context: &evidence.Descriptor{
			GroupIDs: []string{"call-chain:alpha"},
			Roles:    []evidence.Role{evidence.RoleStructural},
		},
	}

	merged := mergeContextCandidates(10, []retrieve.Candidate{primary}, []retrieve.Candidate{structural})
	if len(merged) != 1 || merged[0].Context == nil {
		t.Fatalf("unexpected merge: %+v", merged)
	}
	if !slices.Contains(merged[0].Context.Intents, evidence.IntentMechanism) ||
		!slices.Contains(merged[0].Context.GroupIDs, "call-chain:alpha") ||
		!slices.Contains(merged[0].Context.Roles, evidence.RoleStructural) {
		t.Fatalf("merged context lost provider metadata: %+v", merged[0].Context)
	}
}

func TestMergeRankedProvidersInterleavesProviderRanks(t *testing.T) {
	candidate := func(id, source string, rank int) retrieve.Candidate {
		return retrieve.Candidate{Chunk: index.Chunk{ID: id}, Source: source, Rank: rank}
	}
	lexical := []retrieve.Candidate{
		candidate("lexical-1", "lexical", 1),
		candidate("lexical-2", "lexical", 2),
	}
	vector := []retrieve.Candidate{
		candidate("vector-1", "vector", 1),
		candidate("vector-2", "vector", 2),
	}

	merged := mergeRankedProviders(4, lexical, vector)
	got := []string{
		merged[0].Chunk.ID, merged[1].Chunk.ID,
		merged[2].Chunk.ID, merged[3].Chunk.ID,
	}
	want := []string{"lexical-1", "vector-1", "lexical-2", "vector-2"}
	if !slices.Equal(got, want) {
		t.Fatalf("interleaved IDs = %v, want %v", got, want)
	}
}

func TestMergeContextProvidersBoundsLexiconWithoutReplacingBaseFront(t *testing.T) {
	candidate := func(id, source string, rank int) retrieve.Candidate {
		return retrieve.Candidate{Chunk: index.Chunk{ID: id}, Source: source, Rank: rank}
	}
	base := make([]retrieve.Candidate, 40)
	for index := range base {
		base[index] = candidate("base-"+string(rune('a'+index)), "vector", index+1)
	}
	lexicon := make([]retrieve.Candidate, 40)
	for index := range lexicon {
		lexicon[index] = candidate("lexicon-"+string(rune('a'+index)), "lexicon", index+1)
	}

	merged := mergeContextProviders(60, nil, base, lexicon)
	if len(merged) != 60 {
		t.Fatalf("merged %d candidates, want 60", len(merged))
	}
	for index := 0; index < baseFrontCandidates; index++ {
		if merged[index].Source != "vector" || merged[index].Chunk.ID != base[index].Chunk.ID {
			t.Fatalf("base front displaced at %d: %+v", index, merged[index])
		}
	}
	for index := baseFrontCandidates; index < baseFrontCandidates+maxLexiconCandidates; index++ {
		if merged[index].Source != "lexicon" {
			t.Fatalf("Lexicon supplement missing at %d: %+v", index, merged[index])
		}
	}
	if merged[baseFrontCandidates+maxLexiconCandidates].Chunk.ID != base[baseFrontCandidates].Chunk.ID {
		t.Fatalf("remaining base candidates did not resume: %+v", merged)
	}
}

func TestContextCandidateSourcesPreservesFirstUseOrder(t *testing.T) {
	candidates := []retrieve.Candidate{
		{Source: "exact"}, {Source: "adjacent"}, {Source: "vector"}, {Source: "exact"},
	}
	got := contextCandidateSources(candidates)
	want := []string{"exact", "adjacent", "vector"}
	if !slices.Equal(got, want) {
		t.Fatalf("sources = %v, want %v", got, want)
	}
}

func TestMergeContextCandidatesAppliesCombinedLimit(t *testing.T) {
	candidate := func(id, source string) retrieve.Candidate {
		return retrieve.Candidate{Chunk: index.Chunk{ID: id}, Source: source, Rank: 1}
	}
	merged := mergeContextCandidates(2,
		[]retrieve.Candidate{candidate("one", "exact")},
		[]retrieve.Candidate{candidate("two", "vector"), candidate("three", "vector")},
	)
	if len(merged) != 2 || merged[0].Chunk.ID != "one" || merged[1].Chunk.ID != "two" {
		t.Fatalf("unexpected limited merge: %+v", merged)
	}
}
