package scan

import (
	"os"
	"path/filepath"
	"sort"

	lexfiles "github.com/Lokee86/lexicon/internal/files"
	"github.com/Lokee86/lexicon/internal/objectstore"
	"github.com/Lokee86/lexicon/internal/state"
)

func languagesInTree(root string) ([]string, error) {
	set := make(map[string]struct{})
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if os.IsNotExist(walkErr) {
			return nil
		}
		if walkErr != nil || entry.IsDir() {
			return walkErr
		}
		for _, language := range lexfiles.Languages(path) {
			set[language] = struct{}{}
		}
		return nil
	})
	return sortedSet(set), err
}

func selectedLanguages(languages []string, enabled func(string) bool) []string {
	selected := make([]string, 0, len(languages))
	for _, language := range languages {
		if enabled(language) {
			selected = append(selected, language)
		}
	}
	return selected
}

func hasLanguage(root, language string) (bool, error) {
	found := false
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		for _, candidate := range lexfiles.Languages(path) {
			if candidate == language {
				found = true
				return filepath.SkipAll
			}
		}
		return nil
	})
	return found, err
}

func snapshotDriftLanguages(
	stateRoot string,
	manifest objectstore.Manifest,
	enabled func(string) bool,
) ([]string, error) {
	required, err := languagesInTree(filepath.Join(stateRoot, "source"))
	if err != nil {
		return nil, err
	}
	requiredSet := make(map[string]struct{}, len(required))
	for _, language := range required {
		if enabled(language) {
			requiredSet[language] = struct{}{}
		}
	}
	present := make(map[string]struct{}, len(manifest.Languages))
	dirty := make(map[string]struct{})
	for _, entry := range manifest.Languages {
		present[entry.Language] = struct{}{}
		if _, ok := requiredSet[entry.Language]; !ok {
			dirty[entry.Language] = struct{}{}
		}
	}
	for language := range requiredSet {
		if _, ok := present[language]; !ok {
			dirty[language] = struct{}{}
		}
	}
	return sortedSet(dirty), nil
}

func languagesForChanges(changes []state.Change) []string {
	return lexfiles.CollectLanguages(changedPaths(changes))
}

func mergeLanguages(groups ...[]string) []string {
	set := make(map[string]struct{})
	for _, group := range groups {
		for _, language := range group {
			set[language] = struct{}{}
		}
	}
	return sortedSet(set)
}

func changedPaths(changes []state.Change) []string {
	paths := make([]string, 0, len(changes)*2)
	for _, change := range changes {
		if change.Old != "" {
			paths = append(paths, change.Old)
		}
		if change.New != "" {
			paths = append(paths, change.New)
		}
	}
	return paths
}

func sortedSet(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
