package repostate

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/Lokee86/grimoire/internal/index"
)

func TestEnsureCurrentOnlyReportsPreparedStateWithoutMutation(t *testing.T) {
	root := t.TempDir()
	writeSource(t, root, "package main\n")
	id := testID('a')
	writeLexicon(t, root, id)
	writeArcana(t, root, id)
	writeGrimoire(t, root, id)
	before := snapshotFiles(t, root)

	status, err := Ensure(context.Background(), Options{Root: root, Mode: CurrentOnly})
	if err != nil {
		t.Fatal(err)
	}
	if status.Lexicon.Status != "current" || status.Arcana.Status != "current" || status.Grimoire.Status != "current" {
		t.Fatalf("unexpected state: %+v", status)
	}
	if !status.DeterministicQueryReady || status.Vectors.Status != "missing" {
		t.Fatalf("unexpected readiness/vector status: %+v", status)
	}
	if after := snapshotFiles(t, root); strings.Join(before, "\n") != strings.Join(after, "\n") {
		t.Fatalf("current-only mutated state: before=%v after=%v", before, after)
	}
}

func TestEnsureRefreshesAbsentStateInOrder(t *testing.T) {
	root := t.TempDir()
	writeSource(t, root, "package main\n")
	calls := make([]string, 0, 3)
	var mu sync.Mutex
	runner := fixtureRunner(t, root, &mu, &calls, false)

	status, err := Ensure(context.Background(), Options{Root: root, Mode: RefreshIfNeeded, Run: runner})
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(calls, ","); got != "lexicon:init,arcana:sync,grimoire:index" {
		t.Fatalf("refresh order = %s", got)
	}
	if !status.DeterministicQueryReady || len(status.Actions) != 3 {
		t.Fatalf("unexpected refresh status: %+v", status)
	}
}

func TestEnsureRefreshesStaleLexiconAndAlignsArcana(t *testing.T) {
	root := t.TempDir()
	writeSource(t, root, "package main\n")
	oldID, newID := testID('b'), testID('c')
	writeLexicon(t, root, oldID)
	writeArcana(t, root, testID('d'))
	writeGrimoire(t, root, oldID)
	writeMarker(t, filepath.Join(root, ".lexicon"), oldID)
	writeMarker(t, filepath.Join(root, ".grimoire"), oldID)
	writeSource(t, root, "package changed\n")

	calls := make([]string, 0, 3)
	var mu sync.Mutex
	runner := fixtureRunnerWithLexicon(t, root, &mu, &calls, newID)
	status, err := Ensure(context.Background(), Options{Root: root, Mode: RefreshIfNeeded, Run: runner})
	if err != nil {
		t.Fatal(err)
	}
	if status.Lexicon.Snapshot != newID || status.Arcana.Snapshot != newID || status.Arcana.Expected != newID {
		t.Fatalf("snapshots are not aligned: %+v", status)
	}
	if got := strings.Join(calls, ","); got != "lexicon:scan,arcana:sync,grimoire:index" {
		t.Fatalf("stale refresh order = %s", got)
	}
}

func TestEnsureReportsCommandFailure(t *testing.T) {
	root := t.TempDir()
	writeSource(t, root, "package main\n")
	status, err := Ensure(context.Background(), Options{
		Root: root, Mode: RefreshIfNeeded,
		Run: func(context.Context, string, ...string) error { return errors.New("runner failed") },
	})
	if err == nil || status.Error == "" || len(status.Actions) != 1 || status.Actions[0].Status != "failed" {
		t.Fatalf("failure was not reported: status=%+v err=%v", status, err)
	}
}

func TestEnsureForceRefreshesCurrentState(t *testing.T) {
	root := t.TempDir()
	writeSource(t, root, "package main\n")
	id := testID('e')
	writeLexicon(t, root, id)
	writeArcana(t, root, id)
	writeGrimoire(t, root, id)
	calls := make([]string, 0, 3)
	var mu sync.Mutex
	status, err := Ensure(context.Background(), Options{Root: root, Mode: ForceRefresh, Run: fixtureRunnerWithLexicon(t, root, &mu, &calls, id)})
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(calls, ","); got != "lexicon:rebuild,arcana:sync,grimoire:index" {
		t.Fatalf("force refresh order = %s", got)
	}
	if !status.DeterministicQueryReady {
		t.Fatalf("force refresh is not ready: %+v", status)
	}
}

func TestEnsureSerializesConcurrentRefreshes(t *testing.T) {
	root := t.TempDir()
	writeSource(t, root, "package main\n")
	calls := make([]string, 0, 3)
	var mu sync.Mutex
	runner := fixtureRunner(t, root, &mu, &calls, false)
	var wait sync.WaitGroup
	statuses := make([]Status, 2)
	errorsFound := make([]error, 2)
	for i := range statuses {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			statuses[index], errorsFound[index] = Ensure(context.Background(), Options{Root: root, Mode: RefreshIfNeeded, Run: runner})
		}(i)
	}
	wait.Wait()
	for i := range statuses {
		if errorsFound[i] != nil || !statuses[i].DeterministicQueryReady {
			t.Fatalf("concurrent refresh %d failed: status=%+v err=%v", i, statuses[i], errorsFound[i])
		}
	}
	if len(calls) != 3 {
		t.Fatalf("concurrent callers performed %d commands: %v", len(calls), calls)
	}
}

func writeSource(t *testing.T, root, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func testID(character byte) string { return "sha256:" + strings.Repeat(string(character), 64) }

func writeLexicon(t *testing.T, root, id string) {
	t.Helper()
	state := filepath.Join(root, ".lexicon")
	if err := os.MkdirAll(filepath.Join(state, "snapshots"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(state, "CURRENT"), []byte(id+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(state, "snapshots", strings.TrimPrefix(id, "sha256:")+".json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(state, "config.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeArcana(t *testing.T, root, id string) {
	t.Helper()
	state := filepath.Join(root, ".arcana")
	directory := filepath.Join(state, "snapshots", strings.TrimPrefix(id, "sha256:"))
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"repository.manifest", "lexicon.snapshot"} {
		value := []byte("ok\n")
		if name == "lexicon.snapshot" {
			value = []byte(id + "\n")
		}
		if err := os.WriteFile(filepath.Join(directory, name), value, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(state, "CURRENT"), []byte(id+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeGrimoire(t *testing.T, root, lexiconID string) {
	t.Helper()
	state := filepath.Join(root, ".grimoire")
	snapshot, _, err := index.Build(root, nil, index.BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err := index.Save(state, snapshot); err != nil {
		t.Fatal(err)
	}
	writeMarker(t, state, lexiconID)
}

func writeMarker(t *testing.T, state, lexiconID string) {
	t.Helper()
	fingerprint, err := sourceFingerprint(filepath.Dir(state))
	if err != nil {
		t.Fatal(err)
	}
	data, err := marshalJSON(stateMarker{SourceFingerprint: fingerprint, LexiconSnapshot: lexiconID})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(state, ".repostate.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func snapshotFiles(t *testing.T, root string) []string {
	t.Helper()
	var result []string
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err == nil && info.Mode().IsRegular() && !strings.Contains(filepath.ToSlash(path), "/.git/") {
			result = append(result, path)
		}
		return nil
	})
	return result
}
