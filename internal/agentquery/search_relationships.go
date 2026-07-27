package agentquery

import (
	"context"
	"fmt"

	"github.com/Lokee86/grimoire/internal/structure"
)

func (engine *Engine) searchRelationships(
	ctx context.Context,
	request Request,
	seeds []structure.Node,
	response *Response,
) []RelationshipMatch {
	if len(seeds) == 0 {
		return nil
	}
	if engine.arcanaSnapshot != "" {
		matches, truncated := engine.arcanaRelationships(ctx, seeds, request.Limit, response)
		markLaneTruncated(response, "relationship_matches", truncated)
		if len(matches) > 0 {
			return matches
		}
	}
	matches, truncated := engine.lexiconRelationships(seeds, request.Limit)
	markLaneTruncated(response, "relationship_matches", truncated)
	return matches
}

func (engine *Engine) arcanaRelationships(
	ctx context.Context,
	seeds []structure.Node,
	limit int,
	response *Response,
) ([]RelationshipMatch, bool) {
	matches := make([]RelationshipMatch, 0, limit)
	seen := make(map[string]bool)
	for _, seed := range seeds {
		resolved, err := engine.arcana.Resolve(ctx, engine.arcanaSnapshot, seed.Name, seed.Path, 4)
		if err != nil {
			response.Warnings = append(response.Warnings, "Arcana relationship discovery unavailable: "+err.Error())
			return matches, false
		}
		for _, subjectValue := range resolved {
			if subjectValue.NodeID == nil {
				continue
			}
			neighbors, err := engine.arcana.Neighbors(
				ctx,
				engine.arcanaSnapshot,
				*subjectValue.NodeID,
				"both",
				nil,
			)
			if err != nil {
				response.Warnings = append(response.Warnings, "Arcana relationship discovery unavailable: "+err.Error())
				return matches, false
			}
			subject := engine.node("arcana", engine.arcanaSnapshotID, subjectValue)
			for _, neighbor := range neighbors {
				object := engine.node("arcana", engine.arcanaSnapshotID, neighbor.Node)
				key := relationshipKey(subject, neighbor.Direction, neighbor.Relation, object)
				if seen[key] {
					continue
				}
				seen[key] = true
				matches = append(matches, RelationshipMatch{
					Rank:      len(matches) + 1,
					Provider:  "arcana",
					Subject:   subject,
					Direction: neighbor.Direction,
					Relation:  neighbor.Relation,
					Certainty: neighbor.Certainty,
					Object:    object,
					Reasons:   []string{"Arcana direct graph relationship"},
				})
				if len(matches) == limit {
					return matches, true
				}
			}
		}
	}
	return matches, false
}

func (engine *Engine) lexiconRelationships(
	seeds []structure.Node,
	limit int,
) ([]RelationshipMatch, bool) {
	if engine.lexicon == nil {
		return nil, false
	}
	matches := make([]RelationshipMatch, 0, limit)
	seen := make(map[string]bool)
	for _, seed := range seeds {
		if seed.Identity == "" {
			continue
		}
		subject := engine.node("lexicon", engine.lexiconSnapshot, seed)
		impacts := engine.lexicon.Impact([]string{seed.Identity}, "both", nil, 1, limit)
		for _, impact := range impacts {
			object := engine.node("lexicon", engine.lexiconSnapshot, impact.Node)
			key := relationshipKey(subject, impact.Direction, impact.Relation, object)
			if seen[key] {
				continue
			}
			seen[key] = true
			spans, evidence := siteRanges(impact.Sites, engine.source.Identity())
			matches = append(matches, RelationshipMatch{
				Rank:      len(matches) + 1,
				Provider:  "lexicon",
				Subject:   subject,
				Direction: impact.Direction,
				Relation:  impact.Relation,
				Certainty: certainty(impact.Relation),
				Object:    object,
				Reasons:   []string{"Lexicon direct relationship fallback"},
				Evidence:  evidence,
				Spans:     spans,
			})
			if len(matches) == limit {
				return matches, true
			}
		}
	}
	return matches, false
}

func relationshipKey(subject Node, direction, relation string, object Node) string {
	return fmt.Sprintf(
		"%s\x00%s\x00%s\x00%s",
		handleKey(subject.Handle),
		direction,
		relation,
		handleKey(object.Handle),
	)
}
