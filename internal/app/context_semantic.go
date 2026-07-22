package app

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/Lokee86/grimoire/internal/embedding"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/retrieve"
	"github.com/Lokee86/grimoire/internal/vectorstore"
)

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
	paths := resolveVectorPaths(statePath)
	chunks := snapshot.AllChunks()
	manifest, err := validateVectorSnapshotManifest(paths.Manifest, snapshot, len(chunks))
	if err != nil {
		return nil, err
	}
	library, err := vectorstore.Load(enginePath)
	if err != nil {
		return nil, err
	}
	defer library.Close()
	engine, err := library.OpenSnapshot(paths.Snapshot)
	if err != nil {
		return nil, err
	}
	defer engine.Close()
	info, err := engine.Info()
	if err != nil {
		return nil, err
	}
	if err := validateVectorEngineInfo(manifest, info); err != nil {
		return nil, err
	}
	return queryVectorCandidates(ctx, engine, info, chunks, query, endpoint, limit, queryOptions)
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
	texts := make([]string, len(plan))
	for index, input := range plan {
		texts[index] = input.Text
	}
	vectors, err := embedding.NewClient(endpoint).EmbedQueries(ctx, texts)
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
	id        string
	score     float32
	bestInput int
	bestRank  int
	matches   int
	reasons   []string
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
			current, exists := merged[hit.ID]
			if !exists {
				merged[hit.ID] = &mergedSemanticHit{
					id: hit.ID, score: hit.Score, bestInput: inputIndex,
					bestRank: rank + 1, matches: 1, reasons: []string{reason},
				}
				continue
			}
			current.matches++
			current.reasons = append(current.reasons, reason)
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
			Rank: rank + 1, Reasons: hit.reasons,
		})
	}
	return candidates, nil
}
