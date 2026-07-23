package arcanagraph

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveSnapshotSynchronizesToExpectedLexiconSnapshot(t *testing.T) {
	root := t.TempDir()
	lexiconState := filepath.Join(root, ".lexicon")
	arcanaState := filepath.Join(root, ".arcana")
	if err := os.MkdirAll(lexiconState, 0o755); err != nil {
		t.Fatal(err)
	}
	snapshotID := "sha256:" + strings.Repeat("b", 64)
	if err := os.WriteFile(filepath.Join(lexiconState, "CURRENT"), []byte(snapshotID+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	runs := 0
	run := func(_ context.Context, command string, arguments ...string) error {
		runs++
		if command != "arcana-test" {
			t.Fatalf("unexpected command %q", command)
		}
		if len(arguments) == 0 || arguments[0] != "sync" {
			t.Fatalf("unexpected arguments %v", arguments)
		}
		digest := strings.TrimPrefix(snapshotID, "sha256:")
		snapshot := filepath.Join(arcanaState, "snapshots", digest)
		if err := os.MkdirAll(snapshot, 0o755); err != nil {
			return err
		}
		for _, name := range []string{"repository.manifest", "lexicon.snapshot"} {
			if err := os.WriteFile(filepath.Join(snapshot, name), []byte(snapshotID+"\n"), 0o644); err != nil {
				return err
			}
		}
		return os.WriteFile(filepath.Join(arcanaState, "CURRENT"), []byte(snapshotID+"\n"), 0o644)
	}
	options := StateOptions{
		Root: root, State: arcanaState, LexiconState: lexiconState,
		ExpectedLexiconSnapshot: snapshotID, Command: "arcana-test", Run: run,
	}
	first, firstID, err := ResolveSnapshot(context.Background(), options)
	if err != nil {
		t.Fatal(err)
	}
	if !snapshotComplete(first) || firstID != snapshotID || runs != 1 {
		t.Fatalf("Arcana snapshot was not synchronized: path=%q id=%q runs=%d", first, firstID, runs)
	}
	second, secondID, err := ResolveSnapshot(context.Background(), options)
	if err != nil {
		t.Fatal(err)
	}
	if second != first || secondID != snapshotID || runs != 1 {
		t.Fatalf("current Arcana snapshot was not reused: first=%q second=%q id=%q runs=%d", first, second, secondID, runs)
	}
}
