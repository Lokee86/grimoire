package agentquery

import (
	"context"
	"fmt"
	"strconv"
)

func (engine *Engine) searchRelationships(
	ctx context.Context,
	request Request,
	seeds []relationshipSeed,
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
	seeds []relationshipSeed,
	limit int,
	response *Response,
) ([]RelationshipMatch, bool) {
	buckets := make([][]RelationshipMatch, 0, len(seeds))
	for _, seed := range seeds {
		resolved, err := engine.arcana.ResolveTyped(ctx, engine.arcanaSnapshot, seed.Node.Name, seed.Node.Kind, seed.Node.Path, 4)
		if err != nil {
			response.Warnings = append(response.Warnings, "Arcana relationship discovery unavailable for seed "+seed.Node.Name+": "+err.Error())
			continue
		}
		bucket := make([]RelationshipMatch, 0, limit)
		seen := make(map[string]bool)
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
				response.Warnings = append(response.Warnings, "Arcana relationship discovery unavailable for seed "+seed.Node.Name+": "+err.Error())
				continue
			}
			subject := engine.node("arcana", engine.arcanaSnapshotID, subjectValue)
			for _, neighbor := range neighbors {
				object := engine.node("arcana", engine.arcanaSnapshotID, neighbor.Node)
				key := relationshipKey(subject, neighbor.Direction, neighbor.Relation, object)
				if seen[key] {
					continue
				}
				seen[key] = true
				seedNode := engine.node(seed.Provider, seed.Snapshot, seed.Node)
				bucket = append(bucket, RelationshipMatch{
					Provider:    "arcana",
					Subject:     subject,
					Direction:   neighbor.Direction,
					Relation:    neighbor.Relation,
					Certainty:   neighbor.Certainty,
					Object:      object,
					Reasons:     relationshipReasons("Arcana direct graph relationship", seed),
					Seed:        &seedNode,
					SeedLane:    seed.Lane,
					SeedRank:    seed.Rank,
					SeedScore:   seed.Score,
					SeedReasons: append([]string(nil), seed.Reasons...),
				})
				if len(bucket) == limit {
					break
				}
			}
			if len(bucket) == limit {
				break
			}
		}
		if len(bucket) > 0 {
			buckets = append(buckets, bucket)
		}
	}
	return interleaveRelationshipBuckets(buckets, limit)
}

func (engine *Engine) lexiconRelationships(
	seeds []relationshipSeed,
	limit int,
) ([]RelationshipMatch, bool) {
	if engine.lexicon == nil {
		return nil, false
	}
	buckets := make([][]RelationshipMatch, 0, len(seeds))
	for _, seed := range seeds {
		if seed.Node.Identity == "" {
			continue
		}
		subject := engine.node("lexicon", engine.lexiconSnapshot, seed.Node)
		impacts := engine.lexicon.Impact([]string{seed.Node.Identity}, "both", nil, 1, limit)
		bucket := make([]RelationshipMatch, 0, len(impacts))
		seen := make(map[string]bool)
		for _, impact := range impacts {
			object := engine.node("lexicon", engine.lexiconSnapshot, impact.Node)
			key := relationshipKey(subject, impact.Direction, impact.Relation, object)
			if seen[key] {
				continue
			}
			seen[key] = true
			spans, evidence := siteRanges(impact.Sites, engine.source.Identity())
			seedNode := engine.node(seed.Provider, seed.Snapshot, seed.Node)
			bucket = append(bucket, RelationshipMatch{
				Provider:    "lexicon",
				Subject:     subject,
				Direction:   impact.Direction,
				Relation:    impact.Relation,
				Certainty:   certainty(impact.Relation),
				Object:      object,
				Reasons:     relationshipReasons("Lexicon direct relationship fallback", seed),
				Evidence:    evidence,
				Spans:       spans,
				Seed:        &seedNode,
				SeedLane:    seed.Lane,
				SeedRank:    seed.Rank,
				SeedScore:   seed.Score,
				SeedReasons: append([]string(nil), seed.Reasons...),
			})
		}
		if len(bucket) > 0 {
			buckets = append(buckets, bucket)
		}
	}
	return interleaveRelationshipBuckets(buckets, limit)
}

func interleaveRelationshipBuckets(buckets [][]RelationshipMatch, limit int) ([]RelationshipMatch, bool) {
	if limit <= 0 {
		return nil, len(buckets) > 0
	}
	indices := make([]int, len(buckets))
	seen := make(map[string]bool)
	ordered := make([]RelationshipMatch, 0, limit)
	for {
		progressed := false
		for bucketIndex, bucket := range buckets {
			for indices[bucketIndex] < len(bucket) {
				match := bucket[indices[bucketIndex]]
				indices[bucketIndex]++
				key := relationshipKey(match.Subject, match.Direction, match.Relation, match.Object)
				if seen[key] {
					continue
				}
				seen[key] = true
				ordered = append(ordered, match)
				progressed = true
				break
			}
		}
		if !progressed {
			break
		}
	}
	truncated := len(ordered) > limit
	if truncated {
		ordered = ordered[:limit]
	}
	for index := range ordered {
		ordered[index].Rank = index + 1
	}
	return ordered, truncated
}

func relationshipReasons(base string, seed relationshipSeed) []string {
	reasons := []string{base}
	if seed.Lane != "" {
		reason := "seeded from " + seed.Lane
		if seed.Rank > 0 {
			reason += " rank " + strconv.Itoa(seed.Rank)
		}
		reasons = append(reasons, reason)
	}
	return reasons
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
