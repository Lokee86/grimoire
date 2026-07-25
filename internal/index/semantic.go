package index

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

const chunkPreparationVersion = "semantic-spans-v1"

// SourceSpan is a language-provider-owned source boundary. Lines are one-based
// and inclusive. Files without valid spans retain generic line-window chunking.
type SourceSpan struct {
	Path      string
	StartLine int
	EndLine   int
	Kind      string
	Name      string
}

func chunkFileWithSourceSpans(path, content string, spans []SourceSpan) ([]Chunk, error) {
	lines := normalizedSourceLines(content)
	if len(lines) == 0 {
		return nil, nil
	}
	spans = canonicalSourceSpans(path, spans, len(lines))
	if len(spans) == 0 {
		return chunkLineRange(nil, path, lines, 0, len(lines))
	}

	chunks := make([]Chunk, 0, len(spans)*2+1)
	cursor := 0
	for _, span := range spans {
		start := span.StartLine - 1
		end := span.EndLine
		var err error
		chunks, err = chunkLineRange(chunks, path, lines, cursor, start)
		if err != nil {
			return nil, err
		}
		firstSemantic := len(chunks)
		chunks, err = appendChunk(chunks, path, lines, start, end)
		if err != nil {
			return nil, err
		}
		for index := firstSemantic; index < len(chunks); index++ {
			chunks[index].SemanticKind = span.Kind
			chunks[index].SemanticName = span.Name
		}
		cursor = end
	}
	return chunkLineRange(chunks, path, lines, cursor, len(lines))
}

func canonicalSourceSpans(path string, spans []SourceSpan, lineCount int) []SourceSpan {
	path = filepath.ToSlash(path)
	byRange := make(map[string]SourceSpan)
	for _, span := range spans {
		span.Path = filepath.ToSlash(span.Path)
		if span.Path != "" && span.Path != path {
			continue
		}
		span.Path = path
		span.Kind = strings.TrimSpace(span.Kind)
		span.Name = strings.TrimSpace(span.Name)
		if span.StartLine <= 0 || span.StartLine > lineCount {
			continue
		}
		if span.EndLine < span.StartLine {
			span.EndLine = span.StartLine
		}
		if span.EndLine > lineCount {
			span.EndLine = lineCount
		}
		key := fmt.Sprintf("%d:%d", span.StartLine, span.EndLine)
		if current, exists := byRange[key]; !exists || preferSourceSpan(span, current) {
			byRange[key] = span
		}
	}

	candidates := make([]SourceSpan, 0, len(byRange))
	for _, span := range byRange {
		candidates = append(candidates, span)
	}
	sortSourceSpans(candidates)

	leaves := make([]SourceSpan, 0, len(candidates))
	for index, span := range candidates {
		containsChild := false
		for otherIndex, other := range candidates {
			if index == otherIndex {
				continue
			}
			if span.StartLine <= other.StartLine && span.EndLine >= other.EndLine &&
				(span.StartLine < other.StartLine || span.EndLine > other.EndLine) {
				containsChild = true
				break
			}
		}
		if !containsChild {
			leaves = append(leaves, span)
		}
	}
	sortSourceSpans(leaves)

	result := make([]SourceSpan, 0, len(leaves))
	for _, span := range leaves {
		if len(result) == 0 || span.StartLine > result[len(result)-1].EndLine {
			result = append(result, span)
			continue
		}
		previous := result[len(result)-1]
		if sourceSpanWidth(span) < sourceSpanWidth(previous) ||
			(sourceSpanWidth(span) == sourceSpanWidth(previous) && preferSourceSpan(span, previous)) {
			result[len(result)-1] = span
		}
	}
	return result
}

func sortSourceSpans(spans []SourceSpan) {
	sort.Slice(spans, func(i, j int) bool {
		if spans[i].StartLine != spans[j].StartLine {
			return spans[i].StartLine < spans[j].StartLine
		}
		if spans[i].EndLine != spans[j].EndLine {
			return spans[i].EndLine < spans[j].EndLine
		}
		if sourceSpanPriority(spans[i].Kind) != sourceSpanPriority(spans[j].Kind) {
			return sourceSpanPriority(spans[i].Kind) > sourceSpanPriority(spans[j].Kind)
		}
		return spans[i].Name < spans[j].Name
	})
}

func preferSourceSpan(left, right SourceSpan) bool {
	leftPriority := sourceSpanPriority(left.Kind)
	rightPriority := sourceSpanPriority(right.Kind)
	if leftPriority != rightPriority {
		return leftPriority > rightPriority
	}
	if left.Name != right.Name {
		return left.Name < right.Name
	}
	return left.Kind < right.Kind
}

func sourceSpanPriority(kind string) int {
	switch strings.ToLower(kind) {
	case "test":
		return 7
	case "method":
		return 6
	case "function":
		return 5
	case "type":
		return 4
	case "constant":
		return 3
	case "variable":
		return 2
	case "import":
		return 1
	default:
		return 0
	}
}

func sourceSpanWidth(span SourceSpan) int {
	return span.EndLine - span.StartLine + 1
}

func chunkPreparationHash(path string, spans []SourceSpan) string {
	normalized := append([]SourceSpan(nil), spans...)
	for index := range normalized {
		normalized[index].Path = filepath.ToSlash(normalized[index].Path)
	}
	sortSourceSpans(normalized)

	hasher := sha256.New()
	_, _ = fmt.Fprintf(hasher, "%s\x00%s\x00", chunkPreparationVersion, filepath.ToSlash(path))
	if len(normalized) == 0 {
		_, _ = hasher.Write([]byte("fallback-line-window"))
	}
	for _, span := range normalized {
		_, _ = fmt.Fprintf(
			hasher,
			"%s\x00%d\x00%d\x00%s\x00%s\x00",
			span.Path, span.StartLine, span.EndLine, span.Kind, span.Name,
		)
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func sourceSpansByPath(spans []SourceSpan) map[string][]SourceSpan {
	result := make(map[string][]SourceSpan)
	for _, span := range spans {
		path := filepath.ToSlash(strings.TrimSpace(span.Path))
		if path == "" || strings.HasPrefix(path, "@") {
			continue
		}
		span.Path = path
		result[path] = append(result[path], span)
	}
	for path := range result {
		sortSourceSpans(result[path])
	}
	return result
}

func semanticChunkCount(chunks []Chunk) int {
	count := 0
	for _, chunk := range chunks {
		if chunk.SemanticKind != "" {
			count++
		}
	}
	return count
}
