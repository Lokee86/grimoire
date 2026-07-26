package knowledgevector

import (
	"context"
	"errors"
	"testing"
)

type cancelledEmbedder struct{}

func (cancelledEmbedder) EmbedDocuments(ctx context.Context, _ []string) ([][]float32, error) {
	return nil, ctx.Err()
}

func TestEmbedEntriesReturnsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := embedEntries(ctx, cancelledEmbedder{}, nil, Paths{}, []Entry{{ID: "section", Source: "hash", Text: "text"}}, 1, 1)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}
