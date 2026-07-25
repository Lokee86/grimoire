package lexiconfacts

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/Lokee86/grimoire/internal/index"
)

// SourceSpans returns declaration-level Lexicon boundaries suitable for source
// preparation. Nested function-owned symbols are omitted so local variables,
// parameters, closures, and other implementation details do not fragment their
// owning declaration.
func (corpus *Corpus) SourceSpans() []index.SourceSpan {
	if corpus == nil {
		return nil
	}
	callableOwned := corpus.callableOwnedNodes()
	result := make([]index.SourceSpan, 0)
	for _, node := range corpus.facts.nodes {
		if node.Span == nil || node.Span.StartLine <= 0 {
			continue
		}
		kind := strings.ToLower(strings.TrimSpace(node.Kind))
		if !sourcePreparationKind(kind) {
			continue
		}
		if callableOwned[node.ID] {
			continue
		}
		path := node.Span.Path
		if path == "" {
			path = node.Path
		}
		path = filepath.ToSlash(strings.TrimSpace(path))
		if path == "" || strings.HasPrefix(path, "@") {
			continue
		}
		end := node.Span.EndLine
		if end < node.Span.StartLine {
			end = node.Span.StartLine
		}
		name := strings.TrimSpace(node.Name)
		if name == "" {
			name = strings.TrimSpace(node.QualifiedName)
		}
		result = append(result, index.SourceSpan{
			Path: path, StartLine: node.Span.StartLine, EndLine: end,
			Kind: kind, Name: name,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Path != result[j].Path {
			return result[i].Path < result[j].Path
		}
		if result[i].StartLine != result[j].StartLine {
			return result[i].StartLine < result[j].StartLine
		}
		if result[i].EndLine != result[j].EndLine {
			return result[i].EndLine < result[j].EndLine
		}
		if result[i].Kind != result[j].Kind {
			return result[i].Kind < result[j].Kind
		}
		return result[i].Name < result[j].Name
	})
	return result
}

func sourcePreparationKind(kind string) bool {
	switch kind {
	case "function", "method", "test", "type":
		return true
	default:
		return false
	}
}

func (corpus *Corpus) callableOwnedNodes() map[string]bool {
	result := make(map[string]bool)
	for _, edge := range corpus.facts.edges {
		if edge.Relation != "defines" {
			continue
		}
		parent, exists := corpus.facts.nodes[edge.Source]
		if !exists {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(parent.Kind)) {
		case "function", "method", "test":
			result[edge.Target] = true
		}
	}
	return result
}
