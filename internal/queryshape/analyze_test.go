package queryshape

import (
	"testing"

	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/retrieve"
	"github.com/Lokee86/grimoire/internal/structure"
)

func TestAnalyzeFocusedExactQuery(t *testing.T) {
	exact := []retrieve.Candidate{{
		Chunk: index.Chunk{Path: "internal/profile/store.go"},
		Reasons: []string{
			"identifier \"CreateProfile\" matches content",
			"error code \"ERR_DB_UNAVAILABLE\" matches content",
			"path \"internal/profile/store.go\" matches path",
		},
	}}
	ranked := []retrieve.Candidate{
		{Chunk: index.Chunk{Path: "internal/profile/store.go"}, Score: 100},
		{Chunk: index.Chunk{Path: "internal/profile/store_test.go"}, Score: 30},
	}

	profile, policy := Analyze(Input{
		Query:           "Why does CreateProfile return ERR_DB_UNAVAILABLE in internal/profile/store.go?",
		RequestedBudget: 4000, Exact: exact, Ranked: ranked, Candidates: ranked,
	})

	if profile.Specificity != LevelHigh || profile.Breadth != LevelLow || profile.Ambiguity != LevelLow {
		t.Fatalf("unexpected profile: %+v", profile)
	}
	if profile.ExactSymbolMatches != 1 || profile.ExactPathMatches != 1 || profile.ExactErrorMatches != 1 {
		t.Fatalf("unexpected exact counts: %+v", profile)
	}
	if policy.Scope != "focused" || policy.BudgetMode != "fixed" || policy.TargetTokens != 4000 || !policy.Shadow {
		t.Fatalf("unexpected policy: %+v", policy)
	}
}

func TestAnalyzeBroadCrossSystemQuery(t *testing.T) {
	ranked := []retrieve.Candidate{
		{Chunk: index.Chunk{Path: "internal/retrieve/search.go"}, Score: 10},
		{Chunk: index.Chunk{Path: "internal/selection/selection.go"}, Score: 9.5},
		{Chunk: index.Chunk{Path: "internal/compiler/compiler.go"}, Score: 9},
		{Chunk: index.Chunk{Path: "docs/architecture/system-overview.md"}, Score: 8.5},
	}
	structural := []structure.Evidence{{
		Node: &structure.Node{Path: "internal/retrieve/search.go"},
		Relationships: []structure.Relationship{
			{Node: structure.Node{Path: "internal/selection/selection.go"}},
			{Node: structure.Node{Path: "internal/compiler/compiler.go"}},
		},
	}}

	profile, policy := Analyze(Input{
		Query:           "Explain the architecture and how retrieval, selection, and context compilation work across the system",
		RequestedBudget: 12000, Ranked: ranked, Candidates: ranked, Structural: structural,
	})

	if profile.Breadth != LevelHigh || profile.Ambiguity != LevelHigh {
		t.Fatalf("unexpected profile: %+v", profile)
	}
	if policy.Scope != "exploratory" || policy.ExpansionRadius != 3 || policy.DiversityRequirement != 3 {
		t.Fatalf("unexpected policy: %+v", policy)
	}
}

func TestAnalyzeRepresentsUnspecifiedBudgetWithoutSelectingOne(t *testing.T) {
	profile, policy := Analyze(Input{Query: "Explain persistence"})
	if profile.Specificity != LevelMedium || profile.Breadth != LevelLow || profile.Ambiguity != LevelMedium {
		t.Fatalf("unexpected profile: %+v", profile)
	}
	if policy.BudgetMode != "automatic-shadow" || policy.TargetTokens != 0 || policy.MaximumTokens != 0 {
		t.Fatalf("unexpected policy: %+v", policy)
	}
}
