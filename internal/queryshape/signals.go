package queryshape

import (
	"strings"
	"unicode"

	"github.com/Lokee86/grimoire/internal/retrieve"
)

type taskSignal struct {
	name    string
	phrases []string
}

var taskSignals = []taskSignal{
	{name: "debugging", phrases: []string{"debug", "broken", "failure", "fails", "error", "panic", "exception", "why"}},
	{name: "execution-flow", phrases: []string{"trace", "call chain", "caller", "callee", "execution flow", "data flow"}},
	{name: "architecture", phrases: []string{"architecture", "ownership", "boundary", "subsystem", "across the system"}},
	{name: "mechanism", phrases: []string{"explain", "how does", "how do", "mechanism"}},
	{name: "modification", phrases: []string{"implement", "add", "change", "modify", "refactor", "remove", "fix"}},
	{name: "verification", phrases: []string{"test", "verify", "regression", "benchmark"}},
}

func recognizedTasks(query string) []string {
	words := queryWords(query)
	var result []string
	for _, signal := range taskSignals {
		for _, phrase := range signal.phrases {
			matched := strings.ContainsRune(phrase, ' ') && strings.Contains(query, phrase)
			if !matched {
				_, matched = words[phrase]
			}
			if matched {
				result = append(result, signal.name)
				break
			}
		}
	}
	return result
}

func queryWords(query string) map[string]struct{} {
	fields := strings.FieldsFunc(query, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_'
	})
	result := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		result[field] = struct{}{}
	}
	return result
}

func exactCounts(candidates []retrieve.Candidate, query string) (symbols, paths, errors, configs, quoted int) {
	seen := make(map[string]struct{})
	errorIntent := containsTask(recognizedTasks(query), "debugging")
	for _, candidate := range candidates {
		for _, reason := range candidate.Reasons {
			key := reason
			if at := strings.Index(reason, " matches "); at >= 0 {
				key = reason[:at]
			}
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			switch {
			case strings.HasPrefix(key, "identifier "):
				symbols++
			case strings.HasPrefix(key, "path "), strings.HasPrefix(key, "filename "):
				paths++
			case strings.HasPrefix(key, "error code "):
				errors++
			case strings.HasPrefix(key, "configuration key "):
				configs++
			case strings.HasPrefix(key, "quoted phrase "):
				quoted++
				if errorIntent {
					errors++
				}
			}
		}
	}
	return symbols, paths, errors, configs, quoted
}

func containsTask(tasks []string, target string) bool {
	for _, task := range tasks {
		if task == target {
			return true
		}
	}
	return false
}
