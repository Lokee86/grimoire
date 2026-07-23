package lexiconfacts

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveExportCachesCurrentImmutableSnapshot(t *testing.T) {
	root := t.TempDir()
	lexiconState := filepath.Join(root, ".lexicon")
	if err := os.MkdirAll(lexiconState, 0o755); err != nil {
		t.Fatal(err)
	}
	snapshotID := "sha256:" + strings.Repeat("a", 64)
	if err := os.WriteFile(filepath.Join(lexiconState, "CURRENT"), []byte(snapshotID+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	runs := 0
	run := func(_ context.Context, command string, arguments ...string) error {
		runs++
		if command != "lexicon-test" {
			t.Fatalf("unexpected command %q", command)
		}
		var output string
		for index, argument := range arguments {
			if argument == "--output" && index+1 < len(arguments) {
				output = arguments[index+1]
			}
		}
		if output == "" {
			t.Fatalf("export output missing from %v", arguments)
		}
		if err := os.MkdirAll(output, 0o755); err != nil {
			return err
		}
		return os.WriteFile(
			filepath.Join(output, "go.jsonl"),
			[]byte("{\"record\":\"lexicon\",\"language\":\"go\"}\n"),
			0o644,
		)
	}
	options := ExportOptions{
		Root: root, GrimoireState: filepath.Join(root, ".grimoire"),
		Command: "lexicon-test", Run: run,
	}
	first, firstID, err := ResolveExport(context.Background(), options)
	if err != nil {
		t.Fatal(err)
	}
	if firstID != snapshotID || !hasJSONLLibraries(first) {
		t.Fatalf("unexpected export result directory=%q snapshot=%q", first, firstID)
	}
	second, secondID, err := ResolveExport(context.Background(), options)
	if err != nil {
		t.Fatal(err)
	}
	if second != first || secondID != snapshotID || runs != 1 {
		t.Fatalf("cache was not reused: first=%q second=%q runs=%d", first, second, runs)
	}
}
