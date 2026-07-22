package files

import (
	"io/fs"
	"sort"

	languageRegistry "github.com/Lokee86/lexicon/internal/languages"
)

var ignoredDirectories = map[string]struct{}{
	// Warlock toolchain state is generated state, never repository source.
	".ddocs": {}, ".lexicon": {}, ".arcana": {}, ".grimoire": {}, ".pitlord": {},
	".cantrip": {}, ".homunculus": {}, ".incubus": {}, ".ritual": {}, ".warlock": {},

	".git": {}, ".worktrees": {}, ".workingtrees": {}, ".astro": {},
	"node_modules": {}, "vendor": {}, "target": {}, "dist": {}, "build": {},
	".venv": {}, "venv": {}, "__pycache__": {}, ".pytest_cache": {},
}

func SupportedLanguages() []string {
	return languageRegistry.Supported()
}

func IgnoredDirectory(name string) bool {
	_, ignored := ignoredDirectories[name]
	return ignored
}

func SkipDir(path string, entry fs.DirEntry) bool {
	return entry.IsDir() && IgnoredDirectory(entry.Name())
}

func Languages(path string) []string {
	return languageRegistry.ForPath(path)
}

func Relevant(path string) bool {
	return len(Languages(path)) > 0
}

func CollectLanguages(paths []string) []string {
	set := make(map[string]struct{})
	for _, path := range paths {
		for _, language := range Languages(path) {
			set[language] = struct{}{}
		}
	}
	result := make([]string, 0, len(set))
	for language := range set {
		result = append(result, language)
	}
	sort.Strings(result)
	return result
}
