package evaluation

import (
	"testing"

	"github.com/Lokee86/grimoire/internal/structure"
)

func TestScoreCaseClassifiesPipelineLosses(t *testing.T) {
	evidence := Evidence{Path: "internal/example.go", Symbols: []string{"Target"}}
	indexed := Candidate{Path: evidence.Path, Text: "func Target() {}"}

	tests := []struct {
		name   string
		query  string
		stages Stages
		want   string
	}{
		{"stale", "find target", Stages{}, FailureStaleOrIncompleteIndex},
		{"provider", "explain subsystem behavior", Stages{Indexed: []Candidate{indexed}}, FailureProviderRetrievalMiss},
		{"ranking", "explain subsystem behavior", Stages{Indexed: []Candidate{indexed}, BroadProbe: []Candidate{indexed}}, FailureRankingCutoffMiss},
		{"exact", "where is Target", Stages{Indexed: []Candidate{indexed}}, FailureExactRecoveryMiss},
		{"merge", "find target", Stages{Indexed: []Candidate{indexed}, Retrieved: []Candidate{indexed}}, FailureCandidateMergeLoss},
		{"curation", "find target", Stages{Indexed: []Candidate{indexed}, Retrieved: []Candidate{indexed}, Merged: []Candidate{indexed}}, FailureCurationLoss},
		{"assembly", "find target", Stages{Indexed: []Candidate{indexed}, Retrieved: []Candidate{indexed}, Merged: []Candidate{indexed}, Curated: []Candidate{indexed}}, FailureAssemblyLoss},
		{"budget", "find target", Stages{Indexed: []Candidate{indexed}, Retrieved: []Candidate{indexed}, Merged: []Candidate{indexed}, Curated: []Candidate{indexed}, Assembled: []Candidate{indexed}}, FailureBudgetFittingLoss},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			run := CaseRun{}
			ScoreCase(Case{Query: test.query, Required: []Evidence{evidence}}, &run, test.stages)
			if len(run.Required) != 1 || run.Required[0].FailureStage != test.want {
				t.Fatalf("got %+v, want %s", run.Required, test.want)
			}
		})
	}
}

func TestScoreCaseStructuralRelationshipAndCallChain(t *testing.T) {
	role := structure.Evidence{
		Provider: "arcana", Kind: "operational_role",
		Node: &structure.Node{Name: "ResolveDamage", Path: "internal/damage.go"},
		Relationships: []structure.Relationship{{
			Direction: "outgoing", Relation: "calls", Certainty: "definite",
			Node: structure.Node{Name: "ApplyShield", Path: "internal/shield.go"},
		}},
	}
	chain := structure.Evidence{
		Provider: "arcana", Kind: "call_chain",
		Chain: &structure.Path{Nodes: []structure.Node{
			{Name: "HandleCollision"}, {Name: "ResolveDamage"}, {Name: "ApplyShield"},
		}},
	}
	entry := Case{RequiredStructural: []StructuralExpectation{
		{Provider: "arcana", Kind: "operational_role", Symbol: "ResolveDamage", Path: "internal/damage.go", Relation: "calls", Direction: "outgoing", Certainty: "definite", TargetSymbol: "ApplyShield", TargetPath: "internal/shield.go"},
		{Provider: "arcana", Kind: "call_chain", Chain: []string{"HandleCollision", "ApplyShield"}},
	}}
	run := CaseRun{StructuralSelections: []StructuralSelection{{Evidence: role}, {Evidence: chain}}}
	stages := Stages{
		StructuralProduced: []structure.Evidence{role, chain},
		StructuralComposed: []structure.Evidence{role, chain},
		StructuralIncluded: []structure.Evidence{role, chain},
	}
	ScoreCase(entry, &run, stages)
	if !run.Pass || run.RequiredStructuralRecall != 1 {
		t.Fatalf("unexpected structural score: %+v", run)
	}
	if run.IrrelevantStructuralRate != 0 {
		t.Fatalf("structural irrelevant rate = %v", run.IrrelevantStructuralRate)
	}
}

func TestScoreCaseClassifiesStructuralPipelineLosses(t *testing.T) {
	expected := StructuralExpectation{Provider: "lexicon", Kind: "symbol", Symbol: "Target"}
	item := structure.Evidence{Provider: "lexicon", Kind: "symbol", Node: &structure.Node{Name: "Target"}}
	tests := []struct {
		name   string
		stages Stages
		want   string
	}{
		{"provider", Stages{}, FailureStructuralProviderMiss},
		{"composition", Stages{StructuralProduced: []structure.Evidence{item}}, FailureStructuralCompositionLoss},
		{"assembly", Stages{StructuralProduced: []structure.Evidence{item}, StructuralComposed: []structure.Evidence{item}}, FailureStructuralAssemblyLoss},
		{"budget", Stages{StructuralProduced: []structure.Evidence{item}, StructuralComposed: []structure.Evidence{item}, StructuralAssembled: []structure.Evidence{item}}, FailureStructuralBudgetFittingLoss},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			run := CaseRun{}
			ScoreCase(Case{RequiredStructural: []StructuralExpectation{expected}}, &run, test.stages)
			if len(run.RequiredStructural) != 1 || run.RequiredStructural[0].FailureStage != test.want {
				t.Fatalf("got %+v, want %s", run.RequiredStructural, test.want)
			}
		})
	}
}

func TestScoreCasePassAndIrrelevance(t *testing.T) {
	required := Evidence{Path: "required.go", Symbols: []string{"Required"}}
	requiredCandidate := Candidate{Path: required.Path, Text: "func Required() {}"}
	irrelevantCandidate := Candidate{Path: "other.go", Text: "package other"}
	run := CaseRun{Selections: []Selection{{Path: required.Path}, {Path: "other.go"}}}
	stages := Stages{
		Indexed: []Candidate{requiredCandidate}, Retrieved: []Candidate{requiredCandidate},
		Merged: []Candidate{requiredCandidate}, Curated: []Candidate{requiredCandidate}, Included: []Candidate{requiredCandidate, irrelevantCandidate},
	}
	ScoreCase(Case{Query: "Required", Required: []Evidence{required}}, &run, stages)
	if !run.Pass || run.RequiredEvidenceRecall != 1 {
		t.Fatalf("unexpected pass result: %+v", run)
	}
	if run.IrrelevantSelectionRate != 0.5 {
		t.Fatalf("irrelevant rate = %v", run.IrrelevantSelectionRate)
	}
}
