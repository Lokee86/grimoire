package interstack

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Lokee86/lexicon/internal/objectstore"
)

func Build(
	sourceRoot string,
	store objectstore.Store,
	manifest objectstore.Manifest,
	outputPath string,
) (*objectstore.Analysis, Summary, error) {
	libraries := make([]Library, 0, len(manifest.Languages))
	for _, entry := range manifest.Languages {
		if entry.Language == Language {
			continue
		}
		data, err := store.ExportLanguage(entry)
		if err != nil {
			return nil, Summary{}, fmt.Errorf("load %s facts for interstack analysis: %w", entry.Language, err)
		}
		library, err := ParseLibrary(data)
		if err != nil {
			return nil, Summary{}, err
		}
		libraries = append(libraries, library)
	}
	result, err := Resolve(sourceRoot, libraries)
	if err != nil {
		return nil, Summary{}, err
	}
	data, err := Encode(result)
	if err != nil {
		return nil, Summary{}, err
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return nil, Summary{}, err
	}
	if err := os.WriteFile(outputPath, data, 0o644); err != nil {
		return nil, Summary{}, err
	}
	analysis, err := objectstore.ReadAnalysis(outputPath, Language)
	if err != nil {
		return nil, Summary{}, err
	}
	return analysis, result.Summary, nil
}
