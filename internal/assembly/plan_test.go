package assembly

import (
	"fmt"
	"testing"

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
		candidate("internal/network/socket.go", "lexical", 4),
	}
	result := Plan(queryshape.RetrievalPolicy{Scope: queryshape.ScopeFocused}, candidates, nil)
	if len(result.Candidates) != 3 {
		t.Fatalf("expected three focused candidates, got %+v", result.Decision)
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
	for index := range 20 {
		region := "damage"
		if index%2 == 1 {
			region = "network"
		}
		candidates = append(candidates, candidate(fmt.Sprintf("internal/%s/file_%02d.go", region, index), "lexical", index+1))
	}
	result := Plan(queryshape.RetrievalPolicy{Scope: queryshape.ScopeBounded}, candidates, nil)
	if len(result.Candidates) != 12 || len(result.Decision.RegionsRepresented) != 2 {
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
	evidence := make([]structure.Evidence, 80)
	result := Plan(queryshape.RetrievalPolicy{Scope: queryshape.ScopeExploratory}, candidates, evidence)
	if len(result.Candidates) != 24 || len(result.Decision.RegionsRepresented) != 3 {
		t.Fatalf("unexpected exploratory plan: %+v", result.Decision)
	}
	if len(result.Structural) != 64 || result.Decision.StructuralSelected != 64 {
		t.Fatalf("unexpected structural cap: %+v", result.Decision)
	}
}

func candidate(path, source string, rank int) retrieve.Candidate {
	return retrieve.Candidate{
		Chunk:  index.Chunk{ID: fmt.Sprintf("chunk-%d", rank), Path: path, StartLine: rank, EndLine: rank},
		Source: source,
		Rank:   rank,
	}
}
