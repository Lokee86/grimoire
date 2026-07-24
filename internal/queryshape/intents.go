package queryshape

import (
	"strings"
	"unicode"

	"github.com/Lokee86/grimoire/internal/evidence"
)

const maxRetrievalIntentEntries = 6

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

// retrievalIntents preserves the complete query as the first entry for a
// mixed request, then adds a bounded, stable set of clause queries.
func retrievalIntents(query string, tasks []string) []RetrievalIntent {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}

	clauses := queryClauses(query)
	intents := mappedIntents(tasks)
	if len(intents) == 0 {
		return []RetrievalIntent{{Intent: evidence.IntentMixed, Query: query, Weight: 1}}
	}
	if len(intents) == 1 && len(clauses) <= 1 {
		return []RetrievalIntent{{Intent: intents[0], Query: query, Weight: 1}}
	}

	result := []RetrievalIntent{{Intent: evidence.IntentMixed, Query: query, Weight: 1}}
	seenQueries := map[string]struct{}{normalizedQuery(query): {}}
	for _, clause := range clauses {
		clauseIntents := mappedIntents(recognizedTasks(strings.ToLower(clause)))
		if len(clauseIntents) == 0 {
			continue
		}
		key := normalizedQuery(clause)
		if key == "" {
			continue
		}
		if _, seen := seenQueries[key]; seen {
			continue
		}
		seenQueries[key] = struct{}{}
		result = append(result, RetrievalIntent{
			Intent: clauseIntents[0], Query: clause,
			Weight: 1 / float64(len(result)+1),
		})
		if len(result) == maxRetrievalIntentEntries {
			break
		}
	}

	// A mixed task may have no useful boundary, but the original query remains
	// a valid bounded retrieval request even when no clause could be emitted.
	if len(result) == 1 {
		result = append(result, RetrievalIntent{
			Intent: intents[0], Query: query, Weight: 0.5,
		})
	}
	return result
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

func queryClauses(query string) []string {
	var clauses []string
	for _, part := range splitStrongPunctuation(query) {
		for _, clause := range splitConjunctions(part) {
			clause = cleanClause(clause)
			if clause != "" {
				clauses = append(clauses, clause)
			}
		}
	}
	return clauses
}

func splitStrongPunctuation(query string) []string {
	var parts []string
	start := 0
	for index, r := range query {
		separator := r == ';' || r == '!' || r == '?' || r == '\n' || r == '\r'
		if r == '.' {
			next := index + 1
			separator = next == len(query) || (next < len(query) && unicode.IsSpace(rune(query[next])))
		}
		if !separator {
			continue
		}
		parts = append(parts, query[start:index])
		start = index + 1
	}
	parts = append(parts, query[start:])
	return parts
}

func splitConjunctions(query string) []string {
	var commaClauses []string
	commaParts := strings.Split(query, ",")
	start := 0
	for index := 1; index < len(commaParts); index++ {
		left := strings.Join(commaParts[start:index], ",")
		right := strings.Join(commaParts[index:], ",")
		if !hasRetrievalCue(left) || !startsWithRetrievalCue(right) {
			continue
		}
		commaClauses = append(commaClauses, left)
		start = index
	}
	commaClauses = append(commaClauses, strings.Join(commaParts[start:], ","))

	var clauses []string
	for _, commaClause := range commaClauses {
		clauses = append(clauses, splitConjunctionWords(commaClause)...)
	}
	return clauses
}

func splitConjunctionWords(query string) []string {
	words := strings.Fields(query)
	if len(words) < 3 {
		return wordsToClauses(words)
	}

	var clauses []string
	start := 0
	for index, word := range words {
		if !isConjunction(strings.ToLower(strings.Trim(word, ",:;"))) {
			continue
		}
		left := strings.Join(words[start:index], " ")
		right := strings.Join(words[index+1:], " ")
		if left == "" || right == "" || !hasRetrievalCue(left) || !startsWithRetrievalCue(right) {
			continue
		}
		clauses = append(clauses, left)
		start = index + 1
	}
	clauses = append(clauses, strings.Join(words[start:], " "))
	return clauses
}

func wordsToClauses(words []string) []string {
	if len(words) == 0 {
		return nil
	}
	return []string{strings.Join(words, " ")}
}

func isConjunction(word string) bool {
	switch word {
	case "and", "also", "then", "while", "but", "plus":
		return true
	default:
		return false
	}
}

func hasRetrievalCue(query string) bool {
	return len(mappedIntents(recognizedTasks(strings.ToLower(query)))) > 0
}

func startsWithRetrievalCue(query string) bool {
	words := strings.Fields(strings.ToLower(query))
	if len(words) == 0 {
		return false
	}
	first := strings.Trim(words[0], " 	,;:!?\"")
	switch first {
	case "where", "locate", "find", "explain", "how", "trace", "follow", "caller", "callee", "architecture", "why", "debug", "mechanism":
		return true
	default:
		return false
	}
}

func cleanClause(clause string) string {
	return strings.TrimSpace(strings.Trim(clause, " \t,;:!?"))
}

func normalizedQuery(query string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(query))), " ")
}
