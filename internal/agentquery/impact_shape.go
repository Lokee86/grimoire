package agentquery

import (
	"fmt"
	"sort"
	"strings"
)

const maxImpactCandidates = 64

func impactCandidateLimit(limit int) int {
	if limit <= 0 {
		return 1
	}
	return min(maxImpactCandidates, max(limit, limit*4))
}

func rankImpactDependents(request Request, candidates []Dependent, limit int) []Dependent {
	if limit <= 0 || len(candidates) == 0 {
		return nil
	}
	terms := traceTerms(strings.Join([]string{request.Query, request.Anchor}, " "))
	includeTests := traceRequestIncludesTests(request)
	merged := make(map[string]Dependent, len(candidates))
	order := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		candidate.Score = impactDependentScore(candidate, terms, includeTests)
		candidate.Reasons = impactDependentReasons(candidate, terms, includeTests)
		key := impactDependentKey(candidate)
		if existing, exists := merged[key]; exists {
			if candidate.Score > existing.Score ||
				(candidate.Score == existing.Score && impactDependentTieKey(candidate) < impactDependentTieKey(existing)) {
				candidate.Evidence = unique(append(candidate.Evidence, existing.Evidence...))
				candidate.Spans = mergeImpactRanges(candidate.Spans, existing.Spans)
				candidate.Reasons = unique(append(candidate.Reasons, existing.Reasons...))
				merged[key] = candidate
			} else {
				existing.Evidence = unique(append(existing.Evidence, candidate.Evidence...))
				existing.Spans = mergeImpactRanges(existing.Spans, candidate.Spans)
				existing.Reasons = unique(append(existing.Reasons, candidate.Reasons...))
				merged[key] = existing
			}
			continue
		}
		merged[key] = candidate
		order = append(order, key)
	}

	ranked := make([]Dependent, 0, len(merged))
	for _, key := range order {
		ranked = append(ranked, merged[key])
	}
	sort.SliceStable(ranked, func(left, right int) bool {
		if ranked[left].Score != ranked[right].Score {
			return ranked[left].Score > ranked[right].Score
		}
		if ranked[left].Depth != ranked[right].Depth {
			return ranked[left].Depth < ranked[right].Depth
		}
		return impactDependentTieKey(ranked[left]) < impactDependentTieKey(ranked[right])
	})
	if len(ranked) > limit {
		ranked = ranked[:limit]
	}
	for index := range ranked {
		ranked[index].Rank = index + 1
	}
	return ranked
}

func impactDistinctCandidateCount(candidates []Dependent) int {
	seen := make(map[string]bool, len(candidates))
	for _, candidate := range candidates {
		seen[impactDependentKey(candidate)] = true
	}
	return len(seen)
}

func impactDependentScore(candidate Dependent, terms []string, includeTests bool) float64 {
	depth := candidate.Depth
	if depth <= 0 {
		depth = 1
	}
	value := strings.ToLower(strings.Join([]string{
		candidate.Node.Name, candidate.Node.QualifiedName, candidate.Node.Path,
		candidate.Node.Kind, candidate.Relation,
	}, " "))
	score := 180 - float64(depth-1)*28
	score += traceRelationScore(candidate.Relation) * 0.7
	score += traceTextScore(value, terms)
	score += tracePathLocationScore(candidate.Node.Path)
	if traceTestPath(candidate.Node.Path) {
		if includeTests {
			score += 12
		} else {
			score -= 120
		}
	} else {
		score += 12
	}
	if strings.EqualFold(candidate.Certainty, "possible") || strings.HasPrefix(strings.ToLower(candidate.Relation), "possible-") {
		score -= 24
	} else {
		score += 12
	}
	return score
}

func impactDependentReasons(candidate Dependent, terms []string, includeTests bool) []string {
	reasons := []string{fmt.Sprintf("depth %d structural impact", max(1, candidate.Depth))}
	value := strings.ToLower(strings.Join([]string{
		candidate.Node.Name, candidate.Node.QualifiedName, candidate.Node.Path,
		candidate.Node.Kind, candidate.Relation,
	}, " "))
	if traceTextScore(value, terms) > 0 {
		reasons = append(reasons, "query terms match impacted behavior")
	}
	if traceTestPath(candidate.Node.Path) {
		if includeTests {
			reasons = append(reasons, "test evidence requested")
		} else {
			reasons = append(reasons, "test-only evidence deprioritized")
		}
	} else {
		reasons = append(reasons, "production source preferred")
	}
	if strings.EqualFold(candidate.Certainty, "possible") || strings.HasPrefix(strings.ToLower(candidate.Relation), "possible-") {
		reasons = append(reasons, "possible relationship")
	} else {
		reasons = append(reasons, "definite relationship")
	}
	return reasons
}

func impactDependentKey(candidate Dependent) string {
	label := candidate.Node.QualifiedName
	if label == "" {
		label = candidate.Node.Name
	}
	return strings.ToLower(strings.Join([]string{
		normalizePath(candidate.Node.Path), label, candidate.Node.Kind,
		candidate.Direction, strings.TrimPrefix(candidate.Relation, "possible-"),
	}, "\x00"))
}

func impactDependentTieKey(candidate Dependent) string {
	label := candidate.Node.QualifiedName
	if label == "" {
		label = candidate.Node.Name
	}
	return strings.ToLower(strings.Join([]string{
		normalizePath(candidate.Node.Path), label, candidate.Relation,
		candidate.Direction, candidate.Node.Handle.Provider,
	}, "\x00"))
}

func mergeImpactRanges(left, right []Range) []Range {
	result := append([]Range(nil), left...)
	seen := make(map[string]bool, len(result)+len(right))
	for _, value := range result {
		seen[impactRangeKey(value)] = true
	}
	for _, value := range right {
		key := impactRangeKey(value)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, value)
	}
	return result
}

func impactRangeKey(value Range) string {
	return fmt.Sprintf("%s:%d:%d:%d:%d", normalizePath(value.Path), value.StartLine, value.StartColumn, value.EndLine, value.EndColumn)
}
