package objectstore

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	lexfiles "github.com/Lokee86/lexicon/internal/files"
)

func (s Store) IngestLanguage(outputPath, sourceRoot, language, analysisConfigID string) (LanguageEntry, error) {
	analysis, err := ReadAnalysis(outputPath, language)
	if err != nil {
		return LanguageEntry{}, err
	}
	return s.BuildFullLanguage(analysis, sourceRoot, language, analysisConfigID, "")
}

func sourceFiles(root, language string) (map[string][]byte, error) {
	result := make(map[string][]byte)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if os.IsNotExist(walkErr) {
			return nil
		}
		if walkErr != nil || entry.IsDir() {
			return walkErr
		}
		if !contains(lexfiles.Languages(path), language) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		result[filepath.ToSlash(relative)] = data
		return nil
	})
	return result, err
}

func sortedMapKeys(values map[string][]byte) []string {
	result := make([]string, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func nonNil(records []json.RawMessage) []json.RawMessage {
	if records == nil {
		return []json.RawMessage{}
	}
	return records
}
