package lexiconfacts

import (
	"sort"
	"strings"
)

func rankNodes(nodes map[string]Node, query string, terms []string) []scoredNode {
	lowerQuery := strings.ToLower(query)
	result := make([]scoredNode, 0)
	for _, node := range nodes {
		if !localNode(node) {
			continue
		}
		score, reasons := scoreNode(node, lowerQuery, terms)
		if score <= 0 {
			continue
		}
		result = append(result, scoredNode{node: node, score: score, reasons: reasons})
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
	seedScores := make(map[string]float64, len(seeds))
	for _, seed := range seeds {
		seedScores[seed.node.ID] = seed.score
	}
	for _, edge := range facts.edges {
		var relatedID string
		var seedScore float64
		if score, exists := seedScores[edge.Source]; exists {
			relatedID, seedScore = edge.Target, score
		} else if score, exists := seedScores[edge.Target]; exists {
			relatedID, seedScore = edge.Source, score
		} else {
			continue
		}
		node, exists := facts.nodes[relatedID]
		if !exists || !localNode(node) {
			continue
		}
		score := seedScore*0.62 + relationBonus(edge.Relation)
		candidate := scoredNode{
			node: node, score: score,
			reasons: []string{"Lexicon " + edge.Relation + " relationship from matched symbol"},
		}
		if existing, exists := scored[relatedID]; exists && existing.score >= candidate.score {
			continue
		}
		scored[relatedID] = candidate
	}
}
