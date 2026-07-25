package diffcontext

import (
	"reflect"
	"testing"

	"github.com/Lokee86/grimoire/internal/index"
)

func TestCandidatesMapOverlappingChunksAndDeduplicate(t *testing.T) {
	snapshot := index.Snapshot{Files: []index.FileRecord{
		{Path: "internal/file.go", Chunks: []index.Chunk{
			{ID: "first", Path: "internal/file.go", StartLine: 1, EndLine: 10, TokenCount: 10},
			{ID: "second", Path: "internal/file.go", StartLine: 11, EndLine: 20, TokenCount: 10},
			{ID: "third", Path: "other.go", StartLine: 1, EndLine: 20, TokenCount: 10},
		}},
	}}
	changes := []Change{
		{Path: "internal\\file.go", StartLine: 8, EndLine: 12},
		{Path: "internal/file.go", StartLine: 9, EndLine: 11},
		{Path: "missing.go", StartLine: 1, EndLine: 2},
	}

	got := Candidates(snapshot, changes)
	if len(got) != 2 {
		t.Fatalf("got %d candidates, want 2: %#v", len(got), got)
	}
	if got[0].Chunk.ID != "first" || got[1].Chunk.ID != "second" {
		t.Fatalf("unexpected candidate order: %#v", got)
	}
	for index, candidate := range got {
		if candidate.Source != "git-diff" || candidate.Rank != index+1 {
			t.Fatalf("unexpected provenance/rank: %#v", candidate)
		}
		if len(candidate.Reasons) != 2 {
			t.Fatalf("overlapping changes were not deduplicated with inspectable reasons: %#v", candidate.Reasons)
		}
	}
}

func TestCandidatesDeterministicRankAndEvidenceBound(t *testing.T) {
	snapshot := index.Snapshot{Files: []index.FileRecord{
		{Path: "b.go", Chunks: []index.Chunk{{ID: "b", Path: "b.go", StartLine: 1, EndLine: 4}}},
		{Path: "a.go", Chunks: []index.Chunk{{ID: "a", Path: "a.go", StartLine: 1, EndLine: 4}}},
	}}
	changes := []Change{{Path: "b.go", StartLine: 1, EndLine: 1}, {Path: "a.go", StartLine: 1, EndLine: 1}}
	left := Candidates(snapshot, changes)
	right := Candidates(snapshot, []Change{changes[1], changes[0]})
	if !reflect.DeepEqual(left, right) {
		t.Fatalf("candidate ordering is not deterministic:\nleft %#v\nright %#v", left, right)
	}
	if left[0].Chunk.ID != "a" || left[0].Rank != 1 || left[1].Rank != 2 || left[0].Score <= left[1].Score {
		t.Fatalf("unexpected deterministic ranks/scores: %#v", left)
	}

	evidence := Evidence([]Change{
		{Path: "a.go", StartLine: 1, EndLine: 1},
		{Path: "b.go", StartLine: 2, EndLine: 3, Deleted: true},
	}, 1)
	if len(evidence) != 1 || evidence[0].Provider != "git-diff" || evidence[0].Kind != "change" || evidence[0].Node == nil {
		t.Fatalf("unexpected bounded change evidence: %#v", evidence)
	}
}
