package main

import (
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

func collectSources(root, extension string, selected []string, incremental bool) ([]string, error) {
	if incremental {
		paths := make([]string, 0, len(selected))
		seen := make(map[string]struct{}, len(selected))
		for _, requested := range selected {
			path, err := normalizeSelectedPath(root, requested)
			if err != nil {
				return nil, err
			}
			if strings.ToLower(filepath.Ext(path)) != extension || excludedPath(path) {
				continue
			}
			if _, ok := seen[path]; ok {
				continue
			}
			if info, err := os.Stat(filepath.Join(root, filepath.FromSlash(path))); err == nil && !info.IsDir() {
				seen[path] = struct{}{}
				paths = append(paths, path)
			} else if err != nil && !os.IsNotExist(err) {
				return nil, err
			}
		}
		sort.Strings(paths)
		return paths, nil
	}

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
		if strings.ToLower(filepath.Ext(entry.Name())) != extension {
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

func normalizeSelectedPath(root, requested string) (string, error) {
	path := filepath.Clean(filepath.FromSlash(requested))
	if filepath.IsAbs(path) {
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return "", err
		}
		path = relative
	}
	if path == ".." || strings.HasPrefix(path, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("source path escapes repository: %s", requested)
	}
	return filepath.ToSlash(path), nil
}

func excludedPath(path string) bool {
	for _, part := range strings.Split(filepath.ToSlash(path), "/") {
		if _, excluded := excludedDirectories[strings.ToLower(part)]; excluded {
			return true
		}
	}
	return false
}
