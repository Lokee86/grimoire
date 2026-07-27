package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Lokee86/grimoire/internal/agentruntime"
	"github.com/Lokee86/grimoire/internal/arcanaevaluation"
	"github.com/Lokee86/grimoire/internal/arcanagraph"
	"github.com/Lokee86/grimoire/internal/embedding"
	"github.com/Lokee86/grimoire/internal/evaluation"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/lexiconfacts"
	"github.com/Lokee86/grimoire/internal/structure"
)

type arcanaEvaluationProviders struct {
	Lexicon  func(string, int) ([]structure.Node, error)
	Semantic func(context.Context, string, int) ([]arcanagraph.SemanticSeed, error)
	Rerank   func(context.Context, string, []structure.Node, []arcanagraph.SemanticSeed, int) ([]arcanagraph.RerankedSeed, error)
	Graph    func(context.Context, []structure.Node) ([]structure.Evidence, error)
}

func runEvalArcana(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("eval arcana", flag.ContinueOnError)
	flags.SetOutput(stderr)
	casesPath := flags.String("cases", "", "frozen judged Arcana semantic graph corpus JSON")
	root := flags.String("root", ".", "repository root")
	state := flags.String("state", "", "prepared Grimoire index repository path")
	lexiconFacts := flags.String("lexicon-facts", "", "explicit directory containing exported Lexicon JSONL libraries")
	lexiconState := flags.String("lexicon-state", "", "Lexicon state directory; defaults to <root>/.lexicon")
	lexiconCommand := flags.String("lexicon-command", "", "Lexicon executable override; discovered when omitted")
	arcanaState := flags.String("arcana-state", "", "Arcana state directory; defaults to <root>/.arcana")
	arcanaCommand := flags.String("arcana-command", "", "Arcana executable override; discovered when omitted")
	endpoint := flags.String("endpoint", embedding.DefaultEndpoint, "OpenAI-compatible embeddings endpoint used by Arcana semantic query")
	topK := flags.Int("top-k", 0, "override the corpus seed limit; zero uses corpus top_k")
	recallAtK := flags.String("recall-at-k", "", "override seed recall cutoffs, for example 1,3,6")
	variant := flags.String("variant", "paired", "evaluation variant label")
	timeout := flags.Duration("timeout", 30*time.Second, "per-mode provider timeout")
	outputDir := flags.String("output-dir", "evaluation/results", "result directory")
	outputPrefix := flags.String("output-prefix", "", "result filename prefix")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*casesPath) == "" || *topK < 0 || *timeout <= 0 {
		return errors.New("--cases, non-negative --top-k, and positive --timeout are required")
	}

	corpus, err := arcanaevaluation.LoadCorpus(*casesPath)
	if err != nil {
		return err
	}
	effectiveTopK := *topK
	if effectiveTopK == 0 {
		effectiveTopK = corpus.TopK
	}
	if effectiveTopK < 1 || effectiveTopK > arcanaevaluation.MaxSeedLimit {
		return fmt.Errorf("--top-k must be between 1 and %d", arcanaevaluation.MaxSeedLimit)
	}
	effectiveRecallAtK := append([]int(nil), corpus.RecallAtK...)
	if strings.TrimSpace(*recallAtK) != "" {
		effectiveRecallAtK, err = parseKnowledgeEvaluationKs(*recallAtK)
		if err != nil {
			return fmt.Errorf("--recall-at-k: %w", err)
		}
	}

	absoluteRoot, err := filepath.Abs(*root)
	if err != nil {
		return fmt.Errorf("resolve Arcana evaluation root: %w", err)
	}
	for _, entry := range corpus.Cases {
		if err := validateArcanaEvaluationCase(absoluteRoot, entry); err != nil {
			return err
		}
	}
	statePath, err := resolveState(absoluteRoot, *state)
	if err != nil {
		return err
	}
	prepared, err := index.Load(statePath)
	if err != nil {
		return fmt.Errorf("load prepared index: %w", err)
	}
	resolvedLexiconCommand := agentruntime.ResolveProviderCommand(absoluteRoot, *lexiconCommand, "lexicon")
	resolvedArcanaCommand := agentruntime.ResolveProviderCommand(absoluteRoot, *arcanaCommand, "arcana")

	setupContext, cancelSetup := context.WithTimeout(context.Background(), *timeout)
	exportDirectory, lexiconSnapshot, err := lexiconfacts.ResolveExport(setupContext, lexiconfacts.ExportOptions{
		Root: absoluteRoot, GrimoireState: statePath, ExplicitDirectory: *lexiconFacts,
		LexiconState: *lexiconState, Command: resolvedLexiconCommand,
	})
	if err != nil {
		cancelSetup()
		return fmt.Errorf("resolve Lexicon evaluation export: %w", err)
	}
	if exportDirectory == "" {
		cancelSetup()
		return errors.New("Arcana evaluation requires an immutable Lexicon export")
	}
	arcanaSnapshot, arcanaSnapshotID, err := arcanagraph.ResolveSnapshot(setupContext, arcanagraph.StateOptions{
		Root: absoluteRoot, State: *arcanaState, LexiconState: *lexiconState,
		ExpectedLexiconSnapshot: lexiconSnapshot, Command: resolvedArcanaCommand,
	})
	cancelSetup()
	if err != nil {
		return fmt.Errorf("resolve Arcana evaluation snapshot: %w", err)
	}
	if arcanaSnapshot == "" || arcanaSnapshotID == "" {
		return errors.New("Arcana evaluation requires an immutable Arcana graph snapshot")
	}
	resolvedArcanaState := filepath.Dir(filepath.Dir(arcanaSnapshot))
	if err := requireArcanaVectorIndex(resolvedArcanaState, arcanaSnapshotID); err != nil {
		return err
	}

	client := arcanagraph.Client{Command: resolvedArcanaCommand}
	providers := arcanaEvaluationProviders{
		Lexicon: func(query string, limit int) ([]structure.Node, error) {
			result, searchErr := lexiconfacts.SearchDetailed(prepared, query, exportDirectory, limit)
			return result.Seeds, searchErr
		},
		Semantic: func(ctx context.Context, query string, limit int) ([]arcanagraph.SemanticSeed, error) {
			return client.RankedSemanticSeeds(ctx, resolvedArcanaState, arcanaSnapshotID, *endpoint, query, limit)
		},
		Rerank: func(
			ctx context.Context,
			query string,
			lexicon []structure.Node,
			semantic []arcanagraph.SemanticSeed,
			limit int,
		) ([]arcanagraph.RerankedSeed, error) {
			return client.RerankSeeds(ctx, arcanaSnapshot, query, lexicon, semantic, limit)
		},
		Graph: func(ctx context.Context, seeds []structure.Node) ([]structure.Evidence, error) {
			return client.Search(ctx, arcanaSnapshot, seeds)
		},
	}

	report := arcanaevaluation.Report{
		Version: arcanaevaluation.FormatVersion, GeneratedAt: time.Now(),
		Repository: corpus.Repository, SourceURL: corpus.SourceURL, Revision: corpus.Revision,
		Scope: corpus.Scope, JudgedAt: corpus.JudgedAt, Root: absoluteRoot, State: statePath,
		LexiconSnapshot: lexiconSnapshot, ArcanaSnapshot: arcanaSnapshotID,
		EmbeddingIdentity: embedding.Identity(), CasesFile: *casesPath, Variant: *variant,
		TopK: effectiveTopK, RecallAtK: effectiveRecallAtK,
		Modes: []string{arcanaevaluation.ModeLexiconSeeds, arcanaevaluation.ModeLexiconVectorSeeds},
		Cases: make([]arcanaevaluation.CaseResult, 0, len(corpus.Cases)*2),
	}
	for _, entry := range corpus.Cases {
		for _, mode := range report.Modes {
			ctx, cancel := context.WithTimeout(context.Background(), *timeout)
			measurement := measureArcanaEvaluationMode(ctx, entry.Query, mode, effectiveTopK, providers)
			cancel()
			report.Cases = append(report.Cases, arcanaevaluation.ScoreCase(entry, measurement, effectiveRecallAtK))
		}
	}
	arcanaevaluation.BuildAggregates(&report)

	prefix := strings.TrimSpace(*outputPrefix)
	if prefix == "" {
		prefix = defaultArcanaEvaluationPrefix(corpus.Repository, *variant, report.GeneratedAt)
	}
	resolvedOutputDir := *outputDir
	if !filepath.IsAbs(resolvedOutputDir) {
		resolvedOutputDir = filepath.Join(absoluteRoot, resolvedOutputDir)
	}
	jsonPath := filepath.Join(resolvedOutputDir, prefix+".json")
	markdownPath := filepath.Join(resolvedOutputDir, prefix+".md")
	if err := arcanaevaluation.Write(report, jsonPath, markdownPath); err != nil {
		return err
	}
	return writeArcanaEvaluationSummary(stdout, report, jsonPath, markdownPath)
}

func measureArcanaEvaluationMode(
	ctx context.Context,
	query, mode string,
	limit int,
	providers arcanaEvaluationProviders,
) arcanaevaluation.Measurement {
	measurement := arcanaevaluation.Measurement{Mode: mode}
	started := time.Now()

	lexiconStarted := time.Now()
	measurement.ProviderCalls++
	lexiconSeeds, err := providers.Lexicon(query, limit)
	measurement.Timings.LexiconSeedMS = durationMS(time.Since(lexiconStarted))
	if err != nil {
		measurement.Error = fmt.Sprintf("Lexicon seed retrieval: %v", err)
		measurement.Timings.TotalMS = measuredDurationMS(started)
		return measurement
	}

	var semanticSeeds []arcanagraph.SemanticSeed
	if mode == arcanaevaluation.ModeLexiconVectorSeeds {
		semanticStarted := time.Now()
		measurement.ProviderCalls++
		semanticSeeds, err = providers.Semantic(ctx, query, arcanagraph.SemanticCandidateLimit(limit))
		measurement.Timings.SemanticSeedMS = durationMS(time.Since(semanticStarted))
		if err != nil {
			measurement.Error = fmt.Sprintf("Arcana semantic seed retrieval: %v", err)
			measurement.Timings.TotalMS = measuredDurationMS(started)
			return measurement
		}
		measurement.VectorUsed = true
	} else if mode != arcanaevaluation.ModeLexiconSeeds {
		measurement.Error = fmt.Sprintf("unknown Arcana evaluation mode %q", mode)
		measurement.Timings.TotalMS = measuredDurationMS(started)
		return measurement
	}

	rerankedSeeds, err := providers.Rerank(ctx, query, lexiconSeeds, semanticSeeds, limit)
	if err != nil {
		measurement.Error = fmt.Sprintf("Arcana hybrid seed reranking: %v", err)
		measurement.Timings.TotalMS = measuredDurationMS(started)
		return measurement
	}
	measurement.Seeds = rankedArcanaSeeds(rerankedSeeds)
	if len(measurement.Seeds) > 0 {
		graphStarted := time.Now()
		measurement.ProviderCalls++
		measurement.Structural, err = providers.Graph(ctx, nodesFromRankedSeeds(measurement.Seeds))
		measurement.Timings.GraphSearchMS = durationMS(time.Since(graphStarted))
		if err != nil {
			measurement.Error = fmt.Sprintf("Arcana graph expansion: %v", err)
		}
	}
	measurement.Timings.TotalMS = measuredDurationMS(started)
	return measurement
}

func rankedArcanaSeeds(seeds []arcanagraph.RerankedSeed) []arcanaevaluation.RankedSeed {
	result := make([]arcanaevaluation.RankedSeed, 0, len(seeds))
	for _, seed := range seeds {
		result = append(result, arcanaevaluation.RankedSeed{Node: seed.Node, Source: seed.Source})
	}
	return result
}

func nodesFromRankedSeeds(seeds []arcanaevaluation.RankedSeed) []structure.Node {
	result := make([]structure.Node, 0, len(seeds))
	for _, seed := range seeds {
		result = append(result, seed.Node)
	}
	return result
}

func arcanaEvaluationSeedKey(seed structure.Node) string {
	return seed.Name + "\x00" + filepath.ToSlash(seed.Path)
}

func measuredDurationMS(started time.Time) float64 {
	value := durationMS(time.Since(started))
	if value < 0.001 {
		return 0.001
	}
	return value
}

func requireArcanaVectorIndex(state, snapshotID string) error {
	digest, found := strings.CutPrefix(strings.TrimSpace(snapshotID), "sha256:")
	if !found || len(digest) != 64 {
		return fmt.Errorf("invalid Arcana snapshot ID %q", snapshotID)
	}
	manifest := filepath.Join(state, "vectors", digest, embedding.Identity(), "manifest.json")
	info, err := os.Stat(manifest)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("Arcana paired evaluation requires vector index %s; run `arcana vectorize` first", manifest)
		}
		return fmt.Errorf("inspect Arcana evaluation vector index: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("Arcana evaluation vector manifest is a directory: %s", manifest)
	}
	return nil
}

func validateArcanaEvaluationCase(root string, entry arcanaevaluation.Case) error {
	for _, group := range [][]arcanaevaluation.SeedExpectation{entry.RequiredSeeds, entry.SupportingSeeds} {
		for _, expected := range group {
			if err := validateExpectedSymbol(root, entry.ID, expected.Path, expected.Name); err != nil {
				return err
			}
		}
	}
	for _, group := range [][]evaluation.StructuralExpectation{
		entry.RequiredStructural, entry.SupportingStructural,
	} {
		for _, expected := range group {
			if expected.Path != "" {
				if err := validateExpectedSymbol(root, entry.ID, expected.Path, expected.Symbol); err != nil {
					return err
				}
			}
			if expected.TargetPath != "" {
				if err := validateExpectedSymbol(root, entry.ID, expected.TargetPath, expected.TargetSymbol); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func defaultArcanaEvaluationPrefix(repository, variant string, generated time.Time) string {
	name := strings.ToLower(strings.NewReplacer(" ", "-", "_", "-", "/", "-").Replace(repository))
	variant = strings.ToLower(strings.NewReplacer(" ", "-", "_", "-", "/", "-").Replace(strings.TrimSpace(variant)))
	return fmt.Sprintf("arcana-%s-%s-%s", name, variant, generated.Format("2006-01-02-150405"))
}

func writeArcanaEvaluationSummary(stdout io.Writer, report arcanaevaluation.Report, jsonPath, markdownPath string) error {
	if _, err := fmt.Fprintln(stdout, "mode\tpass\tseed_recall\tmrr\tstructural_recall\tmedian_ms\tmedian_bytes\tprovider_calls"); err != nil {
		return err
	}
	for _, aggregate := range report.Aggregates {
		if _, err := fmt.Fprintf(stdout, "%s\t%.1f%%\t%.1f%%\t%.3f\t%.1f%%\t%.1f\t%.0f\t%.2f\n",
			aggregate.Mode, aggregate.PassRate*100, aggregate.RequiredSeedRecall*100, aggregate.MRR,
			aggregate.RequiredStructuralRecall*100, aggregate.MedianLatencyMS,
			aggregate.MedianPayloadBytes, aggregate.MeanProviderCalls); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(stdout, "json: %s\nmarkdown: %s\n", jsonPath, markdownPath)
	return err
}
