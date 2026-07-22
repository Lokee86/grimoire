package scan

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Lokee86/lexicon/internal/adapters"
	"github.com/Lokee86/lexicon/internal/state"
)

type retryAnalyzer struct {
	fakeAnalyzer
}

func (r *retryAnalyzer) Run(ctx context.Context, request adapters.Request) error {
	if request.ChangedFiles != nil {
		r.languages = append(r.languages, request.Language)
		r.requests = append(r.requests, request)
		return errors.New("scoped repository is incomplete")
	}
	return r.fakeAnalyzer.Run(ctx, request)
}

func TestScanRequestsOnlyImpactedFilesAfterInitialSnapshot(t *testing.T) {
	source := t.TempDir()
	stateRoot := t.TempDir()
	changed := filepath.Join(source, "a.py")
	unchanged := filepath.Join(source, "b.py")
	for path, data := range map[string]string{changed: "value = 1\n", unchanged: "other = 1\n"} {
		if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	gitRepository, err := state.Ensure(stateRoot)
	if err != nil {
		t.Fatal(err)
	}
	analyzer := &fakeAnalyzer{}
	scanner := New(source, stateRoot, gitRepository, analyzer, io.Discard)
	prepareSnapshot(t, scanner, gitRepository)
	if err := os.WriteFile(changed, []byte("value = 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := scanner.Scan(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(analyzer.requests) != 2 {
		t.Fatalf("requests = %#v", analyzer.requests)
	}
	request := analyzer.requests[1]
	if !reflect.DeepEqual(request.ChangedFiles, []string{"a.py"}) {
		t.Fatalf("changed files = %v", request.ChangedFiles)
	}
	if request.RemovedFiles == nil || len(request.RemovedFiles) != 0 {
		t.Fatalf("removed files = %#v", request.RemovedFiles)
	}
	if _, err := os.Stat(filepath.Join(request.Repository, "a.py")); err != nil {
		t.Fatalf("scoped changed file: %v", err)
	}
	if _, err := os.Stat(filepath.Join(request.Repository, "b.py")); !os.IsNotExist(err) {
		t.Fatalf("unchanged file leaked into scope: %v", err)
	}
}

func TestScopedAdapterFailureRetriesCompleteLanguage(t *testing.T) {
	source := t.TempDir()
	stateRoot := t.TempDir()
	path := filepath.Join(source, "a.py")
	if err := os.WriteFile(path, []byte("value = 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRepository, err := state.Ensure(stateRoot)
	if err != nil {
		t.Fatal(err)
	}
	analyzer := &retryAnalyzer{}
	scanner := New(source, stateRoot, gitRepository, analyzer, io.Discard)
	prepareSnapshot(t, scanner, gitRepository)
	if err := os.WriteFile(path, []byte("value = 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := scanner.Scan(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(analyzer.requests) != 3 {
		t.Fatalf("requests = %#v", analyzer.requests)
	}
	if analyzer.requests[1].ChangedFiles == nil || analyzer.requests[2].ChangedFiles != nil {
		t.Fatalf("expected scoped request then full retry: %#v", analyzer.requests)
	}
}

func prepareSnapshot(t *testing.T, scanner *Scanner, repository *state.Repository) {
	t.Helper()
	if err := scanner.Mirror.SyncAll(scanner.Repository); err != nil {
		t.Fatal(err)
	}
	if err := scanner.analyzeFull(context.Background(), []string{"python"}); err != nil {
		t.Fatal(err)
	}
	if err := repository.StageAll(); err != nil {
		t.Fatal(err)
	}
	if err := repository.CommitState(); err != nil {
		t.Fatal(err)
	}
	if _, err := scanner.publishSnapshot(); err != nil {
		t.Fatal(err)
	}
}
