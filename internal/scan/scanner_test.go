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

type fakeAnalyzer struct {
	languages []string
}

func (f *fakeAnalyzer) Run(_ context.Context, language, _, output string) error {
	f.languages = append(f.languages, language)
	data := "{\"adapter_version\":\"test\",\"language\":\"" + language + "\",\"record\":\"lexicon\",\"repository\":\"test\",\"schema_version\":1}\n"
	return os.WriteFile(output, []byte(data), 0o644)
}

func TestScanRebuildsCorruptLibraryWithoutSourceDiff(t *testing.T) {
	source := t.TempDir()
	stateRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(source, "main.py"), []byte("value = 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mirror := state.Mirror{Root: filepath.Join(stateRoot, "source")}
	if err := mirror.SyncAll(source); err != nil {
		t.Fatal(err)
	}
	library := filepath.Join(stateRoot, "library")
	if err := os.MkdirAll(library, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(library, "python.jsonl"), []byte("corrupt\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRepository, err := state.Ensure(stateRoot)
	if err != nil {
		t.Fatal(err)
	}
	if err := gitRepository.StageAll(); err != nil {
		t.Fatal(err)
	}
	if err := gitRepository.CommitState(); err != nil {
		t.Fatal(err)
	}
	analyzer := &fakeAnalyzer{}
	scanner := New(source, stateRoot, gitRepository, analyzer, io.Discard)
	report, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Changed) != 0 || !reflect.DeepEqual(report.Languages, []string{"python"}) {
		t.Fatalf("report = %#v", report)
	}
	if !reflect.DeepEqual(analyzer.languages, []string{"python"}) {
		t.Fatalf("analyzer calls = %v", analyzer.languages)
	}
}

func TestScanUsesInternalDiffAndAffectedLanguage(t *testing.T) {
	source := t.TempDir()
	stateRoot := t.TempDir()
	path := filepath.Join(source, "main.py")
	if err := os.WriteFile(path, []byte("value = 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRepository, err := state.Ensure(stateRoot)
	if err != nil {
		t.Fatal(err)
	}
	analyzer := &fakeAnalyzer{}
	scanner := New(source, stateRoot, gitRepository, analyzer, io.Discard)
	if err := scanner.Mirror.SyncAll(source); err != nil {
		t.Fatal(err)
	}
	if err := gitRepository.StageAll(); err != nil {
		t.Fatal(err)
	}
	if err := gitRepository.CommitState(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("value = 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	report, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(report.Languages, []string{"python"}) {
		t.Fatalf("languages = %v", report.Languages)
	}
	if !reflect.DeepEqual(analyzer.languages, []string{"python"}) {
		t.Fatalf("analyzer calls = %v", analyzer.languages)
	}
	if len(report.Changed) != 1 || report.Changed[0].New != "main.py" {
		t.Fatalf("changes = %#v", report.Changed)
	}
	if report.SnapshotID == "" {
		t.Fatal("expected published snapshot")
	}
	if _, err := os.Stat(filepath.Join(source, ".lexicon", "CURRENT")); err != nil {
		t.Fatalf("current snapshot: %v", err)
	}
}
