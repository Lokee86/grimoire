package lexiconfacts

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/Lokee86/grimoire/internal/evidence"
	"github.com/Lokee86/grimoire/internal/structure"
)

func evidenceForSeeds(seeds []scoredNode, facts library, limit int) []structure.Evidence {
	if limit <= 0 || len(seeds) == 0 {
		return nil
	}
	if len(seeds) > limit {
		seeds = seeds[:limit]
	}
	result := make([]structure.Evidence, 0, len(seeds))
	for index, seed := range seeds {
		node := structureNode(seed.node)
		context := evidence.Descriptor{
			Identity: sourceRangeIdentity(seed.node),
			Roles:    []evidence.Role{evidence.RoleStructural},
			GroupIDs: []string{nodeGroupID(seed.node)},
		}
		if identity := sourceRangeIdentity(seed.node); identity != "" {
			context.Links = []evidence.Link{{Identity: identity, Relation: "source"}}
		}
		result = append(result, structure.Evidence{
			Provider:      source,
			Kind:          "symbol",
			Rank:          index + 1,
			Score:         seed.score,
			Reasons:       append([]string(nil), seed.reasons...),
			Node:          &node,
			Relationships: relationshipsForSeed(seed.node.ID, facts, 12),
			Context:       &context,
		})
	}
	return result
}

func seedNodes(seeds []scoredNode, limit int) []structure.Node {
	if limit <= 0 || len(seeds) == 0 {
		return nil
	}
	if len(seeds) > limit {
		seeds = seeds[:limit]
	}
	result := make([]structure.Node, len(seeds))
	for index, seed := range seeds {
		result[index] = structureNode(seed.node)
	}
	return result
}

func relationshipsForSeed(seedID string, facts library, limit int) []structure.Relationship {
	result := make([]structure.Relationship, 0)
	for _, edge := range facts.edges {
		var direction, relatedID string
		switch seedID {
		case edge.Source:
			direction, relatedID = "outgoing", edge.Target
		case edge.Target:
			direction, relatedID = "incoming", edge.Source
		default:
			continue
		}
		related, exists := facts.nodes[relatedID]
		if !exists || !localNode(related) {
			continue
		}
		result = append(result, structure.Relationship{
			Direction: direction,
			Relation:  edge.Relation,
			Certainty: relationCertainty(edge.Relation),
			Node:      structureNode(related),
		})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Direction != result[j].Direction {
			return result[i].Direction < result[j].Direction
		}
		if result[i].Relation != result[j].Relation {
			return result[i].Relation < result[j].Relation
		}
		if result[i].Node.Path != result[j].Node.Path {
			return result[i].Node.Path < result[j].Node.Path
		}
		return result[i].Node.Name < result[j].Node.Name
	})
	if len(result) > limit {
		result = result[:limit]
	}
	return result
}

func structureNode(node Node) structure.Node {
	result := structure.Node{
		Identity:      node.ID,
		Kind:          node.Kind,
		Name:          node.Name,
		QualifiedName: node.QualifiedName,
		Path:          filepath.ToSlash(nodePath(node)),
	}
	if node.Span != nil {
		result.Span = &structure.Span{
			Path:      filepath.ToSlash(node.Span.Path),
			StartLine: node.Span.StartLine,
			EndLine:   node.Span.EndLine,
		}
	}
	return result
}

func relationCertainty(relation string) string {
	if strings.HasPrefix(relation, "possible-") || strings.Contains(relation, "possible") {
		return "possible"
	}
	return "definite"
}
