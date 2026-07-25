package lexiconfacts

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/Lokee86/grimoire/internal/structure"
)

const relationshipSiteLimit = 3

type relationshipKey struct {
	direction string
	relation  string
	certainty string
	relatedID string
}

type relationshipAggregate struct {
	value structure.Relationship
}

func relationshipSite(edge Edge, facts library) structure.RelationshipSite {
	attributes := edge.Attributes
	site := structure.RelationshipSite{
		Span:           structureSpan(edge.Span),
		Evidence:       stringSliceAttribute(attributes, "evidence"),
		DefinitionSpan: spanAttribute(attributes, "macro_definition_span"),
		CandidateCount: intAttribute(attributes, "candidate_count"),
		Indirect:       stringAttribute(attributes, "indirect"),
		BodyCallee:     stringAttribute(attributes, "macro_body_callee"),
		ExpansionDepth: intPointerAttribute(attributes, "expansion_depth"),
		MacroCallIndex: intPointerAttribute(attributes, "macro_call_index"),
		Substitutions:  stringMapAttribute(attributes, "substitutions"),
		Arguments:      stringSliceAttribute(attributes, "substituted_arguments"),
		ArgumentIndex:  intPointerAttribute(attributes, "argument_index"),
		Expression:     stringAttribute(attributes, "expression"),
	}
	viaIDs := stringSliceAttribute(attributes, "via")
	if viaCall := stringAttribute(attributes, "via_call"); viaCall != "" {
		viaIDs = append(viaIDs, viaCall)
	}
	site.Via = nodesForIDs(viaIDs, facts)
	return site
}

func relationshipCertainty(edge Edge) string {
	if resolution := stringAttribute(edge.Attributes, "resolution"); resolution == "definite" || resolution == "possible" {
		return resolution
	}
	return relationCertainty(edge.Relation)
}

func appendRelationshipSite(aggregate *relationshipAggregate, site structure.RelationshipSite) {
	aggregate.value.Occurrences++
	if emptyRelationshipSite(site) {
		return
	}
	for _, existing := range aggregate.value.Sites {
		if relationshipSiteKey(existing) == relationshipSiteKey(site) {
			return
		}
	}
	if len(aggregate.value.Sites) >= relationshipSiteLimit {
		aggregate.value.SitesTruncated = true
		return
	}
	aggregate.value.Sites = append(aggregate.value.Sites, site)
}

func relationshipSiteKey(site structure.RelationshipSite) string {
	data, _ := json.Marshal(site)
	return string(data)
}

func emptyRelationshipSite(site structure.RelationshipSite) bool {
	return site.Span == nil && len(site.Evidence) == 0 && len(site.Via) == 0 &&
		site.DefinitionSpan == nil && site.CandidateCount == 0 && site.Indirect == "" &&
		site.BodyCallee == "" && site.ExpansionDepth == nil && site.MacroCallIndex == nil &&
		len(site.Substitutions) == 0 && len(site.Arguments) == 0 &&
		site.ArgumentIndex == nil && site.Expression == ""
}

func relationshipPriority(relationship structure.Relationship) (int, int, string) {
	priority := map[string]int{
		"calls": 0, "possible-calls": 1, "passes-to": 2,
		"reads": 3, "writes": 4, "references": 5,
		"implements": 6, "extends": 7, "includes": 8,
	}
	relationRank, exists := priority[relationship.Relation]
	if !exists {
		relationRank = 20
	}
	directionRank := 1
	if relationship.Direction == "outgoing" {
		directionRank = 0
	}
	nodeKey := fmt.Sprintf(
		"%s\x00%s\x00%s", relationship.Node.Path,
		relationship.Node.QualifiedName, relationship.Node.Name,
	)
	return relationRank, directionRank, nodeKey
}

func sortRelationships(values []structure.Relationship) {
	sort.Slice(values, func(i, j int) bool {
		leftRelation, leftDirection, leftNode := relationshipPriority(values[i])
		rightRelation, rightDirection, rightNode := relationshipPriority(values[j])
		if leftRelation != rightRelation {
			return leftRelation < rightRelation
		}
		if leftDirection != rightDirection {
			return leftDirection < rightDirection
		}
		return leftNode < rightNode
	})
}
