package objectstore

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

func selectExportLanguages(manifest Manifest, requested []string) ([]LanguageEntry, error) {
	entries := make(map[string]LanguageEntry, len(manifest.Languages))
	for _, entry := range manifest.Languages {
		if !validExportLanguage(entry.Language) {
			return nil, fmt.Errorf("invalid snapshot language %q", entry.Language)
		}
		if _, exists := entries[entry.Language]; exists {
			return nil, fmt.Errorf("snapshot contains duplicate language %q", entry.Language)
		}
		entries[entry.Language] = entry
	}

	selectedNames := requested
	if len(selectedNames) == 0 {
		selectedNames = make([]string, 0, len(entries))
		for language := range entries {
			selectedNames = append(selectedNames, language)
		}
	}
	selected := make([]LanguageEntry, 0, len(selectedNames))
	seen := make(map[string]struct{}, len(selectedNames))
	for _, language := range selectedNames {
		entry, ok := entries[language]
		if !ok {
			return nil, fmt.Errorf("snapshot has no %s library", language)
		}
		if _, exists := seen[language]; exists {
			continue
		}
		seen[language] = struct{}{}
		selected = append(selected, entry)
	}
	sort.Slice(selected, func(left, right int) bool {
		return selected[left].Language < selected[right].Language
	})
	return selected, nil
}

func validExportLanguage(language string) bool {
	return language != "" && language != "." && language != ".." &&
		!strings.ContainsAny(language, `/\\`)
}

func validExportPath(path string) bool {
	if strings.Contains(path, "\\") {
		return false
	}
	if filepath.IsAbs(filepath.FromSlash(path)) {
		return false
	}
	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(path)))
	if clean != path || clean == "." || clean == "" {
		return false
	}
	for _, part := range strings.Split(path, "/") {
		if part == ".." {
			return false
		}
	}
	return true
}
