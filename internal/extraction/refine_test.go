package extraction

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Lokee86/grimoire/internal/evidence"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/retrieve"
	"github.com/Lokee86/grimoire/internal/tokenizer"
)

func TestExtractorWithoutDiscoverersPreservesCandidates(t *testing.T) {
	candidate := candidateForText(t, "passthrough", numberedLines(48, map[int]string{
		24: "func CompileAdaptiveWithEvidence() {}",
	}))
	result, err := New(DefaultConfig()).Refine(
		Request{Query: "CompileAdaptiveWithEvidence"}, []retrieve.Candidate{candidate},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].Chunk.ID != candidate.Chunk.ID || result[0].Chunk.Text != candidate.Chunk.Text {
		t.Fatalf("extractor without discoverers changed candidates: %+v", result)
	}
}

func TestRefineExtractsFocusedSpanAndPreservesCandidateMetadata(t *testing.T) {
	text := numberedLines(48, map[int]string{
		31: "func CompileAdaptiveWithEvidence() error {",
		32: "    return compileWithEvidence()",
		33: "}",
	})
	tokens, err := tokenizer.Count(text)
	if err != nil {
		t.Fatal(err)
	}
	candidate := retrieve.Candidate{
		Chunk: index.Chunk{ID: "parent", Path: "internal/compiler/compiler.go", StartLine: 101, EndLine: 148, TokenCount: tokens, Text: text},
		Score: 42, Source: "lexical", Rank: 3,
		Reasons: []string{"BM25 content matches compileadaptivewithevidence"},
		Context: &evidence.Descriptor{
			Identity:        evidence.RangeIdentity("internal/compiler/compiler.go", 101, 148),
			GroupIDs:        []string{"facet:compiler"},
			FacetRanks:      map[string]int{"facet:compiler": 1},
			EstimatedTokens: tokens,
			RedundancyKey:   evidence.RangeIdentity("internal/compiler/compiler.go", 101, 148),
		},
	}

	config := DefaultConfig()
	config.MinTokenSavings = 1
	config.MaxRetainedRatio = 0.95
	result, err := New(config, NewLineWindowDiscoverer(DefaultLineWindowConfig())).Refine(Request{
		Query: "explain the compiler ownership flow",
		FacetQueries: map[string]string{
			"facet:compiler": "CompileAdaptiveWithEvidence compileWithEvidence",
		},
	}, []retrieve.Candidate{candidate})
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("refined candidates = %d, want 1", len(result))
	}
	got := result[0]
	if got.Chunk.StartLine <= candidate.Chunk.StartLine || got.Chunk.EndLine >= candidate.Chunk.EndLine {
		t.Fatalf("range was not narrowed: %+v", got.Chunk)
	}
	if !strings.Contains(got.Chunk.Text, "CompileAdaptiveWithEvidence") {
		t.Fatalf("focused span lost target declaration:\n%s", got.Chunk.Text)
	}
	if got.Score != candidate.Score || got.Source != candidate.Source || got.Rank != candidate.Rank {
		t.Fatalf("retrieval metadata changed: %+v", got)
	}
	if got.Context == nil || got.Context.FacetRanks["facet:compiler"] != 1 {
		t.Fatalf("facet metadata was not preserved: %+v", got.Context)
	}
	if got.Context.Identity == candidate.Context.Identity || got.Context.EstimatedTokens != got.Chunk.TokenCount {
		t.Fatalf("range descriptor was not refreshed: %+v", got.Context)
	}
	foundParent := false
	for _, link := range got.Context.Links {
		if link.Identity == candidate.Context.Identity && link.Relation == "extracted_from" {
			foundParent = true
		}
	}
	if !foundParent {
		t.Fatalf("extracted range lost parent provenance: %+v", got.Context.Links)
	}
}

func TestRefineCountsMissingParentTokenMetadata(t *testing.T) {
	candidate := candidateForText(t, "missing-tokens", numberedLines(48, map[int]string{
		25: "func ResolveDamage() {}",
	}))
	candidate.Chunk.TokenCount = 0
	config := DefaultConfig()
	config.MinTokenSavings = 1
	config.MaxRetainedRatio = 0.95
	result, err := New(config, NewLineWindowDiscoverer(DefaultLineWindowConfig())).Refine(
		Request{Query: "ResolveDamage"}, []retrieve.Candidate{candidate},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].Chunk.ID == candidate.Chunk.ID || result[0].Chunk.TokenCount <= 0 {
		t.Fatalf("candidate with missing token metadata was not refined: %+v", result)
	}
	if result[0].Context == nil || result[0].Context.Identity == "" {
		t.Fatalf("extracted candidate did not receive an identity: %+v", result[0].Context)
	}
}

func TestRefinePreservesOriginalChunkWhenNoQueryTermMatches(t *testing.T) {
	candidate := candidateForText(t, "alpha", numberedLines(48, nil))
	result, err := New(DefaultConfig(), NewLineWindowDiscoverer(DefaultLineWindowConfig())).Refine(
		Request{Query: "CompileAdaptiveWithEvidence"}, []retrieve.Candidate{candidate},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].Chunk.ID != candidate.Chunk.ID || result[0].Chunk.Text != candidate.Chunk.Text {
		t.Fatalf("unmatched candidate changed: %+v", result)
	}
}

func TestRefineCanEmitTwoSeparatedFocusedSpans(t *testing.T) {
	text := numberedLines(80, map[int]string{
		10: "func ResolveDamage() {}",
		67: "func ApplyShieldGate() {}",
	})
	candidate := candidateForText(t, "damage", text)
	config := DefaultConfig()
	config.MaxSpans = 2
	config.MinTokenSavings = 1
	config.MaxRetainedRatio = 0.95
	lineConfig := DefaultLineWindowConfig()
	lineConfig.ContextBefore = 3
	lineConfig.ContextAfter = 4
	lineConfig.MaxSpanLines = 8
	result, err := New(config, NewLineWindowDiscoverer(lineConfig)).Refine(
		Request{Query: "ResolveDamage ApplyShieldGate"},
		[]retrieve.Candidate{candidate},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Fatalf("refined candidates = %d, want 2", len(result))
	}
	if !strings.Contains(result[0].Chunk.Text, "ResolveDamage") || !strings.Contains(result[1].Chunk.Text, "ApplyShieldGate") {
		t.Fatalf("separated evidence was not retained: %#v", result)
	}
	if result[0].Chunk.EndLine >= result[1].Chunk.StartLine {
		t.Fatalf("spans overlap: %+v %+v", result[0].Chunk, result[1].Chunk)
	}
	linked := false
	for _, link := range result[0].Context.Links {
		if link.Identity == result[1].Context.Identity && link.Relation == "extracted_companion" && link.Required {
			linked = true
		}
	}
	if !linked {
		t.Fatalf("separated spans were not linked as a required group: %+v", result[0].Context.Links)
	}
}

func TestRefineKeepsDiscovererPriorityBeforeSourceOrdering(t *testing.T) {
	candidate := candidateForText(t, "priority", numberedLines(90, map[int]string{
		8:  "func WeakEarlyMatch() {}",
		45: "func StrongMiddleMatch() {}",
		80: "func StrongLateMatch() {}",
	}))
	config := DefaultConfig()
	config.MaxSpans = 2
	config.MinTokenSavings = 1
	config.MaxRetainedRatio = 0.95
	discoverer := fixedDiscoverer{spans: []Span{
		{StartLine: 76, EndLine: 83, Reason: "strongest"},
		{StartLine: 41, EndLine: 48, Reason: "second strongest"},
		{StartLine: 4, EndLine: 11, Reason: "weakest"},
	}}
	result, err := New(config, discoverer).Refine(
		Request{Query: "StrongMiddleMatch StrongLateMatch"}, []retrieve.Candidate{candidate},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Fatalf("refined candidates = %d, want 2", len(result))
	}
	if !strings.Contains(result[0].Chunk.Text, "StrongMiddleMatch") || !strings.Contains(result[1].Chunk.Text, "StrongLateMatch") {
		t.Fatalf("extractor retained the wrong priority spans: %#v", result)
	}
	if strings.Contains(result[0].Chunk.Text, "WeakEarlyMatch") || strings.Contains(result[1].Chunk.Text, "WeakEarlyMatch") {
		t.Fatalf("weak early span displaced a stronger span: %#v", result)
	}
}

func TestNormalizeSpansMergesTransitiveOverlap(t *testing.T) {
	spans := normalizeSpans([]Span{
		{StartLine: 0, EndLine: 2, Reason: "first"},
		{StartLine: 6, EndLine: 8, Reason: "second"},
		{StartLine: 2, EndLine: 6, Reason: "bridge"},
	}, 20)
	if len(spans) != 1 || spans[0].StartLine != 0 || spans[0].EndLine != 8 {
		t.Fatalf("transitive overlap was not merged: %+v", spans)
	}
}

func TestRefineRequiresBothSizeThresholds(t *testing.T) {
	text := strings.Join([]string{
		"func ResolveDamage() {}",
		strings.Repeat("// compact context\n", 30),
	}, "\n")
	candidate := candidateForText(t, "compact", text)
	config := DefaultConfig()
	config.MinChunkLines = 24
	config.MinChunkTokens = candidate.Chunk.TokenCount + 1
	config.MinTokenSavings = 1
	result, err := New(config, NewLineWindowDiscoverer(DefaultLineWindowConfig())).Refine(
		Request{Query: "ResolveDamage"}, []retrieve.Candidate{candidate},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].Chunk.ID != candidate.Chunk.ID {
		t.Fatalf("candidate below one size threshold should remain unchanged: %+v", result)
	}
}

func TestRefineFallsBackWhenSavingsAreTooSmall(t *testing.T) {
	text := numberedLines(30, map[int]string{15: "func ResolveDamage() {}"})
	candidate := candidateForText(t, "small", text)
	config := DefaultConfig()
	config.MinChunkLines = 1
	config.MinChunkTokens = 1
	config.MinTokenSavings = candidate.Chunk.TokenCount
	result, err := New(config, NewLineWindowDiscoverer(DefaultLineWindowConfig())).Refine(
		Request{Query: "ResolveDamage"}, []retrieve.Candidate{candidate},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].Chunk.ID != candidate.Chunk.ID {
		t.Fatalf("low-savings refinement should preserve original: %+v", result)
	}
}

type fixedDiscoverer struct {
	spans []Span
}

func (discoverer fixedDiscoverer) Discover(DiscoveryRequest) ([]Span, error) {
	return append([]Span(nil), discoverer.spans...), nil
}

func candidateForText(t *testing.T, id, text string) retrieve.Candidate {
	t.Helper()
	tokens, err := tokenizer.Count(text)
	if err != nil {
		t.Fatal(err)
	}
	return retrieve.Candidate{
		Chunk:  index.Chunk{ID: id, Path: id + ".go", StartLine: 1, EndLine: strings.Count(text, "\n") + 1, TokenCount: tokens, Text: text},
		Source: "lexical", Rank: 1,
	}
}

func numberedLines(count int, replacements map[int]string) string {
	lines := make([]string, count)
	for index := range lines {
		lineNumber := index + 1
		if replacement, exists := replacements[lineNumber]; exists {
			lines[index] = replacement
			continue
		}
		lines[index] = fmt.Sprintf("// filler line %02d with enough repeated context to consume tokens", lineNumber)
	}
	return strings.Join(lines, "\n")
}
