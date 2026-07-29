package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type repositorySnapshot struct {
	directories []string
	name        string
	root        string
	sources     []parsedFile
}

func discoverRepository(repository string) (repositorySnapshot, error) {
	root, err := filepath.Abs(repository)
	if err != nil {
		return repositorySnapshot{}, fmt.Errorf("resolve repository: %w", err)
	}
	info, err := os.Stat(root)
	if err != nil {
		return repositorySnapshot{}, fmt.Errorf("stat repository: %w", err)
	}
	if !info.IsDir() {
		return repositorySnapshot{}, errors.New("repository path is not a directory")
	}

	var sources []parsedFile
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		if entry.IsDir() {
			if excludedDirectory(strings.ToLower(entry.Name())) {
				return filepath.SkipDir
			}
			return nil
		}
		extension := strings.ToLower(filepath.Ext(entry.Name()))
		if !lotusScriptExtension(extension) {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", relative, err)
		}
		content, ok := lotusScriptContent(extension, raw)
		if !ok {
			return nil
		}
		sources = append(sources, parsedFile{
			content: content, contentHash: contentID(raw), path: filepath.ToSlash(relative),
		})
		return nil
	})
	if err != nil {
		return repositorySnapshot{}, fmt.Errorf("scan repository: %w", err)
	}
	sort.Slice(sources, func(left, right int) bool { return sources[left].path < sources[right].path })
	return repositorySnapshot{
		directories: sourceDirectories(sources),
		name:        filepath.Base(filepath.Clean(root)),
		root:        root,
		sources:     sources,
	}, nil
}

func sourceDirectories(sources []parsedFile) []string {
	set := map[string]struct{}{".": {}}
	for _, source := range sources {
		directory := filepath.ToSlash(filepath.Dir(filepath.FromSlash(source.path)))
		for directory != "." && directory != "" {
			set[directory] = struct{}{}
			directory = filepath.ToSlash(filepath.Dir(filepath.FromSlash(directory)))
		}
	}
	result := make([]string, 0, len(set))
	for directory := range set {
		result = append(result, directory)
	}
	sort.Slice(result, func(left, right int) bool {
		if result[left] == "." {
			return true
		}
		if result[right] == "." {
			return false
		}
		return result[left] < result[right]
	})
	return result
}

func excludedDirectory(name string) bool {
	switch name {
	case ".git", ".worktrees", ".workingtrees", ".ddocs", ".lexicon", ".arcana", ".grimoire", ".pitlord", ".cantrip", ".homunculus", ".incubus", ".ritual", ".warlock", "node_modules", "target", "__pycache__", ".pytest_cache", ".bundle", "vendor", "build", "dist", "bin", "obj":
		return true
	default:
		return false
	}
}
