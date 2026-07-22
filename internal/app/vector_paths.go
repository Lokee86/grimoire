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
	Ingest   string
	Manifest string
}

func resolveVectorPaths(state string) vectorStatePaths {
	root := filepath.Join(state, "vectors", embedding.Identity())
	return vectorStatePaths{
		Root:     root,
		Store:    filepath.Join(root, "store"),
		Snapshot: filepath.Join(root, "snapshot.gvs"),
		Ingest:   filepath.Join(root, "ingest.next.jsonl"),
		Manifest: filepath.Join(root, "manifest.next.jsonl"),
	}
}

func vectorSource(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}
