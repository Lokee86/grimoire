package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

var excludedDirectories = map[string]struct{}{
	".arcana": {}, ".bundle": {}, ".cache": {}, ".cantrip": {}, ".ddocs": {},
	".git": {}, ".gradle": {}, ".grimoire": {}, ".homunculus": {}, ".idea": {},
	".incubus": {}, ".lexicon": {}, ".mvn": {}, ".next": {}, ".pitlord": {},
	".ritual": {}, ".warlock": {}, ".workingtrees": {}, ".worktrees": {},
	"bin": {}, "build": {}, "coverage": {}, "dist": {}, "generated": {},
	"generated-sources": {}, "node_modules": {}, "obj": {}, "out": {}, "packages": {},
	"target": {}, "temp": {}, "tmp": {}, "vendor": {},
}

type repositorySource struct {
	absolute string
	content  []byte
	path     string
	valid    bool
}

type repositoryManifest struct {
	content []byte
	kind    string
	path    string
	valid   bool
}

type repositorySnapshot struct {
	directories []string
	manifests   []repositoryManifest
	name        string
	root        string
	sources     []repositorySource
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
		return repositorySnapshot{}, fmt.Errorf("repository path is not a directory")
	}
	name := filepath.Base(filepath.Clean(root))
	if name == "" || name == "." || name == string(filepath.Separator) {
		return repositorySnapshot{}, fmt.Errorf("repository has no directory name")
	}

	snapshot := repositorySnapshot{name: name, root: root}
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
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
			relative, err := normalizedRelative(root, path)
			if err != nil {
				return err
			}
			snapshot.directories = append(snapshot.directories, relative)
			return nil
		}
		manifestKind := manifestKind(entry.Name())
		if strings.ToLower(filepath.Ext(entry.Name())) != ".java" && manifestKind == "" {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		relative, err := normalizedRelative(root, path)
		if err != nil {
			return err
		}
		valid := utf8.Valid(content) && !bytes.ContainsRune(content, '\x00')
		if manifestKind != "" {
			snapshot.manifests = append(snapshot.manifests, repositoryManifest{
				content: content, kind: manifestKind, path: relative, valid: valid,
			})
			return nil
		}
		snapshot.sources = append(snapshot.sources, repositorySource{
			absolute: path,
			content:  content,
			path:     relative,
			valid:    valid,
		})
		return nil
	})
	if err != nil {
		return repositorySnapshot{}, fmt.Errorf("scan repository: %w", err)
	}
	sort.Strings(snapshot.directories)
	sort.Slice(snapshot.sources, func(left, right int) bool {
		return snapshot.sources[left].path < snapshot.sources[right].path
	})
	sort.Slice(snapshot.manifests, func(left, right int) bool {
		return snapshot.manifests[left].path < snapshot.manifests[right].path
	})
	return snapshot, nil
}

func manifestKind(name string) string {
	switch name {
	case "pom.xml":
		return "maven"
	case "build.gradle":
		return "gradle-groovy"
	case "build.gradle.kts":
		return "gradle-kotlin"
	default:
		return ""
	}
}

func normalizedRelative(root, absolute string) (string, error) {
	relative, err := filepath.Rel(root, absolute)
	if err != nil {
		return "", err
	}
	relative = filepath.ToSlash(relative)
	if relative == "" || relative == "." || relative == ".." || strings.HasPrefix(relative, "../") || strings.HasPrefix(relative, "/") {
		return "", fmt.Errorf("path escapes repository: %s", absolute)
	}
	return relative, nil
}

func parentPath(path string) string {
	parent := filepath.ToSlash(filepath.Dir(filepath.FromSlash(path)))
	if parent == "." {
		return ""
	}
	return parent
}
