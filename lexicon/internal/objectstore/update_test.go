package objectstore

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestIncrementalLanguageUpdateReusesUnchangedObjectsAndSharedFacts(t *testing.T) {
	store := Store{Root: t.TempDir()}
	source := t.TempDir()
	for path, contents := range map[string]string{
		"a.py": "value = 1\n",
		"b.py": "other = 1\n",
	} {
		if err := os.WriteFile(filepath.Join(source, path), []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	fullPath := filepath.Join(t.TempDir(), "full.jsonl")
	writeAnalysisStream(t, fullPath, []string{
		`{"adapter_version":"test","language":"python","mode":"full","record":"lexicon","repository":"repo","schema_version":1}`,
		`{"id":"repo","kind":"repository","name":"repo","path":".","qualified_name":"repo","record":"node"}`,
		`{"id":"a-old","kind":"file","name":"a.py","owner":"a.py","path":"a.py","qualified_name":"a.py","record":"node"}`,
		`{"id":"b","kind":"file","name":"b.py","owner":"b.py","path":"b.py","qualified_name":"b.py","record":"node"}`,
	})
	full, err := ReadAnalysis(fullPath, "python")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := store.BuildFullLanguage(full, source, "python", "sha256:config", "sha256:adapter")
	if err != nil {
		t.Fatal(err)
	}
	oldA := fileEntry(t, entry, "a.py")
	oldB := fileEntry(t, entry, "b.py")
	oldShared := entry.SharedObjectID

	if err := os.WriteFile(filepath.Join(source, "a.py"), []byte("value = 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	incrementalPath := filepath.Join(t.TempDir(), "incremental.jsonl")
	writeAnalysisStream(t, incrementalPath, []string{
		`{"adapter_version":"test","changed_files":["a.py"],"language":"python","mode":"incremental","record":"lexicon","removed_files":[],"repository":"repo","schema_version":1,"shared_complete":true}`,
		`{"id":"scoped-repo","kind":"repository","name":"repo","path":".","qualified_name":"repo","record":"node"}`,
		`{"id":"a-new","kind":"file","name":"a.py","owner":"a.py","path":"a.py","qualified_name":"a.py","record":"node"}`,
	})
	incremental, err := ReadAnalysis(incrementalPath, "python")
	if err != nil {
		t.Fatal(err)
	}
	updated, err := store.BuildIncrementalLanguage(
		entry,
		incremental,
		source,
		"sha256:config",
		"sha256:adapter",
		[]string{"a.py"},
		[]string{},
		false,
	)
	if err != nil {
		t.Fatal(err)
	}
	newA := fileEntry(t, updated, "a.py")
	newB := fileEntry(t, updated, "b.py")
	if newA.ObjectID == oldA.ObjectID || newA.ContentID == oldA.ContentID {
		t.Fatalf("changed object was reused: old=%#v new=%#v", oldA, newA)
	}
	if newB != oldB {
		t.Fatalf("unchanged object changed: old=%#v new=%#v", oldB, newB)
	}
	if updated.SharedObjectID != oldShared {
		t.Fatalf("partial shared facts replaced: old=%s new=%s", oldShared, updated.SharedObjectID)
	}
}

func TestFullLanguagePreservesSharedRecordOrder(t *testing.T) {
	store := Store{Root: t.TempDir()}
	source := t.TempDir()
	if err := os.WriteFile(filepath.Join(source, "a.py"), []byte("value = 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "full.jsonl")
	writeAnalysisStream(t, path, []string{
		`{"adapter_version":"test","language":"python","mode":"full","record":"lexicon","repository":"repo","schema_version":1}`,
		`{"id":"a-unknown","kind":"file","name":"missing.py","owner":"missing.py","path":"missing.py","qualified_name":"missing.py","record":"node"}`,
		`{"id":"b-repository","kind":"repository","name":"repo","path":".","qualified_name":"repo","record":"node"}`,
		`{"id":"c-unknown","kind":"file","name":"other.py","owner":"other.py","path":"other.py","qualified_name":"other.py","record":"node"}`,
		`{"id":"d-owned","kind":"file","name":"a.py","owner":"a.py","path":"a.py","qualified_name":"a.py","record":"node"}`,
	})
	analysis, err := ReadAnalysis(path, "python")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := store.BuildFullLanguage(analysis, source, "python", "sha256:config", "sha256:adapter")
	if err != nil {
		t.Fatal(err)
	}
	shared, err := store.LoadObject(entry.SharedObjectID)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"a-unknown", "b-repository", "c-unknown"}
	if len(shared.Records) != len(want) {
		t.Fatalf("shared records = %d, want %d", len(shared.Records), len(want))
	}
	for index, raw := range shared.Records {
		var record struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(raw, &record); err != nil {
			t.Fatal(err)
		}
		if record.ID != want[index] {
			t.Fatalf("shared record %d = %q, want %q", index, record.ID, want[index])
		}
	}
}

func fileEntry(t *testing.T, entry LanguageEntry, path string) FileEntry {
	t.Helper()
	for _, file := range entry.Files {
		if file.Path == path {
			return file
		}
	}
	t.Fatalf("missing file entry %s", path)
	return FileEntry{}
}

func writeAnalysisStream(t *testing.T, path string, lines []string) {
	t.Helper()
	data := []byte{}
	for _, line := range lines {
		data = append(data, line...)
		data = append(data, '\n')
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}
