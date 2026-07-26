package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/Lokee86/grimoire/internal/embedding"
	"github.com/Lokee86/grimoire/internal/knowledge"
	"github.com/Lokee86/grimoire/internal/knowledgevector"
)

func runKnowledgeVector(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return errors.New("expected knowledge vector command: build or info")
	}
	switch args[0] {
	case "build":
		return runKnowledgeVectorBuild(args[1:], stdout, stderr)
	case "info":
		return runKnowledgeVectorInfo(args[1:], stdout, stderr)
	default:
		return fmt.Errorf("unknown knowledge vector command %q", args[0])
	}
}

func runKnowledgeVectorBuild(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("knowledge vector build", flag.ContinueOnError)
	flags.SetOutput(stderr)
	root := flags.String("root", ".", "repository root")
	state := flags.String("state", "", "knowledge state directory")
	endpoint := flags.String("endpoint", embedding.DefaultEndpoint, "OpenAI-compatible embeddings endpoint")
	enginePath := flags.String("engine", "", "Rust vector engine DLL")
	batchSize := flags.Int("batch-size", 8, "knowledge sections embedded per request")
	concurrency := flags.Int("batch-concurrency", 1, "concurrent embedding requests")
	timeout := flags.Duration("timeout", 30*time.Minute, "complete knowledge vector build timeout")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *batchSize <= 0 || *concurrency <= 0 || *timeout <= 0 {
		return errors.New("--batch-size, --batch-concurrency, and --timeout must be positive")
	}
	statePath, err := resolveKnowledgeState(*root, *state)
	if err != nil {
		return err
	}
	index, err := knowledge.Load(statePath)
	if err != nil {
		return fmt.Errorf("load knowledge index: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	result, err := knowledgevector.Build(ctx, index, knowledgevector.BuildOptions{
		State: statePath, Endpoint: *endpoint, EnginePath: *enginePath,
		BatchSize: *batchSize, Concurrency: *concurrency,
	})
	if err != nil {
		return err
	}
	return writeJSON(stdout, result)
}

func runKnowledgeVectorInfo(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("knowledge vector info", flag.ContinueOnError)
	flags.SetOutput(stderr)
	root := flags.String("root", ".", "repository root")
	state := flags.String("state", "", "knowledge state directory")
	enginePath := flags.String("engine", "", "Rust vector engine DLL")
	if err := flags.Parse(args); err != nil {
		return err
	}
	statePath, err := resolveKnowledgeState(*root, *state)
	if err != nil {
		return err
	}
	index, err := knowledge.Load(statePath)
	if err != nil {
		return fmt.Errorf("load knowledge index: %w", err)
	}
	return writeJSON(stdout, knowledgevector.Inspect(statePath, index, *enginePath))
}
