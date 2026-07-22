package scan

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Lokee86/lexicon/internal/adapters"
	"github.com/Lokee86/lexicon/internal/state"
)

type fakeAnalyzer struct {
	languages []string
	requests  []adapters.Request
}

func (f *fakeAnalyzer) Run(_ context.Context, request adapters.Request) error {
	f.languages = append(f.languages, request.Language)
	f.requests = append(f.requests, request)
	header := map[string]any{
		"adapter_version": "test", "language": request.Language, "record": "lexicon",
		"repository": "test", "schema_version": 1,
	}
	if request.ChangedFiles != nil || request.RemovedFiles != nil {
		header["mode"] = "incremental"
		header["changed_files"] = request.ChangedFiles
		header["removed_files"] = request.RemovedFiles
		header["shared_complete"] = true
	}
	data, err := json.Marshal(header)
	if err != nil {
		return err
	}
	return os.WriteFile(request.Output, append(data, '\n'), 0o644)
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

func TestScanForcesFullRebuildWhenAdapterFingerprintDrifts(t *testing.T) {
	source := t.TempDir()
	stateRoot := t.TempDir()
	adapterRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(source, "main.py"), []byte("value = 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(adapterRoot, "python"), 0o755); err != nil {
		t.Fatal(err)
	}
	adapterFile := filepath.Join(adapterRoot, "python", "adapter.py")
	if err := os.WriteFile(adapterFile, []byte("value = 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRepository, err := state.Ensure(stateRoot)
	if err != nil {
		t.Fatal(err)
	}
	analyzer := &fakeAnalyzer{}
	scanner := New(source, stateRoot, gitRepository, analyzer, io.Discard)
	scanner.AdapterRoot = adapterRoot
	prepareSnapshot(t, scanner, gitRepository)
	if err := os.WriteFile(adapterFile, []byte("value = 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	report, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Changed) != 0 || !reflect.DeepEqual(report.Languages, []string{"python"}) {
		t.Fatalf("report = %#v", report)
	}
	if len(analyzer.requests) != 2 || analyzer.requests[1].ChangedFiles != nil {
		t.Fatalf("requests = %#v", analyzer.requests)
	}
	_, manifest, err := scanner.Store.Current()
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Languages[0].AdapterFingerprint == "" {
		t.Fatal("published snapshot omitted adapter fingerprint")
	}
}
