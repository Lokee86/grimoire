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
		candidates := make([]Result, 0, len(matches))
		seen := make(map[string]bool)
		for _, match := range matches {
			if isDocumentationPath(match.Node.Path) {
				continue
			}
			candidate := Result{
				Provider: "lexicon",
				Kind:     match.Node.Kind,
				Node:     engine.node("lexicon", engine.lexiconSnapshot, match.Node),
				Excerpt:  engine.nodeExcerpt(match.Node),
				Score:    match.Score,
				Reasons:  append([]string(nil), match.Reasons...),
			}
			key := resultSemanticKey(candidate)
			if seen[key] {
				continue
			}
			seen[key] = true
			symbolAvailable++
			candidates = append(candidates, candidate)
		}
		response.SymbolMatches = selectDiverseResults(candidates, request.Limit)
	}

	if len(response.SymbolMatches) == 0 && engine.arcanaSnapshot != "" {
		arcana, closeArcana := engine.openArcanaQuery(ctx, response)
		defer closeArcana()
		if arcana != nil {
			nodes, err := arcana.Resolve(ctx, engine.arcanaSnapshot, request.Query, "", candidateLimit)
			if err != nil {
				response.Warnings = append(response.Warnings, "Arcana symbol discovery unavailable: "+err.Error())
			} else {
				candidates := make([]Result, 0, len(nodes))
				seen := make(map[string]bool)
				symbolAvailable = 0
				for _, value := range nodes {
					if isDocumentationPath(value.Path) {
						continue
					}
					candidate := Result{
						Provider: "arcana",
						Kind:     value.Kind,
						Node:     engine.node("arcana", engine.arcanaSnapshotID, value),
						Excerpt:  engine.nodeExcerpt(value),
						Reasons:  []string{"Arcana graph symbol match"},
					}
					key := resultSemanticKey(candidate)
					if seen[key] {
						continue
					}
					seen[key] = true
					symbolAvailable++
					candidates = append(candidates, candidate)
				}
				response.SymbolMatches = selectDiverseResults(candidates, request.Limit)
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
	eligibleResults := make([]Result, 0, len(candidates))
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
		eligibleResults = append(eligibleResults, Result{
			Provider: provider,
			Kind:     sourceKind(candidate.Chunk),
			Node:     node,
			Excerpt:  chunkExcerpt(candidate.Chunk),
			Score:    candidate.Score,
			Reasons:  append([]string{reason}, candidate.Reasons...),
		})
	}
	return selectDiverseResults(eligibleResults, limit), eligible, suppressed
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
