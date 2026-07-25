package extraction

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Lokee86/grimoire/internal/spandiscovery"
)

// LanguageDiscoverer selects query-relevant declaration and section boundaries
// from the language-aware span catalog. Unsupported files return no spans so
// the extractor preserves the prepared chunk.
type LanguageDiscoverer struct{}

func NewLanguageDiscoverer() LanguageDiscoverer {
	return LanguageDiscoverer{}
}

func (LanguageDiscoverer) Discover(request DiscoveryRequest) ([]Span, error) {
	discovered := spandiscovery.Discover(request.Chunk.Path, request.Chunk.Text)
	if len(discovered) == 0 || len(request.Terms) == 0 {
		return nil, nil
	}
	lines := strings.Split(request.Chunk.Text, "\n")
	type scoredSpan struct {
		span  Span
		score int
	}
	scored := make([]scoredSpan, 0, len(discovered))
	for _, candidate := range discovered {
		start := max(0, candidate.StartLine-1)
		end := min(len(lines)-1, candidate.EndLine-1)
		if start > end {
			continue
		}
		score := languageSpanScore(candidate, lines[start:end+1], request.Terms)
		if score < 4 {
			continue
		}
		scored = append(scored, scoredSpan{
			span: Span{
				StartLine: start,
				EndLine:   end,
				Reason: fmt.Sprintf(
					"language-aware %s %s", candidate.Kind, candidate.Name,
				),
			},
			score: score,
		})
	}
	sort.SliceStable(scored, func(left, right int) bool {
		if scored[left].score != scored[right].score {
			return scored[left].score > scored[right].score
		}
		leftLength := scored[left].span.EndLine - scored[left].span.StartLine
		rightLength := scored[right].span.EndLine - scored[right].span.StartLine
		if leftLength != rightLength {
			return leftLength < rightLength
		}
		return scored[left].span.StartLine < scored[right].span.StartLine
	})

	result := make([]Span, 0, len(scored))
	for _, candidate := range scored {
		if overlapsAny(candidate.span, result) {
			continue
		}
		result = append(result, candidate.span)
	}
	return result, nil
}

func languageSpanScore(span spandiscovery.Span, lines []string, terms []string) int {
	name := strings.ToLower(span.Name)
	declaration := ""
	if len(lines) > 0 {
		declaration = strings.ToLower(lines[0])
	}
	content := strings.ToLower(strings.Join(lines, "\n"))
	score := 0
	for _, term := range terms {
		switch {
		case name != "" && name == term:
			score += 16
		case name != "" && len(term) >= 4 && (strings.Contains(name, term) || strings.Contains(term, name)):
			score += 10
		case strings.Contains(declaration, term):
			score += 6
		case strings.Contains(content, term):
			score += termWeight(term)
		}
	}
	return score
}

func overlapsAny(candidate Span, selected []Span) bool {
	for _, existing := range selected {
		if candidate.StartLine <= existing.EndLine && existing.StartLine <= candidate.EndLine {
			return true
		}
	}
	return false
}
