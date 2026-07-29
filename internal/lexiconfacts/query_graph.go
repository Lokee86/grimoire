package lexiconfacts

import (
	"strings"

	"github.com/Lokee86/grimoire/internal/structure"
)

func (corpus *Corpus) Trace(
	fromIDs, toIDs []string,
	direction string,
	relations []string,
	maxDepth, limit int,
) []QueryPath {
	if corpus == nil || maxDepth <= 0 || limit <= 0 {
		return nil
	}
	targets := stringSet(toIDs)
	relationSet := stringSet(relations)
	branchLimit := min(max(limit, 8), 16)
	queueLimit := min(max(limit*8, 64), 512)
	if len(targets) > 0 {
		branchLimit = 64
		queueLimit = 4096
	}
	type queuedPath struct {
		current string
		nodes   []Node
		edges   []PathEdge
		visited map[string]bool
	}
	var result []QueryPath
	queue := make([]queuedPath, 0, len(fromIDs))
	globallyVisited := make(map[string]bool, len(fromIDs)+limit*4)
	for _, fromID := range fromIDs {
		start, exists := corpus.facts.nodes[fromID]
		if !exists || globallyVisited[fromID] {
			continue
		}
		globallyVisited[fromID] = true
		queue = append(queue, queuedPath{
			current: fromID,
			nodes:   []Node{start},
			visited: map[string]bool{fromID: true},
		})
	}
	for len(queue) > 0 && len(result) < limit {
		current := queue[0]
		queue = queue[1:]
		if len(current.edges) >= maxDepth {
			continue
		}
		branches := 0
		for _, next := range corpus.queryNeighbors(current.current, direction, relationSet) {
			if current.visited[next.relatedID] || globallyVisited[next.relatedID] {
				continue
			}
			node, exists := corpus.facts.nodes[next.relatedID]
			if !exists {
				continue
			}
			nextNodes := appendNode(current.nodes, node)
			nextEdges := appendPathEdge(current.edges, pathEdge(next.edge, next.direction, corpus.facts))
			isTarget := len(targets) > 0 && targets[next.relatedID]
			shouldAppend := isTarget || (len(targets) == 0 && queryEdgesHaveBehavior(nextEdges))
			if shouldAppend {
				result = append(result, makeQueryPath(nextNodes, nextEdges))
				if len(result) >= limit {
					break
				}
			}
			if isTarget || len(nextEdges) >= maxDepth {
				continue
			}
			nextVisited := make(map[string]bool, len(current.visited)+1)
			for id := range current.visited {
				nextVisited[id] = true
			}
			nextVisited[next.relatedID] = true
			globallyVisited[next.relatedID] = true
			if len(queue) < queueLimit {
				queue = append(queue, queuedPath{
					current: next.relatedID,
					nodes:   nextNodes,
					edges:   nextEdges,
					visited: nextVisited,
				})
			}
			branches++
			if branches >= branchLimit {
				break
			}
		}
	}
	return result
}

func (corpus *Corpus) Impact(
	startIDs []string,
	direction string,
	relations []string,
	maxDepth, limit int,
) []ImpactNode {
	if corpus == nil || maxDepth <= 0 || limit <= 0 {
		return nil
	}
	relationSet := stringSet(relations)
	type queued struct {
		id    string
		depth int
	}
	seen := make(map[string]bool)
	queue := make([]queued, 0, len(startIDs))
	for _, id := range startIDs {
		if _, exists := corpus.facts.nodes[id]; exists && !seen[id] {
			seen[id] = true
			queue = append(queue, queued{id: id})
		}
	}
	var result []ImpactNode
	for len(queue) > 0 && len(result) < limit {
		current := queue[0]
		queue = queue[1:]
		if current.depth >= maxDepth {
			continue
		}
		for _, next := range corpus.queryNeighbors(current.id, direction, relationSet) {
			if seen[next.relatedID] {
				continue
			}
			node, exists := corpus.facts.nodes[next.relatedID]
			if !exists {
				continue
			}
			seen[next.relatedID] = true
			result = append(result, ImpactNode{
				Depth: current.depth + 1, Direction: next.direction,
				Relation: next.edge.Relation, Node: structureNode(node),
				Sites: relationshipSites(next.edge, corpus.facts),
			})
			queue = append(queue, queued{id: next.relatedID, depth: current.depth + 1})
			if len(result) == limit {
				break
			}
		}
	}
	return result
}

func queryEdgesHaveBehavior(edges []PathEdge) bool {
	for _, edge := range edges {
		relation := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(edge.Relation)), "possible-")
		switch relation {
		case "contains", "defines", "imports":
			continue
		default:
			return true
		}
	}
	return false
}

func interstackRelationPriority(relation string) int {
	switch relation {
	case "calls-endpoint", "handled-by", "publishes", "consumes", "reads-config":
		return 0
	case "calls", "possible-calls":
		return 1
	default:
		return 2
	}
}

func pathEdge(edge Edge, direction string, facts library) PathEdge {
	return PathEdge{
		Direction: direction, Relation: edge.Relation,
		Certainty: relationshipCertainty(edge), Sites: relationshipSites(edge, facts),
	}
}

func relationshipSites(edge Edge, facts library) []structure.RelationshipSite {
	site := relationshipSite(edge, facts)
	if emptyRelationshipSite(site) {
		return nil
	}
	return []structure.RelationshipSite{site}
}

func makeQueryPath(nodes []Node, edges []PathEdge) QueryPath {
	result := QueryPath{Nodes: make([]structure.Node, len(nodes)), Edges: append([]PathEdge(nil), edges...)}
	for index, node := range nodes {
		result.Nodes[index] = structureNode(node)
	}
	return result
}

func appendNode(values []Node, value Node) []Node {
	result := make([]Node, len(values)+1)
	copy(result, values)
	result[len(values)] = value
	return result
}

func appendPathEdge(values []PathEdge, value PathEdge) []PathEdge {
	result := make([]PathEdge, len(values)+1)
	copy(result, values)
	result[len(values)] = value
	return result
}

func stringSet(values []string) map[string]bool {
	result := make(map[string]bool, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			result[value] = true
		}
	}
	return result
}
