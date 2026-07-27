package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/grimoire/internal/arcanaevaluation"
	"github.com/Lokee86/grimoire/internal/evaluation"
	"github.com/Lokee86/grimoire/internal/structure"
)

func TestMeasureArcanaEvaluationModePairsSameGraphExpansion(t *testing.T) {
	lexicon := structure.Node{Name: "Lexical", Path: "lexical.go"}
	semantic := structure.Node{Name: "Semantic", Path: "semantic.go"}
	graphCalls := 0
	providers := arcanaEvaluationProviders{
		Lexicon: func(query string, limit int) ([]structure.Node, error) {
			if query != "conceptual task" || limit != 6 {
				t.Fatalf("unexpected Lexicon invocation: query=%q limit=%d", query, limit)
			}
			return []structure.Node{lexicon}, nil
		},
		Semantic: func(_ context.Context, query string, limit int) ([]structure.Node, error) {
			if query != "conceptual task" || limit != 6 {
				t.Fatalf("unexpected semantic invocation: query=%q limit=%d", query, limit)
			}
			return []structure.Node{semantic}, nil
		},
		Graph: func(_ context.Context, seeds []structure.Node) ([]structure.Evidence, error) {
			graphCalls++
			return []structure.Evidence{{Provider: "arcana", Kind: "operational_role", Node: &seeds[0]}}, nil
		},
	}

	baseline := measureArcanaEvaluationMode(context.Background(), "conceptual task", arcanaevaluation.ModeLexiconSeeds, 6, providers)
	vector := measureArcanaEvaluationMode(context.Background(), "conceptual task", arcanaevaluation.ModeLexiconVectorSeeds, 6, providers)
	if baseline.VectorUsed || baseline.ProviderCalls != 2 || len(baseline.Seeds) != 1 || baseline.Seeds[0].Source != "lexicon" {
		t.Fatalf("unexpected baseline measurement: %+v", baseline)
	}
	if !vector.VectorUsed || vector.ProviderCalls != 3 || len(vector.Seeds) != 2 || vector.Seeds[0].Node.Name != "Semantic" || vector.Seeds[0].Source != "vector" {
		t.Fatalf("unexpected vector measurement: %+v", vector)
	}
	if graphCalls != 2 || len(baseline.Structural) != 1 || len(vector.Structural) != 1 {
		t.Fatalf("graph expansion was not paired: calls=%d baseline=%+v vector=%+v", graphCalls, baseline.Structural, vector.Structural)
	}
}

func TestMeasureArcanaEvaluationModeScoresSemanticDeclarationEntry(t *testing.T) {
	semantic := structure.Node{
		Kind: "method", Name: "SemanticSeeds", Path: "internal/arcanagraph/semantic.go",
	}
	providers := arcanaEvaluationProviders{
		Lexicon: func(string, int) ([]structure.Node, error) {
			return nil, nil
		},
		Semantic: func(context.Context, string, int) ([]structure.Node, error) {
			return []structure.Node{semantic}, nil
		},
		Graph: func(_ context.Context, seeds []structure.Node) ([]structure.Evidence, error) {
			if len(seeds) != 1 || seeds[0] != semantic {
				t.Fatalf("semantic seed was not passed unchanged to graph expansion: %+v", seeds)
			}
			return []structure.Evidence{{
				Provider: "arcana", Kind: "operational_role", Node: &semantic,
			}}, nil
		},
	}
	entry := arcanaevaluation.Case{
		ID: "semantic-entry", Query: "conceptual graph entry point",
		RequiredSeeds: []arcanaevaluation.SeedExpectation{{
			Name: "SemanticSeeds", Path: "internal/arcanagraph/semantic.go", Kind: "function",
		}},
		RequiredStructural: []evaluation.StructuralExpectation{{
			Provider: "arcana", Kind: "operational_role",
			Symbol: "SemanticSeeds", Path: "internal/arcanagraph/semantic.go",
		}},
	}

	measurement := measureArcanaEvaluationMode(
		context.Background(), entry.Query,
		arcanaevaluation.ModeLexiconVectorSeeds, 6, providers,
	)
	result := arcanaevaluation.ScoreCase(entry, measurement, []int{1})
	if !measurement.VectorUsed || measurement.ProviderCalls != 3 ||
		len(measurement.Seeds) != 1 || measurement.Seeds[0].Source != "vector" {
		t.Fatalf("semantic entry source was not measured correctly: %+v", measurement)
	}
	if !result.Pass || result.RequiredSeedRecall != 1 || result.RequiredStructuralRecall != 1 {
		t.Fatalf("semantic declaration benefit was not measured: %+v", result)
	}
}

func TestRequireArcanaVectorIndexUsesSnapshotAndEmbeddingIdentity(t *testing.T) {
	state := t.TempDir()
	digest := strings.Repeat("a", 64)
	manifest := filepath.Join(state, "vectors", digest, "qwen3-embedding-0.6b-q8_0-512d", "manifest.json")
	if err := os.MkdirAll(filepath.Dir(manifest), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifest, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := requireArcanaVectorIndex(state, "sha256:"+digest); err != nil {
		t.Fatal(err)
	}
	if err := requireArcanaVectorIndex(state, "sha256:"+strings.Repeat("b", 64)); err == nil || !strings.Contains(err.Error(), "arcana vectorize") {
		t.Fatalf("expected actionable missing-index error, got %v", err)
	}
}

func TestRunDispatchesArcanaEvaluator(t *testing.T) {
	err := Run([]string{"eval", "arcana"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "--cases") {
		t.Fatalf("Arcana evaluator was not dispatched: %v", err)
	}
}

func TestCheckedInArcanaCorpusReferencesCurrentSymbols(t *testing.T) {
	root := filepath.Join("..", "..")
	corpus, err := arcanaevaluation.LoadCorpus(filepath.Join(root, "evaluation", "arcana", "grimoire.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range corpus.Cases {
		if err := validateArcanaEvaluationCase(root, entry); err != nil {
			t.Fatal(err)
		}
	}
}

func TestCheckedInSpaceRocksArcanaCorpusReferencesExternalSymbols(t *testing.T) {
	root := strings.TrimSpace(os.Getenv("GRIMOIRE_ARCANA_SPACE_ROCKS_ROOT"))
	if root == "" {
		t.Skip("set GRIMOIRE_ARCANA_SPACE_ROCKS_ROOT to validate the external Space Rocks checkout")
	}
	corpus, err := arcanaevaluation.LoadCorpus(filepath.Join("..", "..", "evaluation", "arcana", "space-rocks.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range corpus.Cases {
		if err := validateArcanaEvaluationCase(root, entry); err != nil {
			t.Fatal(err)
		}
	}
}
