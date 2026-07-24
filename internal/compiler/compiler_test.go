package compiler

import (
	"strings"
	"testing"

	"github.com/Lokee86/grimoire/internal/assembly"
	"github.com/Lokee86/grimoire/internal/evidence"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/queryshape"
	"github.com/Lokee86/grimoire/internal/retrieve"
	"github.com/Lokee86/grimoire/internal/structure"
	"github.com/Lokee86/grimoire/internal/tokenizer"
)

func TestCompileEnforcesSerializedPackageBudget(t *testing.T) {
	candidates := []retrieve.Candidate{
		candidate(t, "first.go", strings.Repeat("first value ", 80), 10),
		candidate(t, "second.go", strings.Repeat("second value ", 80), 9),
		candidate(t, "third.go", strings.Repeat("third value ", 80), 8),
	}

	full, err := Compile("query", 10_000, index.FormatVersion, tokenizer.Name, []string{"test"}, candidates)
	if err != nil {
		t.Fatal(err)
	}
	if len(full.Selections) != len(candidates) {
		t.Fatalf("expected all selections, got %+v", full.Selections)
	}
	if len(full.RetrievalSources) != 1 || full.RetrievalSources[0] != "test" {
		t.Fatalf("unexpected retrieval sources: %+v", full.RetrievalSources)
	}
	if full.Selections[0].RetrievalSource != "test" || full.Selections[0].RetrievalRank != 1 {
		t.Fatalf("unexpected selection provenance: %+v", full.Selections[0])
	}
	if full.FacetProtection || full.FacetCompanionDepth != 0 || full.FacetsAvailable != 0 {
		t.Fatalf("fixed package fabricated adaptive fitting metadata: %+v", full)
	}
	assertExactPackageCount(t, full)

	tightBudget := full.TokenCount - 20
	tight, err := Compile("query", tightBudget, index.FormatVersion, tokenizer.Name, []string{"test"}, candidates)
	if err != nil {
		t.Fatal(err)
	}
	if len(tight.Selections) >= len(candidates) {
		t.Fatalf("expected at least one omitted selection, got %+v", tight.Selections)
	}
	if tight.OmittedForBudget == 0 {
		t.Fatalf("expected budget omission, got %+v", tight)
	}
	if tight.TokenCount > tightBudget {
		t.Fatalf("package used %d tokens with budget %d", tight.TokenCount, tightBudget)
	}
	assertExactPackageCount(t, tight)
}

func TestCompileWithEvidenceEmitsStructuralFactsBeforeSourceSelections(t *testing.T) {
	node := structure.Node{Name: "ResolveDamage", Kind: "function", Path: "internal/damage.go"}
	evidence := []structure.Evidence{{
		Provider: "lexicon", Kind: "symbol", Rank: 1, Node: &node,
		Relationships: []structure.Relationship{{
			Direction: "outgoing", Relation: "calls", Certainty: "definite",
			Node: structure.Node{Name: "ApplyShield", Path: "internal/shield.go"},
		}},
	}}
	candidates := []retrieve.Candidate{
		candidate(t, "internal/damage.go", strings.Repeat("damage source value ", 120), 10),
	}

	providerState := []structure.ProviderState{{
		Provider: "lexicon", Snapshot: "sha256:abc",
	}}
	full, err := CompileWithEvidence(
		"trace damage", 10_000, index.FormatVersion, tokenizer.Name,
		[]string{"vector"}, providerState, evidence, candidates,
	)
	if err != nil {
		t.Fatal(err)
	}
	if full.Version != PackageVersion || len(full.StructuralEvidence) != 1 {
		t.Fatalf("structural evidence missing from package: %+v", full)
	}
	if len(full.StructuralSources) != 1 || full.StructuralSources[0] != "lexicon" {
		t.Fatalf("unexpected structural sources: %+v", full.StructuralSources)
	}
	if len(full.StructuralState) != 1 || full.StructuralState[0] != providerState[0] {
		t.Fatalf("unexpected structural state: %+v", full.StructuralState)
	}

	retainedWithoutSource := false
	for budget := 1; budget < full.TokenCount; budget++ {
		pkg, compileErr := CompileWithEvidence(
			"trace damage", budget, index.FormatVersion, tokenizer.Name,
			[]string{"vector"}, providerState, evidence, candidates,
		)
		if compileErr != nil {
			continue
		}
		if len(pkg.StructuralEvidence) == 1 && len(pkg.Selections) == 0 {
			retainedWithoutSource = true
			if pkg.OmittedForBudget != 1 {
				t.Fatalf("source omission not recorded: %+v", pkg)
			}
			assertExactPackageCount(t, pkg)
			break
		}
	}
	if !retainedWithoutSource {
		t.Fatal("no budget retained structural evidence before the larger source selection")
	}
}

func TestCompileAdaptiveRetainsAssemblyDecision(t *testing.T) {
	decision := assembly.Decision{
		Scope: queryshape.ScopeFocused, CandidatesConsidered: 3,
		CandidatesSelected: 2, StopReason: "focused evidence coverage satisfied",
	}
	pkg, err := CompileAdaptiveWithEvidence(
		"where", 1000, index.FormatVersion, tokenizer.Name,
		[]string{"exact"}, nil, nil, decision,
		[]retrieve.Candidate{candidate(t, "example.go", "package example", 10)},
	)
	if err != nil {
		t.Fatal(err)
	}
	if pkg.Assembly == nil || pkg.Assembly.Scope != decision.Scope ||
		pkg.Assembly.CandidatesConsidered != decision.CandidatesConsidered ||
		pkg.Assembly.CandidatesSelected != decision.CandidatesSelected ||
		pkg.Assembly.StopReason != decision.StopReason {
		t.Fatalf("assembly decision missing: %+v", pkg.Assembly)
	}
	assertExactPackageCount(t, pkg)
}

func TestCompileAdaptiveProtectsDistinctFacetsDuringBudgetFitting(t *testing.T) {
	candidates := []retrieve.Candidate{
		facetCandidate(t, "a/one.go", strings.Repeat("alpha owner ", 90), 10, evidence.IntentDirectLocation, "facet:a"),
		facetCandidate(t, "a/two.go", strings.Repeat("alpha helper ", 90), 9, evidence.IntentDirectLocation, "facet:a"),
		facetCandidate(t, "a/three.go", strings.Repeat("alpha context ", 90), 8, evidence.IntentDirectLocation, "facet:a"),
		facetCandidate(t, "b/chain.go", strings.Repeat("beta call chain ", 90), 7, evidence.IntentCallChain, "facet:b"),
	}
	decision := assembly.Decision{CoverageAware: true, FacetCoverageDepth: 3, FacetsAvailable: 2}

	var protected Package
	var legacy Package
	found := false
	for budget := 300; budget < 4000; budget++ {
		var err error
		protected, err = CompileAdaptiveWithEvidenceConfig(
			"trace both facets", budget, index.FormatVersion, tokenizer.Name,
			[]string{"test"}, nil, nil, decision, candidates, DefaultConfig(),
		)
		if err != nil {
			continue
		}
		legacy, err = CompileAdaptiveWithEvidenceConfig(
			"trace both facets", budget, index.FormatVersion, tokenizer.Name,
			[]string{"test"}, nil, nil, decision, candidates, LegacyConfig(),
		)
		if err != nil {
			continue
		}
		if len(protected.Selections) == 2 && protected.FacetsProtected == 2 && len(legacy.Selections) == 2 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("no two-selection budget found: protected=%+v legacy=%+v", protected, legacy)
	}
	if protected.Selections[0].ProtectedFacet != "facet:b" || protected.Selections[1].ProtectedFacet != "facet:a" {
		t.Fatalf("protected facets = %+v, want call-chain then location", protected.Selections)
	}
	if legacy.Selections[0].Path != "a/one.go" || legacy.Selections[1].Path != "a/two.go" {
		t.Fatalf("legacy fitting no longer preserves candidate order: %+v", legacy.Selections)
	}
	if protected.FacetsOmittedForBudget != 0 || !protected.FacetProtection {
		t.Fatalf("facet protection summary is incomplete: %+v", protected)
	}
	assertExactPackageCount(t, protected)
}

func TestCompileAdaptiveProtectsSameFileCompanionChunks(t *testing.T) {
	owner := facetCandidate(t, "internal/owner.go", strings.Repeat("owner declaration ", 70), 10, evidence.IntentMechanism, "facet:mechanism")
	owner.Chunk.StartLine, owner.Chunk.EndLine = 1, 40
	owner.ScoreDetails = []retrieve.ScoreDetail{{Name: "BM25 content matches owner", Value: 2}}
	noise := facetCandidate(t, "internal/noise.go", strings.Repeat("unrelated helper ", 70), 9, evidence.IntentMechanism, "facet:mechanism")
	companion := facetCandidate(t, "internal/owner.go", strings.Repeat("owner continuation ", 70), 8, evidence.IntentMechanism, "facet:mechanism")
	companion.Chunk.StartLine, companion.Chunk.EndLine = 42, 80
	companion.ScoreDetails = []retrieve.ScoreDetail{{Name: "BM25 content matches continuation", Value: 2}}
	candidates := []retrieve.Candidate{owner, noise, companion}
	decision := assembly.Decision{CoverageAware: true, FacetCoverageDepth: 3, FacetsAvailable: 1}
	config := Config{ProtectFacets: true, CompanionDepth: 1}

	var protected Package
	var legacy Package
	found := false
	for budget := 300; budget < 3000; budget++ {
		var err error
		protected, err = CompileAdaptiveWithEvidenceConfig(
			"explain the owner", budget, index.FormatVersion, tokenizer.Name,
			[]string{"test"}, nil, nil, decision, candidates, config,
		)
		if err != nil {
			continue
		}
		legacy, err = CompileAdaptiveWithEvidenceConfig(
			"explain the owner", budget, index.FormatVersion, tokenizer.Name,
			[]string{"test"}, nil, nil, decision, candidates, LegacyConfig(),
		)
		if err == nil && len(protected.Selections) == 2 && len(legacy.Selections) == 2 &&
			protected.Selections[1].Path == "internal/owner.go" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("no two-selection budget found: protected=%+v legacy=%+v", protected, legacy)
	}
	if protected.Selections[0].Path != "internal/owner.go" || protected.Selections[1].Path != "internal/owner.go" {
		t.Fatalf("companion chunk was not protected: %+v", protected.Selections)
	}
	if legacy.Selections[1].Path != "internal/noise.go" {
		t.Fatalf("legacy fitting no longer preserves rank order: %+v", legacy.Selections)
	}
}

func TestCompileAdaptiveTriesNextCandidateWhenFacetOwnerDoesNotFit(t *testing.T) {
	candidates := []retrieve.Candidate{
		facetCandidate(t, "large.go", strings.Repeat("large owner ", 500), 10, evidence.IntentMechanism, "facet:mechanism"),
		facetCandidate(t, "small.go", "func smallMechanism() {}", 9, evidence.IntentMechanism, "facet:mechanism"),
	}
	decision := assembly.Decision{CoverageAware: true, FacetCoverageDepth: 2, FacetsAvailable: 1}

	var pkg Package
	found := false
	for budget := 250; budget < 1200; budget++ {
		var err error
		pkg, err = CompileAdaptiveWithEvidenceConfig(
			"how does it work", budget, index.FormatVersion, tokenizer.Name,
			[]string{"test"}, nil, nil, decision, candidates, DefaultConfig(),
		)
		if err == nil && len(pkg.Selections) == 1 && pkg.Selections[0].Path == "small.go" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("smaller facet fallback was not retained: %+v", pkg)
	}
	if pkg.Selections[0].ProtectedFacet != "facet:mechanism" || pkg.FacetsProtected != 1 {
		t.Fatalf("fallback facet was not tracked: %+v", pkg)
	}
}

func TestCompileRejectsBudgetBelowPackageMetadata(t *testing.T) {
	_, err := Compile("query", 1, index.FormatVersion, tokenizer.Name, nil, nil)
	if err == nil {
		t.Fatal("expected a metadata budget error")
	}
}

func candidate(t *testing.T, path, text string, score float64) retrieve.Candidate {
	t.Helper()
	count, err := tokenizer.Count(text)
	if err != nil {
		t.Fatal(err)
	}
	return retrieve.Candidate{
		Chunk: index.Chunk{
			Path: path, StartLine: 1, EndLine: 10,
			TokenCount: count, Text: text,
		},
		Score: score, Source: "test", Rank: 1,
	}
}

func facetCandidate(
	t *testing.T,
	path, text string,
	score float64,
	intent evidence.Intent,
	facet string,
) retrieve.Candidate {
	result := candidate(t, path, text, score)
	result.Context = &evidence.Descriptor{
		Identity:   evidence.RangeIdentity(path, 1, 10),
		Intents:    []evidence.Intent{intent},
		Roles:      []evidence.Role{evidence.RolePrimary},
		GroupIDs:   []string{facet},
		FacetRanks: map[string]int{facet: result.Rank},
	}
	return result
}

func assertExactPackageCount(t *testing.T, pkg Package) {
	t.Helper()
	data, err := Marshal(pkg)
	if err != nil {
		t.Fatal(err)
	}
	count, err := tokenizer.Count(string(data))
	if err != nil {
		t.Fatal(err)
	}
	if count != pkg.TokenCount {
		t.Fatalf("package recorded %d tokens, encoded output has %d", pkg.TokenCount, count)
	}
}
