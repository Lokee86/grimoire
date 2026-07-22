package objectstore

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestExportCurrentReconstructsSortedStandaloneLibrary(t *testing.T) {
	store, entry := exportFixture(t)
	destination := filepath.Join(t.TempDir(), "libraries")
	if err := store.Export("CURRENT", destination, []string{"python"}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(destination, "python.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 6 {
		t.Fatalf("lines = %d, want 6", len(lines))
	}
	var header Header
	if err := json.Unmarshal([]byte(lines[0]), &header); err != nil {
		t.Fatal(err)
	}
	wantHeader := Header{
		Record: "lexicon", SchemaVersion: 1, AdapterVersion: "adapter-1",
		Language: "python", Repository: "repo", Mode: "full",
	}
	if !reflect.DeepEqual(header, wantHeader) {
		t.Fatalf("header = %#v, want %#v", header, wantHeader)
	}
	var records []map[string]any
	for _, line := range lines[1:] {
		var record map[string]any
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatal(err)
		}
		records = append(records, record)
	}
	if got := records[0]["id"]; got != "node-a" {
		t.Fatalf("first record = %v, want node-a", got)
	}
	if got := records[1]["id"]; got != "node-z" {
		t.Fatalf("second record = %v, want node-z", got)
	}
	if got := records[2]["record"]; got != "edge" {
		t.Fatalf("third record = %v, want edge", got)
	}
	if got := records[3]["record"]; got != "unresolved" {
		t.Fatalf("fourth record = %v, want unresolved", got)
	}
	if !strings.Contains(string(data), "shared-record") || !strings.Contains(string(data), "file-record") {
		t.Fatal("export omitted shared or file records")
	}
	if entry.SharedObjectID == "" {
		t.Fatal("fixture did not include a shared object")
	}
	otherDestination := filepath.Join(t.TempDir(), "libraries")
	if err := store.Export("CURRENT", otherDestination, []string{"python"}); err != nil {
		t.Fatal(err)
	}
	other, err := os.ReadFile(filepath.Join(otherDestination, "python.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, other) {
		t.Fatal("repeated export changed the JSONL bytes")
	}
}

func TestExportRejectsUnknownLanguageBeforePublishing(t *testing.T) {
	store, _ := exportFixture(t)
	destination := t.TempDir()
	path := filepath.Join(destination, "python.jsonl")
	original := []byte("old content\n")
	if err := os.WriteFile(path, original, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := store.Export("CURRENT", destination, []string{"ruby"}); err == nil {
		t.Fatal("expected unknown language error")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(data, original) {
		t.Fatalf("destination changed after rejected language: %q", data)
	}
}

func TestExportVerifiesEveryObjectBeforeAtomicPublish(t *testing.T) {
	store, entry := exportFixture(t)
	destination := t.TempDir()
	path := filepath.Join(destination, "python.jsonl")
	original := []byte("previous complete library\n")
	if err := os.WriteFile(path, original, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(store.ObjectPath(entry.Files[0].ObjectID), []byte("corrupt\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := store.Export("CURRENT", destination, []string{"python"}); err == nil {
		t.Fatal("expected corrupt object error")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(data, original) {
		t.Fatalf("destination changed after corrupt object: %q", data)
	}
}

func exportFixture(t *testing.T) (Store, LanguageEntry) {
	t.Helper()
	store := Store{Root: t.TempDir()}
	metadata := FactObject{
		Language: "python", AdapterVersion: "adapter-1", SchemaVersion: 1,
		AnalysisConfigID: "sha256:config",
	}
	shared := metadata
	shared.Records = []json.RawMessage{
		json.RawMessage(`{"record":"unresolved","reason":"shared-record","source":"node-a"}`),
	}
	sharedID, err := store.WriteObject(shared)
	if err != nil {
		t.Fatal(err)
	}
	file := metadata
	file.Owner = "main.py"
	file.SourceContentID = ContentID([]byte("main.py"))
	file.Records = []json.RawMessage{
		json.RawMessage(`{"record":"edge","relation":"calls","source":"node-z","target":"node-a"}`),
		json.RawMessage(`{"id":"node-z","kind":"function","record":"node","path":"main.py"}`),
		json.RawMessage(`{"id":"node-a","kind":"file","path":"main.py","record":"node"}`),
		json.RawMessage(`{"record":"unresolved","reason":"file-record","source":"node-z"}`),
	}
	fileID, err := store.WriteObject(file)
	if err != nil {
		t.Fatal(err)
	}
	entry := LanguageEntry{
		Language: "python", AdapterVersion: "adapter-1", SchemaVersion: 1,
		Repository: "repo", AnalysisConfigID: "sha256:config", SharedObjectID: sharedID,
		Files: []FileEntry{{
			Path: "main.py", Language: "python", ContentID: file.SourceContentID, ObjectID: fileID,
		}},
	}
	if _, err := store.Publish(Manifest{StateCommit: "state", Languages: []LanguageEntry{entry}}); err != nil {
		t.Fatal(err)
	}
	return store, entry
}

func TestExportAcceptsExplicitSnapshotID(t *testing.T) {
	store, _ := exportFixture(t)
	destination := t.TempDir()
	id, _, err := store.Current()
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Export(id, destination, []string{"python"}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(destination, "python.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("export is empty")
	}
}
