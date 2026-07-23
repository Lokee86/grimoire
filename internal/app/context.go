package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/Lokee86/grimoire/internal/compiler"
	"github.com/Lokee86/grimoire/internal/embedding"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/retrieve"
	"github.com/Lokee86/grimoire/internal/selection"
)

func runContext(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("context", flag.ContinueOnError)
	flags.SetOutput(stderr)
	root := flags.String("root", ".", "repository root")
	state := flags.String("state", "", "prepared index repository path")
	query := flags.String("query", "", "task or retrieval query")
	budget := flags.Int("budget", 2000, "maximum o200k_base tokens in the emitted package")
	limit := flags.Int("candidate-limit", 200, "maximum ranked candidates")
	endpoint := flags.String("endpoint", embedding.DefaultEndpoint, "OpenAI-compatible embeddings endpoint")
	enginePath := flags.String("engine", "", "Rust vector engine DLL")
	structureEnabled := flags.Bool("structure", true, "include available Lexicon and Arcana structural evidence")
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
	if *query == "" || *budget <= 0 || *limit <= 0 || *timeout <= 0 || *structureTimeout <= 0 {
		return errors.New("--query and positive --budget, --candidate-limit, --timeout, and --structure-timeout are required")
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

	semanticContext, cancelSemantic := context.WithTimeout(context.Background(), *timeout)
	baseCandidates, err := semanticCandidates(
		semanticContext, snapshot, statePath, *query, *endpoint, *enginePath, *limit, queryOptions,
	)
	cancelSemantic()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "warning: semantic retrieval unavailable; using lexical fallback: %v\n", err)
		baseCandidates = retrieve.Search(snapshot, *query, *limit)
	}

	structural := collectStructuralContext(context.Background(), snapshot, *query, structuralContextOptions{
		Enabled: *structureEnabled, ArcanaEnabled: *structureEnabled,
		Root: *root, GrimoireState: statePath, LexiconFacts: *lexiconFacts,
		LexiconState: *lexiconState, LexiconCommand: *lexiconCommand,
		ArcanaState: *arcanaState, ArcanaCommand: *arcanaCommand,
		Limit: *limit, Timeout: *structureTimeout,
	})
	for _, warning := range structural.Warnings {
		_, _ = fmt.Fprintf(stderr, "warning: %s\n", warning)
	}

	exact := retrieve.Exact(snapshot, *query, min(*limit, maxExactCandidates))
	merged := mergeContextProviders(*limit, exact, baseCandidates, structural.Lexicon.Candidates)
	candidates := selection.Curate(snapshot, merged)

	result, err := compiler.CompileWithEvidence(
		*query, *budget, snapshot.Version, snapshot.Tokenizer,
		contextCandidateSources(candidates), structural.ProviderState, structural.Combined, candidates,
	)
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
