package agentquery

import (
	"context"

	"github.com/Lokee86/grimoire/internal/retrieve"
	"github.com/Lokee86/grimoire/internal/structure"
)

func (engine *Engine) search(ctx context.Context, request Request, response *Response) error {
	candidateLimit := min(200, max(request.Limit*8, request.Limit))

	exactCandidates := retrieve.Exact(engine.source, request.Query, candidateLimit)
	exactMatches, exactTruncated := engine.sourceResults(
		exactCandidates,
		request.Limit,
		"exact",
		"literal source match",
	)
	response.ExactMatches = exactMatches
	markLaneTruncated(response, "exact_matches", exactTruncated)

	exactHandles := make(map[string]string, len(exactMatches))
	for _, result := range exactMatches {
		exactHandles[handleKey(result.Node.Handle)] = result.Node.Handle.Value
	}
	sourceCandidates := retrieve.Search(engine.source, request.Query, candidateLimit)
	sourceMatches, sourceTruncated := engine.sourceResults(
		sourceCandidates,
		request.Limit,
		"lexical",
		"prepared source BM25 match",
	)
	for index := range sourceMatches {
		if duplicate := exactHandles[handleKey(sourceMatches[index].Node.Handle)]; duplicate != "" {
			sourceMatches[index].DuplicateOf = duplicate
			sourceMatches[index].Excerpt = ""
			sourceMatches[index].Reasons = []string{"same source range as exact match"}
		}
	}
	response.SourceMatches = sourceMatches
	markLaneTruncated(response, "source_matches", sourceTruncated)

	seeds := make([]structure.Node, 0, request.Limit)
	if engine.lexicon != nil {
		matches := engine.lexicon.Find(request.Query, candidateLimit)
		eligible := 0
		for _, match := range matches {
			if isDocumentationPath(match.Node.Path) {
				continue
			}
			eligible++
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
			seeds = append(seeds, match.Node)
		}
		for index := range response.SymbolMatches {
			response.SymbolMatches[index].Rank = index + 1
		}
		markLaneTruncated(response, "symbol_matches", eligible > len(response.SymbolMatches))
	}

	if len(response.SymbolMatches) == 0 && engine.arcanaSnapshot != "" {
		nodes, err := engine.arcana.Resolve(ctx, engine.arcanaSnapshot, request.Query, "", candidateLimit)
		if err != nil {
			response.Warnings = append(response.Warnings, "Arcana symbol discovery unavailable: "+err.Error())
		} else {
			eligible := 0
			for _, value := range nodes {
				if isDocumentationPath(value.Path) {
					continue
				}
				eligible++
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
				seeds = append(seeds, value)
			}
			for index := range response.SymbolMatches {
				response.SymbolMatches[index].Rank = index + 1
			}
			markLaneTruncated(response, "symbol_matches", eligible > len(response.SymbolMatches))
		}
	}

	response.RelationshipMatches = engine.searchRelationships(ctx, request, seeds, response)
	return nil
}

func (engine *Engine) sourceResults(
	candidates []retrieve.Candidate,
	limit int,
	provider string,
	reason string,
) ([]Result, bool) {
	results := make([]Result, 0, min(limit, len(candidates)))
	seen := make(map[string]bool)
	eligible := 0
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
	return results, eligible > len(results)
}

func markLaneTruncated(response *Response, lane string, truncated bool) {
	if !truncated {
		return
	}
	response.Truncated = true
	response.TruncatedLanes = append(response.TruncatedLanes, lane)
}
