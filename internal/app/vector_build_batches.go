package app

import (
	"context"
	"errors"
	"sync"

	"github.com/Lokee86/grimoire/internal/vectorstore"
)

type documentEmbedder interface {
	EmbedDocuments(context.Context, []string) ([][]float32, error)
}

type embeddedVectorBatch struct {
	batch   []vectorChunk
	vectors [][]float32
	err     error
}

func embedMissing(
	ctx context.Context,
	client documentEmbedder,
	library *vectorstore.Library,
	paths vectorStatePaths,
	missing []vectorChunk,
	batchSize int,
	concurrency int,
	progress func(int),
) error {
	return embedVectorBatchesWithProgress(ctx, client, missing, batchSize, concurrency, func(batch []vectorChunk, vectors [][]float32) error {
		return ingestVectorBatch(library, paths, batch, vectors)
	}, progress)
}

func embedVectorBatches(
	ctx context.Context,
	client documentEmbedder,
	missing []vectorChunk,
	batchSize int,
	concurrency int,
	ingest func([]vectorChunk, [][]float32) error,
) error {
	return embedVectorBatchesWithProgress(ctx, client, missing, batchSize, concurrency, ingest, nil)
}

func embedVectorBatchesWithProgress(
	ctx context.Context,
	client documentEmbedder,
	missing []vectorChunk,
	batchSize int,
	concurrency int,
	ingest func([]vectorChunk, [][]float32) error,
	progress func(int),
) error {
	if len(missing) == 0 {
		return nil
	}
	if batchSize <= 0 || concurrency <= 0 {
		return errors.New("embedding batch size and concurrency must be positive")
	}
	workerCount := min(concurrency, (len(missing)+batchSize-1)/batchSize)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan []vectorChunk)
	results := make(chan embeddedVectorBatch, workerCount)
	var workers sync.WaitGroup
	for range workerCount {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for batch := range jobs {
				documents := make([]string, len(batch))
				for index, entry := range batch {
					documents[index] = entry.Chunk.Text
				}
				vectors, err := client.EmbedDocuments(ctx, documents)
				select {
				case results <- embeddedVectorBatch{batch: batch, vectors: vectors, err: err}:
				case <-ctx.Done():
					return
				}
				if err != nil {
					return
				}
			}
		}()
	}
	go func() {
		defer close(jobs)
		for start := 0; start < len(missing); start += batchSize {
			end := min(start+batchSize, len(missing))
			select {
			case jobs <- missing[start:end]:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		workers.Wait()
		close(results)
	}()

	var firstErr error
	for result := range results {
		if result.err != nil {
			if firstErr == nil {
				firstErr = result.err
				cancel()
			}
			continue
		}
		if firstErr != nil {
			continue
		}
		if err := ingest(result.batch, result.vectors); err != nil {
			firstErr = err
			cancel()
			continue
		}
		if progress != nil {
			progress(len(result.batch))
		}
	}
	return firstErr
}
