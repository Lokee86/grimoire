package app

import (
	"fmt"

	"github.com/Lokee86/grimoire/internal/retrieve"
)

// mergeScopedContextProviders keeps lexical discovery authoritative. Structural
// providers may enrich an already discovered source range, but they cannot add
// or reorder source candidates in the context lane.
func mergeScopedContextProviders(
	limit int,
	exact, base, lexicon, arcana []retrieve.Candidate,
) []retrieve.Candidate {
	merged := mergeContextProviders(limit, exact, base, nil, nil)
	positions := make(map[string]int, len(merged))
	for index, candidate := range merged {
		positions[contextCandidateKey(candidate)] = index
	}
	for _, group := range [][]retrieve.Candidate{lexicon, arcana} {
		for _, candidate := range group {
			position, exists := positions[contextCandidateKey(candidate)]
			if !exists {
				continue
			}
			current := &merged[position]
			current.Reasons = appendUniqueReason(
				current.Reasons,
				fmt.Sprintf("inspected by %s rank %d", candidate.Source, candidate.Rank),
			)
			for _, reason := range candidate.Reasons {
				current.Reasons = appendUniqueReason(current.Reasons, reason)
			}
		}
	}
	for index := range merged {
		merged[index].Rank = index + 1
	}
	return merged
}
