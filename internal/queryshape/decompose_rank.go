package queryshape

import (
	"sort"
	"strings"

	"github.com/Lokee86/grimoire/internal/evidence"
)

var implementationTerms = []string{
	"aggregate", "baseline", "budget", "candidate", "command", "compilation",
	"compiler", "configuration", "corpus", "curation", "embedding", "evaluation",
	"evidence", "exact", "failure", "format", "identifier", "index", "latency",
	"lexical", "manifest", "metric", "mode", "model", "package", "path",
	"provider", "query", "ranking", "report", "retrieval", "runner", "score",
	"selection", "semantic", "serialization", "snapshot", "stage", "symbol",
	"token", "validation", "vector", "window",
}

func selectRetrievalClauses(clauses []retrievalClause, limit int) []retrievalClause {
	if limit <= 0 {
		return nil
	}
	seenQueries := make(map[string]struct{})
	unique := make([]retrievalClause, 0, len(clauses))
	for _, clause := range clauses {
		key := normalizedQuery(clause.Query)
		if key == "" || clause.Score <= 0 {
			continue
		}
		if _, exists := seenQueries[key]; exists {
			continue
		}
		seenQueries[key] = struct{}{}
		unique = append(unique, clause)
	}
	sort.SliceStable(unique, func(left, right int) bool {
		if unique[left].Score != unique[right].Score {
			return unique[left].Score > unique[right].Score
		}
		return unique[left].Order < unique[right].Order
	})

	selected := make([]retrievalClause, 0, min(limit, len(unique)))
	selectedQueries := make(map[string]struct{})
	appendClause := func(clause retrievalClause) {
		if len(selected) >= limit {
			return
		}
		key := normalizedQuery(clause.Query)
		if _, exists := selectedQueries[key]; exists {
			return
		}
		selectedQueries[key] = struct{}{}
		selected = append(selected, clause)
	}
	seenIntents := make(map[evidence.Intent]struct{})
	for _, clause := range unique {
		if _, exists := seenIntents[clause.Intent]; exists {
			continue
		}
		seenIntents[clause.Intent] = struct{}{}
		appendClause(clause)
	}
	seenTopics := make(map[string]struct{})
	for _, clause := range selected {
		seenTopics[clause.Topic] = struct{}{}
	}
	for _, clause := range unique {
		if _, exists := seenTopics[clause.Topic]; exists {
			continue
		}
		seenTopics[clause.Topic] = struct{}{}
		appendClause(clause)
	}
	for _, clause := range unique {
		appendClause(clause)
	}
	return selected
}

func scoreRetrievalClause(query string) int {
	lower := strings.ToLower(query)
	words := strings.Fields(query)
	score := 0
	if actionCueIndex(words) >= 0 {
		score += 4
	}
	for _, term := range implementationTerms {
		if strings.Contains(lower, term) {
			score++
		}
	}
	if len(words) >= 4 && len(words) <= 48 {
		score += 2
	}
	if strings.Contains(lower, "/") || containsAnyText(lower, ".go", ".rs", ".py", ".json", ".md") {
		score += 2
	}
	if containsAnyText(lower, "suggested location", "suggested output", "example", "completion criteria", "deliverables") {
		score -= 5
	}
	if len(words) > 64 {
		score -= (len(words) - 64) / 8
	}
	return score
}

func clauseTopic(query string) string {
	lower := strings.ToLower(query)
	switch {
	case containsAnyText(lower, "architecture", "ownership", "boundary", "components", "which package"):
		return "architecture"
	case containsAnyText(lower, "report", "evaluation", "score", "metric", "baseline", "mode", "stage data"):
		return "evaluation"
	case containsAnyText(lower, "budget", "package", "compile", "serialization", "token"):
		return "package"
	case containsAnyText(lower, "query window", "embedding", "semantic", "vector", "lexical", "search", "retrieval"):
		return "retrieval"
	case containsAnyText(lower, "curate", "curation", "selection", "merge", "candidate"):
		return "selection"
	case containsAnyText(lower, "exact", "symbol", "identifier", "configuration key", " path"):
		return "exact"
	case containsAnyText(lower, "trace", "follow", "call chain", "execution flow", "dispatch"):
		return "flow"
	default:
		return string(classifyClauseIntent(query))
	}
}
