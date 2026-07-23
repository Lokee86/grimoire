package app

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type concurrentEmbeddingStub struct {
	active atomic.Int64
	peak   atomic.Int64
}

func (stub *concurrentEmbeddingStub) EmbedDocuments(ctx context.Context, documents []string) ([][]float32, error) {
	active := stub.active.Add(1)
	defer stub.active.Add(-1)
	for {
		peak := stub.peak.Load()
		if active <= peak || stub.peak.CompareAndSwap(peak, active) {
			break
		}
	}
	select {
	case <-time.After(25 * time.Millisecond):
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	vectors := make([][]float32, len(documents))
	for index := range documents {
		vectors[index] = []float32{1}
	}
	return vectors, nil
}

func TestEmbedVectorBatchesUsesBoundedConcurrency(t *testing.T) {
	missing := make([]vectorChunk, 8)
	for index := range missing {
		missing[index].Chunk.ID = string(rune('a' + index))
		missing[index].Chunk.Text = "document"
	}
	stub := &concurrentEmbeddingStub{}
	var mu sync.Mutex
	ingested := make(map[string]bool)
	err := embedVectorBatches(context.Background(), stub, missing, 1, 4, func(batch []vectorChunk, vectors [][]float32) error {
		mu.Lock()
		defer mu.Unlock()
		for _, entry := range batch {
			ingested[entry.Chunk.ID] = true
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if peak := stub.peak.Load(); peak < 2 || peak > 4 {
		t.Fatalf("peak embedding concurrency = %d, want 2..4", peak)
	}
	if len(ingested) != len(missing) {
		t.Fatalf("ingested %d batches, want %d", len(ingested), len(missing))
	}
}
