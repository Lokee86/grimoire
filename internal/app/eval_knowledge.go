package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Lokee86/grimoire/internal/embedding"
	"github.com/Lokee86/grimoire/internal/knowledge"
	"github.com/Lokee86/grimoire/internal/knowledgeevaluation"
	"github.com/Lokee86/grimoire/internal/knowledgevector"
)

func runEvalKnowledge(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("eval knowledge", flag.ContinueOnError)
	flags.SetOutput(stderr)
	casesPath := flags.String("cases", "", "frozen judged documentation corpus JSON")
	root := flags.String("root", ".", "repository root")
	state := flags.String("state", "", "knowledge state directory")
	vectors := flags.Bool("vectors", false, "attempt the current documentation vector snapshot as a BM25 supplement")
	endpoint := flags.String("endpoint", embedding.DefaultEndpoint, "OpenAI-compatible embeddings endpoint")
	enginePath := flags.String("engine", "", "Rust vector engine DLL")
	topK := flags.Int("top-k", 0, "override the corpus result limit; zero uses corpus top_k")
	recallAtK := flags.String("recall-at-k", "", "override corpus recall cutoffs, for example 1,3,5,10")
	variant := flags.String("variant", "bm25", "evaluation variant label")
	timeout := flags.Duration("timeout", 2*time.Minute, "per-case knowledge search timeout")
	outputDir := flags.String("output-dir", "evaluation/results", "result directory")
	outputPrefix := flags.String("output-prefix", "", "result filename prefix")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*casesPath) == "" || *topK < 0 || *timeout <= 0 {
		return errors.New("--cases, non-negative --top-k, and positive --timeout are required")
	}
	corpus, err := knowledgeevaluation.LoadCorpus(*casesPath)
	if err != nil {
		return err
	}
	absoluteRoot, err := filepath.Abs(*root)
	if err != nil {
		return fmt.Errorf("resolve knowledge evaluation root: %w", err)
	}
	statePath, err := resolveKnowledgeState(absoluteRoot, *state)
	if err != nil {
		return err
	}
	index, err := knowledge.Load(statePath)
	if err != nil {
		return fmt.Errorf("load knowledge index: %w", err)
	}
	effectiveTopK := *topK
	if effectiveTopK == 0 {
		effectiveTopK = corpus.TopK
		if effectiveTopK == 0 {
			effectiveTopK = 10
		}
	}
	effectiveRecallAtK := corpus.RecallAtK
	if strings.TrimSpace(*recallAtK) != "" {
		effectiveRecallAtK, err = parseKnowledgeEvaluationKs(*recallAtK)
		if err != nil {
			return fmt.Errorf("--recall-at-k: %w", err)
		}
	}

	report := knowledgeevaluation.Report{
		Version: knowledgeevaluation.FormatVersion, GeneratedAt: time.Now(),
		Repository: corpus.Repository, SourceURL: corpus.SourceURL, Revision: corpus.Revision,
		Scope: corpus.Scope, JudgedAt: corpus.JudgedAt, Root: absoluteRoot, State: statePath,
		CasesFile: *casesPath, Variant: *variant, Vectors: *vectors, TopK: effectiveTopK,
		RecallAtK: append([]int(nil), effectiveRecallAtK...), Cases: make([]knowledgeevaluation.CaseResult, 0, len(corpus.Cases)),
	}
	for _, entry := range corpus.Cases {
		ctx, cancel := context.WithTimeout(context.Background(), *timeout)
		options := knowledge.SearchOptions{TopK: effectiveTopK}
		if *vectors {
			options.Vector = knowledgevector.Ranker{State: statePath, Index: index, Endpoint: *endpoint, EnginePath: *enginePath}
		}
		started := time.Now()
		response, searchErr := knowledge.Search(ctx, index, entry.Query, options)
		cancel()
		if searchErr != nil {
			return fmt.Errorf("knowledge evaluation case %q: %w", entry.ID, searchErr)
		}
		report.Cases = append(report.Cases, knowledgeevaluation.ScoreCase(entry, response, time.Since(started), effectiveRecallAtK))
	}
	report.Aggregate = knowledgeevaluation.BuildAggregate(report.Cases, effectiveRecallAtK)

	prefix := strings.TrimSpace(*outputPrefix)
	if prefix == "" {
		prefix = defaultKnowledgeEvaluationPrefix(corpus.Repository, *variant, report.GeneratedAt)
	}
	resolvedOutputDir := *outputDir
	if !filepath.IsAbs(resolvedOutputDir) {
		resolvedOutputDir = filepath.Join(absoluteRoot, resolvedOutputDir)
	}
	jsonPath := filepath.Join(resolvedOutputDir, prefix+".json")
	markdownPath := filepath.Join(resolvedOutputDir, prefix+".md")
	if err := knowledgeevaluation.Write(report, jsonPath, markdownPath); err != nil {
		return err
	}
	return writeKnowledgeEvaluationSummary(stdout, report, jsonPath, markdownPath)
}

func parseKnowledgeEvaluationKs(value string) ([]int, error) {
	seen := make(map[int]struct{})
	result := make([]int, 0)
	for _, raw := range strings.Split(value, ",") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		var parsed int
		if _, err := fmt.Sscanf(raw, "%d", &parsed); err != nil || parsed <= 0 || fmt.Sprintf("%d", parsed) != raw {
			return nil, fmt.Errorf("expected positive integer cutoffs, got %q", raw)
		}
		if _, exists := seen[parsed]; exists {
			continue
		}
		seen[parsed] = struct{}{}
		result = append(result, parsed)
	}
	if len(result) == 0 {
		return nil, errors.New("at least one recall cutoff is required")
	}
	sort.Ints(result)
	return result, nil
}

func defaultKnowledgeEvaluationPrefix(repository, variant string, generated time.Time) string {
	name := strings.ToLower(strings.NewReplacer(" ", "-", "_", "-", "/", "-").Replace(repository))
	variant = strings.ToLower(strings.NewReplacer(" ", "-", "_", "-", "/", "-").Replace(strings.TrimSpace(variant)))
	return fmt.Sprintf("knowledge-%s-%s-%s", name, variant, generated.Format("2006-01-02-150405"))
}

func writeKnowledgeEvaluationSummary(stdout io.Writer, report knowledgeevaluation.Report, jsonPath, markdownPath string) error {
	if _, err := fmt.Fprintln(stdout, "metric\tvalue"); err != nil {
		return err
	}
	for _, metric := range []struct {
		name  string
		value string
	}{
		{"pass_rate", fmt.Sprintf("%.1f%%", report.Aggregate.PassRate*100)},
		{"required_section_recall", fmt.Sprintf("%.1f%%", report.Aggregate.RequiredSectionRecall*100)},
		{"mrr", fmt.Sprintf("%.3f", report.Aggregate.MRR)},
		{"irrelevant_selection_rate", fmt.Sprintf("%.1f%%", report.Aggregate.IrrelevantSelectionRate*100)},
		{"vector_usage_rate", fmt.Sprintf("%.1f%%", report.Aggregate.VectorUsageRate*100)},
		{"vector_error_cases", fmt.Sprintf("%d", report.Aggregate.VectorErrorCases)},
		{"median_latency_ms", fmt.Sprintf("%.1f", report.Aggregate.MedianLatencyMS)},
		{"p95_latency_ms", fmt.Sprintf("%.1f", report.Aggregate.P95LatencyMS)},
	} {
		if _, err := fmt.Fprintf(stdout, "%s\t%s\n", metric.name, metric.value); err != nil {
			return err
		}
	}
	for _, metric := range report.Aggregate.RecallAtK {
		if _, err := fmt.Fprintf(stdout, "recall_at_%d\t%.1f%%\n", metric.K, metric.Value*100); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(stdout, "json: %s\nmarkdown: %s\n", jsonPath, markdownPath)
	return err
}
