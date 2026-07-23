package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Lokee86/grimoire/internal/compiler"
	"github.com/Lokee86/grimoire/internal/embedding"
	"github.com/Lokee86/grimoire/internal/evaluation"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/structure"
)

var evaluationModes = []string{"fast", "full", "quality", "lexical"}

func runEval(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 || args[0] != "retrieval" {
		return errors.New("expected evaluation command: retrieval")
	}
	flags := flag.NewFlagSet("eval retrieval", flag.ContinueOnError)
	flags.SetOutput(stderr)
	casesPath := flags.String("cases", "", "judged retrieval corpus JSON")
	root := flags.String("root", ".", "repository root")
	state := flags.String("state", "", "prepared index repository path")
	modesValue := flags.String("modes", strings.Join(evaluationModes, ","), "comma-separated modes: fast, full, quality, lexical")
	variant := flags.String("variant", "standalone", "evaluation variant label")
	budgetOverride := flags.Int("budget", 0, "override every case token budget")
	limit := flags.Int("candidate-limit", 200, "maximum ranked candidates")
	probeLimit := flags.Int("probe-limit", 800, "broader diagnostic ranking probe")
	endpoint := flags.String("endpoint", embedding.DefaultEndpoint, "OpenAI-compatible embeddings endpoint")
	enginePath := flags.String("engine", "", "Rust vector engine DLL")
	structuralProvidersValue := flags.String("structural-providers", "none", "structural providers: none, lexicon, or lexicon,arcana")
	lexiconFacts := flags.String("lexicon-facts", "", "explicit directory containing exported Lexicon JSONL libraries")
	lexiconState := flags.String("lexicon-state", "", "Lexicon state directory; defaults to <root>/.lexicon")
	lexiconCommand := flags.String("lexicon-command", "lexicon", "Lexicon executable used to export immutable state")
	arcanaState := flags.String("arcana-state", "", "Arcana state directory; defaults to <root>/.arcana")
	arcanaCommand := flags.String("arcana-command", "arcana", "Arcana executable used to synchronize and query graph state")
	structureTimeout := flags.Duration("structure-timeout", 30*time.Second, "per-case structural-provider timeout")
	timeout := flags.Duration("timeout", 10*time.Second, "per-case retrieval timeout")
	outputDir := flags.String("output-dir", "evaluation/results", "result directory")
	outputPrefix := flags.String("output-prefix", "", "result filename prefix")
	windowTokens := flags.Int("query-window-tokens", embedding.DefaultQueryWindowTokens, "tokens per fast query window")
	batchTokens := flags.Int("query-batch-tokens", embedding.DefaultQueryBatchTokens, "maximum split-query tokens per embedding request")
	batchConcurrency := flags.Int("query-batch-concurrency", embedding.DefaultQueryBatchConcurrency, "maximum concurrent query embedding requests")
	maxTokens := flags.Int("query-max-tokens", embedding.DefaultQueryMaxTokens, "optional maximum query tokens embedded; zero means unlimited")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if strings.TrimSpace(*casesPath) == "" || *limit <= 0 || *probeLimit <= 0 || *timeout <= 0 || *structureTimeout <= 0 {
		return errors.New("--cases and positive --candidate-limit, --probe-limit, --timeout, and --structure-timeout are required")
	}
	modes, err := parseEvaluationModes(*modesValue)
	if err != nil {
		return err
	}
	structuralProviders, structureEnabled, arcanaEnabled, err := parseStructuralProviders(*structuralProvidersValue)
	if err != nil {
		return err
	}
	corpus, err := evaluation.LoadCorpus(*casesPath)
	if err != nil {
		return err
	}
	absoluteRoot, err := filepath.Abs(*root)
	if err != nil {
		return fmt.Errorf("resolve evaluation root: %w", err)
	}
	statePath, err := resolveState(absoluteRoot, *state)
	if err != nil {
		return err
	}
	snapshot, err := index.Load(statePath)
	if err != nil {
		return fmt.Errorf("load prepared index: %w", err)
	}
	actualProbeLimit := min(*probeLimit, len(snapshot.AllChunks()))
	if actualProbeLimit < *limit {
		actualProbeLimit = min(*limit, len(snapshot.AllChunks()))
	}

	report := evaluation.Report{
		Version:             evaluation.FormatVersion,
		GeneratedAt:         time.Now(),
		Repository:          corpus.Repository,
		SourceURL:           corpus.SourceURL,
		Revision:            corpus.Revision,
		Scope:               corpus.Scope,
		JudgedAt:            corpus.JudgedAt,
		Root:                absoluteRoot,
		CasesFile:           *casesPath,
		State:               statePath,
		Variant:             *variant,
		Modes:               modes,
		StructuralProviders: structuralProviders,
	}
	for _, entry := range corpus.Cases {
		expectationErr := validateEvaluationCase(absoluteRoot, entry)
		budget := entry.Budget
		if *budgetOverride > 0 {
			budget = *budgetOverride
		}
		for _, mode := range modes {
			run := evaluation.CaseRun{
				CaseID:   entry.ID,
				Category: entry.Category,
				Mode:     mode,
				Variant:  *variant,
				Budget:   budget,
			}
			if expectationErr != nil {
				run.Error = expectationErr.Error()
				evaluation.ScoreCase(entry, &run, evaluation.Stages{})
				applyExpectationError(&run)
				report.Runs = append(report.Runs, run)
				continue
			}
			queryOptions := embedding.DefaultQueryOptions()
			queryOptions.WindowTokens = *windowTokens
			queryOptions.BatchTokens = *batchTokens
			queryOptions.BatchConcurrency = *batchConcurrency
			queryOptions.MaxTokens = *maxTokens
			if mode != "lexical" {
				queryMode, parseErr := embedding.ParseQueryMode(mode)
				if parseErr != nil {
					return parseErr
				}
				queryOptions.Mode = queryMode
			}
			ctx, cancel := context.WithTimeout(context.Background(), *timeout)
			runStart := time.Now()
			executed, executeErr := evaluateContext(ctx, snapshot, evaluatedContextOptions{
				Mode:       mode,
				Query:      entry.Query,
				Budget:     budget,
				Limit:      *limit,
				ProbeLimit: actualProbeLimit,
				StatePath:  statePath,
				Endpoint:   *endpoint,
				EnginePath: *enginePath,
				Structural: structuralContextOptions{
					Enabled: structureEnabled, ArcanaEnabled: arcanaEnabled,
					Root: absoluteRoot, GrimoireState: statePath, LexiconFacts: *lexiconFacts,
					LexiconState: *lexiconState, LexiconCommand: *lexiconCommand,
					ArcanaState: *arcanaState, ArcanaCommand: *arcanaCommand,
					Limit: *limit, Timeout: *structureTimeout,
				},
				QueryOptions: queryOptions,
			})
			cancel()
			run.Timings = executed.Timings
			run.QueryProfile = executed.QueryProfile
			run.RetrievalPolicy = executed.RetrievalPolicy
			if run.Timings.TotalMS == 0 {
				run.Timings.TotalMS = durationMS(time.Since(runStart))
			}
			if executeErr != nil {
				run.Warnings = append([]string(nil), executed.Warnings...)
				run.Error = executeErr.Error()
				evaluation.ScoreCase(entry, &run, executed.Stages)
				applyEvaluationErrorClassification(&run)
				report.Runs = append(report.Runs, run)
				continue
			}
			run.Warnings = append([]string(nil), executed.Warnings...)
			run.RetrievalSources = append([]string(nil), executed.Package.RetrievalSources...)
			run.StructuralSources = append([]string(nil), executed.Package.StructuralSources...)
			run.StructuralState = append([]structure.ProviderState(nil), executed.Package.StructuralState...)
			run.FinalPackageTokens = executed.Package.TokenCount
			run.CandidateCount = len(executed.Stages.Merged)
			run.CuratedCount = len(executed.Stages.Curated)
			run.OmittedForBudget = executed.Package.OmittedForBudget
			run.OmittedStructuralForBudget = executed.Package.OmittedStructuralForBudget
			run.Selections = packageSelections(entry, executed.Package.Selections)
			run.StructuralSelections = packageStructuralSelections(executed.Package.StructuralEvidence)
			run.SelectedPaths = selectedPaths(run.Selections)
			evaluation.ScoreCase(entry, &run, executed.Stages)
			report.Runs = append(report.Runs, run)
		}
	}
	evaluation.BuildAggregates(&report)

	prefix := strings.TrimSpace(*outputPrefix)
	if prefix == "" {
		prefix = defaultEvaluationPrefix(corpus.Repository, *variant, report.GeneratedAt)
	}
	resolvedOutputDir := *outputDir
	if !filepath.IsAbs(resolvedOutputDir) {
		resolvedOutputDir = filepath.Join(absoluteRoot, resolvedOutputDir)
	}
	jsonPath := filepath.Join(resolvedOutputDir, prefix+".json")
	markdownPath := filepath.Join(resolvedOutputDir, prefix+".md")
	if err := evaluation.Write(report, jsonPath, markdownPath); err != nil {
		return err
	}
	if err := writeEvaluationSummary(stdout, report, jsonPath, markdownPath); err != nil {
		return err
	}
	return nil
}

func validateEvaluationCase(root string, entry evaluation.Case) error {
	for _, group := range [][]evaluation.Evidence{entry.Required, entry.Supporting, entry.Forbidden} {
		for _, evidence := range group {
			if err := validateExpectedSymbol(root, entry.ID, evidence.Path, evidence.Symbols...); err != nil {
				return err
			}
		}
	}
	for _, group := range [][]evaluation.StructuralExpectation{
		entry.RequiredStructural, entry.SupportingStructural, entry.ForbiddenStructural,
	} {
		for _, expected := range group {
			if expected.Path != "" && expected.Symbol != "" {
				if err := validateExpectedSymbol(root, entry.ID, expected.Path, expected.Symbol); err != nil {
					return err
				}
			} else if expected.Path != "" {
				if err := validateExpectedSymbol(root, entry.ID, expected.Path); err != nil {
					return err
				}
			}
			if expected.TargetPath != "" && expected.TargetSymbol != "" {
				if err := validateExpectedSymbol(root, entry.ID, expected.TargetPath, expected.TargetSymbol); err != nil {
					return err
				}
			} else if expected.TargetPath != "" {
				if err := validateExpectedSymbol(root, entry.ID, expected.TargetPath); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func validateExpectedSymbol(root, caseID, relativePath string, symbols ...string) error {
	path := filepath.Join(root, filepath.FromSlash(relativePath))
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("incorrect evaluation expectation for %s: read %s: %w", caseID, relativePath, err)
	}
	for _, symbol := range symbols {
		if !strings.Contains(string(data), symbol) {
			return fmt.Errorf("incorrect evaluation expectation for %s: symbol %s is absent from %s", caseID, symbol, relativePath)
		}
	}
	return nil
}

func applyExpectationError(run *evaluation.CaseRun) {
	run.FailureClassifications = []string{evaluation.FailureIncorrectExpectation}
	run.RequiredNeverRetrieved = len(run.Required)
	run.RequiredStructuralNeverProduced = len(run.RequiredStructural)
	for index := range run.Required {
		run.Required[index].FailureStage = evaluation.FailureIncorrectExpectation
	}
	for index := range run.RequiredStructural {
		run.RequiredStructural[index].FailureStage = evaluation.FailureIncorrectExpectation
	}
}

func parseEvaluationModes(value string) ([]string, error) {
	allowed := make(map[string]struct{}, len(evaluationModes))
	for _, mode := range evaluationModes {
		allowed[mode] = struct{}{}
	}
	seen := make(map[string]struct{})
	var result []string
	for _, raw := range strings.Split(value, ",") {
		mode := strings.ToLower(strings.TrimSpace(raw))
		if mode == "" {
			continue
		}
		if _, valid := allowed[mode]; !valid {
			return nil, fmt.Errorf("unknown evaluation mode %q", mode)
		}
		if _, duplicate := seen[mode]; duplicate {
			continue
		}
		seen[mode] = struct{}{}
		result = append(result, mode)
	}
	if len(result) == 0 {
		return nil, errors.New("at least one evaluation mode is required")
	}
	return result, nil
}

func parseStructuralProviders(value string) ([]string, bool, bool, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || value == "none" {
		return nil, false, false, nil
	}
	seen := make(map[string]struct{})
	var providers []string
	for _, raw := range strings.Split(value, ",") {
		provider := strings.TrimSpace(raw)
		if provider == "" {
			continue
		}
		if provider != "lexicon" && provider != "arcana" {
			return nil, false, false, fmt.Errorf("unknown structural provider %q", provider)
		}
		if _, duplicate := seen[provider]; duplicate {
			continue
		}
		seen[provider] = struct{}{}
		providers = append(providers, provider)
	}
	if _, arcana := seen["arcana"]; arcana {
		if _, lexicon := seen["lexicon"]; !lexicon {
			return nil, false, false, errors.New("Arcana evaluation requires Lexicon")
		}
	}
	_, lexicon := seen["lexicon"]
	_, arcana := seen["arcana"]
	return providers, lexicon, arcana, nil
}

func packageStructuralSelections(evidence []structure.Evidence) []evaluation.StructuralSelection {
	result := make([]evaluation.StructuralSelection, 0, len(evidence))
	for _, item := range evidence {
		result = append(result, evaluation.StructuralSelection{Evidence: item})
	}
	return result
}

func packageSelections(entry evaluation.Case, selections []compiler.Selection) []evaluation.Selection {
	result := make([]evaluation.Selection, 0, len(selections))
	for _, selected := range selections {
		result = append(result, evaluation.Selection{
			Path:            selected.Path,
			StartLine:       selected.StartLine,
			EndLine:         selected.EndLine,
			Symbols:         detectedSymbols(entry, selected.Path, selected.Content),
			RetrievalSource: selected.RetrievalSource,
			ProviderRank:    selected.RetrievalRank,
			TokenCount:      selected.TokenCount,
		})
	}
	return result
}

func detectedSymbols(entry evaluation.Case, path, content string) []string {
	seen := make(map[string]struct{})
	var symbols []string
	for _, group := range [][]evaluation.Evidence{entry.Required, entry.Supporting, entry.Forbidden} {
		for _, evidence := range group {
			if filepath.ToSlash(evidence.Path) != filepath.ToSlash(path) {
				continue
			}
			for _, symbol := range evidence.Symbols {
				if !strings.Contains(content, symbol) {
					continue
				}
				if _, exists := seen[symbol]; exists {
					continue
				}
				seen[symbol] = struct{}{}
				symbols = append(symbols, symbol)
			}
		}
	}
	sort.Strings(symbols)
	return symbols
}

func selectedPaths(selections []evaluation.Selection) []string {
	seen := make(map[string]struct{})
	var paths []string
	for _, selection := range selections {
		if _, exists := seen[selection.Path]; exists {
			continue
		}
		seen[selection.Path] = struct{}{}
		paths = append(paths, selection.Path)
	}
	return paths
}

func applyEvaluationErrorClassification(run *evaluation.CaseRun) {
	classification := evaluation.FailureEmbeddingMiss
	message := strings.ToLower(run.Error)
	if strings.Contains(message, "manifest") || strings.Contains(message, "snapshot") ||
		strings.Contains(message, "prepared index") || strings.Contains(message, "vector result") {
		classification = evaluation.FailureStaleOrIncompleteIndex
	}
	run.FailureClassifications = []string{classification}
	run.RequiredNeverRetrieved = len(run.Required)
	for index := range run.Required {
		if !run.Required[index].Included {
			run.Required[index].FailureStage = classification
		}
	}
}

func defaultEvaluationPrefix(repository, variant string, generated time.Time) string {
	name := strings.ToLower(repository)
	name = strings.NewReplacer(" ", "-", "_", "-", "/", "-").Replace(name)
	variant = strings.ToLower(strings.TrimSpace(variant))
	variant = strings.NewReplacer(" ", "-", "_", "-", "/", "-").Replace(variant)
	return fmt.Sprintf("%s-%s-%s", name, variant, generated.Format("2006-01-02-150405"))
}

func writeEvaluationSummary(stdout io.Writer, report evaluation.Report, jsonPath, markdownPath string) error {
	if _, err := fmt.Fprintln(stdout, "mode\tpass\tsource_required\tr_at_10\tr_at_20\tmrr\tsource_irrelevant\tmedian_ms\tp95_ms"); err != nil {
		return err
	}
	for _, aggregate := range report.ByMode {
		if _, err := fmt.Fprintf(stdout, "%s\t%.1f%%\t%.1f%%\t%.1f%%\t%.1f%%\t%.3f\t%.1f%%\t%.1f\t%.1f\n",
			aggregate.Group, aggregate.PassRate*100, aggregate.RequiredEvidenceRecall*100,
			aggregate.RequiredRecallAt10*100, aggregate.RequiredRecallAt20*100,
			aggregate.MeanReciprocalRank, aggregate.IrrelevantSelectionRate*100,
			aggregate.MedianLatencyMS, aggregate.P95LatencyMS); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(stdout, "json: %s\nmarkdown: %s\n", jsonPath, markdownPath); err != nil {
		return err
	}
	return nil
}
