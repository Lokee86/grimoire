package app

import (
	"fmt"

	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/queryshape"
	"github.com/Lokee86/grimoire/internal/retrieve"
)

const (
	maxLexicalFileScopes  = 12
	maxChunksPerFileScope = 2
	lexicalFrontPerLane   = 4
)

type lexicalDiscoveryResult struct {
	Candidates []retrieve.Candidate
	Scopes     []retrieve.Candidate
	FilePaths  []string
}

func discoverLexically(
	snapshot index.Snapshot,
	intents []queryshape.RetrievalIntent,
	limit int,
	config retrieve.Config,
) lexicalDiscoveryResult {
	chunkCandidates := intentLexicalCandidatesWithConfig(snapshot, intents, limit, config)
	paths := intentFileScopePaths(snapshot, intents, maxLexicalFileScopes, config)
	scopedCandidates := intentScopedChunkCandidates(snapshot, intents, paths, limit, config)
	return lexicalDiscoveryResult{
		Candidates: mergeLexicalLanes(limit, chunkCandidates, scopedCandidates),
		Scopes:     scopedCandidates,
		FilePaths:  paths,
	}
}

func intentFileScopePaths(
	snapshot index.Snapshot,
	intents []queryshape.RetrievalIntent,
	limit int,
	config retrieve.Config,
) []string {
	intents = providerRetrievalIntents(intents)
	queries := make([]string, 0, len(intents))
	for _, planned := range intents {
		queries = append(queries, planned.Query)
	}
	groups := retrieve.SearchFilesManyWithConfig(snapshot, queries, max(limit*4, 32), config)
	result := make([]string, 0, limit)
	seen := make(map[string]struct{}, limit)
	cursors := make([]int, len(groups))
	for len(result) < limit {
		added := false
		for groupIndex, group := range groups {
			for cursors[groupIndex] < len(group) && supportingSourcePath(group[cursors[groupIndex]].Path) {
				cursors[groupIndex]++
			}
			if cursors[groupIndex] >= len(group) {
				continue
			}
			path := group[cursors[groupIndex]].Path
			cursors[groupIndex]++
			if _, exists := seen[path]; exists {
				continue
			}
			seen[path] = struct{}{}
			result = append(result, path)
			added = true
			if len(result) == limit {
				break
			}
		}
		if !added {
			break
		}
	}
	return result
}

func intentScopedChunkCandidates(
	snapshot index.Snapshot,
	intents []queryshape.RetrievalIntent,
	paths []string,
	limit int,
	config retrieve.Config,
) []retrieve.Candidate {
	intents = providerRetrievalIntents(intents)
	queries := make([]string, 0, len(intents))
	for _, planned := range intents {
		queries = append(queries, planned.Query)
	}
	results := retrieve.SearchManyInPathsWithConfig(snapshot, queries, paths, limit*2, config)
	groups := make([]intentCandidateGroup, 0, len(intents))
	for index, planned := range intents {
		candidates := diversifyScopedCandidates(results[index], limit)
		for candidateIndex := range candidates {
			candidates[candidateIndex].Source = "lexical-file"
			candidates[candidateIndex].Reasons = appendUniqueReason(
				candidates[candidateIndex].Reasons,
				fmt.Sprintf("chunk selected inside whole-file BM25 scope %s", candidates[candidateIndex].Chunk.Path),
			)
		}
		candidates = rankCandidatesForIntent(candidates, planned, true)
		groups = append(groups, intentCandidateGroup{Intent: planned, Candidates: candidates})
	}
	return mergeIntentCandidateGroups(limit, groups)
}

func diversifyScopedCandidates(candidates []retrieve.Candidate, limit int) []retrieve.Candidate {
	result := make([]retrieve.Candidate, 0, min(limit, len(candidates)))
	counts := make(map[string]int)
	for _, candidate := range candidates {
		if counts[candidate.Chunk.Path] >= maxChunksPerFileScope {
			continue
		}
		counts[candidate.Chunk.Path]++
		candidate.Rank = len(result) + 1
		result = append(result, candidate)
		if len(result) == limit {
			break
		}
	}
	return result
}

func mergeLexicalLanes(limit int, chunks, files []retrieve.Candidate) []retrieve.Candidate {
	front := make([]retrieve.Candidate, 0, lexicalFrontPerLane*2)
	for index := 0; index < lexicalFrontPerLane; index++ {
		if index < len(chunks) {
			front = append(front, chunks[index])
		}
		if index < len(files) {
			front = append(front, files[index])
		}
	}
	rest := mergeRankedProviders(limit, chunks, files)
	merged := mergeContextCandidates(limit, front, rest)
	for index := range merged {
		merged[index].Rank = index + 1
	}
	return merged
}
