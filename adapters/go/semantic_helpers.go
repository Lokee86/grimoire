package main

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"sort"
)

func (s *scanner) semanticFilePath(set *token.FileSet, file *ast.File) (string, bool) {
	filename := set.PositionFor(file.Pos(), false).Filename
	rel, err := s.relative(filename)
	if err != nil || !isGoFile(rel) {
		return "", false
	}
	return rel, true
}

func (s *scanner) importPathFor(rel string) string {
	dir := filepath.ToSlash(filepath.Dir(rel))
	if dir == "" || dir == "." {
		return s.module
	}
	return s.module + "/" + dir
}

func appendUniqueKey(keys []NodeKey, key NodeKey) []NodeKey {
	for _, existing := range keys {
		if existing == key {
			return keys
		}
	}
	return append(keys, key)
}

func (s *scanner) mergeSemanticCall(key string, incoming semanticCall) {
	existing, exists := s.semanticCalls[key]
	if !exists {
		incoming.edges = uniqueSemanticEdges(incoming.edges)
		s.semanticCalls[key] = incoming
		return
	}
	if incoming.resolved {
		existing.resolved = true
		existing.reason = ""
		existing.namespace = ""
		existing.name = ""
		contract := existing.contract
		if incoming.contract != "" {
			contract = incoming.contract
		}
		existing.contract = contract
		merged := append(append([]semanticEdge(nil), existing.edges...), incoming.edges...)
		concrete := make(map[NodeKey]bool)
		for _, edge := range merged {
			if edge.target != contract && (edge.relation == RelCalls || edge.relation == RelPossibleCalls) {
				concrete[edge.target] = true
			}
		}
		callRelation := RelCalls
		if len(concrete) > 1 {
			callRelation = RelPossibleCalls
		}
		filtered := make([]semanticEdge, 0, len(merged))
		seenTargets := make(map[NodeKey]bool)
		for _, edge := range merged {
			if edge.target == contract && (edge.relation == RelCalls || edge.relation == RelPossibleCalls) {
				continue
			}
			if edge.relation == RelCalls || edge.relation == RelPossibleCalls {
				if seenTargets[edge.target] {
					continue
				}
				seenTargets[edge.target] = true
				edge.relation = callRelation
			}
			filtered = append(filtered, edge)
		}
		existing.edges = uniqueSemanticEdges(filtered)
		if callClassPriority(incoming.class) > callClassPriority(existing.class) {
			existing.class = incoming.class
		}
		s.semanticCalls[key] = existing
		return
	}
	if !existing.resolved && existing.reason == "" {
		s.semanticCalls[key] = incoming
	}
}

func uniqueSemanticEdges(edges []semanticEdge) []semanticEdge {
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].relation != edges[j].relation {
			return edges[i].relation < edges[j].relation
		}
		return edges[i].target < edges[j].target
	})
	result := edges[:0]
	for _, edge := range edges {
		if len(result) == 0 || result[len(result)-1] != edge {
			result = append(result, edge)
		}
	}
	return result
}

func callClassPriority(class callClass) int {
	switch class {
	case callClassInterface:
		return 6
	case callClassDynamic:
		return 5
	case callClassConversion:
		return 4
	case callClassBuiltin:
		return 3
	case callClassExternal:
		return 2
	default:
		return 1
	}
}
