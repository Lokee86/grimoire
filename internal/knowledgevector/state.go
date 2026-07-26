package knowledgevector

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/Lokee86/grimoire/internal/embedding"
	"github.com/Lokee86/grimoire/internal/knowledge"
)

const manifestVersion = 1

type Paths struct {
	Root     string
	Store    string
	Snapshot string
	Manifest string
	Ingest   string
	Records  string
}

type Manifest struct {
	Version           int      `json:"version"`
	KnowledgeIdentity string   `json:"knowledge_identity"`
	SnapshotIdentity  string   `json:"snapshot_identity"`
	Model             string   `json:"model"`
	Dimensions        int      `json:"dimensions"`
	Count             int      `json:"count"`
	Sources           []string `json:"sources,omitempty"`
}

type Entry struct {
	ID     string
	Source string
	Text   string
}

type BuildResult struct {
	Snapshot          string  `json:"snapshot"`
	SnapshotIdentity  string  `json:"snapshot_identity"`
	KnowledgeIdentity string  `json:"knowledge_identity"`
	Model             string  `json:"model"`
	Sections          int     `json:"sections"`
	UniqueVectors     int     `json:"unique_vectors"`
	EmbeddedVectors   int     `json:"embedded_vectors"`
	ReusedVectors     int     `json:"reused_vectors"`
	CachedSnapshot    bool    `json:"cached_snapshot"`
	SnapshotBytes     int64   `json:"snapshot_bytes"`
	DurationMS        float64 `json:"duration_ms"`
}

type Info struct {
	State             string `json:"state"`
	Snapshot          string `json:"snapshot"`
	Available         bool   `json:"available"`
	Current           bool   `json:"current"`
	KnowledgeIdentity string `json:"knowledge_identity,omitempty"`
	ExpectedIdentity  string `json:"expected_identity,omitempty"`
	SnapshotIdentity  string `json:"snapshot_identity,omitempty"`
	Model             string `json:"model,omitempty"`
	Dimensions        int    `json:"dimensions,omitempty"`
	Count             int    `json:"count,omitempty"`
	SnapshotBytes     int64  `json:"snapshot_bytes,omitempty"`
	Error             string `json:"error,omitempty"`
}

func ResolvePaths(state string) Paths {
	root := filepath.Join(state, "vectors", embedding.Identity())
	return Paths{
		Root: root, Store: filepath.Join(root, "store"),
		Snapshot: filepath.Join(root, "snapshot.gvs"),
		Manifest: filepath.Join(root, "snapshot.manifest.json"),
		Ingest:   filepath.Join(root, "ingest.next.jsonl"),
		Records:  filepath.Join(root, "records.next.jsonl"),
	}
}

func Available(state string) bool {
	paths := ResolvePaths(state)
	manifest, manifestErr := os.Stat(paths.Manifest)
	snapshot, snapshotErr := os.Stat(paths.Snapshot)
	return manifestErr == nil && snapshotErr == nil && manifest.Mode().IsRegular() && snapshot.Mode().IsRegular()
}

func IndexIdentity(index knowledge.Index) string {
	hash := sha256.New()
	_, _ = fmt.Fprintf(hash, "%d\x00", index.Version)
	for _, document := range index.Documents {
		_, _ = fmt.Fprintf(hash, "%s\x00%s\x00%s\x00", document.Path, document.Kind, document.Hash)
		for _, section := range document.Sections {
			_, _ = fmt.Fprintf(hash, "%s\x00%s\x00", section.ID, section.Hash)
		}
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func Entries(index knowledge.Index) []Entry {
	result := make([]Entry, 0)
	for _, document := range index.Documents {
		for _, section := range document.Sections {
			if section.Text == "" {
				continue
			}
			result = append(result, Entry{ID: section.ID, Source: section.Hash, Text: section.Text})
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}
