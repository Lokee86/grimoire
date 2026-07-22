package embedding

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type queryBatch struct {
	start  int
	inputs []QueryInput
}

func (client *Client) EmbedQueryPlan(
	ctx context.Context,
	plan []QueryInput,
	options QueryOptions,
) ([][]float32, error) {
	batches, err := queryBatches(plan, options)
	if err != nil {
		return nil, err
	}
	if len(batches) == 1 {
		return client.embedQueryBatch(ctx, batches[0])
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	vectors := make([][]float32, len(plan))
	jobs := make(chan queryBatch, len(batches))
	workerCount := min(options.BatchConcurrency, len(batches))
	var wait sync.WaitGroup
	var firstError error
	var errorOnce sync.Once

	for range workerCount {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for batch := range jobs {
				batchVectors, batchErr := client.embedQueryBatch(ctx, batch)
				if batchErr != nil {
					errorOnce.Do(func() {
						firstError = batchErr
						cancel()
					})
					continue
				}
				copy(vectors[batch.start:], batchVectors)
			}
		}()
	}
	for _, batch := range batches {
		jobs <- batch
	}
	close(jobs)
	wait.Wait()
	if firstError != nil {
		return nil, firstError
	}
	return vectors, nil
}

func (client *Client) embedQueryBatch(ctx context.Context, batch queryBatch) ([][]float32, error) {
	texts := make([]string, len(batch.inputs))
	for index, input := range batch.inputs {
		texts[index] = input.Text
	}
	vectors, err := client.EmbedQueries(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("embed query batch starting at input %d: %w", batch.start+1, err)
	}
	if len(vectors) != len(batch.inputs) {
		return nil, fmt.Errorf(
			"query batch starting at input %d returned %d vectors for %d inputs",
			batch.start+1, len(vectors), len(batch.inputs),
		)
	}
	return vectors, nil
}

func queryBatches(plan []QueryInput, options QueryOptions) ([]queryBatch, error) {
	if len(plan) == 0 {
		return nil, errors.New("query plan is empty")
	}
	if err := options.Validate(); err != nil {
		return nil, err
	}
	if options.Mode == QueryModeFull || len(plan) == 1 {
		return []queryBatch{{inputs: plan}}, nil
	}

	batches := make([]queryBatch, 0, len(plan))
	windowStart := 0
	if options.Mode == QueryModeQuality {
		batches = append(batches, queryBatch{inputs: plan[:1]})
		windowStart = 1
	}
	windowsPerBatch := options.BatchTokens / options.WindowTokens
	for start := windowStart; start < len(plan); start += windowsPerBatch {
		end := min(start+windowsPerBatch, len(plan))
		batches = append(batches, queryBatch{start: start, inputs: plan[start:end]})
	}
	return batches, nil
}
