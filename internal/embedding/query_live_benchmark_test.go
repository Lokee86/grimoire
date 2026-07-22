package embedding

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Lokee86/grimoire/internal/tokenizer"
)

func BenchmarkLiveQueryEmbeddingModes(b *testing.B) {
	endpoint := os.Getenv("GRIMOIRE_EMBEDDING_BENCHMARK_ENDPOINT")
	if endpoint == "" {
		b.Skip("set GRIMOIRE_EMBEDDING_BENCHMARK_ENDPOINT to a live embeddings endpoint")
	}
	client := NewClient(endpoint)
	for _, tokenCount := range []int{16, 32, 64, 128} {
		for _, mode := range []QueryMode{QueryModeFast, QueryModeFull, QueryModeQuality} {
			b.Run(fmt.Sprintf("tokens_%d/%s", tokenCount, mode), func(b *testing.B) {
				plans := benchmarkPlans(b, tokenCount, mode)
				b.ResetTimer()
				for index := range b.N {
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					_, err := client.EmbedQueryPlan(ctx, plans[index].inputs, plans[index].options)
					cancel()
					if err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

func BenchmarkLiveQueryRequestBatching(b *testing.B) {
	endpoint := os.Getenv("GRIMOIRE_EMBEDDING_BENCHMARK_ENDPOINT")
	if endpoint == "" {
		b.Skip("set GRIMOIRE_EMBEDDING_BENCHMARK_ENDPOINT to a live embeddings endpoint")
	}
	client := NewClient(endpoint)
	for _, tokenCount := range []int{128, 256, 512} {
		for _, strategy := range []string{"all_windows", "sequential_64", "bounded_64"} {
			b.Run(fmt.Sprintf("tokens_%d/%s", tokenCount, strategy), func(b *testing.B) {
				plans := benchmarkPlans(b, tokenCount, QueryModeFast)
				b.ResetTimer()
				for index := range b.N {
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					err := embedBenchmarkStrategy(ctx, client, plans[index], strategy)
					cancel()
					if err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

func embedBenchmarkStrategy(ctx context.Context, client *Client, plan benchmarkPlan, strategy string) error {
	if strategy == "bounded_64" {
		_, err := client.EmbedQueryPlan(ctx, plan.inputs, plan.options)
		return err
	}
	batches, err := queryBatches(plan.inputs, plan.options)
	if err != nil {
		return err
	}
	if strategy == "all_windows" {
		all := queryBatch{inputs: plan.inputs}
		_, err = client.embedQueryBatch(ctx, all)
		return err
	}
	for _, batch := range batches {
		if _, err = client.embedQueryBatch(ctx, batch); err != nil {
			return err
		}
	}
	return nil
}

type benchmarkPlan struct {
	inputs  []QueryInput
	options QueryOptions
}

func benchmarkPlans(b *testing.B, tokenCount int, mode QueryMode) []benchmarkPlan {
	b.Helper()
	plans := make([]benchmarkPlan, b.N)
	salt := time.Now().UnixNano()
	for iteration := range b.N {
		query := benchmarkQuery(b, tokenCount, fmt.Sprintf("%d-%s-%d", salt, mode, iteration))
		options := DefaultQueryOptions()
		options.Mode = mode
		options.MaxTokens = tokenCount
		inputs, err := PlanQuery(query, options)
		if err != nil {
			b.Fatal(err)
		}
		plans[iteration] = benchmarkPlan{inputs: inputs, options: options}
	}
	return plans
}

func benchmarkQuery(b *testing.B, tokenCount int, variant string) string {
	b.Helper()
	tokens := make([]uint, 0, tokenCount)
	for window := 0; len(tokens) < tokenCount; window++ {
		segment := fmt.Sprintf(
			"variant %s window %d repository vector snapshot freshness fallback exact token budget consumer contract ",
			variant, window,
		)
		segmentTokens, err := tokenizer.Encode(segment)
		if err != nil {
			b.Fatal(err)
		}
		tokens = append(tokens, segmentTokens[:min(DefaultQueryWindowTokens, len(segmentTokens))]...)
	}
	query, err := tokenizer.Decode(tokens[:tokenCount])
	if err != nil {
		b.Fatal(err)
	}
	return query
}
