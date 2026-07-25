package queryshape

import (
	"strings"

	"github.com/Lokee86/grimoire/internal/evidence"
)

const taskPlanContextWeight = 0.30

type taskPlanStep struct {
	name   string
	intent evidence.Intent
	focus  string
	weight float64
}

// PlanTaskRetrieval selects one deterministic task plan and returns the
// bounded retrieval passes that providers execute. Long or explicitly
// structured prompts retain their clause decomposition and are only annotated
// with the selected task plan.
func PlanTaskRetrieval(query string) TaskPlan {
	query = strings.TrimSpace(query)
	if query == "" {
		return TaskPlan{}
	}
	lower := strings.ToLower(query)
	tasks := recognizedTasks(lower)
	kind, confidence := classifyTaskPlan(tasks)
	intents := retrievalIntents(query, tasks)
	if shouldExpandTaskPlan(query, kind, intents) {
		intents = expandedTaskPlanIntents(query, kind)
	} else {
		intents = annotateTaskPlanIntents(intents, kind)
	}
	return TaskPlan{Kind: kind, Confidence: confidence, Intents: intents}
}

func classifyTaskPlan(tasks []string) (TaskPlanKind, Level) {
	switch {
	case containsTask(tasks, "execution-flow"):
		return TaskPlanTracing, LevelHigh
	case containsTask(tasks, "impact"):
		return TaskPlanImpact, LevelHigh
	case containsTask(tasks, "architecture"):
		return TaskPlanArchitecture, LevelHigh
	case containsTask(tasks, "debugging"):
		return TaskPlanDebugging, LevelHigh
	case containsTask(tasks, "verification"):
		return TaskPlanVerification, LevelHigh
	case containsTask(tasks, "modification"):
		return TaskPlanImplementation, LevelHigh
	case containsTask(tasks, "mechanism"):
		return TaskPlanImplementation, LevelMedium
	case containsTask(tasks, "location"):
		return TaskPlanDirectLocation, LevelHigh
	default:
		return TaskPlanGeneral, LevelLow
	}
}

func shouldExpandTaskPlan(query string, kind TaskPlanKind, intents []RetrievalIntent) bool {
	if kind == TaskPlanGeneral || kind == TaskPlanDirectLocation || looksStructuredQuery(query) {
		return false
	}
	return len(intents) <= 2 && len(strings.Fields(query)) <= 40
}

func expandedTaskPlanIntents(query string, kind TaskPlanKind) []RetrievalIntent {
	steps := taskPlanSteps(kind)
	if len(steps) == 0 {
		return annotateTaskPlanIntents(retrievalIntents(query, recognizedTasks(strings.ToLower(query))), kind)
	}
	result := []RetrievalIntent{{
		Intent: evidence.IntentMixed, Query: query, Weight: taskPlanContextWeight,
		Task: kind, Step: "task-context",
	}}
	for _, step := range steps {
		planned := retrievalIntent(step.intent, taskFocusedQuery(query, step.focus), step.weight, true)
		planned.Task = kind
		planned.Step = step.name
		planned.FacetID = evidence.StableID(
			"facet", string(kind), step.name, normalizedQuery(query),
		)
		result = append(result, planned)
		if len(result) == maxRetrievalIntentEntries {
			break
		}
	}
	return result
}

func annotateTaskPlanIntents(intents []RetrievalIntent, kind TaskPlanKind) []RetrievalIntent {
	result := append([]RetrievalIntent(nil), intents...)
	for index := range result {
		result[index].Task = kind
		if result[index].Step != "" {
			continue
		}
		if result[index].Intent == evidence.IntentMixed {
			result[index].Step = "task-context"
			continue
		}
		result[index].Step = "query-facet"
	}
	return result
}

func taskFocusedQuery(query, focus string) string {
	return strings.TrimSpace(query) + "\nRetrieval focus: " + focus + "."
}

func taskPlanSteps(kind TaskPlanKind) []taskPlanStep {
	switch kind {
	case TaskPlanImplementation:
		return []taskPlanStep{
			{name: "implementation", intent: evidence.IntentMechanism, focus: "definition, implementation body, owned state, and lifecycle", weight: 1.00},
			{name: "execution-context", intent: evidence.IntentCallChain, focus: "entry points, callers, callees, data flow, and side effects", weight: 0.95},
			{name: "contracts", intent: evidence.IntentArchitecture, focus: "interfaces, types, configuration, serialization, and subsystem ownership", weight: 0.90},
			{name: "verification", intent: evidence.IntentMechanism, focus: "tests, fixtures, contracts, regressions, and expected behavior", weight: 0.90},
		}
	case TaskPlanImpact:
		return []taskPlanStep{
			{name: "change-site", intent: evidence.IntentDirectLocation, focus: "changed symbol, declaration, implementation body, and owned state", weight: 1.00},
			{name: "dependents", intent: evidence.IntentCallChain, focus: "incoming callers, transitive dependents, consumers, and downstream effects", weight: 1.00},
			{name: "contracts", intent: evidence.IntentArchitecture, focus: "interfaces, implementations, configuration, serialization, and compatibility boundaries", weight: 0.95},
			{name: "regression-coverage", intent: evidence.IntentMechanism, focus: "tests, failure handling, contracts, and regression coverage", weight: 0.95},
		}
	case TaskPlanTracing:
		return []taskPlanStep{
			{name: "entry-point", intent: evidence.IntentDirectLocation, focus: "entry point, starting symbol, and dispatch boundary", weight: 1.00},
			{name: "call-path", intent: evidence.IntentCallChain, focus: "ordered callers, callees, intermediate transitions, and terminal call", weight: 1.00},
			{name: "state-flow", intent: evidence.IntentMechanism, focus: "arguments, state mutations, data flow, side effects, and terminal behavior", weight: 0.95},
			{name: "trace-tests", intent: evidence.IntentMechanism, focus: "tests and fixtures that execute the same path", weight: 0.85},
		}
	case TaskPlanDebugging:
		return []taskPlanStep{
			{name: "failure-site", intent: evidence.IntentDirectLocation, focus: "error message, failing condition, panic, exception, and return site", weight: 1.00},
			{name: "propagation", intent: evidence.IntentCallChain, focus: "callers, execution path, inputs, propagation, and recovery", weight: 0.95},
			{name: "state-and-config", intent: evidence.IntentMechanism, focus: "state mutations, guards, configuration, invariants, and error handling", weight: 0.95},
			{name: "reproduction", intent: evidence.IntentMechanism, focus: "tests, fixtures, reproductions, regressions, and expected behavior", weight: 0.95},
		}
	case TaskPlanArchitecture:
		return []taskPlanStep{
			{name: "ownership", intent: evidence.IntentArchitecture, focus: "subsystem ownership, package boundaries, public entry points, and responsibilities", weight: 1.00},
			{name: "contracts", intent: evidence.IntentDirectLocation, focus: "interfaces, types, registries, factories, adapters, and public contracts", weight: 0.95},
			{name: "representative-implementation", intent: evidence.IntentMechanism, focus: "representative implementations, lifecycle, state ownership, and configuration", weight: 0.90},
			{name: "cross-boundary-flow", intent: evidence.IntentCallChain, focus: "cross-boundary calls, dependencies, events, and data flow", weight: 0.95},
		}
	case TaskPlanVerification:
		return []taskPlanStep{
			{name: "verification-evidence", intent: evidence.IntentMechanism, focus: "tests, specs, benchmarks, fixtures, assertions, and expected outcomes", weight: 1.00},
			{name: "implementation-under-test", intent: evidence.IntentDirectLocation, focus: "implementation under test, public contract, and behavior boundary", weight: 0.95},
			{name: "integration-paths", intent: evidence.IntentCallChain, focus: "callers, integration paths, failure paths, and side effects", weight: 0.90},
			{name: "compatibility", intent: evidence.IntentArchitecture, focus: "interfaces, configuration, serialization, and compatibility boundaries", weight: 0.85},
		}
	default:
		return nil
	}
}

func applyTaskPlanPolicy(policy *RetrievalPolicy) {
	switch policy.TaskPlan {
	case TaskPlanImplementation:
		policy.RequiredEvidence = []string{"implementation", "execution-context", "contracts-and-configuration", "verification"}
		policy.StopConditions = []string{"implementation represented", "direct execution context represented", "contracts or configuration represented", "verification evidence represented"}
	case TaskPlanImpact:
		policy.ExpansionRadius = max(policy.ExpansionRadius, 3)
		policy.DiversityRequirement = max(policy.DiversityRequirement, 3)
		policy.RequiredEvidence = []string{"change-site", "incoming-and-transitive-dependents", "contracts-and-compatibility", "regression-coverage"}
		policy.StopConditions = []string{"change site represented", "dependent paths represented", "compatibility boundaries represented", "regression evidence represented"}
	case TaskPlanTracing:
		policy.ExpansionRadius = max(policy.ExpansionRadius, 3)
		policy.RequiredEvidence = []string{"entry-point", "ordered-call-path", "state-or-data-flow", "terminal-behavior"}
		policy.StopConditions = []string{"entry point represented", "intermediate path represented", "terminal behavior represented", "remaining paths are redundant"}
	case TaskPlanDebugging:
		policy.ExpansionRadius = max(policy.ExpansionRadius, 2)
		policy.RequiredEvidence = []string{"failure-site", "propagation-path", "state-and-configuration", "reproduction-or-regression"}
		policy.StopConditions = []string{"failure site represented", "propagation path represented", "relevant state or configuration represented", "reproduction evidence represented"}
	case TaskPlanArchitecture:
		policy.ExpansionRadius = max(policy.ExpansionRadius, 3)
		policy.DiversityRequirement = max(policy.DiversityRequirement, 3)
		policy.RequiredEvidence = []string{"ownership-and-boundaries", "public-contracts", "representative-implementations", "cross-boundary-relationships"}
		policy.StopConditions = []string{"major ownership regions represented", "public contracts represented", "representative implementations covered", "cross-boundary relationships covered"}
	case TaskPlanVerification:
		policy.DiversityRequirement = max(policy.DiversityRequirement, 2)
		policy.RequiredEvidence = []string{"tests-or-benchmarks", "implementation-under-test", "integration-and-failure-paths", "contracts-and-compatibility"}
		policy.StopConditions = []string{"verification evidence represented", "implementation under test represented", "integration or failure paths represented", "contracts represented"}
	}
}
