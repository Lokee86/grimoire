package agentquery

import (
	"context"

	"github.com/Lokee86/grimoire/internal/retrieve"
)

func (engine *Engine) search(ctx context.Context, request Request, response *Response) error {
	candidateLimit := min(200, max(request.Limit*8, request.Limit))

	exactCandidates := retrieve.Exact(engine.source, request.Query, candidateLimit)
	exactMatches, exactAvailable, _ := engine.sourceResults(
		exactCandidates,
		request.Limit,
		"exact",
		"literal source match",
		nil,
	)
	exactPreviewed := applyResultPreviews(exactMatches, request.Detail)
	response.ExactMatches = exactMatches
	recordLaneCoverage(response, "exact_matches", exactAvailable, len(exactMatches), exactPreviewed, 0)

	exactHandles := make(map[string]string, len(exactMatches))
	for _, result := range exactMatches {
		exactHandles[handleKey(result.Node.Handle)] = result.Node.Handle.Value
	}
	sourceCandidates := retrieve.Search(engine.source, request.Query, candidateLimit)
	sourceMatches, sourceAvailable, suppressedDuplicates := engine.sourceResults(
		sourceCandidates,
		request.Limit,
		"lexical",
		"prepared source BM25 match",
		exactHandles,
	)
	sourcePreviewed := applyResultPreviews(sourceMatches, request.Detail)
	response.SourceMatches = sourceMatches
	recordLaneCoverage(response, "source_matches", sourceAvailable, len(sourceMatches), sourcePreviewed, suppressedDuplicates)

	symbolAvailable := 0
	if engine.lexicon != nil {
		matches := engine.lexicon.Find(request.Query, candidateLimit)
		for _, match := range matches {
			if isDocumentationPath(match.Node.Path) {
				continue
			}
			symbolAvailable++
			if len(response.SymbolMatches) >= request.Limit {
				continue
			}
			response.SymbolMatches = append(response.SymbolMatches, Result{
				Provider: "lexicon",
				Kind:     match.Node.Kind,
				Node:     engine.node("lexicon", engine.lexiconSnapshot, match.Node),
				Excerpt:  engine.nodeExcerpt(match.Node),
				Score:    match.Score,
				Reasons:  append([]string(nil), match.Reasons...),
			})
		}
		for index := range response.SymbolMatches {
			response.SymbolMatches[index].Rank = index + 1
		}
	}

	if len(response.SymbolMatches) == 0 && engine.arcanaSnapshot != "" {
		nodes, err := engine.arcana.Resolve(ctx, engine.arcanaSnapshot, request.Query, "", candidateLimit)
		if err != nil {
			response.Warnings = append(response.Warnings, "Arcana symbol discovery unavailable: "+err.Error())
		} else {
			symbolAvailable = 0
			for _, value := range nodes {
				if isDocumentationPath(value.Path) {
					continue
				}
				symbolAvailable++
				if len(response.SymbolMatches) >= request.Limit {
					continue
				}
				response.SymbolMatches = append(response.SymbolMatches, Result{
					Provider: "arcana",
					Kind:     value.Kind,
					Node:     engine.node("arcana", engine.arcanaSnapshotID, value),
					Excerpt:  engine.nodeExcerpt(value),
					Reasons:  []string{"Arcana graph symbol match"},
				})
			}
			for index := range response.SymbolMatches {
				response.SymbolMatches[index].Rank = index + 1
			}
		}
	}
	symbolPreviewed := applyResultPreviews(response.SymbolMatches, request.Detail)
	recordLaneCoverage(response, "symbol_matches", symbolAvailable, len(response.SymbolMatches), symbolPreviewed, 0)

	deferRelationshipExpansion(response, searchExpansionCandidateCount(response))
	return nil
}

func (engine *Engine) sourceResults(
	candidates []retrieve.Candidate,
	limit int,
	provider string,
	reason string,
	exclude map[string]string,
) ([]Result, int, int) {
	results := make([]Result, 0, min(limit, len(candidates)))
	seen := make(map[string]bool)
	eligible := 0
	suppressed := 0
	for _, candidate := range candidates {
		if isDocumentationPath(candidate.Chunk.Path) {
			continue
		}
		node := engine.sourceNode(candidate.Chunk)
		key := handleKey(node.Handle)
		if seen[key] {
			continue
		}
		seen[key] = true
		if exclude[key] != "" {
			suppressed++
			continue
		}
		eligible++
		if len(results) >= limit {
			continue
		}
		results = append(results, Result{
			Rank:     len(results) + 1,
			Provider: provider,
			Kind:     sourceKind(candidate.Chunk),
			Node:     node,
			Excerpt:  chunkExcerpt(candidate.Chunk),
			Score:    candidate.Score,
			Reasons:  append([]string{reason}, candidate.Reasons...),
		})
	}
	return results, eligible, suppressed
}

func searchExpansionCandidateCount(response *Response) int {
	if response == nil {
		return 0
	}
	seen := make(map[string]bool)
	for _, lane := range [][]Result{response.ExactMatches, response.SourceMatches, response.SymbolMatches} {
		for _, result := range lane {
			if result.Node.Handle.Value == "" {
				continue
			}
			key := handleKey(result.Node.Handle)
			if seen[key] {
				continue
			}
			seen[key] = true
		}
	}
	return len(seen)
}

func markLaneTruncated(response *Response, lane string, truncated bool) {
	if !truncated {
		return
	}
	response.Truncated = true
	for _, existing := range response.TruncatedLanes {
		if existing == lane {
			return
		}
	}
	response.TruncatedLanes = append(response.TruncatedLanes, lane)
}
