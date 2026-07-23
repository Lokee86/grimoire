package assembly

import (
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
	selected := make([]retrieve.Candidate, 0, min(config.maximumCandidates, len(candidates)))
	regions := newOrderedSet()
	roles := newOrderedSet()
	hasExact := false
	candidateTokens := 0
	considered := 0
	stopReason := "candidate set exhausted"

	for _, candidate := range candidates {
		considered++
		if policy.Scope == queryshape.ScopeFocused && !focusedCandidate(candidate, anchorRegion) {
			continue
		}
		selected = append(selected, candidate)
		candidateTokens += max(candidate.Chunk.TokenCount, 1)
		regions.Add(queryshape.PathRegion(candidate.Chunk.Path))
		roles.Add(candidateRole(candidate.Chunk.Path))
		if candidate.Source == "exact" {
			hasExact = true
		}
		if len(selected) >= config.maximumCandidates {
			stopReason = string(policy.Scope) + " candidate cap reached"
			break
		}
		if coverageSatisfied(policy, config, len(selected), candidateTokens, regions.Len(), hasExact, candidates) {
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
			StopReason:           stopReason,
		},
	}
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
