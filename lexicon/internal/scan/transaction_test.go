package scan

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/Lokee86/lexicon/internal/objectstore"
	"github.com/Lokee86/lexicon/internal/state"
)

func TestRecoverPendingPublicationWithoutStateCommit(t *testing.T) {
	source := t.TempDir()
	stateRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(source, "a.py"), []byte("value = 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	repository, err := state.Ensure(stateRoot)
	if err != nil {
		t.Fatal(err)
	}
	scanner := New(source, stateRoot, repository, &fakeAnalyzer{}, io.Discard)
	prepareSnapshot(t, scanner, repository)
	head, err := repository.Head()
	if err != nil {
		t.Fatal(err)
	}
	_, current, err := scanner.Store.Current()
	if err != nil {
		t.Fatal(err)
	}
	candidate := current.WithoutLanguage("python")
	if err := scanner.Store.WritePending(head, false, candidate); err != nil {
		t.Fatal(err)
	}
	if err := scanner.recoverPending(); err != nil {
		t.Fatal(err)
	}
	_, recovered, err := scanner.Store.Current()
	if err != nil {
		t.Fatal(err)
	}
	if recovered.StateCommit != head || len(recovered.Languages) != 0 {
		t.Fatalf("recovered manifest = %#v", recovered)
	}
	if _, err := scanner.Store.Pending(); !errors.Is(err, objectstore.ErrNoPendingPublication) {
		t.Fatalf("pending publication retained: %v", err)
	}
}

func TestLoadManifestMigratesCommittedLegacyLibraryAheadOfCurrent(t *testing.T) {
	source := t.TempDir()
	stateRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(source, "a.py"), []byte("value = 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	repository, err := state.Ensure(stateRoot)
	if err != nil {
		t.Fatal(err)
	}
	scanner := New(source, stateRoot, repository, &fakeAnalyzer{}, io.Discard)
	prepareSnapshot(t, scanner, repository)
	baseHead, err := repository.Head()
	if err != nil {
		t.Fatal(err)
	}
	library := filepath.Join(stateRoot, "library")
	if err := os.MkdirAll(library, 0o755); err != nil {
		t.Fatal(err)
	}
	legacy := "{\"adapter_version\":\"test\",\"language\":\"python\",\"mode\":\"full\",\"record\":\"lexicon\",\"repository\":\"repo\",\"schema_version\":1}\n" +
		"{\"id\":\"a\",\"kind\":\"file\",\"name\":\"a.py\",\"owner\":\"a.py\",\"path\":\"a.py\",\"qualified_name\":\"a.py\",\"record\":\"node\"}\n"
	if err := os.WriteFile(filepath.Join(library, "python.jsonl"), []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := repository.StageAll(); err != nil {
		t.Fatal(err)
	}
	if err := repository.CommitState(); err != nil {
		t.Fatal(err)
	}
	newHead, err := repository.Head()
	if err != nil {
		t.Fatal(err)
	}
	if newHead == baseHead {
		t.Fatal("legacy state commit did not advance")
	}
	manifest, err := scanner.loadManifest()
	if err != nil {
		t.Fatal(err)
	}
	if manifest.StateCommit != newHead {
		t.Fatalf("state commit = %s, want %s", manifest.StateCommit, newHead)
	}
	if _, ok := manifest.Language("python"); !ok {
		t.Fatal("migrated manifest has no python analysis")
	}
}

func TestRecoverPendingPublicationAfterStateCommit(t *testing.T) {
	source := t.TempDir()
	stateRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(source, "a.py"), []byte("value = 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	repository, err := state.Ensure(stateRoot)
	if err != nil {
		t.Fatal(err)
	}
	scanner := New(source, stateRoot, repository, &fakeAnalyzer{}, io.Discard)
	prepareSnapshot(t, scanner, repository)
	baseHead, err := repository.Head()
	if err != nil {
		t.Fatal(err)
	}
	_, candidate, err := scanner.Store.Current()
	if err != nil {
		t.Fatal(err)
	}
	if err := scanner.Store.WritePending(baseHead, true, candidate); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateRoot, "source", "a.py"), []byte("value = 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := repository.StageSource(); err != nil {
		t.Fatal(err)
	}
	if err := repository.CommitState(); err != nil {
		t.Fatal(err)
	}
	newHead, err := repository.Head()
	if err != nil {
		t.Fatal(err)
	}
	if newHead == baseHead {
		t.Fatal("state commit did not advance")
	}
	if err := scanner.recoverPending(); err != nil {
		t.Fatal(err)
	}
	_, recovered, err := scanner.Store.Current()
	if err != nil {
		t.Fatal(err)
	}
	if recovered.StateCommit != newHead {
		t.Fatalf("state commit = %s, want %s", recovered.StateCommit, newHead)
	}
}
