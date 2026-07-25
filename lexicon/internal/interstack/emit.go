package interstack

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

func Encode(result Result) ([]byte, error) {
	repository := result.Repository
	if repository == "" {
		repository = "repository"
	}
	header := map[string]any{
		"adapter_version": AdapterVersion,
		"language":        Language,
		"mode":            "full",
		"record":          "lexicon",
		"repository":      repository,
		"schema_version":  1,
	}
	var output bytes.Buffer
	if err := writeJSONLine(&output, header); err != nil {
		return nil, err
	}
	nodes := append([]Node(nil), result.Nodes...)
	sort.Slice(nodes, func(left, right int) bool {
		if nodes[left].ID != nodes[right].ID {
			return nodes[left].ID < nodes[right].ID
		}
		if nodes[left].Kind != nodes[right].Kind {
			return nodes[left].Kind < nodes[right].Kind
		}
		return nodes[left].QualifiedName < nodes[right].QualifiedName
	})
	for _, node := range nodes {
		record := map[string]any{
			"id":             node.ID,
			"kind":           node.Kind,
			"name":           node.Name,
			"path":           node.Path,
			"qualified_name": node.QualifiedName,
			"record":         "node",
		}
		if len(node.Attributes) > 0 {
			record["attributes"] = node.Attributes
		}
		if node.Span != nil {
			record["span"] = node.Span
		}
		if err := writeJSONLine(&output, record); err != nil {
			return nil, err
		}
	}
	edges := append([]factEdge(nil), result.Edges...)
	sort.Slice(edges, func(left, right int) bool {
		return edgeSortKey(edges[left]) < edgeSortKey(edges[right])
	})
	for _, edge := range edges {
		record := map[string]any{
			"record":   "edge",
			"relation": edge.Relation,
			"source":   edge.Source,
			"target":   edge.Target,
		}
		if edge.Span != nil {
			record["span"] = edge.Span
		}
		if len(edge.Attributes) > 0 {
			record["attributes"] = edge.Attributes
		}
		if err := writeJSONLine(&output, record); err != nil {
			return nil, err
		}
	}
	unresolved := append([]factUnresolved(nil), result.Unresolved...)
	sort.Slice(unresolved, func(left, right int) bool {
		return unresolvedSortKey(unresolved[left]) < unresolvedSortKey(unresolved[right])
	})
	for _, item := range unresolved {
		record := map[string]any{
			"expression": item.Expression,
			"reason":     item.Reason,
			"record":     "unresolved",
			"relation":   item.Relation,
			"source":     item.Source,
		}
		if item.Span != nil {
			record["span"] = item.Span
		}
		if len(item.Attributes) > 0 {
			record["attributes"] = item.Attributes
		}
		if err := writeJSONLine(&output, record); err != nil {
			return nil, err
		}
	}
	return output.Bytes(), nil
}

func writeJSONLine(output *bytes.Buffer, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	output.Write(data)
	output.WriteByte('\n')
	return nil
}

func stableNodeID(kind, identity string) string {
	payload := "lexicon:v1\x00" + Language + "\x00" + kind + "\x00" + identity
	digest := sha256.Sum256([]byte(payload))
	return "sha256:" + hex.EncodeToString(digest[:])
}

func edgeSortKey(edge factEdge) string {
	return strings.Join([]string{edge.Source, edge.Target, edge.Relation, spanSortKey(edge.Span)}, "\x00")
}

func unresolvedSortKey(item factUnresolved) string {
	return strings.Join([]string{item.Source, item.Relation, item.Expression, item.Reason, spanSortKey(item.Span)}, "\x00")
}

func spanSortKey(span *Span) string {
	if span == nil {
		return ""
	}
	return fmt.Sprintf("%s:%010d:%010d:%010d:%010d", span.Path, span.StartLine, span.StartColumn, span.EndLine, span.EndColumn)
}
