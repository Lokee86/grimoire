package investigation

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func testSnapshot() Snapshot {
	return Snapshot{Repository: "repo:abc", Providers: map[string]string{"lexicon": "lex:1", "arcana": "arc:2"}}
}

func testResponse(snapshot Snapshot, id string) Response {
	return Response{
		Snapshot:            snapshot,
		Nodes:               []Node{{ID: id, Kind: "function", Path: "pkg/file.go"}},
		SourceRanges:        []SourceRange{{Path: "pkg/file.go", StartLine: 2, EndLine: 4, Text: "return value"}},
		GraphPaths:          []GraphPath{{ID: "path-" + id, Nodes: []string{"a", id}, Edges: []string{"calls"}}},
		Documents:           []Document{{ID: "doc-" + id, URI: "pkg/file.go", Content: "document body"}},
		UnresolvedQuestions: []UnresolvedQuestion{{Question: "What calls " + id + "?"}},
		RejectedBranches:    []Branch{{ID: "reject-" + id, Reason: "not reachable"}},
		AcceptedBranches:    []Branch{{ID: "accept-" + id, Description: "confirmed"}},
	}
}

func TestCreateOpenAndStatus(t *testing.T) {
	root := t.TempDir()
	snapshot := testSnapshot()
	ledger, err := Create(root, "session-1", snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(root, "investigations", "session-1"); ledger.Path() != want {
		t.Fatalf("path = %q, want %q", ledger.Path(), want)
	}
	opened, err := Open(root, "session-1", snapshot)
	if err != nil {
		t.Fatal(err)
	}
	status, err := opened.Status()
	if err != nil {
		t.Fatal(err)
	}
	if status.SessionID != "session-1" || !status.Snapshot.Equal(snapshot) || status.Responses != 0 {
		t.Fatalf("unexpected status: %+v", status)
	}
	if _, err := Create(root, "session-1", snapshot); !errors.Is(err, ErrSessionExists) {
		t.Fatalf("duplicate create error = %v", err)
	}
}

func TestRecordDeduplicatesAndReturnsDelta(t *testing.T) {
	root := t.TempDir()
	snapshot := testSnapshot()
	ledger, err := Create(root, "session-1", snapshot)
	if err != nil {
		t.Fatal(err)
	}
	first := testResponse(snapshot, "one")
	firstDelta, err := ledger.RecordResponse(first)
	if err != nil {
		t.Fatal(err)
	}
	if len(firstDelta.NewNodes) != 1 || len(firstDelta.NewSourceRanges) != 1 || len(firstDelta.NewGraphPaths) != 1 || len(firstDelta.NewDocuments) != 1 {
		t.Fatalf("first delta did not contain new typed evidence: %+v", firstDelta)
	}
	if firstDelta.NewNodes[0].Handle.String() == "one" {
		t.Fatal("handle is not opaque")
	}
	if _, err := ParseNodeHandle(firstDelta.NewNodes[0].Handle.String()); err != nil {
		t.Fatal(err)
	}

	second := Response{Snapshot: snapshot, Nodes: first.Nodes, Documents: []Document{{ID: "doc-two", Content: "new"}}}
	secondDelta, err := ledger.RecordResponse(second)
	if err != nil {
		t.Fatal(err)
	}
	if len(secondDelta.NewNodes) != 0 || len(secondDelta.PriorNodeHandles) != 1 || len(secondDelta.NewDocuments) != 1 {
		t.Fatalf("unexpected second delta: %+v", secondDelta)
	}

	repeated, err := ledger.RecordResponse(first)
	if err != nil {
		t.Fatal(err)
	}
	if len(repeated.NewNodes) != 0 || len(repeated.PriorNodeHandles) != 1 || len(repeated.NewDocuments) != 0 {
		t.Fatalf("repeated evidence was not referenced: %+v", repeated)
	}
	returned, err := ledger.EvidenceAlreadyReturned(first)
	if err != nil {
		t.Fatal(err)
	}
	if !returned {
		t.Fatal("identical evidence was reported as new")
	}
	status, err := ledger.Status()
	if err != nil {
		t.Fatal(err)
	}
	if status.Responses != 2 || status.UniqueNodes != 1 || status.UniqueDocuments != 2 {
		t.Fatalf("unexpected dedup status: %+v", status)
	}
}

func TestSnapshotMismatchAndClose(t *testing.T) {
	root := t.TempDir()
	snapshot := testSnapshot()
	ledger, err := Create(root, "session-1", snapshot)
	if err != nil {
		t.Fatal(err)
	}
	changed := snapshot
	changed.Providers = map[string]string{"arcana": "changed"}
	if _, err := Open(root, "session-1", changed); !errors.Is(err, ErrSnapshotMismatch) {
		t.Fatalf("open mismatch = %v", err)
	}
	if _, err := ledger.RecordResponse(testResponse(changed, "one")); !errors.Is(err, ErrSnapshotMismatch) {
		t.Fatalf("record mismatch = %v", err)
	}
	if err := ledger.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := ledger.RecordResponse(testResponse(snapshot, "two")); !errors.Is(err, ErrSessionClosed) {
		t.Fatalf("record after close = %v", err)
	}
}

func TestCorruptManifestAndRecordFailSafely(t *testing.T) {
	root := t.TempDir()
	snapshot := testSnapshot()
	ledger, err := Create(root, "manifest", snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ledger.RecordResponse(testResponse(snapshot, "one")); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(root, "investigations", "manifest", "manifest.json")
	if err := os.WriteFile(manifestPath, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(root, "manifest"); !errors.Is(err, ErrCorrupt) {
		t.Fatalf("corrupt manifest = %v", err)
	}

	root = t.TempDir()
	ledger, err = Create(root, "record", snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ledger.RecordResponse(testResponse(snapshot, "one")); err != nil {
		t.Fatal(err)
	}
	matches, err := filepath.Glob(filepath.Join(root, "investigations", "record", "records", "node-*.json"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("node records = %v, %v", matches, err)
	}
	if err := os.WriteFile(matches[0], []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(root, "record"); !errors.Is(err, ErrCorrupt) {
		t.Fatalf("corrupt record = %v", err)
	}
}

func TestConcurrentWritersAreSerialized(t *testing.T) {
	root := t.TempDir()
	snapshot := testSnapshot()
	ledger, err := Create(root, "concurrent", snapshot)
	if err != nil {
		t.Fatal(err)
	}
	const writers = 12
	errs := make(chan error, writers)
	var wait sync.WaitGroup
	for i := 0; i < writers; i++ {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			opened, openErr := Open(root, "concurrent", snapshot)
			if openErr != nil {
				errs <- openErr
				return
			}
			_, recordErr := opened.RecordResponse(Response{Snapshot: snapshot, Nodes: []Node{{ID: "node-" + string(rune('a'+index))}}})
			errs <- recordErr
		}(i)
	}
	wait.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	status, err := ledger.Status()
	if err != nil {
		t.Fatal(err)
	}
	if status.Responses != writers || status.UniqueNodes != writers {
		t.Fatalf("concurrent writes lost: %+v", status)
	}
}
