package app

import (
	"fmt"
	"sort"

	"github.com/Lokee86/grimoire/internal/evidence"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/queryshape"
	"github.com/Lokee86/grimoire/internal/retrieve"
)

const maxActiveIntentPasses = 6

type intentCandidateGroup struct {
	Intent     queryshape.RetrievalIntent
	Candidates []retrieve.Candidate
}

func activeRetrievalIntents(query string) []queryshape.RetrievalIntent {
	planned := queryshape.PlanRetrievalIntents(query)
	if len(planned) > maxActiveIntentPasses {
		planned = planned[:maxActiveIntentPasses]
	}
	return planned
}

func intentLexicalCandidates(snapshot index.Snapshot, intents []queryshape.RetrievalIntent, limit int) []retrieve.Candidate {
	return intentLexicalCandidatesWithConfig(snapshot, intents, limit, retrieve.DefaultConfig())
}

func intentLexicalCandidatesWithConfig(
	snapshot index.Snapshot,
	intents []queryshape.RetrievalIntent,
	limit int,
	config retrieve.Config,
) []retrieve.Candidate {
	intents = providerRetrievalIntents(intents)
	queries := make([]string, 0, len(intents))
	for _, planned := range intents {
		queries = append(queries, planned.Query)
	}
	results := retrieve.SearchManyWithConfig(snapshot, queries, limit, config)
	groups := make([]intentCandidateGroup, 0, len(intents))
	for index, planned := range intents {
		candidates := rankCandidatesForIntent(results[index], planned, true)
		groups = append(groups, intentCandidateGroup{Intent: planned, Candidates: candidates})
	}
	return mergeIntentCandidateGroups(limit, groups)
}

func intentExactCandidates(snapshot index.Snapshot, intents []queryshape.RetrievalIntent, limit int) []retrieve.Candidate {
	intents = providerRetrievalIntents(intents)
	groups := make([]intentCandidateGroup, 0, len(intents))
	for _, planned := range intents {
		candidates := retrieve.Exact(snapshot, planned.Query, limit)
		candidates = rankCandidatesForIntent(candidates, planned, false)
		groups = append(groups, intentCandidateGroup{Intent: planned, Candidates: candidates})
	}
	return mergeIntentCandidateGroups(limit, groups)
}

func providerRetrievalIntents(intents []queryshape.RetrievalIntent) []queryshape.RetrievalIntent {
	if len(intents) <= 1 {
		return intents
	}
	result := make([]queryshape.RetrievalIntent, 0, len(intents))
	for _, planned := range intents {
		if planned.Intent == evidence.IntentMixed && planned.Weight <= 0.25 {
			continue
		}
		result = append(result, planned)
	}
	if len(result) == 0 {
		return intents
	}
	return result
}

func rankCandidatesForIntent(candidates []retrieve.Candidate, planned queryshape.RetrievalIntent, boost bool) []retrieve.Candidate {
	result := append([]retrieve.Candidate(nil), candidates...)
	for index := range result {
		result[index] = annotateCandidateIntent(result[index], planned)
		if !boost {
			continue
		}
		name, value := candidateIntentBoost(result[index], planned)
		if value != 0 {
			result[index].Score += value
			result[index].Reasons = appendUniqueReason(result[index].Reasons, name)
			result[index].ScoreDetails = append(result[index].ScoreDetails, retrieve.ScoreDetail{Name: name, Value: value})
		}
	}
	if boost {
		sort.SliceStable(result, func(left, right int) bool {
			if result[left].Score != result[right].Score {
				return result[left].Score > result[right].Score
			}
			return contextCandidateKey(result[left]) < contextCandidateKey(result[right])
		})
	}
	for index := range result {
		result[index].Rank = index + 1
		if planned.FacetID != "" && result[index].Context != nil {
			if result[index].Context.FacetRanks == nil {
				result[index].Context.FacetRanks = make(map[string]int)
			}
			result[index].Context.FacetRanks[planned.FacetID] = index + 1
		}
	}
	return result
}

func annotateCandidateIntent(candidate retrieve.Candidate, planned queryshape.RetrievalIntent) retrieve.Candidate {
	role := evidence.RolePrimary
	if planned.Intent == evidence.IntentMixed {
		role = evidence.RoleContext
	} else if supportingSourcePath(candidate.Chunk.Path) {
		role = evidence.RoleSupporting
	}
	identity := evidence.RangeIdentity(candidate.Chunk.Path, candidate.Chunk.StartLine, candidate.Chunk.EndLine)
	descriptor := evidence.Descriptor{
		Identity: identity, Intents: []evidence.Intent{planned.Intent}, Roles: []evidence.Role{role},
		EstimatedTokens: candidate.Chunk.TokenCount, RedundancyKey: identity,
	}
	if planned.FacetID != "" {
		rank := candidate.Rank
		if rank <= 0 {
			rank = 1
		}
		descriptor.GroupIDs = []string{planned.FacetID}
		descriptor.FacetRanks = map[string]int{planned.FacetID: rank}
	}
	candidate.Context = mergeCandidateContext(candidate.Context, &descriptor)
	candidate.Reasons = appendUniqueReason(candidate.Reasons, fmt.Sprintf("retrieval intent %s", planned.Intent))
	return candidate
}
