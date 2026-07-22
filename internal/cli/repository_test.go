package cli

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverRepositoryFindsNearestConfiguration(t *testing.T) {
	outer := t.TempDir()
	inner := filepath.Join(outer, "service", "cmd")
	if err := os.MkdirAll(inner, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(outer, ".lexicon"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(outer, "service", ".lexicon"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outer, ".lexicon", "config.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outer, "service", ".lexicon", "config.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	root, err := discoverRepository(inner)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(outer, "service")
	if root != want {
		t.Fatalf("repository root = %q, want %q", root, want)
	}
}

func TestDiscoverRepositoryReportsMissingConfiguration(t *testing.T) {
	_, err := discoverRepository(t.TempDir())
	if !errors.Is(err, errRepositoryNotFound) {
		t.Fatalf("discoverRepository error = %v, want repository-not-found error", err)
	}
}
