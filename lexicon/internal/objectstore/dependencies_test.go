package objectstore

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestImpactedFilesIncludesTransitiveDependentsAndUnresolvedOwners(t *testing.T) {
	store := Store{Root: t.TempDir()}
	metadata := FactObject{
		Language:         "python",
		AdapterVersion:   "test",
		SchemaVersion:    1,
		AnalysisConfigID: "sha256:config",
	}
	files := []struct {
		path    string
		records []json.RawMessage
	}{
		{"a.py", records(
			`{"id":"node-a","kind":"function","record":"node"}`,
			`{"id":"node-a-child","kind":"variable","record":"node"}`,
			`{"record":"edge","relation":"contains","source":"node-a","target":"node-a-child"}`,
		)},
		{"b.py", records(
			`{"id":"node-b","kind":"function","record":"node"}`,
			`{"record":"edge","relation":"calls","source":"node-b","target":"node-a"}`,
		)},
		{"c.py", records(
			`{"id":"node-c","kind":"function","record":"node"}`,
			`{"record":"edge","relation":"calls","source":"node-c","target":"node-b"}`,
		)},
		{"d.py", records(
			`{"id":"node-d","kind":"function","record":"node"}`,
			`{"reason":"missing-target","record":"unresolved","source":"node-d"}`,
		)},
		{"e.py", records(
			`{"id":"node-e","kind":"function","record":"node"}`,
			`{"reason":"builtin-target","record":"unresolved","source":"node-e"}`,
		)},
	}
	entries := make([]FileEntry, 0, len(files))
	for _, file := range files {
		object := metadata
		object.Owner = file.path
		object.SourceContentID = ContentID([]byte(file.path))
		object.Records = file.records
		id, err := store.WriteObject(object)
		if err != nil {
			t.Fatal(err)
		}
		entries = append(entries, FileEntry{
			Path: file.path, Language: "python", ContentID: object.SourceContentID, ObjectID: id,
		})
	}
	_, err := store.Publish(Manifest{
		StateCommit: "state",
		Languages: []LanguageEntry{{
			Language: "python", AdapterVersion: "test", SchemaVersion: 1,
			Repository: "repo", AnalysisConfigID: "sha256:config", Files: entries,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	impacted, err := store.ImpactedFiles("python", []string{"a.py"})
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"a.py", "b.py", "c.py", "d.py", "e.py"}
	if !reflect.DeepEqual(impacted, expected) {
		t.Fatalf("impacted = %v, want %v", impacted, expected)
	}
	emit, context, err := store.DependencyScope("python", []string{"b.py"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(emit, []string{"b.py", "c.py", "d.py", "e.py"}) {
		t.Fatalf("emit = %v", emit)
	}
	if !reflect.DeepEqual(context, []string{"a.py", "b.py", "c.py", "d.py", "e.py"}) {
		t.Fatalf("context = %v", context)
	}
	full, err := store.DirectChangesRequireFull("python", []string{"a.py"})
	if err != nil || full {
		t.Fatalf("local leaf required full analysis: full=%v err=%v", full, err)
	}
	full, err = store.DirectChangesRequireFull("python", []string{"e.py"})
	if err != nil || full {
		t.Fatalf("stable builtin unresolved required full analysis: full=%v err=%v", full, err)
	}
	for _, path := range []string{"b.py", "d.py"} {
		full, err = store.DirectChangesRequireFull("python", []string{path})
		if err != nil || !full {
			t.Fatalf("%s did not require full analysis: full=%v err=%v", path, full, err)
		}
	}
}

func records(values ...string) []json.RawMessage {
	result := make([]json.RawMessage, len(values))
	for index, value := range values {
		result[index] = json.RawMessage(value)
	}
	return result
}
