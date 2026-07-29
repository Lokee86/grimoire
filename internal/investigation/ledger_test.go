package investigation

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
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

func TestSourceRangeHandleIgnoresRetrievalOccurrenceMetadata(t *testing.T) {
	root := t.TempDir()
	snapshot := testSnapshot()
	ledger, err := Create(root, "stable-source-handle", snapshot)
	if err != nil {
		t.Fatal(err)
	}

	first := Response{Snapshot: snapshot, SourceRanges: []SourceRange{{
		Path:      `pkg\file.go`,
		StartLine: 12,
		EndLine:   18,
		Text:      "first excerpt",
		Metadata: map[string]string{
			"retrieval_lane": "symbol_matches",
			"rank":           "1",
			"score":          "0.95",
		},
	}}}
	firstDelta, err := ledger.RecordResponse(first)
	if err != nil {
		t.Fatal(err)
	}
	if len(firstDelta.NewSourceRanges) != 1 {
		t.Fatalf("first source range was not new: %+v", firstDelta)
	}
	firstHandle := firstDelta.NewSourceRanges[0].Handle.String()

	second := Response{Snapshot: snapshot, SourceRanges: []SourceRange{{
		Path:      "pkg/file.go",
		StartLine: 12,
		EndLine:   18,
		Text:      "second excerpt",
		Metadata: map[string]string{
			"retrieval_lane": "relationship_matches",
			"rank":           "7",
			"score":          "0.42",
			"match_reasons":  "different query",
		},
	}}}
	secondDelta, err := ledger.RecordResponse(second)
	if err != nil {
		t.Fatal(err)
	}
	if len(secondDelta.NewSourceRanges) != 0 || len(secondDelta.PriorSourceRanges) != 1 {
		t.Fatalf("same canonical source range was not reused: %+v", secondDelta)
	}
	if got := secondDelta.PriorSourceRanges[0].String(); got != firstHandle {
		t.Fatalf("reused handle = %q, want %q", got, firstHandle)
	}
}

func TestRetrievalHitsPreserveOccurrenceTuples(t *testing.T) {
	root := t.TempDir()
	snapshot := testSnapshot()
	ledger, err := Create(root, "retrieval-tuples", snapshot)
	if err != nil {
		t.Fatal(err)
	}
	response := Response{
		Snapshot:     snapshot,
		Nodes:        []Node{{ID: "seed-node", Kind: "function"}},
		SourceRanges: []SourceRange{{Path: "pkg/file.go", StartLine: 20, EndLine: 24, Text: "func Target()"}},
		RetrievalHits: []RetrievalHit{
			{
				Evidence: EvidenceRef{Kind: "source", Index: 0},
				Lane:     "symbol_matches", Provider: "lexicon", Rank: 1, Score: 0.91,
				Reasons: []string{"qualified name match"},
			},
			{
				Evidence: EvidenceRef{Kind: "source", Index: 0},
				Lane:     "relationship_matches", Provider: "arcana", Rank: 4, Score: 0.37,
				Reasons: []string{"callee of selected seed"}, Relation: "calls", Direction: "outgoing",
				Seed: &RetrievalSeed{
					Evidence: EvidenceRef{Kind: "node", Index: 0},
					Lane:     "exact_matches", Provider: "lexicon", Rank: 2, Score: 0.88,
					Reasons: []string{"literal source match"},
				},
			},
		},
	}

	delta, err := ledger.RecordResponse(response)
	if err != nil {
		t.Fatal(err)
	}
	if len(delta.RetrievalHits) != 2 {
		t.Fatalf("retrieval hit count = %d, want 2", len(delta.RetrievalHits))
	}
	first, second := delta.RetrievalHits[0], delta.RetrievalHits[1]
	if first.EvidenceHandle != second.EvidenceHandle {
		t.Fatalf("same evidence resolved to different handles: %q != %q", first.EvidenceHandle, second.EvidenceHandle)
	}
	if first.Lane != "symbol_matches" || first.Rank != 1 || first.Score != 0.91 || first.Reasons[0] != "qualified name match" {
		t.Fatalf("first occurrence tuple changed: %+v", first)
	}
	if second.Lane != "relationship_matches" || second.Rank != 4 || second.Score != 0.37 || second.Relation != "calls" || second.Seed == nil {
		t.Fatalf("second occurrence tuple changed: %+v", second)
	}
	if second.Seed.EvidenceHandle != delta.NewNodes[0].Handle.String() || second.Seed.Lane != "exact_matches" || second.Seed.Rank != 2 || second.Seed.Score != 0.88 {
		t.Fatalf("seed occurrence tuple changed: %+v", second.Seed)
	}

	replayed, err := ledger.RecordResponse(response)
	if err != nil {
		t.Fatal(err)
	}
	if len(replayed.RetrievalHits) != 2 || replayed.RetrievalHits[1].Rank != 4 {
		t.Fatalf("replayed retrieval tuples changed: %+v", replayed.RetrievalHits)
	}
}

func TestResolveRecordedHandles(t *testing.T) {
	root := t.TempDir()
	snapshot := testSnapshot()
	ledger, err := Create(root, "resolve", snapshot)
	if err != nil {
		t.Fatal(err)
	}
	delta, err := ledger.RecordResponse(testResponse(snapshot, "grimoire:v1:node-one"))
	if err != nil {
		t.Fatal(err)
	}
	if got, err := HandleKind(delta.NewNodes[0].Handle.String()); err != nil || got != "node" {
		t.Fatalf("node handle kind = %q, %v", got, err)
	}
	node, err := ledger.ResolveNodeHandle(delta.NewNodes[0].Handle.String())
	if err != nil {
		t.Fatal(err)
	}
	if node.ID != "grimoire:v1:node-one" || node.Path != "pkg/file.go" {
		t.Fatalf("resolved node = %#v", node)
	}
	source, err := ledger.ResolveSourceRangeHandle(delta.NewSourceRanges[0].Handle.String())
	if err != nil {
		t.Fatal(err)
	}
	if source.Path != "pkg/file.go" || source.StartLine != 2 || source.EndLine != 4 {
		t.Fatalf("resolved source = %#v", source)
	}
	path, err := ledger.ResolveGraphPathHandle(delta.NewGraphPaths[0].Handle.String())
	if err != nil {
		t.Fatal(err)
	}
	if path.ID != "path-grimoire:v1:node-one" || len(path.Nodes) != 2 {
		t.Fatalf("resolved path = %#v", path)
	}
	document, err := ledger.ResolveDocumentHandle(delta.NewDocuments[0].Handle.String())
	if err != nil {
		t.Fatal(err)
	}
	if document.ID != "doc-grimoire:v1:node-one" || document.URI != "pkg/file.go" {
		t.Fatalf("resolved document = %#v", document)
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

func TestLockContentionRecognizesPlatformDirectoryRaces(t *testing.T) {
	if !lockContention(os.ErrExist) {
		t.Fatal("existing lock directory was not recognized as contention")
	}
	if runtime.GOOS == "windows" && !lockContention(os.ErrPermission) {
		t.Fatal("Windows access-denied directory race was not recognized as contention")
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
