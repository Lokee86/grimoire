package app

import (
	"fmt"
	"sort"

	"github.com/Lokee86/grimoire/internal/evidence"
	"github.com/Lokee86/grimoire/internal/queryshape"
	"github.com/Lokee86/grimoire/internal/retrieve"
)

const (
	intentFrontSeed       = 2
	intentReservedPerPass = 2
)

type fusedIntentCandidate struct {
	candidate     retrieve.Candidate
	key           string
	fusedScore    float64
	bestRank      int
	bestGroup     int
	reasons       []string
	context       *evidence.Descriptor
	fusionDetails []retrieve.ScoreDetail
}

func mergeIntentCandidateGroups(limit int, groups []intentCandidateGroup) []retrieve.Candidate {
	if limit <= 0 || len(groups) == 0 {
		return nil
	}
	if len(groups) == 1 {
		return singleIntentCandidates(limit, groups[0].Candidates)
	}

	byKey := fuseIntentCandidateGroups(groups)
	fused := make([]*fusedIntentCandidate, 0, len(byKey))
	for _, candidate := range byKey {
		fused = append(fused, candidate)
	}
	sort.Slice(fused, func(left, right int) bool {
		if fused[left].fusedScore != fused[right].fusedScore {
			return fused[left].fusedScore > fused[right].fusedScore
		}
		if fused[left].bestRank != fused[right].bestRank {
			return fused[left].bestRank < fused[right].bestRank
		}
		return fused[left].key < fused[right].key
	})
	return reserveIntentCoverage(limit, groups, fused, byKey)
}

func singleIntentCandidates(limit int, candidates []retrieve.Candidate) []retrieve.Candidate {
	result := append([]retrieve.Candidate(nil), candidates...)
	if len(result) > limit {
		result = result[:limit]
	}
	for index := range result {
		result[index].Rank = index + 1
	}
	return result
}

func fuseIntentCandidateGroups(groups []intentCandidateGroup) map[string]*fusedIntentCandidate {
	byKey := make(map[string]*fusedIntentCandidate)
	for groupIndex, group := range groups {
		weight := group.Intent.Weight
		if weight <= 0 {
			weight = 1
		}
		for position, candidate := range group.Candidates {
			rank := candidate.Rank
			if rank <= 0 {
				rank = position + 1
			}
			key := contextCandidateKey(candidate)
			contribution := weight / (rrfRankConstant + float64(rank))
			reason := fmt.Sprintf("intent fusion from %s rank %d", group.Intent.Intent, rank)
			current := byKey[key]
			if current == nil {
				current = &fusedIntentCandidate{
					candidate: candidate, key: key, bestRank: rank, bestGroup: groupIndex,
					reasons: append([]string(nil), candidate.Reasons...),
					context: mergeCandidateContext(nil, candidate.Context),
				}
				current.candidate.ScoreDetails = append([]retrieve.ScoreDetail(nil), candidate.ScoreDetails...)
				byKey[key] = current
			}
			current.fusedScore += contribution
			current.reasons = appendUniqueReason(current.reasons, reason)
			for _, candidateReason := range candidate.Reasons {
				current.reasons = appendUniqueReason(current.reasons, candidateReason)
			}
			current.context = mergeCandidateContext(current.context, candidate.Context)
			current.fusionDetails = append(current.fusionDetails, retrieve.ScoreDetail{Name: reason, Value: contribution})
			if rank < current.bestRank || rank == current.bestRank && groupIndex < current.bestGroup {
				current.candidate = candidate
				current.candidate.ScoreDetails = append([]retrieve.ScoreDetail(nil), candidate.ScoreDetails...)
				current.bestRank = rank
				current.bestGroup = groupIndex
			}
		}
	}
	return byKey
}

func reserveIntentCoverage(
	limit int,
	groups []intentCandidateGroup,
	fused []*fusedIntentCandidate,
	byKey map[string]*fusedIntentCandidate,
) []retrieve.Candidate {
	ordered := make([]retrieve.Candidate, 0, min(limit, len(fused)))
	seen := make(map[string]struct{}, limit)
	appendCandidate := func(item *fusedIntentCandidate) {
		if item == nil || len(ordered) >= limit {
			return
		}
		if _, exists := seen[item.key]; exists {
			return
		}
		candidate := item.candidate
		candidate.Score = item.fusedScore
		candidate.Reasons = item.reasons
		candidate.Context = item.context
		candidate.ScoreDetails = append(candidate.ScoreDetails, item.fusionDetails...)
		seen[item.key] = struct{}{}
		ordered = append(ordered, candidate)
	}
	cursors := make([]int, len(groups))
	appendNextGroupCandidate := func(groupIndex int) {
		group := groups[groupIndex]
		if group.Intent.Intent == evidence.IntentMixed {
			return
		}
		for cursors[groupIndex] < len(group.Candidates) {
			candidate := group.Candidates[cursors[groupIndex]]
			cursors[groupIndex]++
			before := len(ordered)
			appendCandidate(byKey[contextCandidateKey(candidate)])
			if len(ordered) > before {
				return
			}
		}
	}

	for groupIndex := range groups {
		appendNextGroupCandidate(groupIndex)
	}
	for index := 0; index < min(intentFrontSeed, len(fused)); index++ {
		appendCandidate(fused[index])
	}
	for round := 1; round < intentReservedPerPass; round++ {
		for groupIndex := range groups {
			appendNextGroupCandidate(groupIndex)
		}
	}
	for _, item := range fused {
		appendCandidate(item)
	}
	for index := range ordered {
		ordered[index].Rank = index + 1
	}
	return ordered
}

func structuralRetrievalIntent(query string, intents []queryshape.RetrievalIntent) queryshape.RetrievalIntent {
	for _, target := range []evidence.Intent{
		evidence.IntentCallChain, evidence.IntentArchitecture,
		evidence.IntentMechanism, evidence.IntentDirectLocation,
	} {
		for _, planned := range intents {
			if planned.Intent == target {
				return planned
			}
		}
	}
	if len(intents) > 0 {
		return intents[0]
	}
	return queryshape.RetrievalIntent{Intent: evidence.IntentMixed, Query: query, Weight: 1}
}

func annotateStructuralIntent(result structuralContextResult, planned queryshape.RetrievalIntent) structuralContextResult {
	for index := range result.Lexicon.Candidates {
		result.Lexicon.Candidates[index] = annotateCandidateIntent(result.Lexicon.Candidates[index], planned)
	}
	for index := range result.Lexicon.Evidence {
		result.Lexicon.Evidence[index].Context = mergeStructuralIntent(result.Lexicon.Evidence[index].Context, planned)
	}
	for index := range result.Arcana {
		result.Arcana[index].Context = mergeStructuralIntent(result.Arcana[index].Context, planned)
	}
	for index := range result.Combined {
		result.Combined[index].Context = mergeStructuralIntent(result.Combined[index].Context, planned)
	}
	return result
}

func mergeStructuralIntent(existing *evidence.Descriptor, planned queryshape.RetrievalIntent) *evidence.Descriptor {
	descriptor := evidence.Descriptor{
		Intents: []evidence.Intent{planned.Intent},
		Roles:   []evidence.Role{evidence.RoleStructural},
	}
	if planned.FacetID != "" {
		descriptor.GroupIDs = []string{planned.FacetID}
	}
	return mergeCandidateContext(existing, &descriptor)
}
