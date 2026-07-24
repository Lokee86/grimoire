package app

import (
	"fmt"
	"sort"

	"github.com/Lokee86/grimoire/internal/evidence"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/retrieve"
	"github.com/Lokee86/grimoire/internal/selection"
)

const (
	maxExactCandidates   = 32
	maxLexiconCandidates = 24
	baseFrontCandidates  = 32
	rrfRankConstant      = 60.0
)

func curateContextCandidates(
	snapshot index.Snapshot,
	query string,
	base []retrieve.Candidate,
	limit int,
) []retrieve.Candidate {
	exactLimit := min(limit, maxExactCandidates)
	exact := retrieve.Exact(snapshot, query, exactLimit)
	merged := mergeContextCandidates(limit, exact, base)
	return selection.Curate(snapshot, merged)
}

func mergeContextCandidates(limit int, groups ...[]retrieve.Candidate) []retrieve.Candidate {
	if limit <= 0 {
		return nil
	}
	merged := make([]retrieve.Candidate, 0, limit)
	positions := make(map[string]int, limit)
	for _, group := range groups {
		for _, candidate := range group {
			key := contextCandidateKey(candidate)
			if existing, found := positions[key]; found {
				reason := fmt.Sprintf("also retrieved by %s rank %d", candidate.Source, candidate.Rank)
				merged[existing].Reasons = appendUniqueReason(merged[existing].Reasons, reason)
				merged[existing].Context = mergeCandidateContext(merged[existing].Context, candidate.Context)
				continue
			}
			if len(merged) >= limit {
				continue
			}
			positions[key] = len(merged)
			candidate.Reasons = append([]string(nil), candidate.Reasons...)
			merged = append(merged, candidate)
		}
	}
	return merged
}

func mergeRankedProviders(limit int, groups ...[]retrieve.Candidate) []retrieve.Candidate {
	if limit <= 0 {
		return nil
	}
	type fusedCandidate struct {
		candidate     retrieve.Candidate
		key           string
		fusedScore    float64
		bestRank      int
		bestProvider  int
		reasons       []string
		context       *evidence.Descriptor
		fusionDetails []retrieve.ScoreDetail
	}

	byKey := make(map[string]*fusedCandidate)
	for providerIndex, group := range groups {
		for position, candidate := range group {
			providerRank := candidate.Rank
			if providerRank <= 0 {
				providerRank = position + 1
			}
			key := contextCandidateKey(candidate)
			contribution := 1 / (rrfRankConstant + float64(providerRank))
			reason := fmt.Sprintf("reciprocal-rank fusion from %s rank %d", candidate.Source, providerRank)
			detail := retrieve.ScoreDetail{Name: reason, Value: contribution}

			current, found := byKey[key]
			if !found {
				current = &fusedCandidate{
					candidate:    candidate,
					key:          key,
					bestRank:     providerRank,
					bestProvider: providerIndex,
					reasons:      append([]string(nil), candidate.Reasons...),
					context:      mergeCandidateContext(nil, candidate.Context),
				}
				current.candidate.Reasons = append([]string(nil), candidate.Reasons...)
				current.candidate.ScoreDetails = append([]retrieve.ScoreDetail(nil), candidate.ScoreDetails...)
				byKey[key] = current
			}

			current.fusedScore += contribution
			current.reasons = appendUniqueReason(current.reasons, reason)
			for _, candidateReason := range candidate.Reasons {
				current.reasons = appendUniqueReason(current.reasons, candidateReason)
			}
			current.context = mergeCandidateContext(current.context, candidate.Context)
			current.fusionDetails = append(current.fusionDetails, detail)

			if providerRank < current.bestRank ||
				providerRank == current.bestRank && providerIndex < current.bestProvider {
				current.candidate = candidate
				current.candidate.Reasons = append([]string(nil), candidate.Reasons...)
				current.candidate.ScoreDetails = append([]retrieve.ScoreDetail(nil), candidate.ScoreDetails...)
				current.bestRank = providerRank
				current.bestProvider = providerIndex
			}
		}
	}

	ordered := make([]*fusedCandidate, 0, len(byKey))
	for _, candidate := range byKey {
		ordered = append(ordered, candidate)
	}
	sort.Slice(ordered, func(left, right int) bool {
		a, b := ordered[left], ordered[right]
		if a.fusedScore != b.fusedScore {
			return a.fusedScore > b.fusedScore
		}
		if a.bestRank != b.bestRank {
			return a.bestRank < b.bestRank
		}
		return a.key < b.key
	})
	if len(ordered) > limit {
		ordered = ordered[:limit]
	}

	merged := make([]retrieve.Candidate, 0, len(ordered))
	for _, fused := range ordered {
		candidate := fused.candidate
		candidate.Score = fused.fusedScore
		candidate.Reasons = fused.reasons
		candidate.Context = fused.context
		candidate.ScoreDetails = append(candidate.ScoreDetails, fused.fusionDetails...)
		merged = append(merged, candidate)
	}
	return merged
}

func mergeContextProviders(
	limit int,
	exact, base, lexicon []retrieve.Candidate,
) []retrieve.Candidate {
	frontCount := min(baseFrontCandidates, len(base))
	lexiconCount := min(maxLexiconCandidates, len(lexicon))
	return mergeContextCandidates(
		limit,
		exact,
		base[:frontCount],
		lexicon[:lexiconCount],
		base[frontCount:],
	)
}

func contextCandidateKey(candidate retrieve.Candidate) string {
	if candidate.Chunk.ID != "" {
		return "id:" + candidate.Chunk.ID
	}
	return fmt.Sprintf(
		"range:%s:%d:%d", candidate.Chunk.Path,
		candidate.Chunk.StartLine, candidate.Chunk.EndLine,
	)
}

func contextCandidateSources(candidates []retrieve.Candidate) []string {
	seen := make(map[string]struct{})
	sources := make([]string, 0, 3)
	for _, candidate := range candidates {
		if candidate.Source == "" {
			continue
		}
		if _, exists := seen[candidate.Source]; exists {
			continue
		}
		seen[candidate.Source] = struct{}{}
		sources = append(sources, candidate.Source)
	}
	return sources
}

func appendUniqueReason(reasons []string, reason string) []string {
	for _, existing := range reasons {
		if existing == reason {
			return reasons
		}
	}
	return append(reasons, reason)
}

func mergeCandidateContext(left, right *evidence.Descriptor) *evidence.Descriptor {
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
