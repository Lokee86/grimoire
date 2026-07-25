package lexiconfacts

import (
	"encoding/json"
	"path/filepath"

	"github.com/Lokee86/grimoire/internal/structure"
)

func nodesForIDs(ids []string, facts library) []structure.Node {
	seen := make(map[string]struct{}, len(ids))
	result := make([]structure.Node, 0, len(ids))
	for _, id := range ids {
		if id == "" {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		node, exists := facts.nodes[id]
		if !exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, structureNode(node))
	}
	return result
}

func structureSpan(span *Span) *structure.Span {
	if span == nil {
		return nil
	}
	return &structure.Span{
		Path: filepath.ToSlash(span.Path), StartLine: span.StartLine,
		StartColumn: span.StartColumn, EndLine: span.EndLine, EndColumn: span.EndColumn,
	}
}

func spanAttribute(attributes map[string]any, key string) *structure.Span {
	value, exists := attributes[key]
	if !exists {
		return nil
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	path, _ := object["path"].(string)
	start := numberValue(object["start_line"])
	startColumn := numberValue(object["start_column"])
	end := numberValue(object["end_line"])
	endColumn := numberValue(object["end_column"])
	if path == "" && start == 0 && end == 0 {
		return nil
	}
	return &structure.Span{
		Path: filepath.ToSlash(path), StartLine: start, StartColumn: startColumn,
		EndLine: end, EndColumn: endColumn,
	}
}

func stringAttribute(attributes map[string]any, key string) string {
	value, _ := attributes[key].(string)
	return value
}

func intAttribute(attributes map[string]any, key string) int {
	return numberValue(attributes[key])
}

func intPointerAttribute(attributes map[string]any, key string) *int {
	value, exists := attributes[key]
	if !exists {
		return nil
	}
	number := numberValue(value)
	return &number
}

func numberValue(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case float64:
		return int(typed)
	case json.Number:
		parsed, _ := typed.Int64()
		return int(parsed)
	}
	return 0
}

func stringSliceAttribute(attributes map[string]any, key string) []string {
	value, exists := attributes[key]
	if !exists {
		return nil
	}
	var result []string
	switch typed := value.(type) {
	case []string:
		result = append(result, typed...)
	case []any:
		for _, item := range typed {
			if text, ok := item.(string); ok && text != "" {
				result = append(result, text)
			}
		}
	case string:
		if typed != "" {
			result = append(result, typed)
		}
	}
	return result
}

func stringMapAttribute(attributes map[string]any, key string) map[string]string {
	value, exists := attributes[key]
	if !exists {
		return nil
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	result := make(map[string]string, len(object))
	for name, value := range object {
		if text, ok := value.(string); ok {
			result[name] = text
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
