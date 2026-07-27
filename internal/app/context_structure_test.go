package app

import (
	"slices"
	"testing"

	"github.com/Lokee86/grimoire/internal/evidence"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/queryshape"
	"github.com/Lokee86/grimoire/internal/retrieve"
	"github.com/Lokee86/grimoire/internal/structure"
)

func TestParseContextStructuralProvidersAllowsArcanaWithoutEmittingLexicon(t *testing.T) {
	emitLexicon, arcanaEnabled, err := parseContextStructuralProviders("arcana")
	if err != nil {
		t.Fatal(err)
	}
	if emitLexicon || !arcanaEnabled {
		t.Fatalf("emitLexicon=%v arcanaEnabled=%v", emitLexicon, arcanaEnabled)
	}
}

func TestParseContextStructuralProvidersRejectsUnknownProvider(t *testing.T) {
	if _, _, err := parseContextStructuralProviders("lexicon,unknown"); err == nil {
		t.Fatal("expected unsupported provider error")
	}
}

func TestParseStructuralScopeMode(t *testing.T) {
	for input, want := range map[string]structuralScopeMode{
		"": structuralScopeLexical, "lexical": structuralScopeLexical, "GLOBAL": structuralScopeGlobal,
	} {
		got, err := parseStructuralScopeMode(input)
		if err != nil {
			t.Fatalf("parse %q: %v", input, err)
		}
		if got != want {
			t.Fatalf("parse %q = %q, want %q", input, got, want)
		}
	}
	if _, err := parseStructuralScopeMode("mixed"); err == nil {
		t.Fatal("expected unsupported structural scope error")
	}
}

func TestParseArcanaSemanticMode(t *testing.T) {
	for input, want := range map[string]arcanaSemanticMode{
		"": arcanaSemanticAuto, "auto": arcanaSemanticAuto, "ON": arcanaSemanticOn, "off": arcanaSemanticOff,
	} {
		got, err := parseArcanaSemanticMode(input)
		if err != nil {
			t.Fatalf("parse %q: %v", input, err)
		}
		if got != want {
			t.Fatalf("parse %q = %q, want %q", input, got, want)
		}
	}
	if _, err := parseArcanaSemanticMode("sometimes"); err == nil {
		t.Fatal("expected unsupported Arcana semantic mode error")
	}
}

func TestArcanaSemanticAutoSkipsExplicitLexiconSeed(t *testing.T) {
	seeds := []structure.Node{{Name: "SearchDetailed", Path: "internal/lexiconfacts/facts.go"}}
	if shouldUseArcanaSemantic(arcanaSemanticAuto, "Trace SearchDetailed callers", seeds) {
		t.Fatal("auto mode should skip semantic expansion for an explicitly named compound identifier")
	}
	if shouldUseArcanaSemantic(arcanaSemanticAuto, "Inspect internal/lexiconfacts/facts.go", seeds) {
		t.Fatal("auto mode should skip semantic expansion for an explicitly named seed path")
	}
}

func TestArcanaSemanticAutoExpandsConceptualQuery(t *testing.T) {
	seeds := []structure.Node{{Name: "Search", Path: "internal/retrieve/search.go"}}
	if !shouldUseArcanaSemantic(arcanaSemanticAuto, "where is reconnect recovery handled?", seeds) {
		t.Fatal("auto mode should use semantic expansion when Lexicon has no explicit query anchor")
	}
	if shouldUseArcanaSemantic(arcanaSemanticOff, "where is reconnect recovery handled?", seeds) {
		t.Fatal("off mode should disable semantic expansion")
	}
	if !shouldUseArcanaSemantic(arcanaSemanticOn, "SearchDetailed", seeds) {
		t.Fatal("on mode should force semantic expansion")
	}
}

func TestMergeArcanaSeedsBalancesSemanticAndLexiconRecall(t *testing.T) {
	lexicon := []structure.Node{
		{Name: "LexiconOne", Path: "lexicon/one.go"},
		{Name: "Shared", Path: "shared.go"},
		{Name: "LexiconThree", Path: "lexicon/three.go"},
		{Name: "LexiconFour", Path: "lexicon/four.go"},
	}
	semantic := []structure.Node{
		{Name: "SemanticOne", Path: "semantic/one.go"},
		{Name: "Shared", Path: "shared.go"},
		{Name: "SemanticThree", Path: "semantic/three.go"},
		{Name: "SemanticFour", Path: "semantic/four.go"},
	}

	result := mergeArcanaSeeds(lexicon, semantic, 6)
	want := []string{
		"SemanticOne", "LexiconOne", "Shared",
		"SemanticThree", "LexiconThree", "SemanticFour",
	}
	if len(result) != len(want) {
		t.Fatalf("merged %d seeds, want %d: %+v", len(result), len(want), result)
	}
	for index := range want {
		if result[index].Name != want[index] {
			t.Fatalf("seed %d=%q, want %q: %+v", index, result[index].Name, want[index], result)
		}
	}
}

func TestAnnotateStructuralIntentIncludesArcanaSourceCandidates(t *testing.T) {
	result := structuralContextResult{ArcanaCandidates: []retrieve.Candidate{{
		Chunk: index.Chunk{ID: "candidate", Path: "service.go", StartLine: 3, EndLine: 8, TokenCount: 12},
		Rank:  2,
		Context: &evidence.Descriptor{
			Intents:  []evidence.Intent{evidence.IntentCallChain},
			Roles:    []evidence.Role{evidence.RoleSupporting},
			GroupIDs: []string{"graph-group"},
		},
	}}}
	planned := queryshape.RetrievalIntent{
		Intent: evidence.IntentArchitecture, FacetID: "facet-architecture", Query: "what depends on service",
	}

	annotated := annotateStructuralIntent(result, planned)
	candidate := annotated.ArcanaCandidates[0]
	if candidate.Context == nil ||
		!slices.Contains(candidate.Context.Intents, evidence.IntentCallChain) ||
		!slices.Contains(candidate.Context.Intents, evidence.IntentArchitecture) ||
		!slices.Contains(candidate.Context.GroupIDs, "graph-group") ||
		!slices.Contains(candidate.Context.GroupIDs, "facet-architecture") ||
		candidate.Context.FacetRanks["facet-architecture"] != 2 {
		t.Fatalf("Arcana candidate intent metadata = %+v", candidate.Context)
	}
	if !slices.Contains(candidate.Reasons, "retrieval intent architecture") {
		t.Fatalf("Arcana candidate intent reason missing: %+v", candidate.Reasons)
	}
}

func TestInterleaveStructuralEvidencePreservesProviderRanks(t *testing.T) {
	lexicon := []structure.Evidence{
		{Provider: "lexicon", Rank: 1},
		{Provider: "lexicon", Rank: 2},
		{Provider: "lexicon", Rank: 3},
	}
	arcana := []structure.Evidence{
		{Provider: "arcana", Rank: 1},
		{Provider: "arcana", Rank: 2},
	}
	result := interleaveStructuralEvidence(lexicon, arcana)
	wantProviders := []string{"lexicon", "arcana", "lexicon", "arcana", "lexicon"}
	wantRanks := []int{1, 1, 2, 2, 3}
	if len(result) != len(wantProviders) {
		t.Fatalf("interleaved %d items, want %d", len(result), len(wantProviders))
	}
	for index := range result {
		if result[index].Provider != wantProviders[index] || result[index].Rank != wantRanks[index] {
			t.Fatalf("unexpected item %d: %+v", index, result[index])
		}
	}
}
