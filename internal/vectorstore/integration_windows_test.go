//go:build windows

package vectorstore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRustABIBuildOpenSearchClose(t *testing.T) {
	library, err := Load("")
	if err != nil {
		t.Skipf("Rust vector DLL is not built: %v", err)
	}
	defer library.Close()

	root := t.TempDir()
	store := filepath.Join(root, "store")
	ingest := filepath.Join(root, "ingest.jsonl")
	manifest := filepath.Join(root, "manifest.jsonl")
	snapshot := filepath.Join(root, "snapshot.gvs")
	if err := os.WriteFile(ingest, []byte(
		"{\"source\":\"a\",\"vector\":[1,0]}\n"+
			"{\"source\":\"b\",\"vector\":[0,1]}\n",
	), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifest, []byte(
		"{\"id\":\"alpha\",\"source\":\"a\"}\n"+
			"{\"id\":\"beta\",\"source\":\"b\"}\n",
	), 0o644); err != nil {
		t.Fatal(err)
	}
	if count, err := library.IngestJSONL(store, "test-model", ingest); err != nil || count != 2 {
		t.Fatalf("ingest count=%d err=%v", count, err)
	}
	if _, err := library.MaterializeJSONL(store, "test-model", manifest, snapshot); err != nil {
		t.Fatal(err)
	}
	engine, err := library.OpenSnapshot(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	info, err := engine.Info()
	if err != nil || info.Dimensions != 2 || info.Count != 2 {
		t.Fatalf("info=%+v err=%v", info, err)
	}
	hits, err := engine.Search([]float32{1, 0}, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 2 || hits[0].ID != "alpha" || hits[1].ID != "beta" {
		t.Fatalf("unexpected hits: %+v", hits)
	}
	if err := engine.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Search([]float32{1, 0}, 1); err == nil {
		t.Fatal("search after close unexpectedly succeeded")
	}
}
