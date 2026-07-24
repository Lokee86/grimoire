package objectstore

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestIngestLanguageCreatesOwnedAndSharedObjects(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "main.py"), []byte("value = 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(root, "python.jsonl")
	data := "{\"adapter_version\":\"1.2.3\",\"language\":\"python\",\"record\":\"lexicon\",\"repository\":\"example\",\"schema_version\":1}\n" +
		"{\"id\":\"sha256:file\",\"kind\":\"file\",\"name\":\"main.py\",\"path\":\"main.py\",\"qualified_name\":\"main.py\",\"record\":\"node\"}\n" +
		"{\"id\":\"sha256:repo\",\"kind\":\"repository\",\"name\":\"example\",\"path\":\"\",\"qualified_name\":\"example\",\"record\":\"node\"}\n"
	if err := os.WriteFile(output, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	store := Store{Root: root}
	entry, err := store.IngestLanguage(output, source, "python", "sha256:config")
	if err != nil {
		t.Fatal(err)
	}
	if len(entry.Files) != 1 || entry.Files[0].Path != "main.py" {
		t.Fatalf("files = %#v", entry.Files)
	}
	if entry.SharedObjectID == "" {
		t.Fatal("expected shared object")
	}
	for _, id := range []string{entry.Files[0].ObjectID, entry.SharedObjectID} {
		if _, err := os.Stat(store.ObjectPath(id)); err != nil {
			t.Fatalf("object %s: %v", id, err)
		}
		if _, err := store.LoadObject(id); err != nil {
			t.Fatalf("load object %s: %v", id, err)
		}
	}
	again, err := store.IngestLanguage(output, source, "python", "sha256:config")
	if err != nil {
		t.Fatal(err)
	}
	if again.Files[0].ObjectID != entry.Files[0].ObjectID || again.SharedObjectID != entry.SharedObjectID {
		t.Fatal("identical input produced different object IDs")
	}
}

func TestSnapshotContentVerification(t *testing.T) {
	store := Store{Root: t.TempDir()}
	id, err := store.Publish(Manifest{StateCommit: "abc"})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(store.snapshotPath(id), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Load(id); err == nil {
		t.Fatal("expected corrupted snapshot to fail verification")
	}
}

func TestPublishCurrentSnapshot(t *testing.T) {
	store := Store{Root: t.TempDir()}
	manifest := Manifest{StateCommit: "abc", Languages: []LanguageEntry{}}
	id, err := store.Publish(manifest)
	if err != nil {
		t.Fatal(err)
	}
	currentID, current, err := store.Current()
	if err != nil {
		t.Fatal(err)
	}
	if currentID != id || current.StateCommit != "abc" {
		t.Fatalf("current = %s %#v", currentID, current)
	}
	data, err := os.ReadFile(store.snapshotPath(id))
	if err != nil {
		t.Fatal(err)
	}
	var decoded Manifest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Version != SnapshotVersion {
		t.Fatalf("version = %d", decoded.Version)
	}
}
