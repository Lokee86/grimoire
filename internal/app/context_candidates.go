package app

import (
	"fmt"

	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/retrieve"
	"github.com/Lokee86/grimoire/internal/selection"
)

const (
	maxExactCandidates   = 32
	maxLexiconCandidates = 24
	baseFrontCandidates  = 32
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
	merged := make([]retrieve.Candidate, 0, limit)
	positions := make(map[string]int, limit)
	for rank := 0; len(merged) < limit; rank++ {
		advanced := false
		for _, group := range groups {
			if rank >= len(group) {
				continue
			}
			advanced = true
			candidate := group[rank]
			key := contextCandidateKey(candidate)
			if existing, found := positions[key]; found {
				reason := fmt.Sprintf("also retrieved by %s rank %d", candidate.Source, candidate.Rank)
				merged[existing].Reasons = appendUniqueReason(merged[existing].Reasons, reason)
				continue
			}
			positions[key] = len(merged)
			candidate.Reasons = append([]string(nil), candidate.Reasons...)
			merged = append(merged, candidate)
			if len(merged) >= limit {
				break
			}
		}
		if !advanced {
			break
		}
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
