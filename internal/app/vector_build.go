package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Lokee86/grimoire/internal/embedding"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/vectorstore"
)

type vectorChunk struct {
	Chunk  index.Chunk
	Source string
}

type vectorBuildResult struct {
	Snapshot         string  `json:"snapshot"`
	Identity         string  `json:"identity"`
	PreparedIdentity string  `json:"prepared_identity"`
	Model            string  `json:"model"`
	Chunks           int     `json:"chunks"`
	UniqueVectors    int     `json:"unique_vectors"`
	Embedded         int     `json:"embedded"`
	EmbeddedVectors  int     `json:"embedded_vectors"`
	Reused           int     `json:"reused"`
	ObjectChecks     int     `json:"object_checks"`
	CachedSnapshot   bool    `json:"cached_snapshot"`
	DurationMS       float64 `json:"duration_ms"`
	SnapshotBytes    int64   `json:"snapshot_bytes"`
	PeakMemoryBytes  uint64  `json:"peak_memory_bytes"`
}

func runVectorBuild(args []string, stdout, stderr io.Writer) error {
	started := time.Now()
	memory := startMemoryPeakSampler()
	memoryStopped := false
	defer func() {
		if !memoryStopped {
			memory.stopAndRead()
		}
	}()
	flags := flag.NewFlagSet("vector build", flag.ContinueOnError)
	flags.SetOutput(stderr)
	root := flags.String("root", ".", "repository root")
	state := flags.String("state", "", "prepared index repository path")
	endpoint := flags.String("endpoint", embedding.DefaultEndpoint, "OpenAI-compatible embeddings endpoint")
	enginePath := flags.String("engine", "", "Rust vector engine DLL")
	batchSize := flags.Int("batch-size", 4, "documents embedded per request")
	batchConcurrency := flags.Int("batch-concurrency", 1, "concurrent embedding requests")
	timeout := flags.Duration("timeout", 30*time.Minute, "complete vector build timeout")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *batchSize <= 0 || *batchConcurrency <= 0 || *timeout <= 0 {
		return errors.New("--batch-size, --batch-concurrency, and --timeout must be positive")
	}
	statePath, err := resolveState(*root, *state)
	if err != nil {
		return err
	}
	paths := resolveVectorPaths(statePath)
	if err := os.MkdirAll(paths.Root, 0o755); err != nil {
		return err
	}
	defer os.Remove(paths.Ingest)
	defer os.Remove(paths.Records)

	if base, baseErr := index.RebuildBase(statePath); baseErr == nil && base.Identity() != "" {
		if manifest, snapshotInfo, reusable := reusableVectorSnapshot(paths, base.Identity(), 0); reusable {
			peakMemory := memory.stopAndRead()
			memoryStopped = true
			uniqueVectors := len(manifest.Sources)
			if uniqueVectors == 0 {
				uniqueVectors = manifest.Count
			}
			_, _ = fmt.Fprintf(stderr, "vector build: snapshot is current; reused %d chunks in %s\n", manifest.Count, formatVectorDuration(time.Since(started)))
			return writeJSON(stdout, vectorBuildResult{
				Snapshot: paths.Snapshot, Identity: manifest.SnapshotIdentity, PreparedIdentity: base.Identity(),
				Model: embedding.Identity(), Chunks: manifest.Count, UniqueVectors: uniqueVectors,
				Reused: manifest.Count, CachedSnapshot: true, DurationMS: durationMS(time.Since(started)),
				SnapshotBytes: snapshotInfo.Size(), PeakMemoryBytes: peakMemory,
			})
		}
	}

	snapshot, err := index.Load(statePath)
	if err != nil {
		return fmt.Errorf("load prepared index: %w", err)
	}
	chunks := snapshot.AllChunks()
	if len(chunks) == 0 {
		return errors.New("prepared index has no chunks")
	}
	preparedIdentity := snapshot.Identity()
	if preparedIdentity == "" {
		return errors.New("prepared index has no published identity")
	}
	_, _ = fmt.Fprintf(stderr, "vector build: prepared %d chunks\n", len(chunks))

	library, err := vectorstore.Load(*enginePath)
	if err != nil {
		return err
	}
	defer library.Close()

	all, unique, sourceCounts := vectorEntries(chunks)
	previousSources := reusableVectorSources(paths)
	missing := make([]vectorChunk, 0)
	objectChecks := 0
	for _, entry := range unique {
		if _, reused := previousSources[entry.Source]; reused {
			continue
		}
		objectChecks++
		exists, existsErr := library.ObjectExists(paths.Store, embedding.Identity(), entry.Source)
		if existsErr != nil {
			return existsErr
		}
		if !exists {
			missing = append(missing, entry)
		}
	}
	embeddedChunks := 0
	for _, entry := range missing {
		embeddedChunks += sourceCounts[entry.Source]
	}
	_, _ = fmt.Fprintf(
		stderr,
		"vector build: %d unique vectors; %d cached, %d to embed, %d object checks\n",
		len(unique), len(unique)-len(missing), len(missing), objectChecks,
	)

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	if len(missing) > 0 {
		progress := newVectorEmbeddingProgress(stderr, len(missing))
		if err := embedMissing(
			ctx,
			embedding.NewClient(*endpoint),
			library,
			paths,
			missing,
			*batchSize,
			*batchConcurrency,
			progress.complete,
		); err != nil {
			return err
		}
	}
	_, _ = fmt.Fprintf(stderr, "vector build: materializing %d chunk records\n", len(all))
	if err := writeVectorRecords(paths.Records, all); err != nil {
		return err
	}
	identity, err := library.MaterializeJSONL(paths.Store, embedding.Identity(), paths.Records, paths.Snapshot)
	if err != nil {
		return err
	}
	manifest := vectorSnapshotManifest{
		Version:          vectorSnapshotManifestVersion,
		PreparedIdentity: preparedIdentity,
		SnapshotIdentity: identity,
		Model:            embedding.Identity(),
		Dimensions:       embedding.Dimensions,
		Count:            len(all),
		Sources:          vectorSources(unique),
	}
	if err := writeVectorSnapshotManifest(paths.Manifest, manifest); err != nil {
		return err
	}
	snapshotInfo, err := os.Stat(paths.Snapshot)
	if err != nil {
		return err
	}
	peakMemory := memory.stopAndRead()
	memoryStopped = true
	_, _ = fmt.Fprintf(stderr, "vector build: complete in %s\n", formatVectorDuration(time.Since(started)))
	return writeJSON(stdout, vectorBuildResult{
		Snapshot: paths.Snapshot, Identity: identity, PreparedIdentity: preparedIdentity,
		Model: embedding.Identity(), Chunks: len(all), UniqueVectors: len(unique),
		Embedded: embeddedChunks, EmbeddedVectors: len(missing), Reused: len(all) - embeddedChunks,
		ObjectChecks: objectChecks, DurationMS: durationMS(time.Since(started)),
		SnapshotBytes: snapshotInfo.Size(), PeakMemoryBytes: peakMemory,
	})
}
