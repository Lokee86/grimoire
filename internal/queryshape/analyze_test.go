package queryshape

import (
	"strings"
	"testing"

	"github.com/Lokee86/grimoire/internal/evidence"
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

func TestAnalyzeRecommendsAutomaticBudgetWhenUnspecified(t *testing.T) {
	profile, policy := Analyze(Input{Query: "Explain persistence"})
	if profile.Specificity != LevelMedium || profile.Breadth != LevelLow || profile.Ambiguity != LevelMedium {
		t.Fatalf("unexpected profile: %+v", profile)
	}
	if policy.BudgetMode != "automatic-shadow" || policy.TargetTokens != BoundedTargetTokens || policy.MaximumTokens != BoundedMaximumTokens {
		t.Fatalf("unexpected policy: %+v", policy)
	}
}

func TestAnalyzeDirectLocationRetrievalIntent(t *testing.T) {
	query := "Where is the profile store initialized?"
	_, policy := Analyze(Input{Query: query, RequestedBudget: 4000})

	if len(policy.Intents) != 1 {
		t.Fatalf("direct-location query emitted intents: %+v", policy.Intents)
	}
	intent := policy.Intents[0]
	if intent.Intent != evidence.IntentDirectLocation || intent.Query != query || intent.Weight <= 0 {
		t.Fatalf("unexpected direct-location intent: %+v", intent)
	}
}

func TestAnalyzeCallChainRetrievalIntent(t *testing.T) {
	query := "Trace the call chain from ResolveProfile to the database adapter"
	_, policy := Analyze(Input{Query: query, RequestedBudget: 4000})

	if len(policy.Intents) != 1 || policy.Intents[0].Intent != evidence.IntentCallChain {
		t.Fatalf("unexpected call-chain intents: %+v", policy.Intents)
	}
	if policy.Intents[0].Query != query || policy.Intents[0].Weight <= 0 {
		t.Fatalf("call-chain intent did not preserve query: %+v", policy.Intents[0])
	}
}

func TestAnalyzeArchitectureRetrievalIntent(t *testing.T) {
	query := "Explain the architecture and ownership boundaries of retrieval"
	_, policy := Analyze(Input{Query: query, RequestedBudget: 4000})

	if len(policy.Intents) != 1 || policy.Intents[0].Intent != evidence.IntentArchitecture {
		t.Fatalf("unexpected architecture intents: %+v", policy.Intents)
	}
	if policy.Intents[0].Query != query || policy.Intents[0].Weight <= 0 {
		t.Fatalf("architecture intent did not preserve query: %+v", policy.Intents[0])
	}
}

func TestAnalyzeBoundsLongMixedRetrievalIntents(t *testing.T) {
	query := "Where is profile loading? Explain how persistence selects a store; trace the call chain into compilation, explain the architecture ownership boundaries, where are verification tests, how does caching work, trace the execution flow again, and explain subsystem context."
	_, policy := Analyze(Input{Query: query, RequestedBudget: 9000})

	if len(policy.Intents) < 2 || len(policy.Intents) > maxRetrievalIntentEntries {
		t.Fatalf("mixed query was not bounded: %d intents (%+v)", len(policy.Intents), policy.Intents)
	}
	if policy.Intents[0].Intent != evidence.IntentMixed || policy.Intents[0].Query != query {
		t.Fatalf("mixed query did not preserve the original first: %+v", policy.Intents)
	}
	seen := make(map[string]struct{}, len(policy.Intents))
	for _, intent := range policy.Intents {
		if strings.TrimSpace(intent.Query) == "" || intent.Weight <= 0 {
			t.Fatalf("invalid mixed retrieval intent: %+v", intent)
		}
		key := strings.ToLower(strings.Join(strings.Fields(intent.Query), " "))
		if _, exists := seen[key]; exists {
			t.Fatalf("duplicate mixed retrieval query %q: %+v", intent.Query, policy.Intents)
		}
		seen[key] = struct{}{}
	}
	for _, want := range []evidence.Intent{
		evidence.IntentDirectLocation,
		evidence.IntentMechanism,
		evidence.IntentCallChain,
		evidence.IntentArchitecture,
	} {
		found := false
		for _, intent := range policy.Intents {
			if intent.Intent == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("mixed query omitted %s intent: %+v", want, policy.Intents)
		}
	}
}
