package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Lokee86/grimoire/internal/embedding"
	"github.com/Lokee86/grimoire/internal/index"
)

func TestVectorEntriesDeduplicatesSharedText(t *testing.T) {
	chunks := []index.Chunk{
		{ID: "one", Text: "same"},
		{ID: "two", Text: "same"},
		{ID: "three", Text: "different"},
	}
	all, unique, counts := vectorEntries(chunks)
	if len(all) != 3 {
		t.Fatalf("all entries = %d, want 3", len(all))
	}
	if len(unique) != 2 {
		t.Fatalf("unique entries = %d, want 2", len(unique))
	}
	if counts[vectorSource("same")] != 2 {
		t.Fatalf("shared source count = %d, want 2", counts[vectorSource("same")])
	}
}

func TestReusableVectorSnapshotMatchesPreparedIdentity(t *testing.T) {
	paths := resolveVectorPaths(t.TempDir())
	if err := os.MkdirAll(paths.Root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(paths.Snapshot, []byte("snapshot"), 0o644); err != nil {
		t.Fatal(err)
	}
	manifest := vectorSnapshotManifest{
		Version:          vectorSnapshotManifestVersion,
		PreparedIdentity: "prepared",
		SnapshotIdentity: "vectors",
		Model:            embedding.Identity(),
		Dimensions:       embedding.Dimensions,
		Count:            3,
		Sources:          []string{"a", "b"},
	}
	if err := writeVectorSnapshotManifest(paths.Manifest, manifest); err != nil {
		t.Fatal(err)
	}
	got, info, ok := reusableVectorSnapshot(paths, "prepared", 3)
	if !ok {
		t.Fatal("expected reusable snapshot")
	}
	if got.SnapshotIdentity != "vectors" || info.Size() == 0 {
		t.Fatalf("unexpected reusable snapshot: %#v, size=%d", got, info.Size())
	}
	if _, _, ok := reusableVectorSnapshot(paths, "changed", 3); ok {
		t.Fatal("changed prepared identity reused stale snapshot")
	}
}

func TestReusableVectorSourcesRequiresPublishedSnapshot(t *testing.T) {
	state := t.TempDir()
	paths := resolveVectorPaths(state)
	if err := os.MkdirAll(filepath.Dir(paths.Manifest), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := vectorSnapshotManifest{
		Version:          vectorSnapshotManifestVersion,
		PreparedIdentity: "old",
		SnapshotIdentity: "vectors",
		Model:            embedding.Identity(),
		Dimensions:       embedding.Dimensions,
		Count:            2,
		Sources:          []string{"one", "two"},
	}
	if err := writeVectorSnapshotManifest(paths.Manifest, manifest); err != nil {
		t.Fatal(err)
	}
	if sources := reusableVectorSources(paths); sources != nil {
		t.Fatalf("sources reused without snapshot: %#v", sources)
	}
	if err := os.WriteFile(paths.Snapshot, []byte("snapshot"), 0o644); err != nil {
		t.Fatal(err)
	}
	sources := reusableVectorSources(paths)
	if len(sources) != 2 {
		t.Fatalf("reusable sources = %d, want 2", len(sources))
	}
}
