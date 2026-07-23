package assembly

import (
	"sort"

	"github.com/Lokee86/grimoire/internal/queryshape"
	"github.com/Lokee86/grimoire/internal/retrieve"
	"github.com/Lokee86/grimoire/internal/structure"
)

type scopeConfig struct {
	minimumCandidates   int
	maximumCandidates   int
	minimumRegions      int
	tokenPoolMultiplier int
	structuralLimit     int
}

// Plan selects a sufficient deterministic evidence set for an active automatic
// retrieval policy. Candidate order remains the curated order supplied by the caller.
func Plan(
	policy queryshape.RetrievalPolicy,
	candidates []retrieve.Candidate,
	evidence []structure.Evidence,
) Result {
	config := configFor(policy.Scope)
	anchorRegion := focusedAnchorRegion(candidates)
	ordered := prioritizeCandidates(policy.Scope, candidates, anchorRegion)
	selected := make([]retrieve.Candidate, 0, min(config.maximumCandidates, len(candidates)))
	regions := newOrderedSet()
	roles := newOrderedSet()
	groups := newOrderedSet()
	hasExact := false
	candidateTokens := 0
	considered := 0
	stopReason := "candidate set exhausted"

	for index, candidate := range ordered {
		considered++
		if policy.Scope == queryshape.ScopeFocused && !focusedCandidate(candidate, anchorRegion) {
			continue
		}
		selected = append(selected, candidate)
		candidateTokens += candidateTokenCount(candidate)
		regions.Add(queryshape.PathRegion(candidate.Chunk.Path))
		roles.Add(candidateRole(candidate.Chunk.Path))
		for _, groupID := range candidateGroups(candidate) {
			groups.Add(groupID)
		}
		if candidate.Source == "exact" {
			hasExact = true
		}
		if len(selected) >= config.maximumCandidates {
			stopReason = string(policy.Scope) + " candidate cap reached"
			break
		}
		if coverageSatisfied(policy, config, len(selected), candidateTokens, regions.Len(), hasExact, candidates) &&
			!hasEligibleGroupCompanion(ordered[index+1:], policy.Scope, anchorRegion, groups) {
			stopReason = string(policy.Scope) + " evidence coverage satisfied"
			break
		}
	}

	structuralLimit := min(config.structuralLimit, len(evidence))
	structural := append([]structure.Evidence(nil), evidence[:structuralLimit]...)
	return Result{
		Candidates: selected,
		Structural: structural,
		Decision: Decision{
			Scope:                policy.Scope,
			CandidatesConsidered: considered,
			CandidatesSelected:   len(selected),
			CandidateTokens:      candidateTokens,
			StructuralConsidered: len(evidence),
			StructuralSelected:   len(structural),
			RegionsRepresented:   regions.Values(),
			RolesRepresented:     roles.Values(),
			GroupsRepresented:    sortedValues(groups.Values()),
			StopReason:           stopReason,
		},
	}
}

func prioritizeCandidates(scope queryshape.Scope, candidates []retrieve.Candidate, anchorRegion string) []retrieve.Candidate {
	remaining := append([]retrieve.Candidate(nil), candidates...)
	ordered := make([]retrieve.Candidate, 0, len(candidates))
	activeGroups := make(map[string]struct{})
	for len(remaining) > 0 {
		index := 0
		if len(activeGroups) > 0 {
			for candidateIndex, candidate := range remaining {
				if scope == queryshape.ScopeFocused && !focusedCandidate(candidate, anchorRegion) {
					continue
				}
				if sharesGroup(candidate, activeGroups) {
					index = candidateIndex
					break
				}
			}
		}
		candidate := remaining[index]
		ordered = append(ordered, candidate)
		remaining = append(remaining[:index], remaining[index+1:]...)
		if scope == queryshape.ScopeFocused && !focusedCandidate(candidate, anchorRegion) {
			continue
		}
		for _, groupID := range candidateGroups(candidate) {
			activeGroups[groupID] = struct{}{}
		}
	}
	return ordered
}

func hasEligibleGroupCompanion(
	candidates []retrieve.Candidate,
	scope queryshape.Scope,
	anchorRegion string,
	groups *orderedSet,
) bool {
	activeGroups := make(map[string]struct{}, groups.Len())
	for _, groupID := range groups.Values() {
		activeGroups[groupID] = struct{}{}
	}
	if len(activeGroups) == 0 {
		return false
	}
	for _, candidate := range candidates {
		if scope == queryshape.ScopeFocused && !focusedCandidate(candidate, anchorRegion) {
			continue
		}
		if sharesGroup(candidate, activeGroups) {
			return true
		}
	}
	return false
}

func sharesGroup(candidate retrieve.Candidate, groups map[string]struct{}) bool {
	for _, groupID := range candidateGroups(candidate) {
		if _, ok := groups[groupID]; ok {
			return true
		}
	}
	return false
}

func candidateGroups(candidate retrieve.Candidate) []string {
	if candidate.Context == nil {
		return nil
	}
	groups := make([]string, 0, len(candidate.Context.GroupIDs))
	for _, groupID := range candidate.Context.GroupIDs {
		if groupID != "" {
			groups = append(groups, groupID)
		}
	}
	return groups
}

func candidateTokenCount(candidate retrieve.Candidate) int {
	tokens := candidate.Chunk.TokenCount
	if candidate.Context != nil && candidate.Context.EstimatedTokens > 0 {
		tokens = candidate.Context.EstimatedTokens
	}
	return max(tokens, 1)
}

func sortedValues(values []string) []string {
	sorted := append([]string(nil), values...)
	sort.Strings(sorted)
	return sorted
}

func configFor(scope queryshape.Scope) scopeConfig {
	switch scope {
	case queryshape.ScopeFocused:
		return scopeConfig{minimumCandidates: 3, maximumCandidates: 32, minimumRegions: 1, tokenPoolMultiplier: 4, structuralLimit: 24}
	case queryshape.ScopeExploratory:
		return scopeConfig{minimumCandidates: 24, maximumCandidates: 128, minimumRegions: 3, tokenPoolMultiplier: 4, structuralLimit: 128}
	default:
		return scopeConfig{minimumCandidates: 12, maximumCandidates: 160, minimumRegions: 2, tokenPoolMultiplier: 12, structuralLimit: 64}
	}
}

func coverageSatisfied(
	policy queryshape.RetrievalPolicy,
	config scopeConfig,
	selected, candidateTokens, regions int,
	hasExact bool,
	all []retrieve.Candidate,
) bool {
	if selected < config.minimumCandidates || regions < config.minimumRegions ||
		candidateTokens < policy.TargetTokens*config.tokenPoolMultiplier {
		return false
	}
	if policy.Scope != queryshape.ScopeFocused {
		return true
	}
	return hasExact || !containsExact(all)
}

func containsExact(candidates []retrieve.Candidate) bool {
	for _, candidate := range candidates {
		if candidate.Source == "exact" {
			return true
		}
	}
	return false
}

func focusedAnchorRegion(candidates []retrieve.Candidate) string {
	for _, candidate := range candidates {
		if candidate.Source == "exact" {
			return queryshape.PathRegion(candidate.Chunk.Path)
		}
	}
	if len(candidates) == 0 {
		return ""
	}
	return queryshape.PathRegion(candidates[0].Chunk.Path)
}

func focusedCandidate(candidate retrieve.Candidate, anchorRegion string) bool {
	if candidate.Source == "exact" || candidate.Source == "adjacent" {
		return true
	}
	return anchorRegion == "" || queryshape.PathRegion(candidate.Chunk.Path) == anchorRegion
}
