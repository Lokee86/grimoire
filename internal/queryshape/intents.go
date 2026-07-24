package queryshape

import (
	"strings"

	"github.com/Lokee86/grimoire/internal/evidence"
)

const maxRetrievalIntentEntries = 7

// PlanRetrievalIntents derives the bounded query-only retrieval plan used
// before candidate generation. It intentionally depends only on the request
// text so retrieval can consume intents before full query profiling.
func PlanRetrievalIntents(query string) []RetrievalIntent {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}
	return retrievalIntents(query, recognizedTasks(strings.ToLower(query)))
}

func retrievalIntents(query string, tasks []string) []RetrievalIntent {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}

	intents := mappedIntents(tasks)
	clauses := decomposeRetrievalQuery(query)
	if len(clauses) <= 1 && len(intents) == 1 && !looksStructuredQuery(query) {
		return []RetrievalIntent{retrievalIntent(intents[0], query, 1, true)}
	}
	if len(clauses) == 0 {
		if len(intents) == 1 {
			return []RetrievalIntent{retrievalIntent(intents[0], query, 1, true)}
		}
		return []RetrievalIntent{retrievalIntent(evidence.IntentMixed, query, 1, true)}
	}

	result := []RetrievalIntent{retrievalIntent(
		evidence.IntentMixed,
		query,
		mixedQueryWeight(query),
		false,
	)}
	seenQueries := map[string]struct{}{normalizedQuery(query): {}}
	for _, clause := range clauses {
		key := normalizedQuery(clause.Query)
		if key == "" {
			continue
		}
		if _, seen := seenQueries[key]; seen {
			continue
		}
		seenQueries[key] = struct{}{}
		result = append(result, retrievalIntent(
			clause.Intent,
			clause.Query,
			clauseWeight(clause.Score),
			true,
		))
		if len(result) == maxRetrievalIntentEntries {
			break
		}
	}
	if len(result) == 1 && len(intents) > 0 {
		result = append(result, retrievalIntent(intents[0], query, 1, true))
	}
	return result
}

func retrievalIntent(intent evidence.Intent, query string, weight float64, facet bool) RetrievalIntent {
	planned := RetrievalIntent{Intent: intent, Query: query, Weight: weight}
	if facet {
		planned.FacetID = evidence.StableID("facet", string(intent), normalizedQuery(query))
	}
	return planned
}

func mappedIntents(tasks []string) []evidence.Intent {
	var result []evidence.Intent
	for _, task := range tasks {
		var intent evidence.Intent
		switch task {
		case "location":
			intent = evidence.IntentDirectLocation
		case "execution-flow":
			intent = evidence.IntentCallChain
		case "architecture":
			intent = evidence.IntentArchitecture
		case "mechanism", "debugging":
			intent = evidence.IntentMechanism
		default:
			continue
		}
		if !containsIntent(result, intent) {
			result = append(result, intent)
		}
	}
	if len(result) > 1 && containsSpecificIntent(result) {
		filtered := result[:0]
		for _, intent := range result {
			if intent != evidence.IntentMechanism {
				filtered = append(filtered, intent)
			}
		}
		result = filtered
	}
	return result
}

func classifyClauseIntent(query string) evidence.Intent {
	lower := strings.ToLower(query)
	switch {
	case containsAnyText(lower, "trace ", "follow ", "call chain", "execution flow", "data flow"):
		return evidence.IntentCallChain
	case containsAnyText(lower, "architecture", "ownership", " owns ", "boundary", "which package", "which components"):
		return evidence.IntentArchitecture
	case containsAnyText(lower, "where ", "locate ", "find ", "exact ", "symbol ", "identifier ", "configuration key", " path "):
		return evidence.IntentDirectLocation
	case containsAnyText(lower, "score", "report", "metric", "baseline", "runner", "evaluation format", "corpus model"):
		return evidence.IntentMechanism
	}
	intents := mappedIntents(recognizedTasks(lower))
	if len(intents) > 0 {
		return intents[0]
	}
	return evidence.IntentMechanism
}

func containsSpecificIntent(intents []evidence.Intent) bool {
	return containsIntent(intents, evidence.IntentDirectLocation) ||
		containsIntent(intents, evidence.IntentCallChain) ||
		containsIntent(intents, evidence.IntentArchitecture)
}

func containsIntent(intents []evidence.Intent, target evidence.Intent) bool {
	for _, intent := range intents {
		if intent == target {
			return true
		}
	}
	return false
}

func mixedQueryWeight(query string) float64 {
	words := len(strings.Fields(query))
	switch {
	case looksStructuredQuery(query), words > 180:
		return 0.15
	case words > 40:
		return 0.25
	default:
		return 0.4
	}
}

func clauseWeight(score int) float64 {
	if score < 0 {
		score = 0
	}
	if score > 10 {
		score = 10
	}
	return 0.75 + float64(score)*0.025
}

func normalizedQuery(query string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(query))), " ")
}
