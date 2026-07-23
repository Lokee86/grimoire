package lexiconfacts

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ExportOptions resolves an immutable Lexicon snapshot into a cached standalone
// JSONL export that Grimoire can inspect without reading Lexicon's private
// mutable implementation state.
type ExportOptions struct {
	Root              string
	GrimoireState     string
	ExplicitDirectory string
	LexiconState      string
	Command           string
	Run               func(context.Context, string, ...string) error
}

// ResolveExport returns an explicit facts directory or a cached export of the
// repository's current immutable Lexicon snapshot. Repositories without
// Lexicon state return empty values without error.
func ResolveExport(ctx context.Context, options ExportOptions) (string, string, error) {
	root := options.Root
	if root == "" {
		root = "."
	}
	if explicit := strings.TrimSpace(options.ExplicitDirectory); explicit != "" {
		if !filepath.IsAbs(explicit) {
			explicit = filepath.Join(root, explicit)
		}
		if !hasJSONLLibraries(explicit) {
			return "", "", fmt.Errorf("no Lexicon JSONL exports found in %s", explicit)
		}
		return explicit, "", nil
	}

	lexiconState := strings.TrimSpace(options.LexiconState)
	if lexiconState == "" {
		lexiconState = filepath.Join(root, ".lexicon")
	} else if !filepath.IsAbs(lexiconState) {
		lexiconState = filepath.Join(root, lexiconState)
	}
	snapshotID, err := readSnapshotID(filepath.Join(lexiconState, "CURRENT"))
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", nil
		}
		return "", "", fmt.Errorf("read Lexicon CURRENT: %w", err)
	}
	digest, err := snapshotDigest(snapshotID)
	if err != nil {
		return "", "", err
	}

	cacheRoot := options.GrimoireState
	if cacheRoot == "" {
		cacheRoot = filepath.Join(root, ".grimoire")
	}
	destination := filepath.Join(cacheRoot, "providers", "lexicon", digest)
	if validCachedExport(destination, snapshotID) {
		return destination, snapshotID, nil
	}

	command := strings.TrimSpace(options.Command)
	if command == "" {
		command = "lexicon"
	}
	run := options.Run
	if run == nil {
		run = runCommand
	}
	parent := filepath.Dir(destination)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return "", "", fmt.Errorf("create Lexicon export cache: %w", err)
	}
	temporary := fmt.Sprintf("%s.tmp-%d", destination, os.Getpid())
	_ = os.RemoveAll(temporary)
	defer os.RemoveAll(temporary)
	if err := run(
		ctx, command,
		"export", "--repo", root, "--output", temporary, "--snapshot", snapshotID,
	); err != nil {
		return "", "", fmt.Errorf("export Lexicon snapshot %s: %w", snapshotID, err)
	}
	if !hasJSONLLibraries(temporary) {
		return "", "", fmt.Errorf("Lexicon export %s produced no JSONL libraries", snapshotID)
	}
	if err := os.WriteFile(filepath.Join(temporary, ".snapshot"), []byte(snapshotID+"\n"), 0o644); err != nil {
		return "", "", fmt.Errorf("write Lexicon export marker: %w", err)
	}
	if err := os.Rename(temporary, destination); err != nil {
		if validCachedExport(destination, snapshotID) {
			return destination, snapshotID, nil
		}
		if removeErr := os.RemoveAll(destination); removeErr != nil {
			return "", "", fmt.Errorf("replace Lexicon export cache: %w", removeErr)
		}
		if renameErr := os.Rename(temporary, destination); renameErr != nil {
			return "", "", fmt.Errorf("publish Lexicon export cache: %w", renameErr)
		}
	}
	return destination, snapshotID, nil
}

func runCommand(ctx context.Context, command string, arguments ...string) error {
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

func readSnapshotID(path string) (string, error) {
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

func snapshotDigest(snapshotID string) (string, error) {
	digest, found := strings.CutPrefix(snapshotID, "sha256:")
	if !found || len(digest) != 64 {
		return "", fmt.Errorf("invalid Lexicon snapshot ID %q", snapshotID)
	}
	for _, character := range digest {
		if !strings.ContainsRune("0123456789abcdef", character) {
			return "", fmt.Errorf("invalid Lexicon snapshot ID %q", snapshotID)
		}
	}
	return digest, nil
}

func validCachedExport(directory, snapshotID string) bool {
	marker, err := os.ReadFile(filepath.Join(directory, ".snapshot"))
	return err == nil && strings.TrimSpace(string(marker)) == snapshotID && hasJSONLLibraries(directory)
}

func hasJSONLLibraries(directory string) bool {
	matches, err := filepath.Glob(filepath.Join(directory, "*.jsonl"))
	return err == nil && len(matches) > 0
}
