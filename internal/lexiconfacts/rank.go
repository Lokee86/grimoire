package lexiconfacts

import (
	"math"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Lokee86/grimoire/internal/evidence"
)

func rankNodes(facts library, query string, terms []string) []scoredNode {
	return rankNodesWithDegrees(facts, graphDegrees(facts.edges), query, terms)
}

func rankNodesWithDegrees(facts library, degrees map[string]int, query string, terms []string) []scoredNode {
	lowerQuery := strings.ToLower(query)
	result := make([]scoredNode, 0)
	for _, node := range facts.nodes {
		if !localNode(node) {
			continue
		}
		score, reasons := scoreNode(node, lowerQuery, terms)
		if score <= 0 {
			continue
		}
		result = append(result, scoredNode{
			node: node, score: score, reasons: reasons, primary: true,
			graph: &evidence.GraphSignals{
				Distance:        0,
				ModuleProximity: 1,
				SymbolRole:      node.Kind,
				Centrality:      normalizedCentrality(degrees[node.ID]),
			},
		})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].score != result[j].score {
			return result[i].score > result[j].score
		}
		if result[i].node.Path != result[j].node.Path {
			return result[i].node.Path < result[j].node.Path
		}
		return result[i].node.QualifiedName < result[j].node.QualifiedName
	})
	return result
}

func scoreNode(node Node, query string, terms []string) (float64, []string) {
	name := strings.ToLower(node.Name)
	qualified := strings.ToLower(node.QualifiedName)
	path := strings.ToLower(nodePath(node))
	normalizedQuery := strings.TrimSpace(query)
	var score float64
	var reasons []string
	exactName := len(name) >= 2 && normalizedQuery == name
	explicitName := len(name) >= 2 && strings.Contains(normalizedQuery, name) && (!lowSignalKind(node.Kind) || exactName)
	if explicitName {
		score += 32
		reasons = append(reasons, "query names Lexicon symbol "+node.Name)
	}
	if qualified != "" && normalizedQuery == qualified {
		score += 48
		reasons = append(reasons, "query names Lexicon qualified symbol")
	}
	if path != "" && normalizedQuery == path {
		score += 48
		reasons = append(reasons, "query names Lexicon source path")
	}
	nameTerms := identifierTerms(node.Name)
	matchedTerms := 0
	for _, term := range terms {
		matched := true
		switch {
		case containsString(nameTerms, term):
			score += 9
			reasons = append(reasons, "symbol name matches "+term)
		case strings.Contains(name, term):
			score += 6
			reasons = append(reasons, "symbol name contains "+term)
		case strings.Contains(qualified, term):
			score += 3
			reasons = append(reasons, "qualified symbol matches "+term)
		case strings.Contains(path, term):
			score += 2
			reasons = append(reasons, "symbol path matches "+term)
		default:
			matched = false
		}
		if matched {
			matchedTerms++
		}
	}
	if score == 0 {
		return 0, nil
	}
	if matchedTerms > 1 {
		score += float64((matchedTerms - 1) * 4)
		reasons = append(reasons, "matches multiple query terms")
	}
	score += symbolKindWeight(node.Kind)
	if symbolKindWeight(node.Kind) > 0 {
		reasons = append(reasons, "implementation-level "+node.Kind)
	}
	if strings.Contains(path, "/test") || strings.Contains(path, "_test.") || strings.Contains(path, "/spec") {
		score -= 8
		reasons = append(reasons, "test-path penalty")
	}
	if strings.Contains(path, "/legacy/") {
		score -= 10
		reasons = append(reasons, "legacy-path penalty")
	}
	if lowSignalKind(node.Kind) && !exactName && matchedTerms < 2 {
		return 0, nil
	}
	if score < 9 {
		return 0, nil
	}
	return score, uniqueStrings(reasons)
}

func symbolKindWeight(kind string) float64 {
	switch strings.ToLower(kind) {
	case "http-endpoint", "message-channel", "config-key":
		return 12
	case "function", "method", "class", "type", "interface", "trait", "module", "service", "signal", "event":
		return 6
	case "file", "namespace", "package":
		return 2
	case "parameter", "local", "variable", "field", "property":
		return -6
	default:
		return 0
	}
}

func lowSignalKind(kind string) bool {
	switch strings.ToLower(kind) {
	case "parameter", "local", "variable", "field", "property":
		return true
	default:
		return false
	}
}

type adjacentRelationship struct {
	relatedID string
	direction string
	edge      Edge
}

func expandRelationships(scored map[string]scoredNode, seeds []scoredNode, facts library) {
	expandRelationshipsWithGraph(scored, seeds, facts, graphDegrees(facts.edges), relationshipAdjacency(facts.edges))
}

func expandRelationshipsWithGraph(
	scored map[string]scoredNode,
	seeds []scoredNode,
	facts library,
	degrees map[string]int,
	adjacency map[string][]adjacentRelationship,
) {
	for _, seed := range seeds {
		for _, direct := range adjacency[seed.node.ID] {
			directNode, exists := facts.nodes[direct.relatedID]
			if !exists || !localNode(directNode) {
				continue
			}
			firstRelation := direct.direction + ":" + direct.edge.Relation
			firstScore := seed.score*0.62 + relationBonus(direct.edge.Relation)
			mergeRelationshipCandidate(scored, scoredNode{
				node: directNode, score: firstScore,
				reasons: []string{"Lexicon " + direct.edge.Relation + " relationship from matched symbol"},
				graph:   relationshipGraph(seed.node, directNode, degrees, 1, []string{firstRelation}),
			})
			if !interstackConnector(directNode.Kind) {
				continue
			}
			for _, second := range adjacency[direct.relatedID] {
				if second.relatedID == seed.node.ID {
					continue
				}
				secondNode, exists := facts.nodes[second.relatedID]
				if !exists || !localNode(secondNode) || interstackConnector(secondNode.Kind) {
					continue
				}
				secondRelation := second.direction + ":" + second.edge.Relation
				relations := []string{firstRelation, secondRelation}
				mergeRelationshipCandidate(scored, scoredNode{
					node:  secondNode,
					score: firstScore*0.62 + relationBonus(second.edge.Relation),
					reasons: []string{
						"Lexicon interstack path through " + directNode.Name + " from matched symbol",
					},
					graph: relationshipGraph(seed.node, secondNode, degrees, 2, relations),
				})
			}
		}
	}
}

func relationshipAdjacency(edges []Edge) map[string][]adjacentRelationship {
	result := make(map[string][]adjacentRelationship)
	for _, edge := range edges {
		result[edge.Source] = append(result[edge.Source], adjacentRelationship{
			relatedID: edge.Target, direction: "outgoing", edge: edge,
		})
		result[edge.Target] = append(result[edge.Target], adjacentRelationship{
			relatedID: edge.Source, direction: "incoming", edge: edge,
		})
	}
	return result
}

func interstackConnector(kind string) bool {
	switch kind {
	case "http-endpoint", "message-channel", "config-key":
		return true
	default:
		return false
	}
}

func relationshipGraph(
	seed, node Node,
	degrees map[string]int,
	distance int,
	relations []string,
) *evidence.GraphSignals {
	return &evidence.GraphSignals{
		Distance:        distance,
		Relations:       relations,
		ModuleProximity: moduleProximity(nodePath(seed), nodePath(node)),
		SymbolRole:      node.Kind,
		Centrality:      normalizedCentrality(degrees[node.ID]),
	}
}

func mergeRelationshipCandidate(scored map[string]scoredNode, candidate scoredNode) {
	relatedID := candidate.node.ID
	if existing, exists := scored[relatedID]; exists {
		merged := evidence.Merge(
			evidence.Descriptor{Graph: existing.graph},
			evidence.Descriptor{Graph: candidate.graph},
		)
		if existing.score >= candidate.score {
			existing.reasons = uniqueStrings(append(existing.reasons, candidate.reasons...))
			existing.graph = merged.Graph
			scored[relatedID] = existing
			return
		}
		candidate.reasons = uniqueStrings(append(candidate.reasons, existing.reasons...))
		candidate.primary = existing.primary
		candidate.graph = merged.Graph
	}
	scored[relatedID] = candidate
}

func graphDegrees(edges []Edge) map[string]int {
	degrees := make(map[string]int)
	for _, edge := range edges {
		degrees[edge.Source]++
		degrees[edge.Target]++
	}
	return degrees
}

func normalizedCentrality(degree int) float64 {
	if degree <= 0 {
		return 0
	}
	const saturationDegree = 12
	value := math.Log1p(float64(degree)) / math.Log1p(saturationDegree)
	return min(value, 1)
}

func moduleProximity(left, right string) float64 {
	left = filepath.ToSlash(filepath.Clean(left))
	right = filepath.ToSlash(filepath.Clean(right))
	if left == "." || right == "." || left == "" || right == "" {
		return 0
	}
	if left == right {
		return 1
	}
	leftParts := strings.Split(filepath.ToSlash(filepath.Dir(left)), "/")
	rightParts := strings.Split(filepath.ToSlash(filepath.Dir(right)), "/")
	limit := min(len(leftParts), len(rightParts))
	shared := 0
	for shared < limit && leftParts[shared] == rightParts[shared] {
		shared++
	}
	if shared == 0 {
		return 0
	}
	return float64(shared) / float64(max(len(leftParts), len(rightParts)))
}
