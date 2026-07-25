package extraction

import (
	"sort"
	"strings"
	"unicode"
)

// LineWindowConfig controls the generic fallback discoverer. It is deliberately
// conservative: language-aware ownership belongs in a separate discoverer.
type LineWindowConfig struct {
	ContextBefore int
	ContextAfter  int
	MaxSpanLines  int
	MergeGap      int
	MinimumScore  int
}

func DefaultLineWindowConfig() LineWindowConfig {
	return LineWindowConfig{
		ContextBefore: 8,
		ContextAfter:  12,
		MaxSpanLines:  24,
		MergeGap:      4,
		MinimumScore:  2,
	}
}

type LineWindowDiscoverer struct {
	config LineWindowConfig
}

func NewLineWindowDiscoverer(config LineWindowConfig) LineWindowDiscoverer {
	defaults := DefaultLineWindowConfig()
	if config.ContextBefore < 0 {
		config.ContextBefore = defaults.ContextBefore
	}
	if config.ContextAfter < 0 {
		config.ContextAfter = defaults.ContextAfter
	}
	if config.MaxSpanLines <= 0 {
		config.MaxSpanLines = defaults.MaxSpanLines
	}
	if config.MergeGap < 0 {
		config.MergeGap = defaults.MergeGap
	}
	if config.MinimumScore <= 0 {
		config.MinimumScore = defaults.MinimumScore
	}
	return LineWindowDiscoverer{config: config}
}

func (discoverer LineWindowDiscoverer) Discover(request DiscoveryRequest) ([]Span, error) {
	lines := strings.Split(request.Chunk.Text, "\n")
	if len(lines) == 0 || len(request.Terms) == 0 {
		return nil, nil
	}
	type scoredLine struct {
		line  int
		score int
	}
	matches := make([]scoredLine, 0)
	queryPhrase := strings.ToLower(strings.TrimSpace(request.Query))
	for lineIndex, line := range lines {
		lower := strings.ToLower(line)
		score := 0
		if len(queryPhrase) >= 8 && strings.Contains(lower, queryPhrase) {
			score += 8
		}
		for _, term := range request.Terms {
			if !strings.Contains(lower, term) {
				continue
			}
			score += termWeight(term)
		}
		if score >= discoverer.config.MinimumScore {
			matches = append(matches, scoredLine{line: lineIndex, score: score})
		}
	}
	if len(matches) == 0 {
		return nil, nil
	}
	sort.SliceStable(matches, func(left, right int) bool {
		if matches[left].score != matches[right].score {
			return matches[left].score > matches[right].score
		}
		return matches[left].line < matches[right].line
	})

	spans := make([]Span, 0, len(matches))
	for _, match := range matches {
		if coveredBySpan(match.line, spans, discoverer.config.MergeGap) {
			continue
		}
		start := max(0, match.line-discoverer.config.ContextBefore)
		end := min(len(lines)-1, match.line+discoverer.config.ContextAfter)
		if end-start+1 > discoverer.config.MaxSpanLines {
			end = start + discoverer.config.MaxSpanLines - 1
			if end < match.line {
				end = match.line
				start = max(0, end-discoverer.config.MaxSpanLines+1)
			}
		}
		spans = append(spans, Span{
			StartLine: start,
			EndLine:   end,
			Reason:    "query term window",
		})
	}
	return spans, nil
}

func coveredBySpan(line int, spans []Span, gap int) bool {
	for _, span := range spans {
		if line >= span.StartLine-gap && line <= span.EndLine+gap {
			return true
		}
	}
	return false
}

func termWeight(term string) int {
	if strings.ContainsAny(term, "_./-") || len(term) >= 10 {
		return 4
	}
	if len(term) >= 6 {
		return 2
	}
	return 1
}

func queryTerms(query string) []string {
	seen := make(map[string]struct{})
	terms := make([]string, 0, 16)
	appendTerm := func(term string) {
		term = strings.ToLower(strings.TrimSpace(term))
		term = strings.Trim(term, "._/-")
		if len(term) < 3 || ignoredQueryTerm(term) {
			return
		}
		if _, exists := seen[term]; exists {
			return
		}
		seen[term] = struct{}{}
		terms = append(terms, term)
	}

	var token []rune
	flush := func() {
		if len(token) == 0 {
			return
		}
		value := string(token)
		appendTerm(value)
		for _, part := range splitIdentifier(value) {
			appendTerm(part)
		}
		token = token[:0]
	}
	for _, current := range query {
		if unicode.IsLetter(current) || unicode.IsDigit(current) || strings.ContainsRune("_./-", current) {
			token = append(token, current)
			continue
		}
		flush()
	}
	flush()
	return terms
}

func splitIdentifier(value string) []string {
	var parts []string
	var current []rune
	flush := func() {
		if len(current) > 0 {
			parts = append(parts, string(current))
			current = current[:0]
		}
	}
	previousLowerOrDigit := false
	for _, character := range value {
		if character == '_' || character == '-' || character == '/' || character == '.' {
			flush()
			previousLowerOrDigit = false
			continue
		}
		if unicode.IsUpper(character) && previousLowerOrDigit {
			flush()
		}
		current = append(current, unicode.ToLower(character))
		previousLowerOrDigit = unicode.IsLower(character) || unicode.IsDigit(character)
	}
	flush()
	return parts
}

func ignoredQueryTerm(term string) bool {
	_, ignored := genericQueryTerms[term]
	return ignored
}

var genericQueryTerms = map[string]struct{}{
	"about": {}, "after": {}, "before": {}, "build": {}, "change": {}, "code": {},
	"does": {}, "explain": {}, "file": {}, "find": {}, "from": {}, "function": {},
	"help": {}, "into": {}, "mechanism": {}, "method": {}, "process": {}, "show": {},
	"system": {}, "that": {}, "their": {}, "these": {}, "this": {}, "through": {},
	"what": {}, "when": {}, "where": {}, "which": {}, "with": {}, "works": {},
}
