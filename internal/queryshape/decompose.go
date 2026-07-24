package queryshape

import (
	"strings"
	"unicode"

	"github.com/Lokee86/grimoire/internal/evidence"
)

const maxSpecificRetrievalClauses = maxRetrievalIntentEntries - 1

type retrievalClause struct {
	Query  string
	Intent evidence.Intent
	Topic  string
	Score  int
	Order  int
}

var actionCueWords = map[string]struct{}{
	"add": {}, "aggregate": {}, "aggregates": {}, "build": {}, "builds": {},
	"compare": {}, "compares": {}, "compile": {}, "compiles": {}, "convert": {},
	"create": {}, "creates": {}, "curate": {}, "curates": {}, "debug": {},
	"define": {}, "determine": {}, "dispatch": {}, "dispatches": {}, "establish": {},
	"execute": {}, "executes": {}, "explain": {}, "exposes": {}, "find": {},
	"fit": {}, "fits": {}, "follow": {}, "how": {}, "identify": {},
	"implement": {}, "implements": {}, "ingest": {}, "ingests": {}, "load": {},
	"loads": {}, "locate": {}, "materialize": {}, "materializes": {}, "merge": {},
	"merges": {}, "parse": {}, "parses": {}, "plan": {}, "plans": {},
	"publish": {}, "publishes": {}, "record": {}, "records": {}, "recover": {},
	"recovers": {}, "report": {}, "reports": {}, "reuse": {}, "reuses": {},
	"route": {}, "routes": {}, "run": {}, "runs": {}, "score": {}, "scores": {},
	"search": {}, "searches": {}, "select": {}, "selects": {}, "serialize": {},
	"serializes": {}, "trace": {}, "validate": {}, "validates": {}, "where": {},
	"which": {}, "write": {}, "writes": {},
}

func decomposeRetrievalQuery(query string) []retrievalClause {
	if sections, structured := markdownRetrievalClauses(query); structured {
		return selectRetrievalClauses(sections, maxSpecificRetrievalClauses)
	}
	cleaned := stripFencedBlocks(query)
	clauses := proseRetrievalClauses(cleaned)
	return selectRetrievalClauses(clauses, maxSpecificRetrievalClauses)
}

func proseRetrievalClauses(query string) []retrievalClause {
	var result []retrievalClause
	order := 0
	for _, sentence := range splitStrongPunctuation(query) {
		for _, part := range splitCommaActions(sentence) {
			part = cleanClauseText(part)
			words := strings.Fields(part)
			cue := actionCueIndex(words)
			if cue < 0 {
				continue
			}
			if cue > 5 {
				part = strings.Join(words[cue:], " ")
			}
			part = compactWords(part, 48)
			if len(strings.Fields(part)) < 3 {
				continue
			}
			result = append(result, retrievalClause{
				Query: part, Intent: classifyClauseIntent(part), Topic: clauseTopic(part),
				Score: scoreRetrievalClause(part), Order: order,
			})
			order++
		}
	}
	return mergeAdjacentClauseTopics(result)
}

func splitStrongPunctuation(query string) []string {
	var parts []string
	start := 0
	for index, r := range query {
		separator := r == ';' || r == '!' || r == '?' || r == '\n' || r == '\r'
		if r == '.' {
			next := index + 1
			separator = next == len(query) || next < len(query) && unicode.IsSpace(rune(query[next]))
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

func splitCommaActions(query string) []string {
	var result []string
	for _, commaPart := range strings.Split(query, ",") {
		result = append(result, splitActionConjunctions(commaPart)...)
	}
	return result
}

func splitActionConjunctions(query string) []string {
	words := strings.Fields(query)
	if len(words) == 0 {
		return nil
	}
	var result []string
	start := 0
	for index := 0; index+1 < len(words); index++ {
		if !isClauseConjunction(normalizedWord(words[index])) || !isActionCue(words[index+1]) {
			continue
		}
		// Keep coordinated verbs together when they share a trailing object,
		// as in "merges and curates candidates".
		if index-start < 2 {
			continue
		}
		result = append(result, strings.Join(words[start:index], " "))
		start = index + 1
	}
	if start < len(words) {
		result = append(result, strings.Join(words[start:], " "))
	}
	return result
}

func mergeAdjacentClauseTopics(clauses []retrievalClause) []retrievalClause {
	if len(clauses) < 2 {
		return clauses
	}
	result := make([]retrievalClause, 0, len(clauses))
	for _, clause := range clauses {
		last := len(result) - 1
		if last < 0 || clause.Topic == "" || result[last].Topic != clause.Topic ||
			len(strings.Fields(result[last].Query))+len(strings.Fields(clause.Query)) > 48 {
			result = append(result, clause)
			continue
		}
		result[last].Query += " and " + clause.Query
		result[last].Score = max(result[last].Score, clause.Score) + 2
		if result[last].Intent == evidence.IntentMechanism && clause.Intent != evidence.IntentMechanism {
			result[last].Intent = clause.Intent
		}
	}
	return result
}

func actionCueIndex(words []string) int {
	for index, word := range words {
		if isActionCue(word) {
			return index
		}
	}
	return -1
}

func isActionCue(word string) bool {
	_, exists := actionCueWords[normalizedWord(word)]
	return exists
}

func isClauseConjunction(word string) bool {
	switch word {
	case "and", "also", "then", "while", "but", "plus":
		return true
	default:
		return false
	}
}

func normalizedWord(word string) string {
	return strings.Trim(strings.ToLower(word), " \t\r\n,;:!?()[]{}\"'`")
}

func cleanClauseText(query string) string {
	query = strings.TrimSpace(strings.Trim(query, " \t,;:!?"))
	words := strings.Fields(query)
	for len(words) > 0 && isClauseConjunction(normalizedWord(words[0])) {
		words = words[1:]
	}
	return strings.Join(words, " ")
}

func compactWords(query string, limit int) string {
	words := strings.Fields(query)
	if len(words) <= limit {
		return strings.Join(words, " ")
	}
	return strings.Join(words[:limit], " ")
}

func containsAnyText(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}
