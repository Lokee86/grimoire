package app

import (
	"slices"
	"testing"

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
