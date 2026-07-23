package app

import (
	"bufio"
	"encoding/json"
	"os"
)

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
