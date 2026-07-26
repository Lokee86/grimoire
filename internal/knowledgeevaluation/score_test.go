package knowledgeevaluation

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/Lokee86/grimoire/internal/knowledge"
)

func TestScoreCaseReportsRecallMRRAndIrrelevantSelections(t *testing.T) {
	entry := Case{
		ID: "search", Query: "how does retrieval fall back", Category: "failure-fallback",
		Required:   []SectionExpectation{{Path: "docs/knowledge.md", Heading: "Fallback"}},
		Supporting: []SectionExpectation{{Path: "docs/knowledge.md", Heading: "Search"}},
		Forbidden:  []SectionExpectation{{Path: "docs/planning.md", Heading: "Future"}},
	}
	response := knowledge.SearchResponse{Results: []knowledge.Result{
		{Handle: "knowledge://docs/knowledge.md#noise", Path: "docs/knowledge.md", Heading: "Noise", Score: 4},
		{Handle: "knowledge://docs/knowledge.md#fallback", Path: "docs/knowledge.md", Heading: "Fallback", Score: 3},
		{Handle: "knowledge://docs/knowledge.md#search", Path: "docs/knowledge.md", Heading: "Search", Score: 2},
	}}
	result := ScoreCase(entry, response, 12*time.Millisecond, []int{1, 2, 3})
	if !result.Pass || result.RequiredSectionRecall != 1 || result.MRR != 0.5 {
		t.Fatalf("unexpected score: %+v", result)
	}
	if result.RecallAtK[0].Value != 0 || result.RecallAtK[1].Value != 1 || result.IrrelevantSelections != 1 {
		t.Fatalf("unexpected ranking metrics: %+v", result)
	}
	if result.RequiredMissing != nil || len(result.RequiredMatched) != 1 {
		t.Fatalf("unexpected required matches: %+v/%+v", result.RequiredMatched, result.RequiredMissing)
	}
}

func TestScoreCaseRejectsForbiddenAndRecordsVectorFailure(t *testing.T) {
	entry := Case{ID: "fallback", Query: "fallback", Category: "failure-fallback", Required: []SectionExpectation{{Path: "docs/knowledge.md", Heading: "Fallback"}}, Forbidden: []SectionExpectation{{Path: "docs/planning.md", Heading: "Future"}}}
	response := knowledge.SearchResponse{Results: []knowledge.Result{
		{Path: "docs/planning.md", Heading: "Future"},
		{Path: "docs/knowledge.md", Heading: "Fallback"},
	}, VectorError: "vector state unavailable"}
	result := ScoreCase(entry, response, time.Millisecond, []int{1})
	if result.Pass || result.ForbiddenSelections != 1 || result.VectorUsed || result.VectorError == "" {
		t.Fatalf("unexpected forbidden/vector score: %+v", result)
	}
}

func TestLoadCorpusNormalizesRecallCutoffsAndRejectsDuplicateIDs(t *testing.T) {
	corpus := Corpus{Version: FormatVersion, Repository: "fixture", Cases: []Case{
		{ID: "one", Query: "one", Category: "architecture", Required: []SectionExpectation{{Path: "docs/a.md", Heading: "A"}}},
		{ID: "two", Query: "two", Category: "ownership", Required: []SectionExpectation{{Path: "docs/b.md", Heading: "B"}}},
	}, RecallAtK: []int{5, 1, 5, 0}}
	data, err := json.Marshal(corpus)
	if err != nil {
		t.Fatal(err)
	}
	path := t.TempDir() + "/corpus.json"
	if err := writeBytes(path, data); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadCorpus(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.RecallAtK) != 2 || loaded.RecallAtK[0] != 1 || loaded.RecallAtK[1] != 5 {
		t.Fatalf("unexpected cutoffs: %v", loaded.RecallAtK)
	}
	corpus.Cases[1].ID = corpus.Cases[0].ID
	data, _ = json.Marshal(corpus)
	if err := writeBytes(path, data); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCorpus(path); err == nil {
		t.Fatal("expected duplicate case ID error")
	}
}

func TestBuildAggregateUsesDeterministicLatencyPercentiles(t *testing.T) {
	cases := []CaseResult{
		{Pass: true, RequiredSectionRecall: 1, MRR: 1, ResultCount: 2, IrrelevantSelections: 1, LatencyMS: 30},
		{Pass: false, RequiredSectionRecall: 0, MRR: 0, ResultCount: 1, IrrelevantSelections: 1, LatencyMS: 10, VectorUsed: true},
		{Pass: true, RequiredSectionRecall: 1, MRR: 0.5, ResultCount: 1, LatencyMS: 20},
	}
	aggregate := BuildAggregate(cases, []int{1})
	if aggregate.Passes != 2 || aggregate.PassRate != 2.0/3.0 || aggregate.MedianLatencyMS != 20 || aggregate.P95LatencyMS != 30 {
		t.Fatalf("unexpected aggregate: %+v", aggregate)
	}
	if aggregate.IrrelevantSelectionRate != 2.0/4.0 || aggregate.VectorUsageRate != 1.0/3.0 {
		t.Fatalf("unexpected aggregate rates: %+v", aggregate)
	}
}

func writeBytes(path string, data []byte) error {
	return os.WriteFile(path, data, 0o644)
}
