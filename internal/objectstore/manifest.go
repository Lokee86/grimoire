package objectstore

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func (s Store) BuildManifest(stateRoot, stateCommit, analysisConfigID string) (Manifest, error) {
	libraryRoot := filepath.Join(stateRoot, "library")
	entries, err := os.ReadDir(libraryRoot)
	if os.IsNotExist(err) {
		entries = nil
	} else if err != nil {
		return Manifest{}, err
	}
	languages := make([]LanguageEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".jsonl" {
			continue
		}
		language := strings.TrimSuffix(entry.Name(), ".jsonl")
		languageEntry, err := s.IngestLanguage(
			filepath.Join(libraryRoot, entry.Name()),
			filepath.Join(stateRoot, "source"),
			language,
			analysisConfigID,
		)
		if err != nil {
			return Manifest{}, fmt.Errorf("ingest %s library: %w", language, err)
		}
		languages = append(languages, languageEntry)
	}
	sort.Slice(languages, func(left, right int) bool {
		return languages[left].Language < languages[right].Language
	})
	return Manifest{Version: SnapshotVersion, StateCommit: stateCommit, Languages: languages}, nil
}
