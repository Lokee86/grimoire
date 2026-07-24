package assembly

import (
	"testing"

	"github.com/Lokee86/grimoire/internal/evidence"
	"github.com/Lokee86/grimoire/internal/queryshape"
	"github.com/Lokee86/grimoire/internal/retrieve"
)

func TestPrioritizeFacetCoverageCoversEveryFacetBeforeRepeats(t *testing.T) {
	candidates := []retrieve.Candidate{
		facetCandidate("internal/a/expensive.go", 1, 900, evidence.RolePrimary, "facet:a"),
		facetCandidate("internal/a/cheap.go", 2, 100, evidence.RolePrimary, "facet:a"),
		facetCandidate("internal/b/owner.go", 3, 200, evidence.RolePrimary, "facet:b"),
		facetCandidate("internal/c/test.go", 4, 50, evidence.RoleSupporting, "facet:c"),
		candidate("internal/noise.go", "lexical", 5),
	}

	ordered, available, claims := prioritizeFacetCoverage(candidates, 1)
	if available != 3 {
		t.Fatalf("available facets = %d, want 3", available)
	}
	want := []string{
		"internal/a/expensive.go",
		"internal/b/owner.go",
		"internal/c/test.go",
	}
	for index, path := range want {
		if ordered[index].Chunk.Path != path {
			t.Fatalf("candidate %d = %s, want %s; order=%+v", index, ordered[index].Chunk.Path, path, ordered)
		}
		if claims[coverageCandidateKey(ordered[index])] == "" {
			t.Fatalf("candidate %s has no facet claim", path)
		}
	}
}

func TestPrioritizeFacetCoverageRequiresDistinctCandidates(t *testing.T) {
	broad := facetCandidate("internal/shared.go", 1, 100, evidence.RolePrimary, "facet:a", "facet:b", "facet:c")
	broad.Context.FacetRanks = map[string]int{"facet:a": 1, "facet:b": 1, "facet:c": 1}
	candidates := []retrieve.Candidate{
		broad,
		facetCandidate("internal/b.go", 2, 100, evidence.RolePrimary, "facet:b"),
		facetCandidate("internal/c.go", 3, 100, evidence.RolePrimary, "facet:c"),
	}

	ordered, available, claims := prioritizeFacetCoverage(candidates, 1)
	if available != 3 {
		t.Fatalf("available facets = %d, want 3", available)
	}
	claimed := make(map[string]struct{})
	for _, candidate := range ordered[:3] {
		claim := claims[coverageCandidateKey(candidate)]
		if claim == "" {
			t.Fatalf("candidate %s has no distinct claim: %+v", candidate.Chunk.Path, claims)
		}
		claimed[claim] = struct{}{}
	}
	if len(claimed) != 3 {
		t.Fatalf("claims = %+v, want three distinct facets", claims)
	}
}

func TestPrioritizeFacetCoverageReservesConfiguredDepth(t *testing.T) {
	candidates := []retrieve.Candidate{
		facetCandidate("a/one.go", 1, 100, evidence.RolePrimary, "facet:a"),
		facetCandidate("a/two.go", 2, 110, evidence.RolePrimary, "facet:a"),
		facetCandidate("a/three.go", 3, 120, evidence.RolePrimary, "facet:a"),
		facetCandidate("b/one.go", 4, 100, evidence.RolePrimary, "facet:b"),
		facetCandidate("b/two.go", 5, 110, evidence.RolePrimary, "facet:b"),
		candidate("noise.go", "lexical", 6),
	}

	ordered, _, claims := prioritizeFacetCoverage(candidates, 2)
	counts := map[string]int{}
	for _, selected := range ordered[:4] {
		counts[claims[coverageCandidateKey(selected)]]++
	}
	if counts["facet:a"] != 2 || counts["facet:b"] != 2 {
		t.Fatalf("front coverage = %+v, want depth two for both facets", counts)
	}
}

func TestDefaultConfigUsesValidatedFacetDepth(t *testing.T) {
	config := DefaultConfig()
	if !config.CoverageAware || config.FacetDepth != 3 {
		t.Fatalf("unexpected production assembly config: %+v", config)
	}
}

func TestPlanWithLegacyConfigPreservesRankedOrder(t *testing.T) {
	candidates := []retrieve.Candidate{
		facetCandidate("internal/x/first.go", 1, 500, evidence.RolePrimary, "facet:a"),
		facetCandidate("internal/x/second.go", 2, 100, evidence.RolePrimary, "facet:a"),
		facetCandidate("internal/x/third.go", 3, 100, evidence.RolePrimary, "facet:b"),
	}
	result := PlanWithConfig(focusedPolicy(100), candidates, nil, LegacyConfig())
	for index, selected := range result.Candidates {
		if selected.Chunk.Path != candidates[index].Chunk.Path {
			t.Fatalf("legacy order changed at %d: got %s want %s", index, selected.Chunk.Path, candidates[index].Chunk.Path)
		}
	}
	if result.Decision.CoverageAware {
		t.Fatal("legacy plan reported coverage-aware assembly")
	}
}

func facetCandidate(path string, rank, tokens int, role evidence.Role, facets ...string) retrieve.Candidate {
	result := candidate(path, "lexical", rank)
	result.Chunk.TokenCount = tokens
	facetRanks := make(map[string]int, len(facets))
	for _, facet := range facets {
		facetRanks[facet] = rank
	}
	result.Context = &evidence.Descriptor{
		GroupIDs:        facets,
		FacetRanks:      facetRanks,
		Roles:           []evidence.Role{role},
		EstimatedTokens: tokens,
	}
	return result
}

func focusedPolicy(target int) queryshape.RetrievalPolicy {
	return queryshape.RetrievalPolicy{Scope: queryshape.ScopeFocused, TargetTokens: target}
}
