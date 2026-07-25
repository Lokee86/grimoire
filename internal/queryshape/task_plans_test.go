package queryshape

import (
	"strings"
	"testing"

	"github.com/Lokee86/grimoire/internal/evidence"
)

func TestPlanTaskRetrievalSelectsBoundedTaskPlans(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		kind      TaskPlanKind
		firstStep string
		intent    evidence.Intent
	}{
		{name: "implementation", query: "How does ResolveDamage work?", kind: TaskPlanImplementation, firstStep: "implementation", intent: evidence.IntentMechanism},
		{name: "impact", query: "What breaks if I change ResolveDamage?", kind: TaskPlanImpact, firstStep: "change-site", intent: evidence.IntentDirectLocation},
		{name: "tracing", query: "Trace ResolveDamage through the server.", kind: TaskPlanTracing, firstStep: "entry-point", intent: evidence.IntentDirectLocation},
		{name: "debugging", query: "Why does ResolveDamage return zero damage?", kind: TaskPlanDebugging, firstStep: "failure-site", intent: evidence.IntentDirectLocation},
		{name: "architecture", query: "Which package owns damage resolution?", kind: TaskPlanArchitecture, firstStep: "ownership", intent: evidence.IntentArchitecture},
		{name: "verification", query: "What tests cover ResolveDamage?", kind: TaskPlanVerification, firstStep: "verification-evidence", intent: evidence.IntentMechanism},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plan := PlanTaskRetrieval(test.query)
			if plan.Kind != test.kind || plan.Confidence == LevelLow {
				t.Fatalf("plan = %+v, want kind %s with non-low confidence", plan, test.kind)
			}
			if len(plan.Intents) != 5 {
				t.Fatalf("intent count = %d, want 5: %+v", len(plan.Intents), plan.Intents)
			}
			if plan.Intents[0].Intent != evidence.IntentMixed || plan.Intents[0].Step != "task-context" {
				t.Fatalf("missing task context pass: %+v", plan.Intents)
			}
			first := plan.Intents[1]
			if first.Task != test.kind || first.Step != test.firstStep || first.Intent != test.intent {
				t.Fatalf("first task step = %+v", first)
			}
			if !strings.Contains(first.Query, "Retrieval focus:") || first.FacetID == "" {
				t.Fatalf("task step is not an executable focused facet: %+v", first)
			}
		})
	}
}

func TestPlanTaskRetrievalPreservesStructuredDecomposition(t *testing.T) {
	query := "Trace grimoire context through command dispatch, query planning, semantic search, exact recovery, curation, and package serialization."
	plan := PlanTaskRetrieval(query)
	if plan.Kind != TaskPlanTracing {
		t.Fatalf("plan kind = %q, want tracing", plan.Kind)
	}
	if len(plan.Intents) != maxRetrievalIntentEntries {
		t.Fatalf("structured intent count = %d, want %d: %+v", len(plan.Intents), maxRetrievalIntentEntries, plan.Intents)
	}
	for _, intent := range plan.Intents {
		if intent.Task != TaskPlanTracing {
			t.Fatalf("structured facet lost plan annotation: %+v", intent)
		}
		if strings.Contains(intent.Query, "Retrieval focus:") {
			t.Fatalf("structured decomposition was replaced by generic expansion: %+v", intent)
		}
	}
}

func TestTaskPlanPolicyDeclaresTaskSpecificEvidence(t *testing.T) {
	_, policy := Analyze(Input{Query: "What breaks if I change ResolveDamage?", RequestedBudget: 6000})
	if policy.TaskPlan != TaskPlanImpact || policy.TaskPlanConfidence != LevelHigh {
		t.Fatalf("impact policy metadata missing: %+v", policy)
	}
	if policy.ExpansionRadius < 3 || policy.DiversityRequirement < 3 {
		t.Fatalf("impact policy is not broad enough: %+v", policy)
	}
	joined := strings.Join(policy.RequiredEvidence, " ")
	for _, required := range []string{"change-site", "dependents", "regression-coverage"} {
		if !strings.Contains(joined, required) {
			t.Errorf("missing required evidence %q: %+v", required, policy.RequiredEvidence)
		}
	}
}

func TestPlanTaskRetrievalLeavesFocusedLocationUnexpanded(t *testing.T) {
	plan := PlanTaskRetrieval("Where is ResolveDamage defined?")
	if plan.Kind != TaskPlanDirectLocation || len(plan.Intents) != 1 {
		t.Fatalf("focused location plan changed: %+v", plan)
	}
	if plan.Intents[0].Task != TaskPlanDirectLocation || plan.Intents[0].Step != "query-facet" {
		t.Fatalf("focused location plan was not annotated: %+v", plan.Intents)
	}
}
