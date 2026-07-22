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
	flags := flag.NewFlagSet("vector build", flag.ContinueOnError)
	flags.SetOutput(stderr)
	root := flags.String("root", ".", "repository root")
	state := flags.String("state", "", "prepared index repository path")
	endpoint := flags.String("endpoint", embedding.DefaultEndpoint, "OpenAI-compatible embeddings endpoint")
	enginePath := flags.String("engine", "", "Rust vector engine DLL")
	batchSize := flags.Int("batch-size", 16, "documents embedded per request")
	timeout := flags.Duration("timeout", 30*time.Minute, "complete vector build timeout")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *batchSize <= 0 || *timeout <= 0 {
		return errors.New("--batch-size and --timeout must be positive")
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
	defer os.Remove(paths.Manifest)

	library, err := vectorstore.Load(*enginePath)
	if err != nil {
		return err
	}
	defer library.Close()

	all := make([]vectorChunk, 0, len(chunks))
	missing := make([]vectorChunk, 0)
	for _, chunk := range chunks {
		entry := vectorChunk{Chunk: chunk, Source: vectorSource(chunk.Text)}
		all = append(all, entry)
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
		if err := embedMissing(ctx, embedding.NewClient(*endpoint), library, paths, missing, *batchSize); err != nil {
			return err
		}
	}
	if err := writeVectorManifest(paths.Manifest, all); err != nil {
		return err
	}
	identity, err := library.MaterializeJSONL(paths.Store, embedding.Identity(), paths.Manifest, paths.Snapshot)
	if err != nil {
		return err
	}
	return writeJSON(stdout, struct {
		Snapshot string `json:"snapshot"`
		Identity string `json:"identity"`
		Model    string `json:"model"`
		Chunks   int    `json:"chunks"`
		Embedded int    `json:"embedded"`
		Reused   int    `json:"reused"`
	}{paths.Snapshot, identity, embedding.Identity(), len(all), len(missing), len(all) - len(missing)})
}

func embedMissing(
	ctx context.Context,
	client *embedding.Client,
	library *vectorstore.Library,
	paths vectorStatePaths,
	missing []vectorChunk,
	batchSize int,
) error {
	file, err := os.Create(paths.Ingest)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(file)
	encoder := json.NewEncoder(writer)
	for start := 0; start < len(missing); start += batchSize {
		end := min(start+batchSize, len(missing))
		documents := make([]string, end-start)
		for index, entry := range missing[start:end] {
			documents[index] = entry.Chunk.Text
		}
		vectors, embedErr := client.EmbedDocuments(ctx, documents)
		if embedErr != nil {
			file.Close()
			return embedErr
		}
		for index, vector := range vectors {
			if err := encoder.Encode(struct {
				Source string    `json:"source"`
				Vector []float32 `json:"vector"`
			}{missing[start+index].Source, vector}); err != nil {
				file.Close()
				return err
			}
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
	if count != uint64(len(missing)) {
		return fmt.Errorf("vector engine ingested %d records, expected %d", count, len(missing))
	}
	return nil
}

func writeVectorManifest(path string, entries []vectorChunk) error {
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
