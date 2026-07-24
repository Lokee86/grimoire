package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var excludedDirectories = map[string]struct{}{
	".arcana": {}, ".bundle": {}, ".cantrip": {}, ".ddocs": {}, ".git": {}, ".godot": {},
	".grimoire": {}, ".homunculus": {}, ".import": {}, ".incubus": {}, ".lexicon": {},
	".next": {}, ".pitlord": {}, ".pytest_cache": {}, ".ritual": {}, ".venv": {},
	".warlock": {}, ".workingtrees": {}, ".worktrees": {}, "__pycache__": {}, "bin": {},
	"build": {}, "coverage": {}, "dist": {}, "node_modules": {}, "obj": {}, "target": {},
	"tmp": {}, "vendor": {}, "venv": {},
}

var sourceExtensions = map[string]struct{}{
	".c": {}, ".cc": {}, ".cp": {}, ".cpp": {}, ".cxx": {}, ".c++": {},
	".h": {}, ".hh": {}, ".hpp": {}, ".hxx": {}, ".h++": {}, ".inc": {},
	".inl": {}, ".ipp": {}, ".tpp": {},
}

var explicitCPPHeaders = map[string]struct{}{
	".hh": {}, ".hpp": {}, ".hxx": {}, ".h++": {}, ".inl": {}, ".ipp": {}, ".tpp": {},
}

func collectSources(root string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		if entry.IsDir() {
			if _, excluded := excludedDirectories[strings.ToLower(entry.Name())]; excluded {
				return filepath.SkipDir
			}
			return nil
		}
		if !isSourcePath(entry.Name()) {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		paths = append(paths, filepath.ToSlash(relative))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan repository: %w", err)
	}
	sort.Strings(paths)
	return paths, nil
}

func isSourcePath(path string) bool {
	if filepath.Ext(path) == ".C" {
		return true
	}
	_, ok := sourceExtensions[strings.ToLower(filepath.Ext(path))]
	return ok
}

func classifyLanguage(path string, content []byte, compileLanguages map[string]string) string {
	path = filepath.ToSlash(path)
	if language := compileLanguages[path]; language != "" {
		return language
	}
	extension := filepath.Ext(path)
	if extension == ".C" {
		return "cpp"
	}
	extension = strings.ToLower(extension)
	if extension == ".c" {
		return "c"
	}
	if _, ok := explicitCPPHeaders[extension]; ok {
		return "cpp"
	}
	if extension != ".h" && extension != ".inc" {
		return "cpp"
	}
	text := "\n" + string(content) + "\n"
	for _, marker := range []string{"namespace ", "class ", "template<", "template <", "constexpr ", "std::", "public:", "private:", "protected:", " override", " virtual ", "nullptr", "decltype(", "using namespace ", "::"} {
		if strings.Contains(text, marker) {
			return "cpp"
		}
	}
	return "c"
}

type compileCommand struct {
	Arguments []string `json:"arguments"`
	Command   string   `json:"command"`
	Directory string   `json:"directory"`
	File      string   `json:"file"`
}

func loadCompileLanguages(root string) map[string]string {
	path := filepath.Join(root, "compile_commands.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return map[string]string{}
	}
	var commands []compileCommand
	if json.Unmarshal(data, &commands) != nil {
		return map[string]string{}
	}
	result := map[string]string{}
	for _, command := range commands {
		file := filepath.FromSlash(command.File)
		if !filepath.IsAbs(file) {
			base := command.Directory
			if base == "" {
				base = root
			}
			file = filepath.Join(base, file)
		}
		relative, err := filepath.Rel(root, file)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			continue
		}
		result[filepath.ToSlash(relative)] = commandLanguage(command)
	}
	return result
}

func commandLanguage(command compileCommand) string {
	joined := strings.ToLower(command.Command + " " + strings.Join(command.Arguments, " "))
	for _, marker := range []string{"-x c++", "clang++", "g++", " c++", "cpp"} {
		if strings.Contains(joined, marker) {
			return "cpp"
		}
	}
	return "c"
}
