package embedding

import (
	"context"
	"fmt"
	"os"
	"strings"
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
				inputs := benchmarkInputs(b, tokenCount, mode)
				b.ResetTimer()
				for index := range b.N {
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					_, err := client.EmbedQueries(ctx, inputs[index])
					cancel()
					if err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

func benchmarkInputs(b *testing.B, tokenCount int, mode QueryMode) [][]string {
	b.Helper()
	inputs := make([][]string, b.N)
	salt := time.Now().UnixNano()
	for iteration := range b.N {
		query := benchmarkQuery(b, tokenCount, fmt.Sprintf("%d-%s-%d", salt, mode, iteration))
		plan, err := PlanQuery(query, QueryOptions{
			Mode: mode, WindowTokens: DefaultQueryWindowTokens, MaxTokens: tokenCount,
		})
		if err != nil {
			b.Fatal(err)
		}
		inputs[iteration] = make([]string, len(plan))
		for index, item := range plan {
			inputs[iteration][index] = item.Text
		}
	}
	return inputs
}

func benchmarkQuery(b *testing.B, tokenCount int, variant string) string {
	b.Helper()
	source := "benchmark variant " + variant + " " + strings.Repeat(
		"repository vector snapshot freshness fallback exact token budget consumer contract ", 64,
	)
	tokens, err := tokenizer.Encode(source)
	if err != nil {
		b.Fatal(err)
	}
	if len(tokens) < tokenCount {
		b.Fatalf("benchmark source has %d tokens, need %d", len(tokens), tokenCount)
	}
	query, err := tokenizer.Decode(tokens[:tokenCount])
	if err != nil {
		b.Fatal(err)
	}
	return query
}
