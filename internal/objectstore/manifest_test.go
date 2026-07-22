package objectstore

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Lokee86/lexicon/internal/adapters"
)

func TestBuildManifestPersistsAdapterFingerprint(t *testing.T) {
	root := t.TempDir()
	stateRoot := filepath.Join(root, "repo")
	adapterRoot := filepath.Join(root, "adapters")
	if err := os.MkdirAll(filepath.Join(stateRoot, "source", "library"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(adapterRoot, "python"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeManifestFixture(t, filepath.Join(stateRoot, "source", "main.py"), "value = 1\n")
	writeManifestFixture(t, filepath.Join(stateRoot, "library", "python.jsonl"), "{\"adapter_version\":\"test\",\"language\":\"python\",\"record\":\"lexicon\",\"repository\":\"test\",\"schema_version\":1}\n")
	writeManifestFixture(t, filepath.Join(adapterRoot, "python", "adapter.py"), "value = 1\n")

	store := Store{Root: root}
	manifest, err := store.BuildManifest(stateRoot, "commit", "config", adapterRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Languages) != 1 {
		t.Fatalf("languages = %#v", manifest.Languages)
	}
	expected, err := adapters.Fingerprint(adapterRoot, "python")
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Languages[0].AdapterFingerprint != expected {
		t.Fatalf("fingerprint = %q, want %q", manifest.Languages[0].AdapterFingerprint, expected)
	}

	id, err := store.Publish(manifest)
	if err != nil {
		t.Fatal(err)
	}
	_, published, err := store.Current()
	if err != nil {
		t.Fatal(err)
	}
	if published.Languages[0].AdapterFingerprint != expected {
		t.Fatalf("published fingerprint = %q, want %q", published.Languages[0].AdapterFingerprint, expected)
	}
	if id == "" {
		t.Fatal("expected snapshot ID")
	}
}

func writeManifestFixture(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}
