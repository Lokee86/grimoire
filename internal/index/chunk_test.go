package index

import (
	"strings"
	"testing"

	"github.com/Lokee86/grimoire/internal/tokenizer"
)

func TestChunkFileSplitsOversizedSingleLine(t *testing.T) {
	content := strings.Repeat("oversized_token ", defaultMaxChunkTokens*3)
	chunks, err := chunkFile("bundle.js", content)
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) < 3 {
		t.Fatalf("chunks = %d, want at least 3", len(chunks))
	}
	seen := make(map[string]struct{}, len(chunks))
	for _, chunk := range chunks {
		if chunk.StartLine != 1 || chunk.EndLine != 1 {
			t.Fatalf("single-line chunk has range %d-%d", chunk.StartLine, chunk.EndLine)
		}
		if _, exists := seen[chunk.ID]; exists {
			t.Fatalf("duplicate chunk ID %q", chunk.ID)
		}
		seen[chunk.ID] = struct{}{}
		assertChunkTokenLimit(t, chunk)
	}
}

func TestChunkFileSplitsOversizedMultilineContent(t *testing.T) {
	line := strings.Repeat("source_token ", 250)
	content := strings.Repeat(line+"\n", 20)
	chunks, err := chunkFile("large.go", content)
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) < 2 {
		t.Fatalf("chunks = %d, want multiple chunks", len(chunks))
	}
	previousEnd := 0
	for _, chunk := range chunks {
		assertChunkTokenLimit(t, chunk)
		if chunk.StartLine <= previousEnd {
			t.Fatalf("chunk ranges overlap or regress: previous end %d, current %d-%d", previousEnd, chunk.StartLine, chunk.EndLine)
		}
		previousEnd = chunk.EndLine
	}
}

func assertChunkTokenLimit(t *testing.T, chunk Chunk) {
	t.Helper()
	if chunk.TokenCount <= 0 || chunk.TokenCount > defaultMaxChunkTokens {
		t.Fatalf("chunk token count = %d, limit = %d", chunk.TokenCount, defaultMaxChunkTokens)
	}
	exact, err := tokenizer.Count(chunk.Text)
	if err != nil {
		t.Fatal(err)
	}
	if exact != chunk.TokenCount {
		t.Fatalf("stored token count = %d, exact = %d", chunk.TokenCount, exact)
	}
}
