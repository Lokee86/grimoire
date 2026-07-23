package lexiconfacts

import (
	"strings"
	"unicode"
)

var stopTerms = map[string]struct{}{
	"about": {}, "after": {}, "also": {}, "and": {}, "are": {}, "does": {},
	"explain": {}, "find": {}, "from": {}, "how": {}, "into": {}, "its": {},
	"the": {}, "their": {}, "through": {}, "what": {}, "where": {}, "which": {},
	"with": {},
}

func queryTerms(query string) []string {
	fields := strings.FieldsFunc(strings.ToLower(query), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_'
	})
	seen := make(map[string]struct{}, len(fields))
	result := make([]string, 0, len(fields))
	for _, field := range fields {
		if len(field) < 3 {
			continue
		}
		if _, stop := stopTerms[field]; stop {
			continue
		}
		if _, exists := seen[field]; exists {
			continue
		}
		seen[field] = struct{}{}
		result = append(result, field)
	}
	return result
}

func identifierTerms(value string) []string {
	var terms []string
	var current []rune
	runes := []rune(value)
	flush := func() {
		if len(current) >= 2 {
			terms = append(terms, strings.ToLower(string(current)))
		}
		current = nil
	}
	for index, r := range runes {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			flush()
			continue
		}
		if index > 0 && unicode.IsUpper(r) && len(current) > 0 && unicode.IsLower(current[len(current)-1]) {
			flush()
		}
		current = append(current, r)
	}
	flush()
	return terms
}

func nodePath(node Node) string {
	if node.Span != nil && node.Span.Path != "" {
		return node.Span.Path
	}
	if node.Owner != "" {
		return node.Owner
	}
	return node.Path
}

func localNode(node Node) bool {
	path := nodePath(node)
	return path != "" && !strings.HasPrefix(path, "@")
}

func relationBonus(relation string) float64 {
	switch relation {
	case "calls", "implements", "overrides":
		return 8
	case "contains", "imports", "references":
		return 4
	default:
		return 2
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
