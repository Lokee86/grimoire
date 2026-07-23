package app

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Lokee86/grimoire/internal/index"
)

const Version = "0.1.0-dev"

type stringListFlag []string

func (values *stringListFlag) String() string {
	return strings.Join(*values, ",")
}

func (values *stringListFlag) Set(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return errors.New("exclude path must not be empty")
	}
	*values = append(*values, value)
	return nil
}

func Run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return errors.New("expected command: index, context, eval, model, vector, or version")
	}

	switch args[0] {
	case "index":
		return runIndex(args[1:], stdout, stderr)
	case "context":
		return runContext(args[1:], stdout, stderr)
	case "eval":
		return runEval(args[1:], stdout, stderr)
	case "model":
		return runModel(args[1:], stdout, stderr)
	case "vector":
		return runVector(args[1:], stdout, stderr)
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
	state := flags.String("state", "", "prepared index repository path")
	ignoreFile := flags.String("ignore-file", "", "root-relative or absolute ignore file; defaults to .gitignore hierarchy")
	maxFileBytes := flags.Int64("max-file-bytes", 0, "maximum indexed file size")
	var excludePaths stringListFlag
	flags.Var(&excludePaths, "exclude", "root-relative or absolute path to exclude; may be repeated")
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

	excluded := append([]string{statePath}, excludePaths...)
	snapshot, stats, err := index.Build(*root, previous, index.BuildOptions{
		MaxFileBytes: *maxFileBytes,
		IgnoreFile:   *ignoreFile,
		ExcludePaths: excluded,
	})
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

func resolveState(root, state string) (string, error) {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve root: %w", err)
	}
	if state == "" {
		return filepath.Join(absoluteRoot, ".grimoire"), nil
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
	if errors.Is(err, index.ErrIncompatibleIndex) {
		base, rebuildErr := index.RebuildBase(path)
		if rebuildErr != nil {
			return nil, fmt.Errorf("prepare index rebuild: %w", rebuildErr)
		}
		return &base, nil
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
