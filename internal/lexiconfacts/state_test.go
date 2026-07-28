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
	lexiconState := filepath.Join(root, ".warlock", "tools", "lexicon")
	if err := os.MkdirAll(lexiconState, 0o755); err != nil {
		t.Fatal(err)
	}
	snapshotID := "sha256:" + strings.Repeat("a", 64)
	if err := os.WriteFile(filepath.Join(lexiconState, "CURRENT"), []byte(snapshotID+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	runs := 0
	run := func(_ context.Context, command ExportCommand) error {
		runs++
		if command.Executable != "lexicon-test" {
			t.Fatalf("unexpected command %q", command.Executable)
		}
		if actual := environmentValue(command.Environment, "LEXICON_STATE_DIR"); actual != lexiconState {
			t.Fatalf("Lexicon state environment = %q, want %q", actual, lexiconState)
		}
		var output string
		for index, argument := range command.Arguments {
			if argument == "--output" && index+1 < len(command.Arguments) {
				output = command.Arguments[index+1]
			}
		}
		if output == "" {
			t.Fatalf("export output missing from %v", command.Arguments)
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
		LexiconState: lexiconState, Command: "lexicon-test", Run: run,
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

func environmentValue(environment []string, name string) string {
	for _, entry := range environment {
		key, value, found := strings.Cut(entry, "=")
		if found && strings.EqualFold(key, name) {
			return value
		}
	}
	return ""
}
