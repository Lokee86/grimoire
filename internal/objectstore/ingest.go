package objectstore

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	lexfiles "github.com/Lokee86/lexicon/internal/files"
)

func (s Store) IngestLanguage(outputPath, sourceRoot, language, analysisConfigID string) (LanguageEntry, error) {
	header, records, err := parseOutput(outputPath)
	if err != nil {
		return LanguageEntry{}, err
	}
	if err := validateFullHeader(header, language, outputPath); err != nil {
		return LanguageEntry{}, err
	}
	files, err := sourceFiles(sourceRoot, language)
	if err != nil {
		return LanguageEntry{}, err
	}
	owners := nodeOwners(records)
	groups := make(map[string][]json.RawMessage)
	shared := make([]json.RawMessage, 0)
	for _, record := range records {
		owner := recordOwner(record.value, owners)
		if _, ok := files[owner]; owner != "" && ok {
			groups[owner] = append(groups[owner], record.raw)
		} else {
			shared = append(shared, record.raw)
		}
	}
	entry := LanguageEntry{
		Language:         language,
		AdapterVersion:   header.AdapterVersion,
		SchemaVersion:    header.SchemaVersion,
		Repository:       header.Repository,
		AnalysisConfigID: analysisConfigID,
	}
	paths := sortedMapKeys(files)
	entry.Files = make([]FileEntry, 0, len(paths))
	for _, path := range paths {
		contentID := ContentID(files[path])
		objectID, err := s.WriteObject(FactObject{
			Language: language, Owner: path, SourceContentID: contentID,
			AdapterVersion: header.AdapterVersion, SchemaVersion: header.SchemaVersion,
			AnalysisConfigID: analysisConfigID, Records: nonNil(groups[path]),
		})
		if err != nil {
			return LanguageEntry{}, err
		}
		entry.Files = append(entry.Files, FileEntry{
			Path: path, Language: language, ContentID: contentID, ObjectID: objectID,
		})
	}
	if len(shared) > 0 {
		entry.SharedObjectID, err = s.WriteObject(FactObject{
			Language: language, AdapterVersion: header.AdapterVersion,
			SchemaVersion: header.SchemaVersion, AnalysisConfigID: analysisConfigID,
			Records: shared,
		})
		if err != nil {
			return LanguageEntry{}, err
		}
	}
	return entry, nil
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
