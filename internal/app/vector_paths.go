package app

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"

	"github.com/Lokee86/grimoire/internal/embedding"
)

type vectorStatePaths struct {
	Root     string
	Store    string
	Snapshot string
	Manifest string
	Ingest   string
	Records  string
}

func resolveVectorPaths(state string) vectorStatePaths {
	root := filepath.Join(state, "vectors", embedding.Identity())
	return vectorStatePaths{
		Root:     root,
		Store:    filepath.Join(root, "store"),
		Snapshot: filepath.Join(root, "snapshot.gvs"),
		Manifest: filepath.Join(root, "snapshot.manifest.json"),
		Ingest:   filepath.Join(root, "ingest.next.jsonl"),
		Records:  filepath.Join(root, "records.next.jsonl"),
	}
}

func vectorSource(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}
