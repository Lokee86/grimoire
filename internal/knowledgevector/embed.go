package knowledgevector

import (
	"context"
	"errors"
	"sync"

	"github.com/Lokee86/grimoire/internal/vectorstore"
)

type documentEmbedder interface {
	EmbedDocuments(context.Context, []string) ([][]float32, error)
}

type embeddedBatch struct {
	entries []Entry
	vectors [][]float32
	err     error
}

func embedEntries(ctx context.Context, client documentEmbedder, library *vectorstore.Library, paths Paths, entries []Entry, batchSize, concurrency int) error {
	if len(entries) == 0 {
		return nil
	}
	if batchSize <= 0 || concurrency <= 0 {
		return errors.New("knowledge embedding batch size and concurrency must be positive")
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	workerCount := min(concurrency, (len(entries)+batchSize-1)/batchSize)
	jobs := make(chan []Entry)
	results := make(chan embeddedBatch, workerCount)
	var workers sync.WaitGroup
	for range workerCount {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for batch := range jobs {
				texts := make([]string, len(batch))
				for index := range batch {
					texts[index] = batch[index].Text
				}
				vectors, err := client.EmbedDocuments(ctx, texts)
				select {
				case results <- embeddedBatch{entries: batch, vectors: vectors, err: err}:
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
		for start := 0; start < len(entries); start += batchSize {
			end := min(start+batchSize, len(entries))
			select {
			case jobs <- entries[start:end]:
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
		if err := ingestBatch(library, paths, result.entries, result.vectors); err != nil {
			firstErr = err
			cancel()
		}
	}
	if firstErr != nil {
		return firstErr
	}
	return ctx.Err()
}
