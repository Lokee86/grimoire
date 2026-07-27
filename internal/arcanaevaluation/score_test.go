package arcanaevaluation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/grimoire/internal/structure"
)

func TestLoadCorpusNormalizesPairingControls(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cases.json")
	data := `{
  "version": 1,
  "repository": "fixture",
  "top_k": 4,
  "recall_at_k": [4, 1, 4],
  "cases": [{
    "id": "seed",
    "query": "find target",
    "category": "direct-location",
    "required_seeds": [{"name": "Target", "path": "target.go"}],
    "required_structural": [{"provider": "arcana", "kind": "operational_role", "symbol": "Target", "path": "target.go"}]
  }]
}`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	corpus, err := LoadCorpus(path)
	if err != nil {
		t.Fatal(err)
	}
	if corpus.TopK != 4 || len(corpus.RecallAtK) != 2 || corpus.RecallAtK[0] != 1 || corpus.RecallAtK[1] != 4 {
		t.Fatalf("unexpected corpus controls: %+v", corpus)
	}
}

func TestScoreCaseMeasuresSeedsStructurePayloadAndCalls(t *testing.T) {
	entry := Case{
		ID: "target", Query: "find target", Category: "direct-location",
		RequiredSeeds:   []SeedExpectation{{Name: "Target", Path: "target.go"}},
		SupportingSeeds: []SeedExpectation{{Name: "Helper", Path: "helper.go"}},
		RequiredStructural: []StructuralExpectation{{
			Provider: "arcana", Kind: "operational_role", Symbol: "Target", Path: "target.go",
			Relation: "calls", Direction: "outgoing", TargetSymbol: "Helper", TargetPath: "helper.go",
		}},
	}
	measurement := Measurement{
		Mode: ModeLexiconVectorSeeds, VectorUsed: true, ProviderCalls: 3,
		Timings: Timings{LexiconSeedMS: 1, SemanticSeedMS: 2, GraphSearchMS: 3, TotalMS: 6},
		Seeds: []RankedSeed{
			{Node: structure.Node{Name: "Noise", Path: "noise.go"}, Source: "vector"},
			{Node: structure.Node{Name: "Target", Path: "target.go"}, Source: "vector"},
			{Node: structure.Node{Name: "Helper", Path: "helper.go"}, Source: "lexicon"},
		},
		Structural: []structure.Evidence{{
			Provider: "arcana", Kind: "operational_role", Node: &structure.Node{Name: "Target", Path: "target.go"},
			Relationships: []structure.Relationship{{
				Direction: "outgoing", Relation: "calls", Node: structure.Node{Name: "Helper", Path: "helper.go"},
			}},
		}},
	}
	result := ScoreCase(entry, measurement, []int{1, 3})
	if !result.Pass || result.RequiredSeedRecall != 1 || result.RequiredStructuralRecall != 1 {
		t.Fatalf("unexpected scored result: %+v", result)
	}
	if result.MRR != 0.5 || result.FirstRequiredSeedRank != 2 || result.RecallAtK[0].Value != 0 || result.RecallAtK[1].Value != 1 {
		t.Fatalf("unexpected ranking metrics: %+v", result)
	}
	if result.PayloadBytes != result.SeedPayloadBytes+result.StructuralPayloadBytes || result.PayloadBytes <= 0 || result.ProviderCalls != 3 {
		t.Fatalf("payload or call metrics missing: %+v", result)
	}
}

func TestScoreCasePreservesExactSymbolRecallAcrossCallableKinds(t *testing.T) {
	entry := Case{
		ID: "semantic-entry", Query: "find conceptual graph entry point",
		RequiredSeeds: []SeedExpectation{{
			Name: "SemanticSeeds", Path: "internal/arcanagraph/semantic.go", Kind: "function",
		}},
		RequiredStructural: []StructuralExpectation{{
			Provider: "arcana", Kind: "operational_role",
			Symbol: "SemanticSeeds", Path: "internal/arcanagraph/semantic.go",
		}},
	}
	method := structure.Node{
		Kind: "method", Name: "SemanticSeeds", Path: "internal/arcanagraph/semantic.go",
	}
	result := ScoreCase(entry, Measurement{
		Mode:  ModeLexiconVectorSeeds,
		Seeds: []RankedSeed{{Node: method, Source: "vector"}},
		Structural: []structure.Evidence{{
			Provider: "arcana", Kind: "operational_role", Node: &method,
		}},
	}, []int{1})

	if !result.Pass || result.RequiredSeedRecall != 1 || result.RecallAtK[0].Value != 1 {
		t.Fatalf("exact name/path seed recall was lost to callable kind normalization: %+v", result)
	}
	if result.RequiredStructuralRecall != 1 || !result.StructuralJudgments[0].Matched {
		t.Fatalf("resolved declaration did not satisfy structural judgment: %+v", result)
	}
}

func TestBuildAggregatesProducesVectorMinusBaselineDeltas(t *testing.T) {
	report := Report{RecallAtK: []int{1}, Cases: []CaseResult{
		{Mode: ModeLexiconSeeds, Pass: false, RequiredSeedRecall: 0, RecallAtK: []RecallMetric{{K: 1, Value: 0}}, MRR: 0, RequiredStructuralRecall: 0.5, Timings: Timings{TotalMS: 4}, PayloadBytes: 100, ProviderCalls: 2},
		{Mode: ModeLexiconVectorSeeds, Pass: true, VectorUsed: true, RequiredSeedRecall: 1, RecallAtK: []RecallMetric{{K: 1, Value: 1}}, MRR: 1, RequiredStructuralRecall: 1, Timings: Timings{TotalMS: 9}, PayloadBytes: 140, ProviderCalls: 3},
	}}
	BuildAggregates(&report)
	if len(report.Aggregates) != 2 || report.Comparison.RequiredSeedRecallDelta != 1 || report.Comparison.MRRDelta != 1 {
		t.Fatalf("unexpected comparison: %+v", report)
	}
	if report.Comparison.RequiredStructuralRecallDelta != 0.5 || report.Comparison.MedianLatencyMSDelta != 5 || report.Comparison.MeanProviderCallsDelta != 1 {
		t.Fatalf("unexpected cost deltas: %+v", report.Comparison)
	}
}

func TestWriteEmitsPairedJSONAndMarkdown(t *testing.T) {
	directory := t.TempDir()
	report := Report{
		Version: FormatVersion, Repository: "fixture", Variant: "paired", ArcanaSnapshot: "sha256:test",
		EmbeddingIdentity: "test-model", RecallAtK: []int{1},
		Cases: []CaseResult{{Mode: ModeLexiconSeeds}, {Mode: ModeLexiconVectorSeeds}},
	}
	jsonPath := filepath.Join(directory, "report.json")
	markdownPath := filepath.Join(directory, "report.md")
	if err := Write(report, jsonPath, markdownPath); err != nil {
		t.Fatal(err)
	}
	markdown, err := os.ReadFile(markdownPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(markdown), "Vector-minus-baseline deltas") || !strings.Contains(string(markdown), "Provider calls") {
		t.Fatalf("paired metrics missing from Markdown: %s", markdown)
	}
	if _, err := os.Stat(jsonPath); err != nil {
		t.Fatal(err)
	}
}
