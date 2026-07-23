package app

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/Lokee86/grimoire/internal/embedding"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/retrieve"
	"github.com/Lokee86/grimoire/internal/vectorstore"
)

type semanticMetrics struct {
	SnapshotValidation time.Duration
	Embedding          time.Duration
	VectorSearch       time.Duration
	CandidateMerge     time.Duration
	DiagnosticProbe    time.Duration
}

type semanticEvaluationResult struct {
	Candidates []retrieve.Candidate
	BroadProbe []retrieve.Candidate
	Metrics    semanticMetrics
}

func semanticCandidates(
	ctx context.Context,
	snapshot index.Snapshot,
	statePath string,
	query string,
	endpoint string,
	enginePath string,
	limit int,
	queryOptions embedding.QueryOptions,
) ([]retrieve.Candidate, error) {
	result, err := semanticCandidatesForEvaluation(
		ctx, snapshot, statePath, query, endpoint, enginePath, limit, 0, queryOptions,
	)
	return result.Candidates, err
}

func semanticCandidatesForEvaluation(
	ctx context.Context,
	snapshot index.Snapshot,
	statePath string,
	query string,
	endpoint string,
	enginePath string,
	limit int,
	probeLimit int,
	queryOptions embedding.QueryOptions,
) (semanticEvaluationResult, error) {
	var result semanticEvaluationResult
	validationStart := time.Now()
	paths := resolveVectorPaths(statePath)
	chunks := snapshot.AllChunks()
	manifest, err := validateVectorSnapshotManifest(paths.Manifest, snapshot, len(chunks))
	if err != nil {
		return result, err
	}
	library, err := vectorstore.Load(enginePath)
	if err != nil {
		return result, err
	}
	defer library.Close()
	engine, err := library.OpenSnapshot(paths.Snapshot)
	if err != nil {
		return result, err
	}
	defer engine.Close()
	info, err := engine.Info()
	if err != nil {
		return result, err
	}
	if err := validateVectorEngineInfo(manifest, info); err != nil {
		return result, err
	}
	result.Metrics.SnapshotValidation = time.Since(validationStart)

	plan, err := embedding.PlanQuery(query, queryOptions)
	if err != nil {
		return result, err
	}
	embeddingStart := time.Now()
	vectors, err := embedding.NewClient(endpoint).EmbedQueryPlan(ctx, plan, queryOptions)
	result.Metrics.Embedding = time.Since(embeddingStart)
	if err != nil {
		return result, err
	}
	for _, vector := range vectors {
		if info.Dimensions != len(vector) {
			return result, fmt.Errorf("vector snapshot has %d dimensions, query has %d", info.Dimensions, len(vector))
		}
	}

	searchStart := time.Now()
	hits, err := searchQueryVectors(engine, vectors, limit)
	result.Metrics.VectorSearch = time.Since(searchStart)
	if err != nil {
		return result, err
	}
	mergeStart := time.Now()
	result.Candidates, err = mergeSemanticHits(chunks, plan, hits, limit)
	result.Metrics.CandidateMerge = time.Since(mergeStart)
	if err != nil {
		return result, err
	}

	if probeLimit > limit {
		probeStart := time.Now()
		probeHits, probeErr := searchQueryVectors(engine, vectors, probeLimit)
		if probeErr != nil {
			return result, probeErr
		}
		result.BroadProbe, probeErr = mergeSemanticHits(chunks, plan, probeHits, probeLimit)
		result.Metrics.DiagnosticProbe = time.Since(probeStart)
		if probeErr != nil {
			return result, probeErr
		}
	} else {
		result.BroadProbe = append([]retrieve.Candidate(nil), result.Candidates...)
	}
	return result, nil
}

func queryVectorCandidates(
	ctx context.Context,
	engine *vectorstore.Engine,
	info vectorstore.Info,
	chunks []index.Chunk,
	query string,
	endpoint string,
	limit int,
	queryOptions embedding.QueryOptions,
) ([]retrieve.Candidate, error) {
	plan, err := embedding.PlanQuery(query, queryOptions)
	if err != nil {
		return nil, err
	}
	vectors, err := embedding.NewClient(endpoint).EmbedQueryPlan(ctx, plan, queryOptions)
	if err != nil {
		return nil, err
	}
	for _, vector := range vectors {
		if info.Dimensions != len(vector) {
			return nil, fmt.Errorf("vector snapshot has %d dimensions, query has %d", info.Dimensions, len(vector))
		}
	}
	hits, err := searchQueryVectors(engine, vectors, limit)
	if err != nil {
		return nil, err
	}
	return mergeSemanticHits(chunks, plan, hits, limit)
}

type querySearchResult struct {
	hits []vectorstore.Hit
	err  error
}

func searchQueryVectors(
	engine *vectorstore.Engine,
	vectors [][]float32,
	limit int,
) ([][]vectorstore.Hit, error) {
	results := make([]querySearchResult, len(vectors))
	var wait sync.WaitGroup
	wait.Add(len(vectors))
	for index := range vectors {
		go func() {
			defer wait.Done()
			results[index].hits, results[index].err = engine.Search(vectors[index], limit)
		}()
	}
	wait.Wait()

	hits := make([][]vectorstore.Hit, len(results))
	for index, result := range results {
		if result.err != nil {
			return nil, fmt.Errorf("search query vector %d: %w", index+1, result.err)
		}
		hits[index] = result.hits
	}
	return hits, nil
}

type mergedSemanticHit struct {
	id           string
	score        float32
	bestInput    int
	bestRank     int
	matches      int
	reasons      []string
	scoreDetails []retrieve.ScoreDetail
}

func mergeSemanticHits(
	chunks []index.Chunk,
	plan []embedding.QueryInput,
	hits [][]vectorstore.Hit,
	limit int,
) ([]retrieve.Candidate, error) {
	merged := make(map[string]*mergedSemanticHit)
	for inputIndex, inputHits := range hits {
		for rank, hit := range inputHits {
			reason := fmt.Sprintf("semantic vector similarity from %s", plan[inputIndex].Label)
			detail := retrieve.ScoreDetail{Name: reason, Value: float64(hit.Score)}
			current, exists := merged[hit.ID]
			if !exists {
				merged[hit.ID] = &mergedSemanticHit{
					id: hit.ID, score: hit.Score, bestInput: inputIndex,
					bestRank: rank + 1, matches: 1, reasons: []string{reason},
					scoreDetails: []retrieve.ScoreDetail{detail},
				}
				continue
			}
			current.matches++
			current.reasons = append(current.reasons, reason)
			current.scoreDetails = append(current.scoreDetails, detail)
			if hit.Score > current.score || hit.Score == current.score && rank+1 < current.bestRank {
				current.score = hit.Score
				current.bestInput = inputIndex
				current.bestRank = rank + 1
			}
		}
	}

	ordered := make([]*mergedSemanticHit, 0, len(merged))
	for _, hit := range merged {
		ordered = append(ordered, hit)
	}
	sort.Slice(ordered, func(left, right int) bool {
		a, b := ordered[left], ordered[right]
		if a.score != b.score {
			return a.score > b.score
		}
		if a.matches != b.matches {
			return a.matches > b.matches
		}
		if a.bestRank != b.bestRank {
			return a.bestRank < b.bestRank
		}
		if a.bestInput != b.bestInput {
			return a.bestInput < b.bestInput
		}
		return a.id < b.id
	})
	if len(ordered) > limit {
		ordered = ordered[:limit]
	}

	byID := make(map[string]index.Chunk, len(chunks))
	for _, chunk := range chunks {
		byID[chunk.ID] = chunk
	}
	candidates := make([]retrieve.Candidate, 0, len(ordered))
	for rank, hit := range ordered {
		chunk, exists := byID[hit.id]
		if !exists {
			return nil, fmt.Errorf("vector result %s is absent from the prepared index", hit.id)
		}
		candidates = append(candidates, retrieve.Candidate{
			Chunk: chunk, Score: float64(hit.score), Source: "vector",
			Rank: rank + 1, Reasons: hit.reasons, ScoreDetails: hit.scoreDetails,
		})
	}
	return candidates, nil
}
