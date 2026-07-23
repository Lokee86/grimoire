package arcanagraph

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// StateOptions resolves or catches up Arcana's immutable graph snapshot for a
// repository. ExpectedLexiconSnapshot binds graph evidence to the same Lexicon
// snapshot used for symbol evidence.
type StateOptions struct {
	Root                    string
	State                   string
	LexiconState            string
	ExpectedLexiconSnapshot string
	Command                 string
	Run                     func(context.Context, string, ...string) error
}

func ResolveSnapshot(ctx context.Context, options StateOptions) (string, string, error) {
	root := options.Root
	if root == "" {
		root = "."
	}
	state := options.State
	if state == "" {
		state = filepath.Join(root, ".arcana")
	} else if !filepath.IsAbs(state) {
		state = filepath.Join(root, state)
	}
	lexiconState := options.LexiconState
	if lexiconState == "" {
		lexiconState = filepath.Join(root, ".lexicon")
	} else if !filepath.IsAbs(lexiconState) {
		lexiconState = filepath.Join(root, lexiconState)
	}

	expected := strings.TrimSpace(options.ExpectedLexiconSnapshot)
	if expected == "" {
		value, err := readCurrent(filepath.Join(lexiconState, "CURRENT"))
		if err == nil {
			expected = value
		} else if !os.IsNotExist(err) {
			return "", "", fmt.Errorf("read Lexicon CURRENT for Arcana: %w", err)
		}
	}
	current, currentErr := readCurrent(filepath.Join(state, "CURRENT"))
	if currentErr != nil && !os.IsNotExist(currentErr) {
		return "", "", fmt.Errorf("read Arcana CURRENT: %w", currentErr)
	}
	if currentErr == nil && (expected == "" || current == expected) {
		snapshot, err := snapshotDirectory(state, current)
		if err == nil && snapshotComplete(snapshot) {
			return snapshot, current, nil
		}
	}
	if expected == "" {
		return "", "", nil
	}

	command := strings.TrimSpace(options.Command)
	if command == "" {
		command = "arcana"
	}
	run := options.Run
	if run == nil {
		run = runStateCommand
	}
	if err := run(ctx, command, "sync", "--lexicon", lexiconState, "--state", state); err != nil {
		return "", "", fmt.Errorf("synchronize Arcana with %s: %w", expected, err)
	}
	current, err := readCurrent(filepath.Join(state, "CURRENT"))
	if err != nil {
		return "", "", fmt.Errorf("read synchronized Arcana CURRENT: %w", err)
	}
	if current != expected {
		return "", "", fmt.Errorf("Arcana synchronized %s, expected %s", current, expected)
	}
	snapshot, err := snapshotDirectory(state, current)
	if err != nil {
		return "", "", err
	}
	if !snapshotComplete(snapshot) {
		return "", "", fmt.Errorf("Arcana snapshot %s is incomplete", snapshot)
	}
	return snapshot, current, nil
}

func snapshotDirectory(state, snapshotID string) (string, error) {
	digest, found := strings.CutPrefix(snapshotID, "sha256:")
	if !found || len(digest) != 64 {
		return "", fmt.Errorf("invalid Arcana snapshot ID %q", snapshotID)
	}
	for _, character := range digest {
		if !strings.ContainsRune("0123456789abcdef", character) {
			return "", fmt.Errorf("invalid Arcana snapshot ID %q", snapshotID)
		}
	}
	return filepath.Join(state, "snapshots", digest), nil
}

func snapshotComplete(snapshot string) bool {
	return fileExists(filepath.Join(snapshot, "repository.manifest")) &&
		fileExists(filepath.Join(snapshot, "lexicon.snapshot"))
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func readCurrent(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	value := strings.TrimSpace(string(data))
	if value == "" {
		return "", fmt.Errorf("%s is empty", path)
	}
	return value, nil
}

func runStateCommand(ctx context.Context, command string, arguments ...string) error {
	var stdout, stderr bytes.Buffer
	process := exec.CommandContext(ctx, command, arguments...)
	process.Stdout = &stdout
	process.Stderr = &stderr
	if err := process.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = strings.TrimSpace(stdout.String())
		}
		if message != "" {
			return fmt.Errorf("%w: %s", err, message)
		}
		return err
	}
	return nil
}
