package app

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Lokee86/grimoire/internal/compiler"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/retrieve"
)

const Version = "0.1.0-dev"

func Run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return errors.New("expected command: index, context, or version")
	}

	switch args[0] {
	case "index":
		return runIndex(args[1:], stdout, stderr)
	case "context":
		return runContext(args[1:], stdout, stderr)
	case "version":
		_, err := fmt.Fprintln(stdout, Version)
		return err
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runIndex(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("index", flag.ContinueOnError)
	flags.SetOutput(stderr)
	root := flags.String("root", ".", "repository root")
	state := flags.String("state", "", "prepared index path")
	maxFileBytes := flags.Int64("max-file-bytes", 0, "maximum indexed file size")
	if err := flags.Parse(args); err != nil {
		return err
	}

	statePath, err := resolveState(*root, *state)
	if err != nil {
		return err
	}
	previous, err := loadOptional(statePath)
	if err != nil {
		return err
	}

	snapshot, stats, err := index.Build(*root, previous, index.BuildOptions{MaxFileBytes: *maxFileBytes})
	if err != nil {
		return err
	}
	if err := index.Save(statePath, snapshot); err != nil {
		return err
	}

	response := struct {
		State string           `json:"state"`
		Files int              `json:"files"`
		Stats index.BuildStats `json:"stats"`
	}{State: statePath, Files: len(snapshot.Files), Stats: stats}
	return writeJSON(stdout, response)
}

func runContext(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("context", flag.ContinueOnError)
	flags.SetOutput(stderr)
	root := flags.String("root", ".", "repository root")
	state := flags.String("state", "", "prepared index path")
	query := flags.String("query", "", "task or retrieval query")
	budget := flags.Int("budget", 2000, "estimated content-token budget")
	limit := flags.Int("candidate-limit", 200, "maximum ranked candidates")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *query == "" {
		return errors.New("--query is required")
	}
	if *budget <= 0 {
		return errors.New("--budget must be positive")
	}

	statePath, err := resolveState(*root, *state)
	if err != nil {
		return err
	}
	snapshot, err := index.Load(statePath)
	if err != nil {
		return fmt.Errorf("load prepared index: %w", err)
	}
	candidates := retrieve.Search(snapshot, *query, *limit)
	result := compiler.Compile(*query, *budget, snapshot.Version, candidates)
	return writeJSON(stdout, result)
}

func resolveState(root, state string) (string, error) {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve root: %w", err)
	}
	if state == "" {
		return filepath.Join(absoluteRoot, ".grimoire", "index.json"), nil
	}
	if filepath.IsAbs(state) {
		return filepath.Clean(state), nil
	}
	return filepath.Join(absoluteRoot, state), nil
}

func loadOptional(path string) (*index.Snapshot, error) {
	snapshot, err := index.Load(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("load existing index: %w", err)
	}
	return &snapshot, nil
}

func writeJSON(writer io.Writer, value any) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}
