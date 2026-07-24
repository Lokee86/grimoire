package assembly

import (
	"fmt"
	"testing"

	"github.com/Lokee86/grimoire/internal/evidence"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/queryshape"
	"github.com/Lokee86/grimoire/internal/retrieve"
	"github.com/Lokee86/grimoire/internal/structure"
)

func TestPlanFocusedStaysNearExactAnchor(t *testing.T) {
	candidates := []retrieve.Candidate{
		candidate("internal/damage/resolve.go", "exact", 1),
		candidate("internal/damage/model.go", "lexical", 2),
		candidate("internal/damage/resolve_test.go", "adjacent", 3),
		candidate("internal/damage/shield.go", "lexical", 4),
		candidate("internal/damage/status.go", "lexical", 5),
		candidate("internal/damage/status_test.go", "lexical", 6),
		candidate("internal/network/socket.go", "lexical", 7),
	}
	result := Plan(queryshape.RetrievalPolicy{
		Scope: queryshape.ScopeFocused, TargetTokens: queryshape.FocusedTargetTokens,
	}, candidates, nil)
	if len(result.Candidates) != 6 {
		t.Fatalf("expected six focused candidates, got %+v", result.Decision)
	}
	if result.Decision.StopReason != "focused evidence coverage satisfied" {
		t.Fatalf("unexpected stop reason: %+v", result.Decision)
	}
	if len(result.Decision.RegionsRepresented) != 1 || result.Decision.RegionsRepresented[0] != "internal/damage" {
		t.Fatalf("unexpected regions: %+v", result.Decision.RegionsRepresented)
	}
}

func TestPlanBoundedStopsAfterTwoRegionCoverage(t *testing.T) {
	var candidates []retrieve.Candidate
	for index := range 40 {
		region := "damage"
		if index%2 == 1 {
			region = "network"
		}
		candidates = append(candidates, candidate(fmt.Sprintf("internal/%s/file_%02d.go", region, index), "lexical", index+1))
	}
	result := Plan(queryshape.RetrievalPolicy{
		Scope: queryshape.ScopeBounded, TargetTokens: queryshape.BoundedTargetTokens,
	}, candidates, nil)
	if len(result.Candidates) != 36 || len(result.Decision.RegionsRepresented) != 2 {
		t.Fatalf("unexpected bounded plan: %+v", result.Decision)
	}
	if result.Decision.StopReason != "bounded evidence coverage satisfied" {
		t.Fatalf("unexpected stop reason: %+v", result.Decision)
	}
}

func TestPlanExploratoryCapsStructuralEvidence(t *testing.T) {
	regions := []string{"damage", "network", "rooms"}
	var candidates []retrieve.Candidate
	for index := range 30 {
		region := regions[index%len(regions)]
		candidates = append(candidates, candidate(fmt.Sprintf("internal/%s/file_%02d.go", region, index), "lexical", index+1))
	}
	evidence := make([]structure.Evidence, 160)
	result := Plan(queryshape.RetrievalPolicy{
		Scope: queryshape.ScopeExploratory, TargetTokens: queryshape.ExploratoryTargetTokens,
	}, candidates, evidence)
	if len(result.Candidates) != 24 || len(result.Decision.RegionsRepresented) != 3 {
		t.Fatalf("unexpected exploratory plan: %+v", result.Decision)
	}
	if len(result.Structural) != 128 || result.Decision.StructuralSelected != 128 {
		t.Fatalf("unexpected structural cap: %+v", result.Decision)
	}
}

func TestPlanPromotesStructuralGroupSourceAfterCuratedPrefix(t *testing.T) {
	var candidates []retrieve.Candidate
	for index := range 12 {
		candidates = append(candidates, candidate(fmt.Sprintf("internal/damage/file_%02d.go", index), "lexical", index+1))
	}
	candidates[10] = candidateWithContext("internal/damage/trace.go", "lexical", 11, []string{"alpha"}, 700)
	structural := []structure.Evidence{{Context: &evidence.Descriptor{GroupIDs: []string{"alpha"}}}}
	result := Plan(queryshape.RetrievalPolicy{
		Scope: queryshape.ScopeFocused, TargetTokens: 750,
	}, candidates, structural)
	if len(result.Candidates) != 9 || result.Candidates[8].Chunk.Path != candidates[10].Chunk.Path {
		t.Fatalf("structural source anchor was not promoted after curated prefix: %+v", result.Candidates)
	}
	if result.Decision.GroupsRepresented != 1 {
		t.Fatalf("represented groups = %d, want 1", result.Decision.GroupsRepresented)
	}
}

func TestPlanDoesNotActivateGroupsFromSourceCandidates(t *testing.T) {
	candidates := []retrieve.Candidate{
		candidateWithContext("internal/damage/first.go", "exact", 1, []string{"alpha"}, 500),
		candidate("internal/damage/second.go", "lexical", 2),
		candidate("internal/damage/third.go", "lexical", 3),
		candidate("internal/damage/fourth.go", "lexical", 4),
		candidateWithContext("internal/damage/companion.go", "lexical", 5, []string{"alpha"}, 500),
	}
	result := Plan(queryshape.RetrievalPolicy{
		Scope: queryshape.ScopeFocused, TargetTokens: 500,
	}, candidates, nil)
	if len(result.Candidates) != 3 {
		t.Fatalf("selected %d candidates, want 3", len(result.Candidates))
	}
	for index := range result.Candidates {
		if result.Candidates[index].Chunk.Path != candidates[index].Chunk.Path {
			t.Fatalf("source group changed curated order: %+v", result.Candidates)
		}
	}
}

func TestPlanStructuralGroupPriorityRespectsCandidateCap(t *testing.T) {
	candidates := make([]retrieve.Candidate, 0, 170)
	for index := range 170 {
		candidates = append(candidates, candidate(fmt.Sprintf("internal/damage/file_%03d.go", index), "lexical", index+1))
	}
	candidates[169].Context = &evidence.Descriptor{GroupIDs: []string{"call-chain"}}
	structural := []structure.Evidence{{Context: &evidence.Descriptor{GroupIDs: []string{"call-chain"}}}}
	result := Plan(queryshape.RetrievalPolicy{
		Scope: queryshape.ScopeBounded, TargetTokens: 1_000_000_000,
	}, candidates, structural)
	if len(result.Candidates) != 160 {
		t.Fatalf("selected %d candidates, want 160", len(result.Candidates))
	}
	if result.Candidates[preservedCandidatePrefix].Chunk.Path != candidates[169].Chunk.Path {
		t.Fatalf("structural source anchor was not retained before cap: %s", result.Candidates[preservedCandidatePrefix].Chunk.Path)
	}
	if result.Decision.GroupsRepresented != 1 {
		t.Fatalf("represented groups = %d, want 1", result.Decision.GroupsRepresented)
	}
}

func TestPlanUngroupedOrderingRemainsCuratedOrder(t *testing.T) {
	candidates := []retrieve.Candidate{
		candidate("internal/damage/first.go", "exact", 1),
		candidate("internal/damage/second.go", "lexical", 2),
		candidate("internal/damage/third.go", "adjacent", 3),
		candidate("internal/damage/fourth.go", "lexical", 4),
	}
	result := Plan(queryshape.RetrievalPolicy{
		Scope: queryshape.ScopeFocused, TargetTokens: 1000,
	}, candidates, nil)
	if len(result.Candidates) != 3 {
		t.Fatalf("selected %d candidates, want 3", len(result.Candidates))
	}
	for index, selected := range result.Candidates {
		if selected.Chunk.Path != candidates[index].Chunk.Path {
			t.Fatalf("candidate %d = %s, want %s", index, selected.Chunk.Path, candidates[index].Chunk.Path)
		}
	}
	if result.Decision.GroupsRepresented != 0 {
		t.Fatalf("represented groups = %d, want 0", result.Decision.GroupsRepresented)
	}
}

func candidate(path, source string, rank int) retrieve.Candidate {
	return retrieve.Candidate{
		Chunk: index.Chunk{
			ID: fmt.Sprintf("chunk-%d", rank), Path: path,
			StartLine: rank, EndLine: rank, TokenCount: 2000,
		},
		Source: source,
		Rank:   rank,
	}
}

func candidateWithContext(path, source string, rank int, groupIDs []string, estimatedTokens int) retrieve.Candidate {
	candidate := candidate(path, source, rank)
	candidate.Context = &evidence.Descriptor{
		GroupIDs:        groupIDs,
		EstimatedTokens: estimatedTokens,
	}
	return candidate
}
