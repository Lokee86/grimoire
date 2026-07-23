package compiler

import (
	"strings"
	"testing"

	"github.com/Lokee86/grimoire/internal/assembly"
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
