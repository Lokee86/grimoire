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
	timeout := flags.Duration("timeout", 2*time.Second, "semantic retrieval timeout")
	modeValue := flags.String("query-embedding-mode", string(embedding.QueryModeFast), "query embedding mode: fast, full, or quality")
	windowTokens := flags.Int("query-window-tokens", embedding.DefaultQueryWindowTokens, "tokens per fast query window")
	batchTokens := flags.Int("query-batch-tokens", embedding.DefaultQueryBatchTokens, "maximum split-query tokens per embedding request")
	batchConcurrency := flags.Int("query-batch-concurrency", embedding.DefaultQueryBatchConcurrency, "maximum concurrent query embedding requests")
	maxTokens := flags.Int("query-max-tokens", embedding.DefaultQueryMaxTokens, "optional maximum query tokens embedded; zero means unlimited")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *query == "" || *budget <= 0 || *limit <= 0 || *timeout <= 0 {
		return errors.New("--query and positive --budget, --candidate-limit, and --timeout are required")
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

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	baseCandidates, err := semanticCandidates(
		ctx, snapshot, statePath, *query, *endpoint, *enginePath, *limit, queryOptions,
	)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "warning: semantic retrieval unavailable; using lexical fallback: %v\n", err)
		baseCandidates = retrieve.Search(snapshot, *query, *limit)
	}
	candidates := curateContextCandidates(snapshot, *query, baseCandidates, *limit)

	result, err := compiler.Compile(
		*query, *budget, snapshot.Version, snapshot.Tokenizer,
		contextCandidateSources(candidates), candidates,
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
