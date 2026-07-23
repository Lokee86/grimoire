package app

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/Lokee86/grimoire/internal/index"
)

func TestRunIndexExcludesRepeatedPaths(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "included.go"), []byte("package fixture\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "evaluation", "retrieval"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "evaluation", "retrieval", "cases.json"), []byte(`{"answer":"included.go"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	state := filepath.Join(root, ".isolated-state")
	if err := Run([]string{
		"index", "--root", root, "--state", state,
		"--exclude", "evaluation/retrieval",
	}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	snapshot, err := index.Load(state)
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Files) != 1 || snapshot.Files[0].Path != "included.go" {
		t.Fatalf("unexpected indexed files: %+v", snapshot.Files)
	}
}
