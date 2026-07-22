package objectstore

import (
	"fmt"
	"sort"
)

func (s Store) exportLanguage(entry LanguageEntry) ([]byte, error) {
	if entry.AdapterVersion == "" || entry.SchemaVersion != 1 || entry.Repository == "" || entry.AnalysisConfigID == "" {
		return nil, fmt.Errorf("invalid metadata")
	}

	records := make([]exportRecord, 0)
	if entry.SharedObjectID != "" {
		object, err := s.LoadObject(entry.SharedObjectID)
		if err != nil {
			return nil, fmt.Errorf("load shared object %s: %w", entry.SharedObjectID, err)
		}
		if err := validateExportObject(object, entry, "shared", "", ""); err != nil {
			return nil, err
		}
		loaded, err := exportRecords(object.Records)
		if err != nil {
			return nil, fmt.Errorf("decode shared object %s: %w", entry.SharedObjectID, err)
		}
		records = append(records, loaded...)
	}

	files := append([]FileEntry(nil), entry.Files...)
	sort.Slice(files, func(left, right int) bool {
		return files[left].Path < files[right].Path
	})
	seenPaths := make(map[string]struct{}, len(files))
	for _, file := range files {
		if file.Path == "" || !validExportPath(file.Path) {
			return nil, fmt.Errorf("invalid file path %q", file.Path)
		}
		if _, exists := seenPaths[file.Path]; exists {
			return nil, fmt.Errorf("duplicate file path %q", file.Path)
		}
		seenPaths[file.Path] = struct{}{}
		if file.Language != entry.Language || file.ContentID == "" || file.ObjectID == "" {
			return nil, fmt.Errorf("invalid metadata for file %q", file.Path)
		}
		object, err := s.LoadObject(file.ObjectID)
		if err != nil {
			return nil, fmt.Errorf("load file object %s: %w", file.ObjectID, err)
		}
		if err := validateExportObject(object, entry, "file", file.Path, file.ContentID); err != nil {
			return nil, err
		}
		loaded, err := exportRecords(object.Records)
		if err != nil {
			return nil, fmt.Errorf("decode file object %s: %w", file.ObjectID, err)
		}
		records = append(records, loaded...)
	}

	sort.Slice(records, func(left, right int) bool {
		if records[left].key == records[right].key {
			return string(records[left].raw) < string(records[right].raw)
		}
		return records[left].key < records[right].key
	})
	return encodeExport(entry, records)
}

func validateExportObject(object FactObject, entry LanguageEntry, kind, owner, contentID string) error {
	if object.Language != entry.Language || object.AdapterVersion != entry.AdapterVersion ||
		object.SchemaVersion != entry.SchemaVersion || object.AnalysisConfigID != entry.AnalysisConfigID {
		return fmt.Errorf("%s object metadata does not match %s manifest", kind, entry.Language)
	}
	if kind == "shared" {
		if object.Owner != "" || object.SourceContentID != "" {
			return fmt.Errorf("shared object has file ownership metadata")
		}
		return nil
	}
	if object.Owner != owner || object.SourceContentID != contentID {
		return fmt.Errorf("file object metadata does not match %s manifest", owner)
	}
	return nil
}
