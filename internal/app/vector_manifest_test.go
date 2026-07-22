package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/grimoire/internal/embedding"
	"github.com/Lokee86/grimoire/internal/index"
)

func TestVectorSnapshotManifestRequiresExactPreparedIdentity(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "alpha.go"), []byte("package alpha\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Run([]string{"index", "--root", root}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	statePath, err := resolveState(root, "")
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := index.Load(statePath)
	if err != nil {
		t.Fatal(err)
	}
	paths := resolveVectorPaths(statePath)
	if err := os.MkdirAll(paths.Root, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := vectorSnapshotManifest{
		Version:          vectorSnapshotManifestVersion,
		PreparedIdentity: "stale-prepared-identity",
		SnapshotIdentity: "vector-identity",
		Model:            embedding.Identity(),
		Dimensions:       embedding.Dimensions,
		Count:            len(snapshot.AllChunks()),
	}
	if err := writeVectorSnapshotManifest(paths.Manifest, manifest); err != nil {
		t.Fatal(err)
	}
	if _, err := validateVectorSnapshotManifest(paths.Manifest, snapshot, manifest.Count); err == nil || !strings.Contains(err.Error(), "current prepared index is "+snapshot.Identity()) {
		t.Fatalf("expected exact identity mismatch, got %v", err)
	}

	manifest.PreparedIdentity = snapshot.Identity()
	if err := writeVectorSnapshotManifest(paths.Manifest, manifest); err != nil {
		t.Fatal(err)
	}
	loaded, err := validateVectorSnapshotManifest(paths.Manifest, snapshot, manifest.Count)
	if err != nil {
		t.Fatal(err)
	}
	if loaded != manifest {
		t.Fatalf("manifest changed during round trip: %+v != %+v", loaded, manifest)
	}
}
