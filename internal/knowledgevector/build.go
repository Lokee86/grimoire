package knowledgevector

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/Lokee86/grimoire/internal/embedding"
	"github.com/Lokee86/grimoire/internal/knowledge"
	"github.com/Lokee86/grimoire/internal/vectorstore"
)

type BuildOptions struct {
	State       string
	Endpoint    string
	EnginePath  string
	BatchSize   int
	Concurrency int
}

func Build(ctx context.Context, index knowledge.Index, options BuildOptions) (BuildResult, error) {
	started := time.Now()
	if options.State == "" {
		return BuildResult{}, errors.New("knowledge vector state is required")
	}
	if options.BatchSize <= 0 {
		options.BatchSize = 8
	}
	if options.Concurrency <= 0 {
		options.Concurrency = 1
	}
	entries := Entries(index)
	if len(entries) == 0 {
		return BuildResult{}, errors.New("knowledge index has no sections")
	}
	paths := ResolvePaths(options.State)
	if err := os.MkdirAll(paths.Root, 0o755); err != nil {
		return BuildResult{}, err
	}
	defer os.Remove(paths.Ingest)
	defer os.Remove(paths.Records)

	library, err := vectorstore.Load(options.EnginePath)
	if err != nil {
		return BuildResult{}, err
	}
	defer library.Close()

	unique := uniqueEntries(entries)
	if current, ok := reusableBuildResult(library, paths, index, len(entries), len(unique), started); ok {
		return current, nil
	}
	missing := make([]Entry, 0)
	for _, entry := range unique {
		exists, existsErr := library.ObjectExists(paths.Store, embedding.Identity(), entry.Source)
		if existsErr != nil {
			return BuildResult{}, existsErr
		}
		if !exists {
			missing = append(missing, entry)
		}
	}
	if err := embedEntries(ctx, embedding.NewClient(options.Endpoint), library, paths, missing, options.BatchSize, options.Concurrency); err != nil {
		return BuildResult{}, err
	}
	if err := writeRecords(paths.Records, entries); err != nil {
		return BuildResult{}, err
	}
	snapshotIdentity, err := library.MaterializeJSONL(paths.Store, embedding.Identity(), paths.Records, paths.Snapshot)
	if err != nil {
		return BuildResult{}, err
	}
	sources := make([]string, 0, len(unique))
	for _, entry := range unique {
		sources = append(sources, entry.Source)
	}
	sort.Strings(sources)
	manifest := Manifest{
		Version: manifestVersion, KnowledgeIdentity: IndexIdentity(index), SnapshotIdentity: snapshotIdentity,
		Model: embedding.Identity(), Dimensions: embedding.Dimensions, Count: len(entries), Sources: sources,
	}
	if err := writeManifest(paths.Manifest, manifest); err != nil {
		return BuildResult{}, err
	}
	info, err := os.Stat(paths.Snapshot)
	if err != nil {
		return BuildResult{}, err
	}
	return BuildResult{
		Snapshot: paths.Snapshot, SnapshotIdentity: snapshotIdentity, KnowledgeIdentity: manifest.KnowledgeIdentity,
		Model: manifest.Model, Sections: len(entries), UniqueVectors: len(unique), EmbeddedVectors: len(missing),
		ReusedVectors: len(unique) - len(missing), SnapshotBytes: info.Size(),
		DurationMS: float64(time.Since(started)) / float64(time.Millisecond),
	}, nil
}

func reusableBuildResult(library *vectorstore.Library, paths Paths, index knowledge.Index, sections, unique int, started time.Time) (BuildResult, bool) {
	manifest, err := ReadManifest(paths.Manifest)
	if err != nil {
		return BuildResult{}, false
	}
	engine, err := library.OpenSnapshot(paths.Snapshot)
	if err != nil {
		return BuildResult{}, false
	}
	defer engine.Close()
	info, err := engine.Info()
	if err != nil || validateManifest(manifest, index, info) != nil {
		return BuildResult{}, false
	}
	fileInfo, err := os.Stat(paths.Snapshot)
	if err != nil {
		return BuildResult{}, false
	}
	return BuildResult{
		Snapshot: paths.Snapshot, SnapshotIdentity: manifest.SnapshotIdentity,
		KnowledgeIdentity: manifest.KnowledgeIdentity, Model: manifest.Model,
		Sections: sections, UniqueVectors: unique, ReusedVectors: unique,
		CachedSnapshot: true, SnapshotBytes: fileInfo.Size(),
		DurationMS: float64(time.Since(started)) / float64(time.Millisecond),
	}, true
}

func uniqueEntries(entries []Entry) []Entry {
	seen := make(map[string]bool, len(entries))
	result := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		if seen[entry.Source] {
			continue
		}
		seen[entry.Source] = true
		result = append(result, entry)
	}
	return result
}

func writeRecords(path string, entries []Entry) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(file)
	encoder := json.NewEncoder(writer)
	for _, entry := range entries {
		if err := encoder.Encode(struct {
			ID     string `json:"id"`
			Source string `json:"source"`
		}{entry.ID, entry.Source}); err != nil {
			file.Close()
			return err
		}
	}
	if err := writer.Flush(); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

func ingestBatch(library *vectorstore.Library, paths Paths, entries []Entry, vectors [][]float32) error {
	if len(entries) != len(vectors) {
		return fmt.Errorf("embedding provider returned %d vectors for %d knowledge sections", len(vectors), len(entries))
	}
	file, err := os.Create(paths.Ingest)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(file)
	encoder := json.NewEncoder(writer)
	for index, vector := range vectors {
		if err := encoder.Encode(struct {
			Source string    `json:"source"`
			Vector []float32 `json:"vector"`
		}{entries[index].Source, vector}); err != nil {
			file.Close()
			return err
		}
	}
	if err := writer.Flush(); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	count, err := library.IngestJSONL(paths.Store, embedding.Identity(), paths.Ingest)
	if err != nil {
		return err
	}
	if count != uint64(len(entries)) {
		return fmt.Errorf("vector engine ingested %d knowledge vectors, expected %d", count, len(entries))
	}
	return nil
}
