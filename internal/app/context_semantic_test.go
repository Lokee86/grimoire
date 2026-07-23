package app

import (
	"math"
	"strings"
	"testing"

	"github.com/Lokee86/grimoire/internal/embedding"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/vectorstore"
)

func TestMergeSemanticHitsCombinesWindowsDeterministically(t *testing.T) {
	chunks := []index.Chunk{
		{ID: "shared", Path: "shared.go"},
		{ID: "first", Path: "first.go"},
		{ID: "second", Path: "second.go"},
	}
	plan := []embedding.QueryInput{
		{Label: "split window 1/2"},
		{Label: "split window 2/2"},
	}
	hits := [][]vectorstore.Hit{
		{{ID: "first", Score: 0.9}, {ID: "shared", Score: 0.8}},
		{{ID: "shared", Score: 0.95}, {ID: "second", Score: 0.85}},
	}
	candidates, err := mergeSemanticHits(chunks, plan, hits, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 3 || candidates[0].Chunk.ID != "shared" || candidates[0].Rank != 1 {
		t.Fatalf("unexpected merged ordering: %+v", candidates)
	}
	if len(candidates[0].Reasons) != 2 ||
		!strings.Contains(candidates[0].Reasons[0], "window 1/2") ||
		!strings.Contains(candidates[0].Reasons[1], "window 2/2") {
		t.Fatalf("missing window provenance: %+v", candidates[0].Reasons)
	}
	if len(candidates[0].ScoreDetails) != 2 || math.Abs(candidates[0].ScoreDetails[1].Value-0.95) > 0.000001 {
		t.Fatalf("missing semantic score attribution: %+v", candidates[0].ScoreDetails)
	}
}

func TestMergeSemanticHitsUsesMatchCountAsScoreTieBreak(t *testing.T) {
	chunks := []index.Chunk{{ID: "one"}, {ID: "many"}}
	plan := []embedding.QueryInput{{Label: "one"}, {Label: "two"}}
	hits := [][]vectorstore.Hit{
		{{ID: "one", Score: 0.8}, {ID: "many", Score: 0.8}},
		{{ID: "many", Score: 0.8}},
	}
	candidates, err := mergeSemanticHits(chunks, plan, hits, 10)
	if err != nil {
		t.Fatal(err)
	}
	if candidates[0].Chunk.ID != "many" {
		t.Fatalf("multi-window match did not win score tie: %+v", candidates)
	}
}
