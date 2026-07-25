package lexiconfacts

import (
	"sort"
	"strings"

	"github.com/Lokee86/grimoire/internal/structure"
)

// Match is a deterministic Lexicon node match with provider-local reasons.
type Match struct {
	Node    structure.Node
	Score   float64
	Reasons []string
}

// PathEdge preserves the relation and occurrence evidence between path nodes.
type PathEdge struct {
	Direction string
	Relation  string
	Certainty string
	Sites     []structure.RelationshipSite
}

// QueryPath is an ordered path through one immutable Lexicon export.
type QueryPath struct {
	Nodes []structure.Node
	Edges []PathEdge
}

// ImpactNode is one bounded graph traversal result.
type ImpactNode struct {
	Depth     int
	Direction string
	Relation  string
	Node      structure.Node
	Sites     []structure.RelationshipSite
}

func (corpus *Corpus) Find(query string, limit int) []Match {
	if corpus == nil || limit <= 0 {
		return nil
	}
	ranked := rankNodes(corpus.facts, query, queryTerms(query))
	if len(ranked) > limit {
		ranked = ranked[:limit]
	}
	result := make([]Match, len(ranked))
	for index, item := range ranked {
		result[index] = Match{
			Node: structureNode(item.node), Score: item.score,
			Reasons: append([]string(nil), item.reasons...),
		}
	}
	return result
}

// Anchors returns compact, deterministic symbols and cross-stack contracts.
func (corpus *Corpus) Anchors(limit int) []structure.Node {
	if corpus == nil || limit <= 0 {
		return nil
	}
	nodes := make([]Node, 0, len(corpus.facts.nodes))
	for _, node := range corpus.facts.nodes {
		if localNode(node) {
			nodes = append(nodes, node)
		}
	}
	sort.Slice(nodes, func(i, j int) bool {
		left, right := anchorPriority(nodes[i]), anchorPriority(nodes[j])
		if left != right {
			return left < right
		}
		if nodes[i].Path != nodes[j].Path {
			return nodes[i].Path < nodes[j].Path
		}
		return nodes[i].ID < nodes[j].ID
	})
	if len(nodes) > limit {
		nodes = nodes[:limit]
	}
	result := make([]structure.Node, len(nodes))
	for index, node := range nodes {
		result[index] = structureNode(node)
	}
	return result
}

func anchorPriority(node Node) int {
	if interstackConnector(node.Kind) {
		return 0
	}
	path := strings.ToLower(node.Path)
	if strings.Contains(path, "test") || strings.Contains(path, "spec") {
		return 2
	}
	if strings.HasSuffix(path, ".md") || strings.Contains(path, "doc") {
		return 3
	}
	switch node.Kind {
	case "class", "interface", "trait", "module", "function", "method":
		return 1
	default:
		return 4
	}
}

// Resolve uses durable node identity first, then exact qualified/name/path matches.
func (corpus *Corpus) Resolve(anchor string, limit int) []structure.Node {
	if corpus == nil || limit <= 0 {
		return nil
	}
	anchor = strings.TrimSpace(anchor)
	var nodes []Node
	if node, exists := corpus.facts.nodes[anchor]; exists {
		nodes = append(nodes, node)
	} else {
		for _, node := range corpus.facts.nodes {
			if node.QualifiedName == anchor || node.Name == anchor || nodePath(node) == anchor {
				nodes = append(nodes, node)
			}
		}
	}
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Path != nodes[j].Path {
			return nodes[i].Path < nodes[j].Path
		}
		return nodes[i].ID < nodes[j].ID
	})
	if len(nodes) > limit {
		nodes = nodes[:limit]
	}
	result := make([]structure.Node, len(nodes))
	for index, node := range nodes {
		result[index] = structureNode(node)
	}
	return result
}

// ResolveSource maps an exact prepared source range to overlapping declarations.
func (corpus *Corpus) ResolveSource(path string, startLine, endLine, limit int) []structure.Node {
	if corpus == nil || limit <= 0 {
		return nil
	}
	path = strings.ReplaceAll(strings.TrimSpace(path), "\\", "/")
	var nodes []Node
	for _, node := range corpus.facts.nodes {
		if node.Span == nil || strings.ReplaceAll(node.Span.Path, "\\", "/") != path {
			continue
		}
		if endLine > 0 && node.Span.StartLine > endLine {
			continue
		}
		if startLine > 0 && node.Span.EndLine < startLine {
			continue
		}
		nodes = append(nodes, node)
	}
	sort.Slice(nodes, func(i, j int) bool {
		leftSize := nodes[i].Span.EndLine - nodes[i].Span.StartLine
		rightSize := nodes[j].Span.EndLine - nodes[j].Span.StartLine
		if leftSize != rightSize {
			return leftSize < rightSize
		}
		return nodes[i].ID < nodes[j].ID
	})
	if len(nodes) > limit {
		nodes = nodes[:limit]
	}
	result := make([]structure.Node, len(nodes))
	for index, node := range nodes {
		result[index] = structureNode(node)
	}
	return result
}
