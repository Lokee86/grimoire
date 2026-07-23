package app

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/Lokee86/grimoire/internal/embedding"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/vectorstore"
)

type vectorChunk struct {
	Chunk  index.Chunk
	Source string
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
	snapshot, err := index.Load(statePath)
	if err != nil {
		return fmt.Errorf("load prepared index: %w", err)
	}
	chunks := snapshot.AllChunks()
	if len(chunks) == 0 {
		return errors.New("prepared index has no chunks")
	}
	paths := resolveVectorPaths(statePath)
	if err := os.MkdirAll(paths.Root, 0o755); err != nil {
		return err
	}
	defer os.Remove(paths.Ingest)
	defer os.Remove(paths.Records)

	library, err := vectorstore.Load(*enginePath)
	if err != nil {
		return err
	}
	defer library.Close()

	all := make([]vectorChunk, 0, len(chunks))
	for _, chunk := range chunks {
		all = append(all, vectorChunk{Chunk: chunk, Source: vectorSource(chunk.Text)})
	}
	missing := make([]vectorChunk, 0)
	for _, entry := range uniqueVectorSources(all) {
		exists, existsErr := library.ObjectExists(paths.Store, embedding.Identity(), entry.Source)
		if existsErr != nil {
			return existsErr
		}
		if !exists {
			missing = append(missing, entry)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	if len(missing) > 0 {
		if err := embedMissing(ctx, embedding.NewClient(*endpoint), library, paths, missing, *batchSize, *batchConcurrency); err != nil {
			return err
		}
	}
	preparedIdentity := snapshot.Identity()
	if preparedIdentity == "" {
		return errors.New("prepared index has no published identity")
	}
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
	return writeJSON(stdout, struct {
		Snapshot         string  `json:"snapshot"`
		Identity         string  `json:"identity"`
		PreparedIdentity string  `json:"prepared_identity"`
		Model            string  `json:"model"`
		Chunks           int     `json:"chunks"`
		Embedded         int     `json:"embedded"`
		Reused           int     `json:"reused"`
		DurationMS       float64 `json:"duration_ms"`
		SnapshotBytes    int64   `json:"snapshot_bytes"`
		PeakMemoryBytes  uint64  `json:"peak_memory_bytes"`
	}{
		Snapshot: paths.Snapshot, Identity: identity, PreparedIdentity: preparedIdentity,
		Model: embedding.Identity(), Chunks: len(all), Embedded: len(missing),
		Reused: len(all) - len(missing), DurationMS: durationMS(time.Since(started)),
		SnapshotBytes: snapshotInfo.Size(), PeakMemoryBytes: peakMemory,
	})
}

func uniqueVectorSources(entries []vectorChunk) []vectorChunk {
	seen := make(map[string]struct{}, len(entries))
	result := make([]vectorChunk, 0, len(entries))
	for _, entry := range entries {
		if _, exists := seen[entry.Source]; exists {
			continue
		}
		seen[entry.Source] = struct{}{}
		result = append(result, entry)
	}
	return result
}

type documentEmbedder interface {
	EmbedDocuments(context.Context, []string) ([][]float32, error)
}

type embeddedVectorBatch struct {
	batch   []vectorChunk
	vectors [][]float32
	err     error
}

func embedMissing(
	ctx context.Context,
	client documentEmbedder,
	library *vectorstore.Library,
	paths vectorStatePaths,
	missing []vectorChunk,
	batchSize int,
	concurrency int,
) error {
	return embedVectorBatches(ctx, client, missing, batchSize, concurrency, func(batch []vectorChunk, vectors [][]float32) error {
		return ingestVectorBatch(library, paths, batch, vectors)
	})
}

func embedVectorBatches(
	ctx context.Context,
	client documentEmbedder,
	missing []vectorChunk,
	batchSize int,
	concurrency int,
	ingest func([]vectorChunk, [][]float32) error,
) error {
	if len(missing) == 0 {
		return nil
	}
	if batchSize <= 0 || concurrency <= 0 {
		return errors.New("embedding batch size and concurrency must be positive")
	}
	workerCount := min(concurrency, (len(missing)+batchSize-1)/batchSize)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan []vectorChunk)
	results := make(chan embeddedVectorBatch, workerCount)
	var workers sync.WaitGroup
	for range workerCount {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for batch := range jobs {
				documents := make([]string, len(batch))
				for index, entry := range batch {
					documents[index] = entry.Chunk.Text
				}
				vectors, err := client.EmbedDocuments(ctx, documents)
				select {
				case results <- embeddedVectorBatch{batch: batch, vectors: vectors, err: err}:
				case <-ctx.Done():
					return
				}
				if err != nil {
					return
				}
			}
		}()
	}
	go func() {
		defer close(jobs)
		for start := 0; start < len(missing); start += batchSize {
			end := min(start+batchSize, len(missing))
			select {
			case jobs <- missing[start:end]:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		workers.Wait()
		close(results)
	}()

	var firstErr error
	for result := range results {
		if result.err != nil {
			if firstErr == nil {
				firstErr = result.err
				cancel()
			}
			continue
		}
		if firstErr != nil {
			continue
		}
		if err := ingest(result.batch, result.vectors); err != nil {
			firstErr = err
			cancel()
		}
	}
	return firstErr
}

func writeVectorRecords(path string, entries []vectorChunk) error {
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
		}{entry.Chunk.ID, entry.Source}); err != nil {
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
