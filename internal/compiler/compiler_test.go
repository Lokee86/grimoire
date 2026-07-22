package compiler

import (
	"testing"

	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/retrieve"
)

func TestCompileFitsWholeChunksWithinBudget(t *testing.T) {
	candidates := []retrieve.Candidate{
		{Chunk: index.Chunk{Path: "first.go", StartLine: 1, EndLine: 10, EstimatedTokens: 7, Text: "first"}, Score: 10},
		{Chunk: index.Chunk{Path: "second.go", StartLine: 1, EndLine: 10, EstimatedTokens: 6, Text: "second"}, Score: 9},
		{Chunk: index.Chunk{Path: "third.go", StartLine: 1, EndLine: 10, EstimatedTokens: 3, Text: "third"}, Score: 8},
	}

	result := Compile("query", 10, index.FormatVersion, candidates)
	if len(result.Selections) != 2 {
		t.Fatalf("expected two selections, got %+v", result.Selections)
	}
	if result.Selections[0].Path != "first.go" || result.Selections[1].Path != "third.go" {
		t.Fatalf("unexpected selections: %+v", result.Selections)
	}
	if result.EstimatedTokens != 10 || result.OmittedForBudget != 1 {
		t.Fatalf("unexpected budget result: %+v", result)
	}
}
