package agentquery

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/structure"
)

func (engine *Engine) node(provider, snapshot string, value structure.Node) Node {
	result := Node{
		Handle: nodeHandle(provider, snapshot, value),
		Kind:   value.Kind, Name: value.Name,
		QualifiedName: value.QualifiedName, Path: normalizePath(value.Path),
	}
	if value.Span != nil {
		span := rangeFromStructure(*value.Span, engine.source.Identity())
		result.Span = &span
	}
	return result
}

func rangeFromStructure(span structure.Span, snapshot string) Range {
	path := normalizePath(span.Path)
	end := span.EndLine
	if end < span.StartLine {
		end = span.StartLine
	}
	return Range{
		Path: path, StartLine: span.StartLine, StartColumn: span.StartColumn,
		EndLine: end, EndColumn: span.EndColumn,
		Handle: sourceHandle(snapshot, path, span.StartLine, end),
	}
}

func (engine *Engine) sourceNode(chunk index.Chunk) Node {
	path := normalizePath(chunk.Path)
	span := Range{
		Path: path, StartLine: chunk.StartLine, EndLine: chunk.EndLine,
		Handle: sourceHandle(engine.source.Identity(), path, chunk.StartLine, chunk.EndLine),
	}
	return Node{
		Handle: span.Handle, Kind: sourceKind(chunk),
		Name: chunk.SemanticName, Path: path, Span: &span,
	}
}

func sourceKind(chunk index.Chunk) string {
	if chunk.SemanticKind != "" {
		return chunk.SemanticKind
	}
	return "source-range"
}

func siteRanges(sites []structure.RelationshipSite, snapshot string) ([]Range, []string) {
	var ranges []Range
	var evidence []string
	for _, site := range sites {
		if site.Span != nil {
			ranges = append(ranges, rangeFromStructure(*site.Span, snapshot))
		}
		evidence = append(evidence, site.Evidence...)
		if site.Expression != "" {
			evidence = append(evidence, "expression: "+site.Expression)
		}
	}
	return ranges, unique(evidence)
}

func unique(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" && !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}

func classifyPath(path string) (string, string) {
	lower := strings.ToLower(filepath.ToSlash(path))
	switch {
	case strings.Contains(lower, "/test") || strings.Contains(lower, "_test.") ||
		strings.Contains(lower, "/spec") || strings.Contains(lower, ".spec."):
		return "test", "repository test anchor"
	case strings.HasSuffix(lower, ".md") || strings.Contains(lower, "/docs/"):
		return "document", "repository documentation anchor"
	case strings.Contains(lower, "contract") || strings.Contains(lower, "schema") ||
		strings.Contains(lower, "config") || strings.Contains(lower, "route"):
		return "contract", "repository contract or configuration anchor"
	default:
		return "file", "repository source anchor"
	}
}

func handleKey(handle Handle) string {
	nodeID := ""
	if handle.NodeID != nil {
		nodeID = fmt.Sprintf("%d", *handle.NodeID)
	}
	return handle.Provider + "\x00" + handle.NodeIdentity + "\x00" +
		nodeID + "\x00" + handle.Path +
		fmt.Sprintf("\x00%d\x00%d", handle.StartLine, handle.EndLine)
}
