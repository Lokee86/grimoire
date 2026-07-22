package compiler

import (
	"strings"
	"testing"

	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/retrieve"
	"github.com/Lokee86/grimoire/internal/tokenizer"
)

func TestCompileEnforcesSerializedPackageBudget(t *testing.T) {
	candidates := []retrieve.Candidate{
		candidate(t, "first.go", strings.Repeat("first value ", 80), 10),
		candidate(t, "second.go", strings.Repeat("second value ", 80), 9),
		candidate(t, "third.go", strings.Repeat("third value ", 80), 8),
	}

	full, err := Compile("query", 10_000, index.FormatVersion, tokenizer.Name, candidates)
	if err != nil {
		t.Fatal(err)
	}
	if len(full.Selections) != len(candidates) {
		t.Fatalf("expected all selections, got %+v", full.Selections)
	}
	assertExactPackageCount(t, full)

	tightBudget := full.TokenCount - 20
	tight, err := Compile("query", tightBudget, index.FormatVersion, tokenizer.Name, candidates)
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

func TestCompileRejectsBudgetBelowPackageMetadata(t *testing.T) {
	_, err := Compile("query", 1, index.FormatVersion, tokenizer.Name, nil)
	if err == nil {
		t.Fatal("expected a metadata budget error")
	}
}

func candidate(t *testing.T, path, text string, score int) retrieve.Candidate {
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
		Score: score,
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
