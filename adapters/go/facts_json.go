package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

func encodeFacts(facts RepositoryFacts) string {
	return encodeFactsScoped(facts, nil, nil)
}

func encodeFactsScoped(facts RepositoryFacts, changedFiles, removedFiles []string) string {
	nodes := append([]NodeFact(nil), facts.Nodes...)
	edges := append([]EdgeFact(nil), facts.Edges...)
	unresolved := append([]UnresolvedReferenceFact(nil), facts.Unresolved...)
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Key != nodes[j].Key {
			return nodes[i].Key < nodes[j].Key
		}
		if nodes[i].Kind != nodes[j].Kind {
			return nodes[i].Kind < nodes[j].Kind
		}
		if nodes[i].Path != nodes[j].Path {
			return nodes[i].Path < nodes[j].Path
		}
		return qualifiedName(nodes[i]) < qualifiedName(nodes[j])
	})
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].Source != edges[j].Source {
			return edges[i].Source < edges[j].Source
		}
		if edges[i].Target != edges[j].Target {
			return edges[i].Target < edges[j].Target
		}
		if edges[i].Relation != edges[j].Relation {
			return edges[i].Relation < edges[j].Relation
		}
		return spanText(edges[i].Span) < spanText(edges[j].Span)
	})
	sort.Slice(unresolved, func(i, j int) bool {
		left, right := unresolved[i], unresolved[j]
		if left.Source != right.Source {
			return left.Source < right.Source
		}
		if left.Relation != right.Relation {
			return left.Relation < right.Relation
		}
		if left.Expression != right.Expression {
			return left.Expression < right.Expression
		}
		if left.Reason != right.Reason {
			return left.Reason < right.Reason
		}
		return spanText(left.Span) < spanText(right.Span)
	})

	nodeByKey := make(map[NodeKey]NodeFact, len(nodes))
	for _, node := range nodes {
		nodeByKey[node.Key] = node
	}

	selected := selectedOwners(changedFiles)
	incremental := changedFiles != nil || removedFiles != nil
	var output strings.Builder
	header := map[string]any{
		"adapter_version": "0.1.0",
		"language":        "go",
		"mode":            "full",
		"record":          "lexicon",
		"repository":      facts.Repository,
		"schema_version":  1,
	}
	if incremental {
		header["mode"] = "incremental"
		header["changed_files"] = sortedPaths(changedFiles)
		header["removed_files"] = sortedPaths(removedFiles)
		header["shared_complete"] = true
	}
	writeJSONRecord(&output, header)
	for _, node := range nodes {
		if incremental && !includeOwner(nodeOwner(node), selected) {
			continue
		}
		record := map[string]any{
			"id":             string(node.Key),
			"kind":           string(node.Kind),
			"name":           node.Name,
			"path":           node.Path,
			"qualified_name": qualifiedName(node),
			"record":         "node",
		}
		if node.ContentID != nil {
			record["content_id"] = string(*node.ContentID)
		}
		if node.Span != nil {
			record["span"] = spanValue(node.Span)
			record["owner"] = node.Span.Path
		} else if node.Kind == KindFile {
			record["owner"] = node.Path
		}
		writeJSONRecord(&output, record)
	}
	for _, edge := range edges {
		owner := edgeOwner(edge, nodeByKey)
		if incremental && !includeOwner(owner, selected) {
			continue
		}
		record := map[string]any{
			"record":   "edge",
			"relation": string(edge.Relation),
			"source":   string(edge.Source),
			"target":   string(edge.Target),
		}
		if edge.Span != nil {
			record["span"] = spanValue(edge.Span)
		}
		if owner != "" {
			record["owner"] = owner
		}
		writeJSONRecord(&output, record)
	}
	for _, reference := range unresolved {
		owner := unresolvedOwner(reference, nodeByKey)
		if incremental && !includeOwner(owner, selected) {
			continue
		}
		record := map[string]any{
			"expression": reference.Expression,
			"reason":     string(reference.Reason),
			"record":     "unresolved",
			"relation":   string(reference.Relation),
			"source":     string(reference.Source),
		}
		if reference.CandidateNamespace != "" {
			record["candidate_namespace"] = reference.CandidateNamespace
		}
		if reference.CandidateName != "" {
			record["candidate_name"] = reference.CandidateName
		}
		if reference.Span != nil {
			record["span"] = spanValue(reference.Span)
		}
		if owner != "" {
			record["owner"] = owner
		}
		writeJSONRecord(&output, record)
	}
	return output.String()
}

func qualifiedName(node NodeFact) string {
	switch node.Kind {
	case KindRepository:
		return node.Name
	case KindDirectory, KindFile, KindImport, KindNamespace:
		return node.Path
	default:
		return node.Path + "::" + node.Name
	}
}

func nodeOwner(node NodeFact) string {
	if node.Span != nil {
		return node.Span.Path
	}
	if node.Kind == KindFile {
		return node.Path
	}
	return ""
}

func edgeOwner(edge EdgeFact, nodes map[NodeKey]NodeFact) string {
	if edge.Span != nil {
		return edge.Span.Path
	}
	return nodeOwner(nodes[edge.Source])
}

func unresolvedOwner(reference UnresolvedReferenceFact, nodes map[NodeKey]NodeFact) string {
	if reference.Span != nil {
		return reference.Span.Path
	}
	return nodeOwner(nodes[reference.Source])
}

func selectedOwners(paths []string) map[string]struct{} {
	selected := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		selected[filepath.ToSlash(path)] = struct{}{}
	}
	return selected
}

func includeOwner(owner string, selected map[string]struct{}) bool {
	if owner == "" {
		return true
	}
	_, ok := selected[filepath.ToSlash(owner)]
	return ok
}

func sortedPaths(paths []string) []string {
	result := append([]string(nil), paths...)
	for index := range result {
		result[index] = filepath.ToSlash(result[index])
	}
	sort.Strings(result)
	return result
}

func spanValue(span *SourceSpan) map[string]any {
	return map[string]any{
		"end_column":   span.EndColumn,
		"end_line":     span.EndLine,
		"path":         span.Path,
		"start_column": span.StartColumn,
		"start_line":   span.StartLine,
	}
}

func writeJSONRecord(output *strings.Builder, record map[string]any) {
	encoded, err := json.Marshal(record)
	if err != nil {
		panic(fmt.Sprintf("encode Lexicon fact: %v", err))
	}
	output.Write(encoded)
	output.WriteByte('\n')
}

func spanText(span *SourceSpan) string {
	if span == nil {
		return ""
	}
	return fmt.Sprintf("%s\x00%d\x00%d\x00%d\x00%d", span.Path, span.StartLine, span.StartColumn, span.EndLine, span.EndColumn)
}
