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
	queries := make([]string, 0, len(intents))
	for _, planned := range intents {
		queries = append(queries, planned.Query)
	}
	results := retrieve.SearchMany(snapshot, queries, limit)
	groups := make([]intentCandidateGroup, 0, len(intents))
	for index, planned := range intents {
		candidates := rankCandidatesForIntent(results[index], planned, true)
		groups = append(groups, intentCandidateGroup{Intent: planned, Candidates: candidates})
	}
	return mergeIntentCandidateGroups(limit, groups)
}

func intentExactCandidates(snapshot index.Snapshot, intents []queryshape.RetrievalIntent, limit int) []retrieve.Candidate {
	groups := make([]intentCandidateGroup, 0, len(intents))
	for _, planned := range intents {
		candidates := retrieve.Exact(snapshot, planned.Query, limit)
		candidates = rankCandidatesForIntent(candidates, planned, false)
		groups = append(groups, intentCandidateGroup{Intent: planned, Candidates: candidates})
	}
	return mergeIntentCandidateGroups(limit, groups)
}

func rankCandidatesForIntent(candidates []retrieve.Candidate, planned queryshape.RetrievalIntent, boost bool) []retrieve.Candidate {
	result := append([]retrieve.Candidate(nil), candidates...)
	for index := range result {
		result[index] = annotateCandidateIntent(result[index], planned)
		if !boost {
			continue
		}
		name, value := candidateIntentBoost(result[index], planned.Intent)
		if value > 0 {
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
	candidate.Context = mergeCandidateContext(candidate.Context, &descriptor)
	candidate.Reasons = appendUniqueReason(candidate.Reasons, fmt.Sprintf("retrieval intent %s", planned.Intent))
	return candidate
}
