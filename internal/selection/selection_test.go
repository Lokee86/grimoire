package selection

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/retrieve"
)

func TestCurateRemovesDuplicateIDsAndOverlappingRanges(t *testing.T) {
	candidates := []retrieve.Candidate{
		candidate("internal/alpha.go", "a", 1, 10, 1, "vector"),
		candidate("internal/alpha.go", "overlap", 5, 12, 2, "vector"),
		candidate("internal/alpha.go", "b", 13, 20, 3, "vector"),
		candidate("other.go", "a", 1, 2, 4, "lexical"),
	}

	curated := Curate(index.Snapshot{}, candidates)
	if got := chunkIDs(curated); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("unexpected curated IDs: %v", got)
	}
	if curated[0].Source != "vector" || curated[0].Rank != 1 {
		t.Fatalf("primary provenance changed: %+v", curated[0])
	}
}

func TestCurateSoftlyDiversifiesFilesAndSubsystems(t *testing.T) {
	candidates := []retrieve.Candidate{
		candidate("internal/alpha/one.go", "one", 1, 2, 1, "vector"),
		candidate("internal/alpha/two.go", "two", 1, 2, 2, "lexical"),
		candidate("internal/beta/three.go", "three", 1, 2, 3, "vector"),
		candidate("internal/gamma/four.go", "four", 1, 2, 4, "vector"),
	}

	curated := Curate(index.Snapshot{}, candidates)
	if got := chunkIDs(curated); !reflect.DeepEqual(got, []string{"one", "three", "four", "two"}) {
		t.Fatalf("diversity did not reorder softly: %v", got)
	}
	if len(curated) != len(candidates) {
		t.Fatalf("diversity discarded unique candidates: %d", len(curated))
	}
}

func TestCurateDoesNotCompareProviderRanks(t *testing.T) {
	candidates := []retrieve.Candidate{
		candidate("internal/exact/result.go", "exact", 1, 2, 99, "exact"),
		candidate("internal/vector/result.go", "vector", 1, 2, 1, "vector"),
		candidate("internal/lexical/result.go", "lexical", 1, 2, 2, "lexical"),
	}

	curated := Curate(index.Snapshot{}, candidates)
	if got := chunkIDs(curated); !reflect.DeepEqual(got, []string{"exact", "vector", "lexical"}) {
		t.Fatalf("provider rank changed merged ordering: %v", got)
	}
	if curated[0].Source != "exact" || curated[0].Rank != 99 {
		t.Fatalf("primary provenance changed: %+v", curated[0])
	}
}

func TestCurateBoundsThreeNeighborAnchorsBeforeRemainingPrimaries(t *testing.T) {
	path := "internal/alpha.go"
	primaryIDs := []string{"p1", "p2", "p3", "p4", "p5", "p6"}
	neighborIDs := []string{"n1", "n2", "n3", "n4", "n5", "n6"}
	chunks := make([]index.Chunk, 0, len(primaryIDs)*2)
	candidates := make([]retrieve.Candidate, 0, len(primaryIDs))
	for index := range primaryIDs {
		primaryLine := index*2 + 1
		chunks = append(chunks, chunk(primaryIDs[index], path, primaryLine, primaryLine))
		chunks = append(chunks, chunk(neighborIDs[index], path, primaryLine+1, primaryLine+1))
		candidates = append(candidates, candidate(path, primaryIDs[index], primaryLine, primaryLine, index+1, "vector"))
	}
	snapshot := index.Snapshot{Files: []index.FileRecord{{Path: path, Chunks: chunks}}}

	curated := Curate(snapshot, candidates)
	want := []string{"p1", "p2", "p3", "n1", "n2", "n3", "p4", "p5", "p6"}
	if got := chunkIDs(curated); !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected bounded ordering: got %v want %v", got, want)
	}
}

func TestCuratePromotesExistingRetrievedNeighbor(t *testing.T) {
	snapshot := index.Snapshot{Files: []index.FileRecord{{
		Path: "internal/alpha.go",
		Chunks: []index.Chunk{
			chunk("primary", "internal/alpha.go", 1, 4),
			chunk("required-neighbor", "internal/alpha.go", 5, 8),
		},
	}}}
	candidates := []retrieve.Candidate{
		candidate("internal/alpha.go", "primary", 1, 4, 1, "vector"),
		candidate("internal/beta.go", "beta", 1, 4, 2, "vector"),
		candidate("internal/gamma.go", "gamma", 1, 4, 3, "vector"),
		candidate("internal/delta.go", "delta", 1, 4, 4, "vector"),
		candidate("internal/epsilon.go", "epsilon", 1, 4, 5, "vector"),
		candidate("internal/zeta.go", "zeta", 1, 4, 6, "vector"),
		candidate("internal/alpha.go", "required-neighbor", 5, 8, 7, "vector"),
	}

	curated := Curate(snapshot, candidates)
	if got := chunkIDs(curated); !reflect.DeepEqual(got, []string{
		"primary", "beta", "gamma", "required-neighbor", "delta", "epsilon", "zeta",
	}) {
		t.Fatalf("retrieved neighbor was not promoted: %v", got)
	}
	if curated[3].Source != "vector" || curated[3].Rank != 7 {
		t.Fatalf("promoted neighbor lost provider provenance: %+v", curated[3])
	}
}

func TestCurateAddsPreparedNeighborsWithReasons(t *testing.T) {
	snapshot := index.Snapshot{Files: []index.FileRecord{{
		Path: "internal/alpha.go",
		Chunks: []index.Chunk{
			chunk("before", "internal/alpha.go", 1, 4),
			chunk("primary", "internal/alpha.go", 5, 8),
			chunk("after", "internal/alpha.go", 9, 12),
		},
	}}}
	primary := candidate("internal/alpha.go", "primary", 5, 8, 1, "vector")
	nextPrimary := candidate("internal/alpha.go", "after", 9, 12, 2, "lexical")

	curated := Curate(snapshot, []retrieve.Candidate{primary, nextPrimary})
	if len(curated) != 3 {
		t.Fatalf("expected two primaries plus one deduplicated neighbor, got %d", len(curated))
	}
	if curated[0].Source != primary.Source || curated[0].Rank != primary.Rank ||
		!reflect.DeepEqual(curated[0].Reasons, primary.Reasons) {
		t.Fatalf("primary provenance changed: %+v", curated[0])
	}
	if curated[1].Source != nextPrimary.Source || curated[1].Rank != nextPrimary.Rank {
		t.Fatalf("second primary provenance changed: %+v", curated[1])
	}
	neighbor := curated[2]
	if neighbor.Source != "adjacent" || len(neighbor.Reasons) != 1 {
		t.Fatalf("unexpected adjacent provenance: %+v", neighbor)
	}
	if !strings.Contains(neighbor.Reasons[0], "previous prepared chunk") {
		t.Fatalf("adjacent reason is not inspectable: %+v", neighbor.Reasons)
	}
	if curated[0].Chunk.ID != "primary" || curated[1].Chunk.ID != "after" || neighbor.Chunk.ID != "before" {
		t.Fatalf("neighbors were not ordered around primary: %v", chunkIDs(curated))
	}
}

func TestDefaultConfigUsesCalibratedValues(t *testing.T) {
	config := DefaultConfig()
	if config.FileRepeatPenalty != 10 || config.SubsystemRepeatPenalty != 18 || config.AdjacentPrimaryLimit != 3 {
		t.Fatalf("unexpected production curation defaults: %+v", config)
	}
}

func TestCurateWithConfigAppliesStrongerSubsystemPressure(t *testing.T) {
	candidates := []retrieve.Candidate{
		candidate("internal/alpha/one.go", "one", 1, 2, 1, "vector"),
		candidate("internal/alpha/two.go", "two", 1, 2, 2, "vector"),
		candidate("internal/alpha/three.go", "three", 1, 2, 3, "vector"),
		candidate("internal/beta/four.go", "four", 1, 2, 4, "vector"),
	}
	config := Config{SubsystemRepeatPenalty: 10}

	curated := CurateWithConfig(index.Snapshot{}, candidates, config)
	if got := chunkIDs(curated); !reflect.DeepEqual(got, []string{"one", "four", "two", "three"}) {
		t.Fatalf("strong subsystem pressure did not promote a new subsystem: %v", got)
	}
}

func TestCurateWithConfigCanDisableAdjacentPromotion(t *testing.T) {
	snapshot := index.Snapshot{Files: []index.FileRecord{{
		Path: "internal/alpha.go",
		Chunks: []index.Chunk{
			chunk("before", "internal/alpha.go", 1, 4),
			chunk("primary", "internal/alpha.go", 5, 8),
			chunk("after", "internal/alpha.go", 9, 12),
		},
	}}}
	config := DefaultConfig()
	config.AdjacentPrimaryLimit = 0

	curated := CurateWithConfig(snapshot, []retrieve.Candidate{
		candidate("internal/alpha.go", "primary", 5, 8, 1, "vector"),
	}, config)
	if got := chunkIDs(curated); !reflect.DeepEqual(got, []string{"primary"}) {
		t.Fatalf("adjacent promotion was not disabled: %v", got)
	}
}

func TestCurateIsDeterministic(t *testing.T) {
	snapshot := index.Snapshot{Files: []index.FileRecord{{
		Path: "internal/alpha.go",
		Chunks: []index.Chunk{
			chunk("a", "internal/alpha.go", 1, 2),
			chunk("b", "internal/alpha.go", 3, 4),
		},
	}}}
	candidates := []retrieve.Candidate{
		candidate("internal/alpha.go", "b", 3, 4, 2, "vector"),
		candidate("internal/alpha.go", "a", 1, 2, 1, "lexical"),
	}

	first := Curate(snapshot, candidates)
	second := Curate(snapshot, candidates)
	if got := chunkIDs(first); !reflect.DeepEqual(got, []string{"b", "a"}) {
		t.Fatalf("incoming order was not preserved: %v", got)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("curation was not deterministic:\nfirst=%+v\nsecond=%+v", first, second)
	}
}

func candidate(path, id string, start, end, rank int, source string) retrieve.Candidate {
	return retrieve.Candidate{
		Chunk: index.Chunk{ID: id, Path: path, StartLine: start, EndLine: end},
		Score: float64(100 - rank), Source: source, Rank: rank,
		Reasons: []string{"provider reason"},
	}
}

func chunk(pathID, path string, start, end int) index.Chunk {
	return index.Chunk{ID: pathID, Path: path, StartLine: start, EndLine: end}
}

func chunkIDs(candidates []retrieve.Candidate) []string {
	ids := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		ids = append(ids, candidate.Chunk.ID)
	}
	return ids
}
