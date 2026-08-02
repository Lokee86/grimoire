package agentquery

import "testing"

func TestRankImpactDependentsPrioritizesRelevantProductionBehavior(t *testing.T) {
	request := Request{Query: "session update", Limit: 3}
	candidates := []Dependent{
		{
			Depth: 1, Direction: "incoming", Relation: "calls", Certainty: "definite",
			Node: Node{Name: "TestSessionUpdate", Path: "internal/session/update_test.go", Kind: "function"},
		},
		{
			Depth: 2, Direction: "incoming", Relation: "contains", Certainty: "definite",
			Node: Node{Name: "Session", Path: "internal/session/model.go", Kind: "type"},
		},
		{
			Depth: 1, Direction: "incoming", Relation: "updates", Certainty: "definite",
			Node: Node{Name: "UpdateSession", QualifiedName: "session.UpdateSession", Path: "internal/session/update.go", Kind: "function"},
		},
	}

	ranked := rankImpactDependents(request, candidates, 3)
	if len(ranked) != 3 {
		t.Fatalf("ranked count = %d, want 3", len(ranked))
	}
	if ranked[0].Node.Name != "UpdateSession" || ranked[0].Rank != 1 || ranked[0].Score <= ranked[1].Score {
		t.Fatalf("relevant production behavior was not ranked first: %+v", ranked)
	}
	if ranked[2].Node.Name != "TestSessionUpdate" {
		t.Fatalf("test-only impact was not deprioritized: %+v", ranked)
	}
	if len(ranked[0].Reasons) == 0 {
		t.Fatalf("ranked impact has no explanation: %+v", ranked[0])
	}
}

func TestRankImpactDependentsMergesProviderDuplicates(t *testing.T) {
	span := Range{Path: "internal/session/update.go", StartLine: 10, EndLine: 20}
	candidates := []Dependent{
		{
			Depth: 1, Direction: "incoming", Relation: "calls", Certainty: "definite",
			Node:     Node{Handle: Handle{Provider: "lexicon"}, Name: "UpdateSession", Path: span.Path, Kind: "function"},
			Evidence: []string{"Lexicon call site"}, Spans: []Range{span},
		},
		{
			Depth: 1, Direction: "incoming", Relation: "calls", Certainty: "definite",
			Node:     Node{Handle: Handle{Provider: "arcana"}, Name: "UpdateSession", Path: span.Path, Kind: "function"},
			Evidence: []string{"Arcana bounded graph traversal"}, Spans: []Range{span},
		},
	}

	ranked := rankImpactDependents(Request{}, candidates, 4)
	if len(ranked) != 1 {
		t.Fatalf("provider duplicate count = %d, want 1: %+v", len(ranked), ranked)
	}
	if len(ranked[0].Evidence) != 2 || len(ranked[0].Spans) != 1 {
		t.Fatalf("duplicate evidence was not merged: %+v", ranked[0])
	}
}

func TestRankImpactDependentsKeepsRequestedTestsRelevant(t *testing.T) {
	candidates := []Dependent{
		{Depth: 1, Relation: "calls", Certainty: "definite", Node: Node{Name: "TestRecovery", Path: "internal/recovery/recovery_test.go"}},
		{Depth: 1, Relation: "calls", Certainty: "definite", Node: Node{Name: "Recover", Path: "internal/recovery/recovery.go"}},
	}
	ranked := rankImpactDependents(Request{Query: "recovery test"}, candidates, 2)
	if ranked[0].Node.Name != "TestRecovery" {
		t.Fatalf("explicit test query did not retain test relevance: %+v", ranked)
	}
}
