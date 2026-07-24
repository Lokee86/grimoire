package queryshape

import (
	"strings"

	"github.com/Lokee86/grimoire/internal/evidence"
)

// expandRetrievalClause adds stable implementation vocabulary for task phrases
// whose natural-language wording rarely appears verbatim in source code.
func expandRetrievalClause(query string, intent evidence.Intent) string {
	lower := strings.ToLower(query)
	focus := lower
	if intent == evidence.IntentCallChain {
		if through := strings.LastIndex(focus, " through "); through >= 0 {
			focus = focus[through+len(" through "):]
		}
	}
	terms := make([]string, 0, 8)
	add := func(values ...string) {
		for _, value := range values {
			if !containsWord(lower, value) && !sliceContains(terms, value) {
				terms = append(terms, value)
			}
		}
	}

	if containsAnyText(focus, "command dispatch", "top-level dispatch", "public context command") {
		add("main", "run", "route", "handler", "switch")
	}
	if containsAnyText(focus, "query planning", "query window", "embedding windows") {
		add("plan", "query", "window")
	}
	if containsAnyText(focus, "semantic search", "semantic retrieval") {
		add("semantic", "vector", "search", "candidate")
	}
	if containsAnyText(focus, "exact recovery", "exact path", "exact candidates") {
		add("exact", "retrieve", "signal", "candidate")
	}
	if containsAnyText(focus, "candidate curation", "curation", "curate") {
		add("curate", "selection", "candidate")
	}
	if containsAnyText(focus, "package serialization", "package compilation", "compile a package", "token-budgeted package") {
		add("compile", "marshal", "serialize", "package", "budget")
	}
	if containsAnyText(focus, "corpus loading", "corpus validation") {
		add("load", "validate", "corpus")
	}
	if containsAnyText(focus, "evaluation execution", "context evaluation", "per-mode") {
		add("evaluate", "context", "mode", "run")
	}
	if containsAnyText(focus, "evidence scoring", "stage scoring") {
		add("score", "evidence", "stage")
	}
	if containsAnyText(focus, "aggregate metrics", "aggregate construction") {
		add("aggregate", "metrics", "runs", "build")
	}
	if containsAnyText(focus, "json/markdown reporting", "result-file writing", "baseline report", "stage data") {
		add("report", "write", "json", "markdown")
	}
	if containsAnyText(focus, "bounded concurrent batches", "missing batches") {
		add("batch", "concurrent", "worker")
	}
	if containsAnyText(focus, "publish its manifest", "snapshot manifest") {
		add("manifest", "write", "read", "validate")
	}
	if containsAnyText(focus, "materialize a complete snapshot") {
		add("materialize", "snapshot", "ingest")
	}

	if intent == evidence.IntentArchitecture {
		add("owner", "package")
	}
	if len(terms) == 0 {
		return query
	}
	return strings.TrimSpace(query + " " + strings.Join(terms, " "))
}

func containsWord(value, word string) bool {
	for _, field := range strings.FieldsFunc(value, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_')
	}) {
		if field == word {
			return true
		}
	}
	return false
}

func sliceContains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
