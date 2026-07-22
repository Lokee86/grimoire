package objectstore

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	lexfiles "github.com/Lokee86/lexicon/internal/files"
)

type rawRecord struct {
	Record string `json:"record"`
	Owner  string `json:"owner"`
	Path   string `json:"path"`
	Kind   string `json:"kind"`
	Source string `json:"source"`
	Span   *struct {
		Path string `json:"path"`
	} `json:"span"`
}

type parsedRecord struct {
	raw   json.RawMessage
	value rawRecord
}

func (s Store) IngestLanguage(outputPath, sourceRoot, language, analysisConfigID string) (LanguageEntry, error) {
	header, records, err := parseOutput(outputPath)
	if err != nil {
		return LanguageEntry{}, err
	}
	if err := validateHeader(header, language, outputPath); err != nil {
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
			Language:         language,
			Owner:            path,
			SourceContentID:  contentID,
			AdapterVersion:   header.AdapterVersion,
			SchemaVersion:    header.SchemaVersion,
			AnalysisConfigID: analysisConfigID,
			Records:          nonNil(groups[path]),
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
			Language:         language,
			AdapterVersion:   header.AdapterVersion,
			SchemaVersion:    header.SchemaVersion,
			AnalysisConfigID: analysisConfigID,
			Records:          shared,
		})
		if err != nil {
			return LanguageEntry{}, err
		}
	}
	return entry, nil
}

func parseOutput(path string) (Header, []parsedRecord, error) {
	file, err := os.Open(path)
	if err != nil {
		return Header{}, nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 32*1024*1024)
	if !scanner.Scan() {
		return Header{}, nil, fmt.Errorf("adapter output is empty: %s", path)
	}
	var header Header
	if err := json.Unmarshal(scanner.Bytes(), &header); err != nil {
		return Header{}, nil, fmt.Errorf("decode adapter header: %w", err)
	}
	if err := validateHeader(header, header.Language, path); err != nil {
		return Header{}, nil, err
	}
	records := make([]parsedRecord, 0)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var compact bytes.Buffer
		if err := json.Compact(&compact, line); err != nil {
			return Header{}, nil, fmt.Errorf("decode adapter record: %w", err)
		}
		raw := append(json.RawMessage(nil), compact.Bytes()...)
		var value rawRecord
		if err := json.Unmarshal(raw, &value); err != nil {
			return Header{}, nil, fmt.Errorf("decode adapter record: %w", err)
		}
		switch value.Record {
		case "node", "edge", "unresolved":
		default:
			return Header{}, nil, fmt.Errorf("unsupported Lexicon record %q in %s", value.Record, path)
		}
		records = append(records, parsedRecord{raw: raw, value: value})
	}
	if err := scanner.Err(); err != nil {
		return Header{}, nil, err
	}
	return header, records, nil
}

func ValidateOutput(path, language string) error {
	header, _, err := parseOutput(path)
	if err != nil {
		return err
	}
	return validateHeader(header, language, path)
}

func validateHeader(header Header, language, path string) error {
	if header.Record != "lexicon" || header.SchemaVersion != 1 || header.AdapterVersion == "" || header.Language == "" || header.Repository == "" {
		return fmt.Errorf("invalid Lexicon adapter header in %s", path)
	}
	if header.Language != language {
		return fmt.Errorf("adapter output language %q does not match %q", header.Language, language)
	}
	if header.Mode != "" && header.Mode != "full" {
		return fmt.Errorf("application requires full adapter output, got mode %q", header.Mode)
	}
	return nil
}

func nodeOwners(records []parsedRecord) map[string]string {
	owners := make(map[string]string)
	for _, record := range records {
		if record.value.Record != "node" {
			continue
		}
		var identity struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(record.raw, &identity) != nil || identity.ID == "" {
			continue
		}
		if owner := directOwner(record.value); owner != "" {
			owners[identity.ID] = owner
		}
	}
	return owners
}

func recordOwner(record rawRecord, owners map[string]string) string {
	if owner := directOwner(record); owner != "" {
		return owner
	}
	return normalizeOwner(owners[record.Source])
}

func directOwner(record rawRecord) string {
	if record.Owner != "" {
		return normalizeOwner(record.Owner)
	}
	if record.Span != nil && record.Span.Path != "" {
		return normalizeOwner(record.Span.Path)
	}
	if record.Record == "node" && record.Kind == "file" {
		return normalizeOwner(record.Path)
	}
	return ""
}

func normalizeOwner(path string) string {
	path = filepath.ToSlash(path)
	if path == "" || filepath.IsAbs(filepath.FromSlash(path)) {
		return ""
	}
	for _, part := range strings.Split(path, "/") {
		if part == ".." {
			return ""
		}
	}
	path = filepath.ToSlash(filepath.Clean(filepath.FromSlash(path)))
	if path == "." || path == "" {
		return ""
	}
	return path
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
