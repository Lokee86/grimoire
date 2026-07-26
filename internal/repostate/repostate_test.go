package repostate

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/Lokee86/grimoire/internal/embedding"
	"github.com/Lokee86/grimoire/internal/index"
)

func TestEnsureCurrentOnlyReportsPreparedStateWithoutMutation(t *testing.T) {
	root := t.TempDir()
	writeSource(t, root, "package main\n")
	id := testID('a')
	writeLexicon(t, root, id)
	writeArcana(t, root, id)
	writeGrimoire(t, root, id)
	writeKnowledge(t, root)
	before := snapshotFiles(t, root)

	status, err := Ensure(context.Background(), Options{Root: root, Mode: CurrentOnly})
	if err != nil {
		t.Fatal(err)
	}
	if status.Lexicon.Status != "current" || status.Arcana.Status != "current" || status.Grimoire.Status != "current" {
		t.Fatalf("unexpected state: %+v", status)
	}
	if !status.DeterministicQueryReady || status.Knowledge.Status != "current" ||
		status.ArcanaVectors.Status != "missing" || status.KnowledgeVectors.Status != "missing" {
		t.Fatalf("unexpected readiness/knowledge/vector status: %+v", status)
	}
	if after := snapshotFiles(t, root); strings.Join(before, "\n") != strings.Join(after, "\n") {
		t.Fatalf("current-only mutated state: before=%v after=%v", before, after)
	}
}

func TestEnsureReportsOnlyValidatedCurrentArcanaVectors(t *testing.T) {
	root := t.TempDir()
	writeSource(t, root, "package main\n")
	id := testID('a')
	writeLexicon(t, root, id)
	writeArcana(t, root, id)
	writeGrimoire(t, root, id)
	writeKnowledge(t, root)
	writeArcanaVectors(t, root, id)

	status, err := Ensure(context.Background(), Options{Root: root, Mode: CurrentOnly})
	if err != nil {
		t.Fatal(err)
	}
	if status.ArcanaVectors.Status != "current" || status.ArcanaVectors.Count != 1 || status.ArcanaVectors.Bytes != int64(embedding.Dimensions*4) {
		t.Fatalf("validated Arcana vector status = %+v", status.ArcanaVectors)
	}

	directory := arcanaVectorDirectory(root, id)
	vectorsPath := filepath.Join(directory, "vectors.f32")
	vectors, err := os.ReadFile(vectorsPath)
	if err != nil {
		t.Fatal(err)
	}
	vectors[0] = 1
	if err := os.WriteFile(vectorsPath, vectors, 0o644); err != nil {
		t.Fatal(err)
	}
	status, err = Ensure(context.Background(), Options{Root: root, Mode: CurrentOnly})
	if err != nil {
		t.Fatal(err)
	}
	if status.ArcanaVectors.Status != "stale" || !strings.Contains(status.ArcanaVectors.Reason, "checksum") {
		t.Fatalf("checksum-mismatched Arcana vector status = %+v", status.ArcanaVectors)
	}

	writeArcanaVectors(t, root, id)
	if err := os.WriteFile(filepath.Join(directory, "nodes.jsonl"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	status, err = Ensure(context.Background(), Options{Root: root, Mode: CurrentOnly})
	if err != nil {
		t.Fatal(err)
	}
	if status.ArcanaVectors.Status != "stale" || !strings.Contains(status.ArcanaVectors.Reason, "node records") {
		t.Fatalf("corrupt Arcana vector status = %+v", status.ArcanaVectors)
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
	if got := strings.Join(calls, ","); got != "lexicon:init,arcana:sync,grimoire:index,grimoire:knowledge" {
		t.Fatalf("refresh order = %s", got)
	}
	if !status.DeterministicQueryReady || status.Knowledge.Status != "current" || len(status.Actions) != 4 {
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
	if got := strings.Join(calls, ","); got != "lexicon:scan,arcana:sync,grimoire:index,grimoire:knowledge" {
		t.Fatalf("stale refresh order = %s", got)
	}
}

func TestEnsureReportsRequiredGrimoireFailureAfterOptionalProviderFailure(t *testing.T) {
	root := t.TempDir()
	writeSource(t, root, "package main\n")
	status, err := Ensure(context.Background(), Options{
		Root: root, Mode: RefreshIfNeeded,
		Run: func(context.Context, string, ...string) error { return errors.New("runner failed") },
	})
	if err == nil || status.Error == "" || len(status.Actions) != 2 || status.Actions[1].Status != "failed" {
		t.Fatalf("required failure was not reported: status=%+v err=%v", status, err)
	}
	if len(status.Warnings) == 0 || !strings.Contains(status.Warnings[0], "Lexicon refresh unavailable") {
		t.Fatalf("optional provider failure was not retained as a warning: %+v", status.Warnings)
	}
}

func TestEnsureFallsBackToSourceWhenLexiconIsUnavailable(t *testing.T) {
	root := t.TempDir()
	writeSource(t, root, "package main\n")
	calls := make([]string, 0, 2)
	status, err := Ensure(context.Background(), Options{
		Root: root, Mode: RefreshIfNeeded,
		Run: func(_ context.Context, command string, arguments ...string) error {
			calls = append(calls, command+":"+arguments[0])
			if command == "lexicon" {
				return errors.New("executable not found")
			}
			if command == "grimoire" && arguments[0] == "index" {
				writeGrimoire(t, root, "")
				return nil
			}
			if command == "grimoire" && arguments[0] == "knowledge" {
				writeKnowledge(t, root)
				return nil
			}
			return errors.New("unexpected command")
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(calls, ","); got != "lexicon:init,grimoire:index,grimoire:knowledge" {
		t.Fatalf("fallback calls = %s", got)
	}
	if !status.DeterministicQueryReady || status.Grimoire.Status != "current" {
		t.Fatalf("source fallback is not ready: %+v", status)
	}
	if len(status.Warnings) == 0 || !strings.Contains(status.Warnings[0], "continuing with source analysis") {
		t.Fatalf("fallback warning missing: %+v", status.Warnings)
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
	if got := strings.Join(calls, ","); got != "lexicon:rebuild,arcana:sync,grimoire:index,grimoire:knowledge" {
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
	if len(calls) != 4 {
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

func writeArcanaVectors(t *testing.T, root, id string) {
	t.Helper()
	snapshotDirectory := filepath.Join(root, ".arcana", "snapshots", strings.TrimPrefix(id, "sha256:"))
	repositoryManifest := "version=1\nsnapshot_id=1111111111111111\ngraph_snapshot_id=2222222222222222\nnode_count=1\n"
	if err := os.WriteFile(filepath.Join(snapshotDirectory, "repository.manifest"), []byte(repositoryManifest), 0o644); err != nil {
		t.Fatal(err)
	}
	directory := arcanaVectorDirectory(root, id)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	record := []byte("{\"node_key\":\"0000000000000001\",\"kind\":\"function\",\"path\":\"main.go\",\"name\":\"main\"}\n")
	if err := os.WriteFile(filepath.Join(directory, "nodes.jsonl"), record, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "vectors.f32"), make([]byte, embedding.Dimensions*4), 0o644); err != nil {
		t.Fatal(err)
	}
	recordsHash, err := fileSHA256(filepath.Join(directory, "nodes.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	vectorsHash, err := fileSHA256(filepath.Join(directory, "vectors.f32"))
	if err != nil {
		t.Fatal(err)
	}
	manifest := arcanaVectorManifest{
		Version: 2, RepositorySnapshotID: "1111111111111111", GraphSnapshotID: "2222222222222222",
		Model: embedding.ModelReference, Identity: embedding.Identity(), Dimensions: embedding.Dimensions,
		ItemCount: 1, RecordsFile: "nodes.jsonl", RecordsSHA256: recordsHash,
		VectorsFile: "vectors.f32", VectorsSHA256: vectorsHash,
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "manifest.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func arcanaVectorDirectory(root, id string) string {
	return filepath.Join(root, ".arcana", "vectors", strings.TrimPrefix(id, "sha256:"), embedding.Identity())
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
