package retrieve

import (
	"testing"

	"github.com/Lokee86/grimoire/internal/index"
)

func TestSearchManyKeepsIntentTermsIsolated(t *testing.T) {
	snapshot := index.Snapshot{Version: index.FormatVersion, Files: []index.FileRecord{
		{Path: "alpha.go", Chunks: []index.Chunk{{ID: "alpha", Path: "alpha.go", StartLine: 1, Text: "package alpha\nfunc AlphaResolver() {}", TokenCount: 4}}},
		{Path: "beta.go", Chunks: []index.Chunk{{ID: "beta", Path: "beta.go", StartLine: 1, Text: "package beta\nfunc BetaCompiler() {}", TokenCount: 4}}},
	}}

	results := SearchMany(snapshot, []string{"alpha resolver", "beta compiler"}, 10)
	if len(results) != 2 {
		t.Fatalf("result groups = %d, want 2", len(results))
	}
	if len(results[0]) != 1 || results[0][0].Chunk.Path != "alpha.go" {
		t.Fatalf("alpha query leaked terms or lost its match: %+v", results[0])
	}
	if len(results[1]) != 1 || results[1][0].Chunk.Path != "beta.go" {
		t.Fatalf("beta query leaked terms or lost its match: %+v", results[1])
	}
}

func TestSearchManyPreservesEmptyQueryPosition(t *testing.T) {
	snapshot := index.Snapshot{Version: index.FormatVersion, Files: []index.FileRecord{{
		Path: "resolver.go", Chunks: []index.Chunk{{ID: "resolver", Path: "resolver.go", StartLine: 1, Text: "resolver", TokenCount: 1}},
	}}}

	results := SearchMany(snapshot, []string{"", "resolver"}, 10)
	if len(results) != 2 || results[0] != nil || len(results[1]) != 1 {
		t.Fatalf("unexpected positional results: %+v", results)
	}
}
