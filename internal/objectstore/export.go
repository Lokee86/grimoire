package objectstore

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Export writes complete standalone JSONL libraries for the selected languages
// from snapshot, or from CURRENT when snapshot is empty or CURRENT.
func (s Store) Export(snapshot, destination string, languages []string) error {
	manifest, err := s.resolveExportSnapshot(snapshot)
	if err != nil {
		return err
	}
	selected, err := selectExportLanguages(manifest, languages)
	if err != nil {
		return err
	}

	type exportOutput struct {
		language string
		data     []byte
	}
	outputs := make([]exportOutput, 0, len(selected))
	for _, entry := range selected {
		data, err := s.exportLanguage(entry)
		if err != nil {
			return fmt.Errorf("export %s library: %w", entry.Language, err)
		}
		outputs = append(outputs, exportOutput{language: entry.Language, data: data})
	}
	for _, output := range outputs {
		if err := writeAtomic(filepath.Join(destination, output.language+".jsonl"), output.data); err != nil {
			return fmt.Errorf("publish %s library: %w", output.language, err)
		}
	}
	return nil
}

func (s Store) resolveExportSnapshot(snapshot string) (Manifest, error) {
	if snapshot == "" || strings.EqualFold(snapshot, "CURRENT") {
		_, manifest, err := s.Current()
		return manifest, err
	}
	return s.Load(snapshot)
}
