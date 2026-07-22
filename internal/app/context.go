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
	"github.com/Lokee86/grimoire/internal/vectorstore"
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
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *query == "" || *budget <= 0 || *limit <= 0 || *timeout <= 0 {
		return errors.New("--query and positive --budget, --candidate-limit, and --timeout are required")
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
	candidates, err := semanticCandidates(ctx, snapshot, statePath, *query, *endpoint, *enginePath, *limit)
	sources := []string{"vector"}
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "warning: semantic retrieval unavailable; using lexical fallback: %v\n", err)
		candidates = retrieve.Search(snapshot, *query, *limit)
		sources = []string{"lexical"}
	}

	result, err := compiler.Compile(
		*query, *budget, snapshot.Version, snapshot.Tokenizer, sources, candidates,
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

func semanticCandidates(
	ctx context.Context,
	snapshot index.Snapshot,
	statePath string,
	query string,
	endpoint string,
	enginePath string,
	limit int,
) ([]retrieve.Candidate, error) {
	paths := resolveVectorPaths(statePath)
	chunks := snapshot.AllChunks()
	manifest, err := validateVectorSnapshotManifest(paths.Manifest, snapshot, len(chunks))
	if err != nil {
		return nil, err
	}
	library, err := vectorstore.Load(enginePath)
	if err != nil {
		return nil, err
	}
	defer library.Close()
	engine, err := library.OpenSnapshot(paths.Snapshot)
	if err != nil {
		return nil, err
	}
	defer engine.Close()
	info, err := engine.Info()
	if err != nil {
		return nil, err
	}
	if err := validateVectorEngineInfo(manifest, info); err != nil {
		return nil, err
	}
	queryVector, err := embedding.NewClient(endpoint).EmbedQuery(ctx, query)
	if err != nil {
		return nil, err
	}
	if info.Dimensions != len(queryVector) {
		return nil, fmt.Errorf("vector snapshot has %d dimensions, query has %d", info.Dimensions, len(queryVector))
	}
	hits, err := engine.Search(queryVector, limit)
	if err != nil {
		return nil, err
	}
	byID := make(map[string]index.Chunk, len(chunks))
	for _, chunk := range chunks {
		byID[chunk.ID] = chunk
	}
	candidates := make([]retrieve.Candidate, 0, len(hits))
	for rank, hit := range hits {
		chunk, exists := byID[hit.ID]
		if !exists {
			return nil, fmt.Errorf("vector result %s is absent from the prepared index", hit.ID)
		}
		candidates = append(candidates, retrieve.Candidate{
			Chunk: chunk, Score: float64(hit.Score), Source: "vector", Rank: rank + 1,
			Reasons: []string{"semantic vector similarity"},
		})
	}
	return candidates, nil
}
