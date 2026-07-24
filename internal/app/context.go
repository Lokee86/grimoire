package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Lokee86/grimoire/internal/assembly"
	"github.com/Lokee86/grimoire/internal/compiler"
	"github.com/Lokee86/grimoire/internal/embedding"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/queryshape"
	"github.com/Lokee86/grimoire/internal/selection"
)

func runContext(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("context", flag.ContinueOnError)
	flags.SetOutput(stderr)
	root := flags.String("root", ".", "repository root")
	state := flags.String("state", "", "prepared index repository path")
	query := flags.String("query", "", "task or retrieval query")
	budget := flags.Int("budget", 0, "maximum o200k_base tokens in the emitted package; zero selects automatically")
	limit := flags.Int("candidate-limit", 200, "maximum ranked candidates")
	endpoint := flags.String("endpoint", embedding.DefaultEndpoint, "OpenAI-compatible embeddings endpoint")
	enginePath := flags.String("engine", "", "Rust vector engine DLL")
	structureEnabled := flags.Bool("structure", true, "include available structural evidence")
	structuralProviders := flags.String("structural-providers", "lexicon,arcana", "structural evidence providers: none, lexicon, arcana, or lexicon,arcana")
	lexiconFacts := flags.String("lexicon-facts", "", "explicit directory containing exported Lexicon JSONL libraries")
	lexiconState := flags.String("lexicon-state", "", "Lexicon state directory; defaults to <root>/.lexicon")
	lexiconCommand := flags.String("lexicon-command", "lexicon", "Lexicon executable used to export the current snapshot")
	arcanaState := flags.String("arcana-state", "", "Arcana state directory; defaults to <root>/.arcana")
	arcanaCommand := flags.String("arcana-command", "arcana", "Arcana executable used to synchronize and query graph state")
	structureTimeout := flags.Duration("structure-timeout", 30*time.Second, "complete structural-provider timeout")
	timeout := flags.Duration("timeout", 2*time.Second, "semantic retrieval timeout")
	modeValue := flags.String("query-embedding-mode", string(embedding.QueryModeFast), "query embedding mode: fast, full, or quality")
	windowTokens := flags.Int("query-window-tokens", embedding.DefaultQueryWindowTokens, "tokens per fast query window")
	batchTokens := flags.Int("query-batch-tokens", embedding.DefaultQueryBatchTokens, "maximum split-query tokens per embedding request")
	batchConcurrency := flags.Int("query-batch-concurrency", embedding.DefaultQueryBatchConcurrency, "maximum concurrent query embedding requests")
	maxTokens := flags.Int("query-max-tokens", embedding.DefaultQueryMaxTokens, "optional maximum query tokens embedded; zero means unlimited")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *query == "" || *budget < 0 || *limit <= 0 || *timeout <= 0 || *structureTimeout <= 0 {
		return errors.New("--query, non-negative --budget, and positive --candidate-limit, --timeout, and --structure-timeout are required")
	}
	emitLexicon, arcanaEnabled, err := parseContextStructuralProviders(*structuralProviders)
	if err != nil {
		return err
	}
	if !*structureEnabled {
		emitLexicon = false
		arcanaEnabled = false
	}
	mode, err := embedding.ParseQueryMode(*modeValue)
	if err != nil {
		return err
	}
	queryOptions := embedding.QueryOptions{
		Mode: mode, WindowTokens: *windowTokens, BatchTokens: *batchTokens,
		BatchConcurrency: *batchConcurrency, MaxTokens: *maxTokens,
	}
	if err := queryOptions.Validate(); err != nil {
		return err
	}

	statePath, err := resolveState(*root, *state)
	if err != nil {
		return err
	}
	snapshot, err := index.Load(statePath)
	if err != nil {
		return fmt.Errorf("load prepared index: %w", err)
	}

	intents := activeRetrievalIntents(*query)
	lexicalCandidates := intentLexicalCandidates(snapshot, intents, *limit)
	semanticContext, cancelSemantic := context.WithTimeout(context.Background(), *timeout)
	semanticCandidates, err := semanticCandidates(
		semanticContext, snapshot, statePath, *query, *endpoint, *enginePath, *limit, queryOptions,
	)
	cancelSemantic()
	baseCandidates := lexicalCandidates
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "warning: semantic retrieval unavailable; using lexical fallback (BM25): %v\n", err)
	} else {
		semanticCandidates = rankCandidatesForIntent(semanticCandidates, intents[0], false)
		baseCandidates = mergeRankedProviders(*limit, lexicalCandidates, semanticCandidates)
	}

	structuralIntent := structuralRetrievalIntent(*query, intents)
	structural := collectStructuralContext(context.Background(), snapshot, structuralIntent.Query, structuralContextOptions{
		Enabled: emitLexicon || arcanaEnabled, ArcanaEnabled: arcanaEnabled, EmitLexicon: emitLexicon,
		Root: *root, GrimoireState: statePath, LexiconFacts: *lexiconFacts,
		LexiconState: *lexiconState, LexiconCommand: *lexiconCommand,
		ArcanaState: *arcanaState, ArcanaCommand: *arcanaCommand,
		Limit: *limit, Timeout: *structureTimeout,
	})
	structural = annotateStructuralIntent(structural, structuralIntent)
	for _, warning := range structural.Warnings {
		_, _ = fmt.Fprintf(stderr, "warning: %s\n", warning)
	}

	exact := intentExactCandidates(snapshot, intents, min(*limit, maxExactCandidates))
	lexiconCandidates := structural.Lexicon.Candidates
	if !emitLexicon {
		lexiconCandidates = nil
	}
	merged := mergeContextProviders(*limit, exact, baseCandidates, lexiconCandidates)
	_, policy := queryshape.Analyze(queryshape.Input{
		Query: *query, RequestedBudget: *budget,
		Exact: exact, Ranked: baseCandidates, Candidates: merged, Structural: structural.Combined,
	})
	effectiveBudget := *budget
	automatic := effectiveBudget == 0
	if automatic {
		policy = queryshape.Activate(policy)
		effectiveBudget = policy.TargetTokens
	}
	candidates := selection.Curate(snapshot, merged)
	evidence := structural.Combined

	var result compiler.Package
	if automatic {
		planned := assembly.Plan(policy, candidates, evidence)
		candidates = planned.Candidates
		evidence = planned.Structural
		result, err = compiler.CompileAdaptiveWithEvidence(
			*query, effectiveBudget, snapshot.Version, snapshot.Tokenizer,
			contextCandidateSources(candidates), structural.ProviderState, evidence,
			planned.Decision, candidates,
		)
	} else {
		result, err = compiler.CompileWithEvidence(
			*query, effectiveBudget, snapshot.Version, snapshot.Tokenizer,
			contextCandidateSources(candidates), structural.ProviderState, evidence, candidates,
		)
	}
	if err != nil {
		return err
	}
	data, err := compiler.Marshal(result)
	if err != nil {
		return err
	}
	_, err = stdout.Write(data)
	return err
}

func parseContextStructuralProviders(value string) (emitLexicon, arcanaEnabled bool, err error) {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" || value == "none" {
		return false, false, nil
	}
	seen := make(map[string]struct{})
	for _, provider := range strings.Split(value, ",") {
		provider = strings.TrimSpace(provider)
		if provider == "" {
			continue
		}
		if _, exists := seen[provider]; exists {
			continue
		}
		seen[provider] = struct{}{}
		switch provider {
		case "lexicon":
			emitLexicon = true
		case "arcana":
			arcanaEnabled = true
		default:
			return false, false, fmt.Errorf("unsupported structural provider %q", provider)
		}
	}
	return emitLexicon, arcanaEnabled, nil
}
