package lexiconfacts

import (
	"path/filepath"
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
	aggregates := make(map[relationshipKey]*relationshipAggregate)
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
		certainty := relationshipCertainty(edge)
		key := relationshipKey{
			direction: direction, relation: edge.Relation,
			certainty: certainty, relatedID: relatedID,
		}
		aggregate := aggregates[key]
		if aggregate == nil {
			aggregate = &relationshipAggregate{value: structure.Relationship{
				Direction: direction, Relation: edge.Relation,
				Certainty: certainty, Node: structureNode(related),
			}}
			aggregates[key] = aggregate
		}
		appendRelationshipSite(aggregate, relationshipSite(edge, facts))
	}
	result := make([]structure.Relationship, 0, len(aggregates))
	for _, aggregate := range aggregates {
		result = append(result, aggregate.value)
	}
	sortRelationships(result)
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
			Path:        filepath.ToSlash(node.Span.Path),
			StartLine:   node.Span.StartLine,
			StartColumn: node.Span.StartColumn,
			EndLine:     node.Span.EndLine,
			EndColumn:   node.Span.EndColumn,
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
