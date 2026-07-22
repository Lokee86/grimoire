package app

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Lokee86/grimoire/internal/embedding"
	"github.com/Lokee86/grimoire/internal/vectorstore"
)

func ingestVectorBatch(
	library *vectorstore.Library,
	paths vectorStatePaths,
	batch []vectorChunk,
	vectors [][]float32,
) error {
	if len(vectors) != len(batch) {
		return fmt.Errorf("embedding provider returned %d vectors for %d chunks", len(vectors), len(batch))
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
		}{batch[index].Source, vector}); err != nil {
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
	if count != uint64(len(batch)) {
		return fmt.Errorf("vector engine ingested %d records, expected %d", count, len(batch))
	}
	return nil
}
