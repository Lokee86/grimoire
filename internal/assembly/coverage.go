package assembly

import (
	"strings"

	"github.com/Lokee86/grimoire/internal/evidence"
	"github.com/Lokee86/grimoire/internal/retrieve"
)

const facetPrefix = "facet:"

type indexedCandidate struct {
	candidate retrieve.Candidate
	position  int
}

// prioritizeFacetCoverage reserves distinct strong candidates for each query
// facet before appending the remaining candidates in their original order. A
// candidate may be relevant to several facets, but it claims only its strongest
// still-open facet so one generic chunk cannot satisfy the whole query alone.
func prioritizeFacetCoverage(candidates []retrieve.Candidate, depth int) ([]retrieve.Candidate, int, map[string]string) {
	if depth <= 0 || len(candidates) < 2 {
		return append([]retrieve.Candidate(nil), candidates...), countAvailableFacets(candidates), nil
	}
	available := countAvailableFacets(candidates)
	if available == 0 {
		return append([]retrieve.Candidate(nil), candidates...), 0, nil
	}

	remaining := make([]indexedCandidate, 0, len(candidates))
	for position, candidate := range candidates {
		remaining = append(remaining, indexedCandidate{candidate: candidate, position: position})
	}
	coverage := make(map[string]int, available)
	claims := make(map[string]string, available*depth)
	ordered := make([]retrieve.Candidate, 0, len(candidates))
	for len(remaining) > 0 {
		best := -1
		for index := range remaining {
			if facetCoverageUnits(remaining[index].candidate, coverage, depth) == 0 {
				continue
			}
			if best < 0 || betterCoverageCandidate(remaining[index], remaining[best], coverage, depth) {
				best = index
			}
		}
		if best < 0 {
			break
		}
		selected := remaining[best]
		remaining = append(remaining[:best], remaining[best+1:]...)
		ordered = append(ordered, selected.candidate)
		facet := bestUncoveredFacet(selected.candidate, coverage, depth)
		if facet != "" {
			coverage[facet]++
			claims[coverageCandidateKey(selected.candidate)] = facet
		}
	}
	for _, item := range remaining {
		ordered = append(ordered, item.candidate)
	}
	return ordered, available, claims
}

func betterCoverageCandidate(left, right indexedCandidate, coverage map[string]int, depth int) bool {
	leftUnits := facetCoverageUnits(left.candidate, coverage, depth)
	rightUnits := facetCoverageUnits(right.candidate, coverage, depth)
	if leftUnits != rightUnits {
		return leftUnits > rightUnits
	}
	leftRank := uncoveredFacetRank(left.candidate, coverage, depth)
	rightRank := uncoveredFacetRank(right.candidate, coverage, depth)
	if leftRank != rightRank {
		return leftRank < rightRank
	}
	leftExact := candidateExactPriority(left.candidate)
	rightExact := candidateExactPriority(right.candidate)
	if leftExact != rightExact {
		return leftExact > rightExact
	}
	leftRole := candidateEvidenceRole(left.candidate)
	rightRole := candidateEvidenceRole(right.candidate)
	if leftRole != rightRole {
		return leftRole > rightRole
	}
	leftTokens := candidateTokenCount(left.candidate)
	rightTokens := candidateTokenCount(right.candidate)
	if leftTokens != rightTokens {
		return leftTokens < rightTokens
	}
	return left.position < right.position
}

func uncoveredFacetRank(candidate retrieve.Candidate, coverage map[string]int, depth int) int {
	facet := bestUncoveredFacet(candidate, coverage, depth)
	if facet == "" || candidate.Context == nil {
		return int(^uint(0) >> 1)
	}
	rank := candidate.Context.FacetRanks[facet]
	if rank <= 0 {
		return int(^uint(0) >> 1)
	}
	return rank
}

func bestUncoveredFacet(candidate retrieve.Candidate, coverage map[string]int, depth int) string {
	if candidate.Context == nil {
		return ""
	}
	bestFacet := ""
	bestRank := int(^uint(0) >> 1)
	for _, facet := range candidateFacets(candidate) {
		if coverage[facet] >= depth {
			continue
		}
		rank := candidate.Context.FacetRanks[facet]
		if rank <= 0 {
			rank = int(^uint(0) >> 1)
		}
		if rank < bestRank || (rank == bestRank && (bestFacet == "" || facet < bestFacet)) {
			bestFacet = facet
			bestRank = rank
		}
	}
	return bestFacet
}

func facetCoverageUnits(candidate retrieve.Candidate, coverage map[string]int, depth int) int {
	if bestUncoveredFacet(candidate, coverage, depth) == "" {
		return 0
	}
	return 1
}

func coverageCandidateKey(candidate retrieve.Candidate) string {
	if candidate.Context != nil && candidate.Context.Identity != "" {
		return candidate.Context.Identity
	}
	return evidence.RangeIdentity(candidate.Chunk.Path, candidate.Chunk.StartLine, candidate.Chunk.EndLine)
}

func candidateEvidenceRole(candidate retrieve.Candidate) int {
	if candidate.Context == nil {
		return 0
	}
	priority := 0
	for _, role := range candidate.Context.Roles {
		switch role {
		case evidence.RolePrimary:
			return 3
		case evidence.RoleSupporting:
			if priority < 2 {
				priority = 2
			}
		case evidence.RoleStructural:
			if priority < 1 {
				priority = 1
			}
		}
	}
	return priority
}

func candidateExactPriority(candidate retrieve.Candidate) int {
	if candidate.Source == "exact" {
		return 2
	}
	if candidate.Context != nil && candidate.Context.ExactMatchStrength > 0 {
		return 1
	}
	return 0
}

func candidateFacets(candidate retrieve.Candidate) []string {
	if candidate.Context == nil {
		return nil
	}
	return facetIDs(candidate.Context.GroupIDs)
}

func facetIDs(groups []string) []string {
	result := make([]string, 0, len(groups))
	for _, group := range groups {
		if isFacetID(group) {
			result = append(result, group)
		}
	}
	return result
}

func isFacetID(group string) bool {
	return strings.HasPrefix(group, facetPrefix)
}

func countAvailableFacets(candidates []retrieve.Candidate) int {
	seen := make(map[string]struct{})
	for _, candidate := range candidates {
		for _, facet := range candidateFacets(candidate) {
			seen[facet] = struct{}{}
		}
	}
	return len(seen)
}
