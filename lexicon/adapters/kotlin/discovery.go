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

type sourceFile struct {
	absolute string
	content  []byte
	path     string
}

type manifestFile struct {
	content []byte
	format  string
	path    string
}

var excludedDirectories = map[string]struct{}{
	".arcana": {}, ".bundle": {}, ".cantrip": {}, ".ddocs": {}, ".git": {},
	".godot": {}, ".gradle": {}, ".grimoire": {}, ".homunculus": {}, ".idea": {},
	".import": {}, ".incubus": {}, ".kotlin": {}, ".lexicon": {}, ".pitlord": {},
	".pytest_cache": {}, ".ritual": {}, ".vs": {}, ".vscode": {}, ".warlock": {},
	".workingtrees": {}, ".worktrees": {}, "__pycache__": {}, "artifacts": {},
	"bin": {}, "build": {}, "coverage": {}, "dist": {}, "generated": {},
	"node_modules": {}, "obj": {}, "out": {}, "packages": {}, "target": {},
	"temp": {}, "tmp": {}, "vendor": {},
}

func discoverRepository(repository string) (string, string, []sourceFile, []manifestFile, error) {
	root, err := filepath.Abs(repository)
	if err != nil {
		return "", "", nil, nil, fmt.Errorf("resolve repository: %w", err)
	}
	info, err := os.Stat(root)
	if err != nil {
		return "", "", nil, nil, fmt.Errorf("stat repository: %w", err)
	}
	if !info.IsDir() {
		return "", "", nil, nil, errors.New("repository path is not a directory")
	}
	repositoryName := filepath.Base(filepath.Clean(root))
	if repositoryName == "." || repositoryName == string(filepath.Separator) || repositoryName == "" {
		return "", "", nil, nil, errors.New("repository path has no stable directory name")
	}

	var sources []sourceFile
	var manifests []manifestFile
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			if isExcludedDirectory(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if !entry.Type().IsRegular() || (!isKotlinSource(entry.Name()) && manifestFormat(entry.Name()) == "") {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relative, err = normalizeRelativePath(relative)
		if err != nil {
			return err
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", relative, err)
		}
		if format := manifestFormat(entry.Name()); format != "" {
			manifests = append(manifests, manifestFile{content: content, format: format, path: relative})
		} else {
			sources = append(sources, sourceFile{absolute: path, content: content, path: relative})
		}
		return nil
	})
	if err != nil {
		return "", "", nil, nil, fmt.Errorf("scan repository: %w", err)
	}
	sort.Slice(sources, func(i, j int) bool { return sources[i].path < sources[j].path })
	sort.Slice(manifests, func(i, j int) bool { return manifests[i].path < manifests[j].path })
	return root, repositoryName, sources, manifests, nil
}

func isExcludedDirectory(name string) bool {
	_, excluded := excludedDirectories[strings.ToLower(name)]
	return excluded
}

func isKotlinSource(name string) bool {
	if manifestFormat(name) != "" {
		return false
	}
	extension := strings.ToLower(filepath.Ext(name))
	return extension == ".kt" || extension == ".kts"
}

func manifestFormat(name string) string {
	switch strings.ToLower(name) {
	case "build.gradle.kts":
		return "gradle-kotlin"
	case "build.gradle":
		return "gradle-groovy"
	case "pom.xml":
		return "maven"
	default:
		return ""
	}
}

func normalizeRelativePath(path string) (string, error) {
	normalized := filepath.ToSlash(path)
	if normalized == "" || normalized == "." || filepath.IsAbs(path) || strings.HasPrefix(normalized, "/") {
		return "", fmt.Errorf("path is not repository-relative: %s", path)
	}
	segments := strings.Split(normalized, "/")
	for _, segment := range segments {
		if segment == "" || segment == "." || segment == ".." {
			return "", fmt.Errorf("path is not normalized: %s", path)
		}
	}
	return normalized, nil
}
