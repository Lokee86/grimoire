package app

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/grimoire/internal/knowledgeevaluation"
)

func TestEvalKnowledgeRunsFrozenCorpusInOrder(t *testing.T) {
	root := t.TempDir()
	writeKnowledgeEvalFixture(t, root)
	if err := Run([]string{"knowledge", "index", "--root", root}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	corpusPath := filepath.Join(root, "cases.json")
	corpus := knowledgeevaluation.Corpus{
		Version: knowledgeevaluation.FormatVersion, Repository: "fixture", TopK: 3, RecallAtK: []int{1, 3},
		Cases: []knowledgeevaluation.Case{
			{ID: "ownership", Query: "which package owns BM25", Category: "ownership-boundary", Required: []knowledgeevaluation.SectionExpectation{{Path: "docs/knowledge.md", Heading: "Ownership"}}},
			{ID: "fallback", Query: "what happens when vectors are unavailable", Category: "failure-fallback", Required: []knowledgeevaluation.SectionExpectation{{Path: "docs/knowledge.md", Heading: "Fallback"}}},
		},
	}
	data, err := json.Marshal(corpus)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(corpusPath, data, 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	if err := Run([]string{"eval", "knowledge", "--root", root, "--cases", corpusPath, "--vectors=false", "--output-dir", "results", "--output-prefix", "knowledge"}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile(filepath.Join(root, "results", "knowledge.json"))
	if err != nil {
		t.Fatal(err)
	}
	var report knowledgeevaluation.Report
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Cases) != 2 || report.Cases[0].CaseID != "ownership" || report.Cases[1].CaseID != "fallback" {
		t.Fatalf("cases lost corpus order: %+v", report.Cases)
	}
	if report.Aggregate.Passes != 2 || report.Aggregate.RecallAtK[0].K != 1 || report.Aggregate.RecallAtK[1].K != 3 {
		t.Fatalf("unexpected aggregate: %+v", report.Aggregate)
	}
	if report.Cases[0].VectorUsed || report.Cases[0].VectorError != "" || report.Cases[0].LatencyMS <= 0 {
		t.Fatalf("unexpected BM25 case metadata: %+v", report.Cases[0])
	}
	if !strings.Contains(stdout.String(), "required_section_recall") || !strings.Contains(stdout.String(), "json:") {
		t.Fatalf("summary missing metrics or paths: %s", stdout.String())
	}
}

func TestEvalKnowledgeRecordsVectorFallbackWithoutReplacingBM25(t *testing.T) {
	root := t.TempDir()
	writeKnowledgeEvalFixture(t, root)
	if err := Run([]string{"knowledge", "index", "--root", root}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	corpusPath := filepath.Join(root, "cases.json")
	corpus := knowledgeevaluation.Corpus{
		Version: knowledgeevaluation.FormatVersion, Repository: "fixture",
		Cases: []knowledgeevaluation.Case{{ID: "fallback", Query: "vectors unavailable", Category: "failure-fallback", Required: []knowledgeevaluation.SectionExpectation{{Path: "docs/knowledge.md", Heading: "Fallback"}}}},
	}
	data, _ := json.Marshal(corpus)
	if err := os.WriteFile(corpusPath, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Run([]string{"eval", "knowledge", "--root", root, "--cases", corpusPath, "--output-dir", "results", "--output-prefix", "vector-fallback"}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "results", "vector-fallback.json"))
	if err != nil {
		t.Fatal(err)
	}
	var report knowledgeevaluation.Report
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Cases) != 1 || report.Cases[0].VectorUsed || report.Cases[0].VectorError == "" || !report.Cases[0].Pass {
		t.Fatalf("vector fallback did not preserve BM25 result: %+v", report.Cases)
	}
}

func writeKnowledgeEvalFixture(t *testing.T, root string) {
	t.Helper()
	path := filepath.Join(root, "docs", "knowledge.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "# Knowledge\n\n## Ownership\nThe knowledge package owns documentation discovery and BM25 ranking.\n\n## Fallback\nMissing documentation vectors preserve BM25 results.\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
