package index

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func Load(path string) (Snapshot, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Snapshot{}, err
	}

	var snapshot Snapshot
	if err := json.Unmarshal(content, &snapshot); err != nil {
		return Snapshot{}, fmt.Errorf("decode index: %w", err)
	}
	if snapshot.Version != FormatVersion {
		return Snapshot{}, fmt.Errorf("unsupported index version %d", snapshot.Version)
	}
	return snapshot, nil
}

func Save(path string, snapshot Snapshot) error {
	if snapshot.Version != FormatVersion {
		return fmt.Errorf("cannot save index version %d", snapshot.Version)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create index directory: %w", err)
	}

	temporary := path + ".tmp"
	file, err := os.Create(temporary)
	if err != nil {
		return fmt.Errorf("create temporary index: %w", err)
	}
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(snapshot); err != nil {
		file.Close()
		os.Remove(temporary)
		return fmt.Errorf("encode index: %w", err)
	}
	if err := file.Close(); err != nil {
		os.Remove(temporary)
		return fmt.Errorf("close temporary index: %w", err)
	}
	if err := os.Rename(temporary, path); err != nil {
		os.Remove(temporary)
		return fmt.Errorf("replace index: %w", err)
	}
	return nil
}
