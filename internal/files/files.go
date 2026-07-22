package files

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

var ignoredDirectories = map[string]struct{}{
	".git": {}, ".lexicon": {}, ".warlock": {}, ".worktrees": {}, ".workingtrees": {}, ".astro": {},
	"node_modules": {}, "vendor": {}, "target": {}, "dist": {}, "build": {},
	".venv": {}, "venv": {}, "__pycache__": {}, ".pytest_cache": {},
}

var extensionLanguages = map[string]string{
	".go": "go", ".py": "python", ".rb": "ruby", ".gemspec": "ruby",
	".gd": "gdscript", ".rs": "rust", ".ts": "typescript", ".tsx": "typescript",
	".mts": "typescript", ".cts": "typescript", ".js": "typescript", ".jsx": "typescript",
	".mjs": "typescript", ".cjs": "typescript",
}

var namedLanguages = map[string][]string{
	"go.mod": {"go"}, "go.sum": {"go"},
	"Cargo.toml": {"rust"}, "Cargo.lock": {"rust"},
	"package.json": {"typescript"}, "package-lock.json": {"typescript"},
	"tsconfig.json": {"typescript"}, "jsconfig.json": {"typescript"},
	"pyproject.toml": {"python"}, "setup.cfg": {"python"}, "requirements.txt": {"python"},
	"Gemfile": {"ruby"}, "Gemfile.lock": {"ruby"}, "project.godot": {"gdscript"},
}

func IgnoredDirectory(name string) bool {
	_, ignored := ignoredDirectories[name]
	return ignored
}

func SkipDir(path string, entry fs.DirEntry) bool {
	return entry.IsDir() && IgnoredDirectory(entry.Name())
}

func Languages(path string) []string {
	name := filepath.Base(path)
	if languages, ok := namedLanguages[name]; ok {
		return append([]string(nil), languages...)
	}
	if language, ok := extensionLanguages[strings.ToLower(filepath.Ext(name))]; ok {
		return []string{language}
	}
	return nil
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
