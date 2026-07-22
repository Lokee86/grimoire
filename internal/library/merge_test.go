package library

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMergeReplacesSelectedOwnersAndSharedFacts(t *testing.T) {
	root := t.TempDir()
	full := filepath.Join(root, "full.jsonl")
	incremental := filepath.Join(root, "incremental.jsonl")
	merged := filepath.Join(root, "merged.jsonl")
	writeTestStream(t, full, []string{
		`{"adapter_version":"test","language":"python","mode":"full","record":"lexicon","repository":"repo","schema_version":1}`,
		`{"id":"repo-old","kind":"repository","name":"repo","path":".","qualified_name":"repo","record":"node"}`,
		`{"id":"a-old","kind":"file","name":"a.py","owner":"a.py","path":"a.py","qualified_name":"a.py","record":"node"}`,
		`{"id":"b","kind":"file","name":"b.py","owner":"b.py","path":"b.py","qualified_name":"b.py","record":"node"}`,
	})
	writeTestStream(t, incremental, []string{
		`{"adapter_version":"test","changed_files":["a.py"],"language":"python","mode":"incremental","record":"lexicon","removed_files":[],"repository":"repo","schema_version":1,"shared_complete":true}`,
		`{"id":"repo-new","kind":"repository","name":"repo","path":".","qualified_name":"repo","record":"node"}`,
		`{"id":"a-new","kind":"file","name":"a.py","owner":"a.py","path":"a.py","qualified_name":"a.py","record":"node"}`,
	})
	if err := Merge(full, incremental, merged); err != nil {
		t.Fatal(err)
	}
	header, records, err := readStream(merged)
	if err != nil {
		t.Fatal(err)
	}
	if header["mode"] != "full" || header["changed_files"] != nil || header["shared_complete"] != nil {
		t.Fatalf("unexpected merged header: %#v", header)
	}
	ids := make(map[string]bool)
	for _, record := range records {
		if id, _ := record["id"].(string); id != "" {
			ids[id] = true
		}
	}
	for _, expected := range []string{"repo-new", "a-new", "b"} {
		if !ids[expected] {
			t.Fatalf("missing %s in %#v", expected, ids)
		}
	}
	for _, stale := range []string{"repo-old", "a-old"} {
		if ids[stale] {
			t.Fatalf("retained stale record %s", stale)
		}
	}
}

func TestMergeRetainsCompleteSharedFactsForPartialScope(t *testing.T) {
	root := t.TempDir()
	full := filepath.Join(root, "full.jsonl")
	incremental := filepath.Join(root, "incremental.jsonl")
	merged := filepath.Join(root, "merged.jsonl")
	writeTestStream(t, full, []string{
		`{"adapter_version":"test","language":"python","record":"lexicon","repository":"repo","schema_version":1}`,
		`{"id":"repo-old","kind":"repository","name":"repo","path":".","qualified_name":"repo","record":"node"}`,
		`{"id":"a-old","kind":"file","owner":"a.py","path":"a.py","qualified_name":"a.py","record":"node"}`,
	})
	writeTestStream(t, incremental, []string{
		`{"adapter_version":"test","changed_files":["a.py"],"language":"python","mode":"incremental","record":"lexicon","removed_files":[],"repository":"repo","schema_version":1,"shared_complete":false}`,
		`{"id":"repo-scoped","kind":"repository","name":"repo","path":".","qualified_name":"repo","record":"node"}`,
		`{"id":"a-new","kind":"file","owner":"a.py","path":"a.py","qualified_name":"a.py","record":"node"}`,
	})
	if err := Merge(full, incremental, merged); err != nil {
		t.Fatal(err)
	}
	_, records, err := readStream(merged)
	if err != nil {
		t.Fatal(err)
	}
	ids := make(map[string]bool)
	for _, record := range records {
		id, _ := record["id"].(string)
		ids[id] = true
	}
	if !ids["repo-old"] || !ids["a-new"] || ids["repo-scoped"] || ids["a-old"] {
		t.Fatalf("unexpected partial merge records: %#v", ids)
	}
}

func TestMergeRejectsUndeclaredOwner(t *testing.T) {
	root := t.TempDir()
	full := filepath.Join(root, "full.jsonl")
	incremental := filepath.Join(root, "incremental.jsonl")
	writeTestStream(t, full, []string{
		`{"adapter_version":"test","language":"python","record":"lexicon","repository":"repo","schema_version":1}`,
	})
	writeTestStream(t, incremental, []string{
		`{"adapter_version":"test","changed_files":["a.py"],"language":"python","mode":"incremental","record":"lexicon","removed_files":[],"repository":"repo","schema_version":1,"shared_complete":true}`,
		`{"id":"b","kind":"file","name":"b.py","owner":"b.py","path":"b.py","qualified_name":"b.py","record":"node"}`,
	})
	if err := Merge(full, incremental, filepath.Join(root, "merged.jsonl")); err == nil {
		t.Fatal("expected undeclared owner to fail")
	}
}

func writeTestStream(t *testing.T, path string, lines []string) {
	t.Helper()
	data := []byte{}
	for _, line := range lines {
		var value map[string]any
		if err := json.Unmarshal([]byte(line), &value); err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		data = append(data, encoded...)
		data = append(data, '\n')
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}
