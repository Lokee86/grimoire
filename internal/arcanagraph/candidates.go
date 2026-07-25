package arcanagraph

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Lokee86/grimoire/internal/evidence"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/retrieve"
	"github.com/Lokee86/grimoire/internal/structure"
)

const candidateSource = "arcana"

type graphSourceMatch struct {
	node       structure.Node
	span       *structure.Span
	score      float64
	reason     string
	intent     evidence.Intent
	role       evidence.Role
	groupIDs   []string
	links      []evidence.Link
	redundancy string
}

// SourceCandidates turns source-bearing Arcana graph evidence into ordinary
// Grimoire retrieval candidates. Structural evidence remains authoritative for
// graph relationships; these candidates only localize the source that explains
// those relationships so normal curation and package assembly can retain it.
func SourceCandidates(
	snapshot index.Snapshot,
	facts []structure.Evidence,
	limit int,
) []retrieve.Candidate {
	if limit <= 0 || len(facts) == 0 {
		return nil
	}
	matches := graphSourceMatches(facts)
	return candidatesForMatches(snapshot, matches, limit)
}

func graphSourceMatches(facts []structure.Evidence) []graphSourceMatch {
	matches := make([]graphSourceMatch, 0, len(facts)*4)
	for _, fact := range facts {
		groups, links := factGroupsAndLinks(fact)
		subject := evidenceSubject(fact)
		switch fact.Kind {
		case "operational_role":
			if fact.Node != nil {
				matches = append(matches, nodeSourceMatch(
					*fact.Node, 100, "Arcana graph subject "+nodeLabel(*fact.Node),
					evidence.IntentMechanism, evidence.RolePrimary, groups, links,
				))
			}
			for _, relationship := range fact.Relationships {
				score := 84.0
				if relationship.Direction == "incoming" {
					score += 2
				}
				if relationship.Certainty == "possible" {
					score -= 12
				}
				reason := fmt.Sprintf(
					"Arcana %s %s graph neighbor of %s",
					nonEmpty(relationship.Direction, "related"),
					nonEmpty(relationship.Relation, "relationship"), subject,
				)
				matchGroups := appendUniqueStrings(
					append([]string(nil), groups...),
					relationGroupID(fact.Node, relationship),
				)
				matches = append(matches, nodeSourceMatch(
					relationship.Node, score, reason, evidence.IntentMechanism,
					evidence.RoleSupporting, matchGroups, links,
				))
			}
		case "impact":
			if fact.Node != nil {
				matches = append(matches, nodeSourceMatch(
					*fact.Node, 98, "Arcana impact subject "+nodeLabel(*fact.Node),
					evidence.IntentArchitecture, evidence.RolePrimary, groups, links,
				))
			}
			for _, dependent := range fact.Dependents {
				depth := max(dependent.Depth, 1)
				score := max(48.0, 82.0-float64(depth-1)*7.0)
				reason := fmt.Sprintf(
					"Arcana impact dependent at depth %d from %s", depth, subject,
				)
				matchGroups := appendUniqueStrings(
					append([]string(nil), groups...),
					impactGroupID(fact.Node, dependent),
				)
				matches = append(matches, nodeSourceMatch(
					dependent.Node, score, reason, evidence.IntentArchitecture,
					evidence.RoleSupporting, matchGroups, links,
				))
			}
		case "call_chain":
			if fact.Chain == nil {
				continue
			}
			for index, node := range fact.Chain.Nodes {
				endpoint := index == 0 || index == len(fact.Chain.Nodes)-1
				score := 90.0
				role := evidence.RoleSupporting
				if endpoint {
					score = 96
					role = evidence.RolePrimary
				}
				reason := fmt.Sprintf(
					"Arcana call-chain node %d of %d", index+1, len(fact.Chain.Nodes),
				)
				matches = append(matches, nodeSourceMatch(
					node, score, reason, evidence.IntentCallChain, role, groups, links,
				))
			}
		case "unresolved":
			if fact.Node != nil {
				matches = append(matches, nodeSourceMatch(
					*fact.Node, 92, "Arcana unresolved-reference owner "+nodeLabel(*fact.Node),
					evidence.IntentMechanism, evidence.RolePrimary, groups, links,
				))
			}
			for _, unresolved := range fact.Unresolved {
				if unresolved.Span == nil {
					continue
				}
				reason := "Arcana unresolved reference"
				if unresolved.Expression != "" {
					reason += " " + unresolved.Expression
				}
				matchGroups := appendUniqueStrings(
					append([]string(nil), groups...),
					evidence.StableID("arcana-unresolved", sourceIdentityFromSpan(unresolved.Span), unresolved.Expression),
				)
				matches = append(matches, graphSourceMatch{
					span: unresolved.Span, score: 70, reason: reason,
					intent: evidence.IntentMechanism, role: evidence.RoleSupporting,
					groupIDs: matchGroups, links: links,
					redundancy: normalizedSourcePath(unresolved.Span.Path) + "::unresolved::" + unresolved.Expression,
				})
			}
		}
	}
	return matches
}

func candidatesForMatches(
	snapshot index.Snapshot,
	matches []graphSourceMatch,
	limit int,
) []retrieve.Candidate {
	chunksByPath := make(map[string][]index.Chunk, len(snapshot.Files))
	for _, file := range snapshot.Files {
		chunksByPath[normalizedSourcePath(file.Path)] = file.Chunks
	}
	byChunk := make(map[string]retrieve.Candidate)
	for _, match := range matches {
		path := sourceMatchPath(match)
		chunks := chunksByPath[path]
		if len(chunks) == 0 {
			continue
		}
		for _, chunk := range sourceMatchChunks(chunks, match.span) {
			candidate := candidateForMatch(match, chunk)
			key := candidateChunkKey(chunk)
			if existing, found := byChunk[key]; found {
				byChunk[key] = mergeGraphSourceCandidate(existing, candidate)
				continue
			}
			byChunk[key] = candidate
		}
	}

	candidates := make([]retrieve.Candidate, 0, len(byChunk))
	for _, candidate := range byChunk {
		candidates = append(candidates, candidate)
	}
	sort.Slice(candidates, func(left, right int) bool {
		if candidates[left].Score != candidates[right].Score {
			return candidates[left].Score > candidates[right].Score
		}
		if candidates[left].Chunk.Path != candidates[right].Chunk.Path {
			return candidates[left].Chunk.Path < candidates[right].Chunk.Path
		}
		return candidates[left].Chunk.StartLine < candidates[right].Chunk.StartLine
	})
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}
	for index := range candidates {
		candidates[index].Rank = index + 1
	}
	return candidates
}

func candidateForMatch(match graphSourceMatch, chunk index.Chunk) retrieve.Candidate {
	identity := sourceIdentityFromSpan(match.span)
	if identity == "" {
		identity = evidence.RangeIdentity(chunk.Path, chunk.StartLine, chunk.EndLine)
	}
	groups := append([]string(nil), match.groupIDs...)
	if match.node.Name != "" || match.node.Identity != "" || match.node.Path != "" {
		groups = appendUniqueStrings(groups, nodeGroupID(match.node))
	}
	links := make([]evidence.Link, 0, len(match.links))
	for _, link := range match.links {
		if link.Identity == "" || link.Identity == identity {
			continue
		}
		links = appendUniqueLinks(links, link)
	}
	redundancy := match.redundancy
	if redundancy == "" {
		redundancy = sourceRedundancyKey(match.node, chunk)
	}
	context := evidence.Descriptor{
		Identity: identity, Intents: []evidence.Intent{match.intent},
		Roles: []evidence.Role{match.role}, GroupIDs: groups,
		EstimatedTokens: max(chunk.TokenCount, 1), RedundancyKey: redundancy,
		Links: links,
	}
	return retrieve.Candidate{
		Chunk: chunk, Score: match.score, Source: candidateSource,
		Reasons:      []string{match.reason},
		ScoreDetails: []retrieve.ScoreDetail{{Name: match.reason, Value: match.score}},
		Context:      &context,
	}
}

func mergeGraphSourceCandidate(left, right retrieve.Candidate) retrieve.Candidate {
	preferred, other := left, right
	if right.Score > left.Score {
		preferred, other = right, left
	}
	preferred.Reasons = appendUniqueReasonStrings(preferred.Reasons, other.Reasons...)
	preferred.ScoreDetails = appendUniqueScoreDetails(preferred.ScoreDetails, other.ScoreDetails...)
	preferred.Context = mergeSourceContext(preferred.Context, other.Context)
	return preferred
}

func nodeSourceMatch(
	node structure.Node,
	score float64,
	reason string,
	intent evidence.Intent,
	role evidence.Role,
	groups []string,
	links []evidence.Link,
) graphSourceMatch {
	return graphSourceMatch{
		node: node, span: node.Span, score: score, reason: reason,
		intent: intent, role: role,
		groupIDs:   append([]string(nil), groups...),
		links:      append([]evidence.Link(nil), links...),
		redundancy: sourceRedundancyKey(node, index.Chunk{}),
	}
}

func factGroupsAndLinks(fact structure.Evidence) ([]string, []evidence.Link) {
	if fact.Context == nil {
		return nil, nil
	}
	return append([]string(nil), fact.Context.GroupIDs...), append([]evidence.Link(nil), fact.Context.Links...)
}

func evidenceSubject(fact structure.Evidence) string {
	if fact.Node != nil {
		return nodeLabel(*fact.Node)
	}
	if fact.Chain != nil && len(fact.Chain.Nodes) > 0 {
		return nodeLabel(fact.Chain.Nodes[0])
	}
	return "matched symbol"
}

func nodeLabel(node structure.Node) string {
	for _, value := range []string{node.QualifiedName, node.Name, node.Path, node.Identity} {
		if value != "" {
			return value
		}
	}
	return "graph node"
}

func sourceMatchPath(match graphSourceMatch) string {
	if match.span != nil && match.span.Path != "" {
		return normalizedSourcePath(match.span.Path)
	}
	return normalizedSourcePath(match.node.Path)
}

func sourceMatchChunks(chunks []index.Chunk, span *structure.Span) []index.Chunk {
	if span == nil || span.StartLine <= 0 {
		return chunks[:1]
	}
	end := max(span.EndLine, span.StartLine)
	matches := make([]index.Chunk, 0, 2)
	for _, chunk := range chunks {
		if chunk.EndLine < span.StartLine || chunk.StartLine > end {
			continue
		}
		matches = append(matches, chunk)
	}
	if len(matches) == 0 {
		return []index.Chunk{nearestSourceChunk(chunks, span.StartLine, end)}
	}
	return matches
}

func nearestSourceChunk(chunks []index.Chunk, start, end int) index.Chunk {
	best := chunks[0]
	bestDistance := sourceChunkDistance(best, start, end)
	for _, chunk := range chunks[1:] {
		distance := sourceChunkDistance(chunk, start, end)
		if distance < bestDistance || distance == bestDistance && chunk.StartLine < best.StartLine {
			best = chunk
			bestDistance = distance
		}
	}
	return best
}

func sourceChunkDistance(chunk index.Chunk, start, end int) int {
	switch {
	case start > chunk.EndLine:
		return start - chunk.EndLine
	case chunk.StartLine > end:
		return chunk.StartLine - end
	default:
		return 0
	}
}

func sourceIdentityFromSpan(span *structure.Span) string {
	if span == nil || span.Path == "" || span.StartLine <= 0 {
		return ""
	}
	return evidence.RangeIdentity(span.Path, span.StartLine, max(span.EndLine, span.StartLine))
}

func sourceRedundancyKey(node structure.Node, chunk index.Chunk) string {
	path := normalizedSourcePath(node.Path)
	if node.Span != nil && node.Span.Path != "" {
		path = normalizedSourcePath(node.Span.Path)
	}
	if path == "" {
		path = normalizedSourcePath(chunk.Path)
	}
	name := node.QualifiedName
	if name == "" {
		name = node.Name
	}
	if name == "" {
		name = node.Identity
	}
	if name == "" {
		return path
	}
	return path + "::" + name
}

func relationGroupID(subject *structure.Node, relationship structure.Relationship) string {
	parts := []string{relationship.Direction, relationship.Relation, relationship.Certainty, nodeLabel(relationship.Node)}
	if subject != nil {
		parts = append([]string{nodeLabel(*subject)}, parts...)
	}
	return evidence.StableID("arcana-relation", parts...)
}

func impactGroupID(subject *structure.Node, dependent structure.DepthNode) string {
	parts := []string{fmt.Sprintf("depth:%d", dependent.Depth), nodeLabel(dependent.Node)}
	if subject != nil {
		parts = append([]string{nodeLabel(*subject)}, parts...)
	}
	return evidence.StableID("arcana-impact", parts...)
}

func candidateChunkKey(chunk index.Chunk) string {
	if chunk.ID != "" {
		return "id:" + chunk.ID
	}
	return fmt.Sprintf("range:%s:%d:%d", chunk.Path, chunk.StartLine, chunk.EndLine)
}

func mergeSourceContext(left, right *evidence.Descriptor) *evidence.Descriptor {
	if left == nil && right == nil {
		return nil
	}
	var leftValue, rightValue evidence.Descriptor
	if left != nil {
		leftValue = *left
	}
	if right != nil {
		rightValue = *right
	}
	merged := evidence.Merge(leftValue, rightValue)
	return &merged
}

func appendUniqueReasonStrings(existing []string, values ...string) []string {
	for _, value := range values {
		if value == "" || containsString(existing, value) {
			continue
		}
		existing = append(existing, value)
	}
	return existing
}

func appendUniqueScoreDetails(existing []retrieve.ScoreDetail, values ...retrieve.ScoreDetail) []retrieve.ScoreDetail {
	for _, value := range values {
		found := false
		for _, prior := range existing {
			if prior == value {
				found = true
				break
			}
		}
		if !found {
			existing = append(existing, value)
		}
	}
	return existing
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func normalizedSourcePath(path string) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}
	return filepath.ToSlash(filepath.Clean(path))
}

func nonEmpty(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
