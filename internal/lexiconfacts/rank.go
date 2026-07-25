package lexiconfacts

import (
	"math"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Lokee86/grimoire/internal/evidence"
)

func rankNodes(facts library, query string, terms []string) []scoredNode {
	lowerQuery := strings.ToLower(query)
	degrees := graphDegrees(facts.edges)
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
	var score float64
	var reasons []string
	if len(name) >= 2 && strings.Contains(query, name) {
		score += 32
		reasons = append(reasons, "query names Lexicon symbol "+node.Name)
	}
	if qualified != "" && strings.Contains(query, qualified) {
		score += 48
		reasons = append(reasons, "query names Lexicon qualified symbol")
	}
	if path != "" && strings.Contains(query, path) {
		score += 48
		reasons = append(reasons, "query names Lexicon source path")
	}
	nameTerms := identifierTerms(node.Name)
	for _, term := range terms {
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
		}
	}
	if score < 9 {
		return 0, nil
	}
	return score, uniqueStrings(reasons)
}

func expandRelationships(scored map[string]scoredNode, seeds []scoredNode, facts library) {
	seedNodes := make(map[string]scoredNode, len(seeds))
	degrees := graphDegrees(facts.edges)
	for _, seed := range seeds {
		seedNodes[seed.node.ID] = seed
	}
	for _, edge := range facts.edges {
		var relatedID, direction string
		var seed scoredNode
		if matched, exists := seedNodes[edge.Source]; exists {
			relatedID, direction, seed = edge.Target, "outgoing", matched
		} else if matched, exists := seedNodes[edge.Target]; exists {
			relatedID, direction, seed = edge.Source, "incoming", matched
		} else {
			continue
		}
		node, exists := facts.nodes[relatedID]
		if !exists || !localNode(node) {
			continue
		}
		relation := direction + ":" + edge.Relation
		candidate := scoredNode{
			node:  node,
			score: seed.score*0.62 + relationBonus(edge.Relation),
			reasons: []string{
				"Lexicon " + edge.Relation + " relationship from matched symbol",
			},
			graph: &evidence.GraphSignals{
				Distance:        1,
				Relations:       []string{relation},
				ModuleProximity: moduleProximity(nodePath(seed.node), nodePath(node)),
				SymbolRole:      node.Kind,
				Centrality:      normalizedCentrality(degrees[node.ID]),
			},
		}
		if existing, exists := scored[relatedID]; exists {
			merged := evidence.Merge(
				evidence.Descriptor{Graph: existing.graph},
				evidence.Descriptor{Graph: candidate.graph},
			)
			if existing.score >= candidate.score {
				existing.reasons = uniqueStrings(append(existing.reasons, candidate.reasons...))
				existing.graph = merged.Graph
				scored[relatedID] = existing
				continue
			}
			candidate.reasons = uniqueStrings(append(candidate.reasons, existing.reasons...))
			candidate.primary = existing.primary
			candidate.graph = merged.Graph
		}
		scored[relatedID] = candidate
	}
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
