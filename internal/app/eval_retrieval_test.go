package app

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/grimoire/internal/evaluation"
)

func TestParseStructuralProviders(t *testing.T) {
	providers, enabled, arcana, err := parseStructuralProviders("lexicon,arcana")
	if err != nil {
		t.Fatal(err)
	}
	if !enabled || !arcana || len(providers) != 2 {
		t.Fatalf("unexpected providers: %v enabled=%t arcana=%t", providers, enabled, arcana)
	}
	if _, _, _, err := parseStructuralProviders("arcana"); err == nil {
		t.Fatal("expected Arcana without Lexicon to fail")
	}
}

func TestEvalRetrievalLexicalWritesResults(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "internal", "damage"), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "package damage\n\nfunc ResolveDamage() int { return 10 }\n"
	if err := os.WriteFile(filepath.Join(root, "internal", "damage", "resolve.go"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Run([]string{"index", "--root", root}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	corpusPath := filepath.Join(root, "cases.json")
	corpus := evaluation.Corpus{
		Version: 1, Repository: "fixture",
		Cases: []evaluation.Case{{
			ID: "resolve-damage", Query: "Where is ResolveDamage?",
			Category: evaluation.CategoryDirectLocation, Budget: 500,
			Required: []evaluation.Evidence{{Path: "internal/damage/resolve.go", Symbols: []string{"ResolveDamage"}}},
		}},
	}
	data, err := json.Marshal(corpus)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(corpusPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	if err := Run([]string{
		"eval", "retrieval", "--root", root, "--cases", corpusPath,
		"--modes", "lexical", "--output-dir", "results", "--output-prefix", "fixture",
	}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	resultPath := filepath.Join(root, "results", "fixture.json")
	resultData, err := os.ReadFile(resultPath)
	if err != nil {
		t.Fatal(err)
	}
	var report evaluation.Report
	if err := json.Unmarshal(resultData, &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Runs) != 1 || !report.Runs[0].Pass {
		t.Fatalf("unexpected report: %+v", report.Runs)
	}
	if report.Runs[0].Timings.TotalMS <= 0 {
		t.Fatalf("total timing was not recorded: %+v", report.Runs[0].Timings)
	}
	if report.Runs[0].QueryProfile.Specificity != "high" || report.Runs[0].RetrievalPolicy.Scope != "focused" || !report.Runs[0].RetrievalPolicy.Shadow {
		t.Fatalf("query profile was not emitted: profile=%+v policy=%+v", report.Runs[0].QueryProfile, report.Runs[0].RetrievalPolicy)
	}
	markdown, err := os.ReadFile(filepath.Join(root, "results", "fixture.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(markdown), "## Query profile shadow output") {
		t.Fatalf("query profile section missing from Markdown: %s", markdown)
	}
}
