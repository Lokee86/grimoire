package extraction

import (
	"testing"

	"github.com/Lokee86/grimoire/internal/retrieve"
)

func TestRefinePreservesCandidateBelowMinimumSpanCount(t *testing.T) {
	candidate := candidateForText(t, "single", numberedLines(48, map[int]string{
		24: "func ResolveDamage() {}",
	}))
	config := DefaultConfig()
	config.MinSpans = 2
	config.MinTokenSavings = 1
	config.MaxRetainedRatio = 0.95
	result, err := New(config, NewLineWindowDiscoverer(DefaultLineWindowConfig())).Refine(
		Request{Query: "ResolveDamage"}, []retrieve.Candidate{candidate},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].Chunk.ID != candidate.Chunk.ID {
		t.Fatalf("single-span candidate should remain unchanged: %+v", result)
	}
}
