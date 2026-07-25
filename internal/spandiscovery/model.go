package spandiscovery

import "sort"

// Kind describes the source construct bounded by a discovered span.
type Kind string

const (
	KindSection  Kind = "section"
	KindType     Kind = "type"
	KindFunction Kind = "function"
	KindMethod   Kind = "method"
	KindBlock    Kind = "block"
)

// Span identifies a meaningful source range. Lines are one-based and inclusive.
type Span struct {
	StartLine int
	EndLine   int
	Kind      Kind
	Name      string
	Language  string
}

// SmallestContaining returns the narrowest discovered span that fully contains
// the requested source range.
func SmallestContaining(spans []Span, startLine, endLine int) (Span, bool) {
	if startLine <= 0 {
		return Span{}, false
	}
	if endLine < startLine {
		endLine = startLine
	}

	var best Span
	found := false
	for _, span := range spans {
		if span.StartLine > startLine || span.EndLine < endLine {
			continue
		}
		if !found || spanLength(span) < spanLength(best) ||
			(spanLength(span) == spanLength(best) && lessSpan(span, best)) {
			best = span
			found = true
		}
	}
	return best, found
}

// Overlapping returns discovered spans that intersect the requested range,
// ordered from narrowest to widest.
func Overlapping(spans []Span, startLine, endLine int) []Span {
	if startLine <= 0 {
		return nil
	}
	if endLine < startLine {
		endLine = startLine
	}

	result := make([]Span, 0)
	for _, span := range spans {
		if span.EndLine < startLine || span.StartLine > endLine {
			continue
		}
		result = append(result, span)
	}
	sort.Slice(result, func(i, j int) bool {
		left, right := result[i], result[j]
		if spanLength(left) != spanLength(right) {
			return spanLength(left) < spanLength(right)
		}
		return lessSpan(left, right)
	})
	return result
}

func normalizeSpans(spans []Span, lineCount int) []Span {
	result := spans[:0]
	for _, span := range spans {
		if span.StartLine <= 0 || span.StartLine > lineCount {
			continue
		}
		if span.EndLine < span.StartLine {
			span.EndLine = span.StartLine
		}
		if span.EndLine > lineCount {
			span.EndLine = lineCount
		}
		result = append(result, span)
	}
	sort.Slice(result, func(i, j int) bool { return lessSpan(result[i], result[j]) })
	return result
}

func lessSpan(left, right Span) bool {
	if left.StartLine != right.StartLine {
		return left.StartLine < right.StartLine
	}
	if left.EndLine != right.EndLine {
		return left.EndLine < right.EndLine
	}
	if left.Kind != right.Kind {
		return left.Kind < right.Kind
	}
	return left.Name < right.Name
}

func spanLength(span Span) int {
	return span.EndLine - span.StartLine + 1
}
