package scan

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Lokee86/lexicon/internal/state"
)

func TestRebuildWithEmptyLanguagesRebuildsAllSourceLanguages(t *testing.T) {
	source := t.TempDir()
	stateRoot := t.TempDir()
	for relative, contents := range map[string]string{
		"main.py": "value = 1\n",
		"main.rb": "value = 1\n",
	} {
		if err := os.WriteFile(filepath.Join(source, relative), []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	gitRepository, err := state.Ensure(stateRoot)
	if err != nil {
		t.Fatal(err)
	}
	analyzer := &fakeAnalyzer{}
	scanner := New(source, stateRoot, gitRepository, analyzer, io.Discard)

	report, err := scanner.Rebuild(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"python", "ruby"}
	if !reflect.DeepEqual(report.Languages, want) || !reflect.DeepEqual(analyzer.languages, want) {
		t.Fatalf("report = %#v, analyzer languages = %v", report, analyzer.languages)
	}
	if report.SnapshotID == "" {
		t.Fatal("rebuild did not publish a snapshot")
	}
}
