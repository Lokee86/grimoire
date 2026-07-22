package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Lokee86/grimoire/internal/embedding"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/vectorstore"
)

func runVector(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return errors.New("expected vector command: build, search, or info")
	}
	switch args[0] {
	case "build":
		return runVectorBuild(args[1:], stdout, stderr)
	case "search":
		return runVectorSearch(args[1:], stdout, stderr)
	case "info":
		return runVectorInfo(args[1:], stdout, stderr)
	default:
		return fmt.Errorf("unknown vector command %q", args[0])
	}
}

func runVectorInfo(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("vector info", flag.ContinueOnError)
	flags.SetOutput(stderr)
	root := flags.String("root", ".", "repository root")
	state := flags.String("state", "", "prepared index repository path")
	enginePath := flags.String("engine", "", "Rust vector engine DLL")
	if err := flags.Parse(args); err != nil {
		return err
	}
	statePath, err := resolveState(*root, *state)
	if err != nil {
		return err
	}
	paths := resolveVectorPaths(statePath)
	response := struct {
		Engine          string           `json:"engine,omitempty"`
		EngineAvailable bool             `json:"engine_available"`
		Snapshot        string           `json:"snapshot"`
		SnapshotExists  bool             `json:"snapshot_exists"`
		Info            vectorstore.Info `json:"info,omitempty"`
		Error           string           `json:"error,omitempty"`
	}{Snapshot: paths.Snapshot}
	if _, statErr := os.Stat(paths.Snapshot); statErr == nil {
		response.SnapshotExists = true
	}
	resolved, findErr := vectorstore.FindLibrary(*enginePath)
	if findErr != nil {
		response.Error = findErr.Error()
		return writeJSON(stdout, response)
	}
	response.Engine, response.EngineAvailable = resolved, true
	if !response.SnapshotExists {
		return writeJSON(stdout, response)
	}
	library, err := vectorstore.Load(resolved)
	if err != nil {
		return err
	}
	defer library.Close()
	engine, err := library.OpenSnapshot(paths.Snapshot)
	if err != nil {
		return err
	}
	defer engine.Close()
	response.Info, err = engine.Info()
	if err != nil {
		return err
	}
	return writeJSON(stdout, response)
}

func runVectorSearch(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("vector search", flag.ContinueOnError)
	flags.SetOutput(stderr)
	root := flags.String("root", ".", "repository root")
	state := flags.String("state", "", "prepared index repository path")
	query := flags.String("query", "", "semantic repository query")
	topK := flags.Int("top-k", 20, "maximum vector results")
	endpoint := flags.String("endpoint", embedding.DefaultEndpoint, "OpenAI-compatible embeddings endpoint")
	enginePath := flags.String("engine", "", "Rust vector engine DLL")
	timeout := flags.Duration("timeout", 2*time.Minute, "query embedding timeout")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *query == "" || *topK <= 0 || *timeout <= 0 {
		return errors.New("--query, positive --top-k, and positive --timeout are required")
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
	queryVector, err := embedding.NewClient(*endpoint).EmbedQuery(ctx, *query)
	if err != nil {
		return err
	}

	library, err := vectorstore.Load(*enginePath)
	if err != nil {
		return err
	}
	defer library.Close()
	engine, err := library.OpenSnapshot(resolveVectorPaths(statePath).Snapshot)
	if err != nil {
		return err
	}
	defer engine.Close()
	info, err := engine.Info()
	if err != nil {
		return err
	}
	if info.Model != embedding.Identity() || info.Dimensions != len(queryVector) {
		return fmt.Errorf("vector snapshot uses %s/%dd, expected %s/%dd", info.Model, info.Dimensions, embedding.Identity(), len(queryVector))
	}
	hits, err := engine.Search(queryVector, *topK)
	if err != nil {
		return err
	}
	chunks := make(map[string]index.Chunk, len(snapshot.AllChunks()))
	for _, chunk := range snapshot.AllChunks() {
		chunks[chunk.ID] = chunk
	}
	type result struct {
		ID        string  `json:"id"`
		Path      string  `json:"path"`
		StartLine int     `json:"start_line"`
		EndLine   int     `json:"end_line"`
		Score     float32 `json:"score"`
	}
	results := make([]result, 0, len(hits))
	for _, hit := range hits {
		chunk, exists := chunks[hit.ID]
		if !exists {
			continue
		}
		results = append(results, result{ID: hit.ID, Path: chunk.Path, StartLine: chunk.StartLine, EndLine: chunk.EndLine, Score: hit.Score})
	}
	return writeJSON(stdout, struct {
		Model   string   `json:"model"`
		Results []result `json:"results"`
	}{Model: info.Model, Results: results})
}
