package app

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Lokee86/grimoire/internal/evidence"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/queryshape"
	"github.com/Lokee86/grimoire/internal/retrieve"
	"github.com/Lokee86/grimoire/internal/tokenizer"
)

func TestExtractContextCandidatesUsesLanguageSpansAndRequiredLinks(t *testing.T) {
	lines := make([]string, 48)
	for index := range lines {
		lines[index] = fmt.Sprintf("// filler %02d with enough context to consume tokens", index+1)
	}
	lines[8] = "func ResolveDamage() {"
	lines[9] = "}"
	lines[37] = "func ApplyShieldGate() {"
	lines[38] = "}"
	text := strings.Join(lines, "\n")
	tokens, err := tokenizer.Count(text)
	if err != nil {
		t.Fatal(err)
	}
	facet := "facet:damage"
	candidate := retrieve.Candidate{
		Chunk: index.Chunk{
			ID: "damage-parent", Path: "internal/damage/resolve.go",
			StartLine: 1, EndLine: len(lines), TokenCount: tokens, Text: text,
		},
		Source: "lexical", Rank: 1,
		Context: &evidence.Descriptor{
			Identity:   evidence.RangeIdentity("internal/damage/resolve.go", 1, len(lines)),
			GroupIDs:   []string{facet},
			FacetRanks: map[string]int{facet: 1},
		},
	}
	result, err := extractContextCandidates(
		"Explain ResolveDamage and ApplyShieldGate",
		[]queryshape.RetrievalIntent{{
			FacetID: facet, Intent: evidence.IntentMechanism,
			Query: "ResolveDamage ApplyShieldGate", Weight: 1,
		}},
		[]retrieve.Candidate{candidate},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Fatalf("extracted candidates = %d, want 2", len(result))
	}
	if result[0].Context == nil || result[1].Context == nil {
		t.Fatalf("extracted candidates lost context: %+v", result)
	}
	linked := false
	for _, link := range result[0].Context.Links {
		if link.Identity == result[1].Context.Identity && link.Required {
			linked = true
		}
	}
	if !linked {
		t.Fatalf("extracted candidates were not linked: %+v", result[0].Context.Links)
	}
}
