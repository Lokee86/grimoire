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

func analyzeRepository(repo string, selections ...[]string) ([]byte, error) {
	var changedFiles, removedFiles []string
	if len(selections) > 0 {
		changedFiles = selections[0]
	}
	if len(selections) > 1 {
		removedFiles = selections[1]
	}
	root, err := filepath.Abs(repo)
	if err != nil {
		return nil, fmt.Errorf("resolve repository: %w", err)
	}
	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("stat repository: %w", err)
	}
	if !info.IsDir() {
		return nil, errors.New("repository path is not a directory")
	}

	files, dirs, projectRoots, err := collectSources(root)
	if err != nil {
		return nil, err
	}
	repositoryName := filepath.Base(filepath.Clean(root))
	facts := &factSet{
		nodeByID:                   make(map[string]map[string]any),
		edgeKeys:                   make(map[string]struct{}),
		unresolvedKeys:             make(map[string]struct{}),
		moduleByPath:               make(map[string]string),
		classByName:                make(map[string][]string),
		methodByClassName:          make(map[string]map[string][]string),
		fileByPath:                 make(map[string]string),
		projectRootByFilePath:      make(map[string]string),
		autoloadOwnerByProjectName: make(map[string]map[string]string),
	}
	addRepositoryFacts(facts, repositoryName, dirs)

	parsed := make([]*parsedFile, 0, len(files))
	for _, path := range files {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		pf, err := parseFile(path, content)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		pf.projectRoot = nearestProjectRoot(path, projectRoots)
		facts.projectRootByFilePath[normalizeSourcePath(path)] = pf.projectRoot
		for index := range pf.declarations {
			if pf.declarations[index].preloadPath != "" {
				pf.declarations[index].preloadPath = projectResourcePath(pf.projectRoot, pf.declarations[index].preloadPath)
			}
		}
		addFileFacts(facts, pf, content, dirs)
		parsed = append(parsed, pf)
	}

	for _, pf := range parsed {
		processDeclarations(facts, pf)
	}
	for _, pf := range parsed {
		processExtends(facts, pf)
	}
	if err := processProjectAutoloads(root, projectRoots, facts); err != nil {
		return nil, err
	}
	model := buildSemanticModel(facts, parsed)
	for _, pf := range parsed {
		processImports(facts, pf)
		processCalls(facts, model, pf)
	}
	return facts.render(repositoryName, changedFiles, removedFiles), nil
}

func addRepositoryFacts(facts *factSet, repositoryName string, dirs []string) {
	repositoryID := nodeID("repository", repositoryName)
	rootDirectoryID := nodeID("directory", ".")
	facts.addNode(node("repository", repositoryName, ".", repositoryName, repositoryID, nil, ""))
	facts.addNode(node("directory", repositoryName, ".", ".", rootDirectoryID, nil, ""))
	facts.addEdge(edge(repositoryID, rootDirectoryID, "contains", nil))

	dirIDs := map[string]string{".": rootDirectoryID}
	for _, dir := range dirs {
		if dir == "." {
			continue
		}
		id := nodeID("directory", dir)
		dirIDs[dir] = id
		facts.addNode(node("directory", filepath.Base(filepath.FromSlash(dir)), dir, dir, id, nil, ""))
	}
	for _, dir := range dirs {
		if dir == "." {
			continue
		}
		parent := filepath.ToSlash(filepath.Dir(filepath.FromSlash(dir)))
		if parent == "" {
			parent = "."
		}
		facts.addEdge(edge(dirIDs[parent], dirIDs[dir], "contains", nil))
	}
}

func addFileFacts(facts *factSet, pf *parsedFile, content []byte, dirs []string) {
	fileID := nodeID("file", pf.path)
	moduleID := nodeID("module", pf.path)
	pf.moduleID = moduleID
	facts.fileByPath[pf.path] = fileID
	facts.moduleByPath[pf.path] = moduleID
	facts.addNode(node("file", filepath.Base(filepath.FromSlash(pf.path)), pf.path, pf.path, fileID, nil, contentID(content)))
	facts.addNode(node("module", strings.TrimSuffix(filepath.Base(filepath.FromSlash(pf.path)), filepath.Ext(pf.path)), pf.path, pf.path, moduleID, nil, ""))
	dirIDs := make(map[string]string, len(dirs))
	for _, dir := range dirs {
		dirIDs[dir] = nodeID("directory", dir)
	}
	dir := filepath.ToSlash(filepath.Dir(filepath.FromSlash(pf.path)))
	facts.addEdge(edge(dirIDs[dir], fileID, "contains", nil))
	facts.addEdge(edge(fileID, moduleID, "contains", nil))
}

func collectSources(root string) ([]string, []string, []string, error) {
	var files []string
	var projectRoots []string
	dirs := map[string]bool{".": true}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		name := strings.ToLower(entry.Name())
		if entry.IsDir() {
			if excludedDirectory(name) {
				return filepath.SkipDir
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			dirs[filepath.ToSlash(rel)] = true
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if strings.EqualFold(entry.Name(), "project.godot") {
			projectRoot := filepath.ToSlash(filepath.Dir(filepath.FromSlash(rel)))
			if projectRoot == "" {
				projectRoot = "."
			}
			projectRoots = append(projectRoots, projectRoot)
		}
		if strings.EqualFold(filepath.Ext(entry.Name()), ".gd") {
			files = append(files, rel)
		}
		return nil
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("scan repository: %w", err)
	}
	if len(projectRoots) == 0 {
		projectRoots = append(projectRoots, ".")
	}
	sort.Strings(files)
	projectRoots = uniqueSorted(projectRoots)
	needed := map[string]bool{".": true}
	for _, file := range files {
		dir := filepath.ToSlash(filepath.Dir(filepath.FromSlash(file)))
		for dir != "." && dir != "" {
			needed[dir] = true
			dir = filepath.ToSlash(filepath.Dir(filepath.FromSlash(dir)))
		}
	}
	filteredDirs := make([]string, 0, len(needed))
	for dir := range needed {
		if dirs[dir] {
			filteredDirs = append(filteredDirs, dir)
		}
	}
	sort.Slice(filteredDirs, func(i, j int) bool {
		if filteredDirs[i] == "." {
			return true
		}
		if filteredDirs[j] == "." {
			return false
		}
		return filteredDirs[i] < filteredDirs[j]
	})
	return files, filteredDirs, projectRoots, nil
}

func nearestProjectRoot(sourcePath string, projectRoots []string) string {
	sourcePath = normalizeSourcePath(sourcePath)
	best := "."
	for _, projectRoot := range projectRoots {
		projectRoot = normalizeSourcePath(projectRoot)
		if projectRoot == "." || sourcePath == projectRoot || strings.HasPrefix(sourcePath, projectRoot+"/") {
			if projectRoot != "." && (best == "." || len(projectRoot) > len(best)) {
				best = projectRoot
			}
		}
	}
	return best
}

func projectResourcePath(projectRoot, resourcePath string) string {
	resourcePath = normalizeSourcePath(resourcePath)
	projectRoot = normalizeSourcePath(projectRoot)
	if projectRoot == "." || projectRoot == "" {
		return resourcePath
	}
	return normalizeSourcePath(projectRoot + "/" + resourcePath)
}

func excludedDirectory(name string) bool {
	switch name {
	case ".git", ".worktrees", ".workingtrees", ".ddocs", ".lexicon", ".arcana", ".grimoire", ".pitlord", ".cantrip", ".homunculus", ".incubus", ".ritual", ".warlock", "node_modules", "target", "__pycache__", ".pytest_cache", ".bundle", "vendor", ".godot", ".import", "build", "dist", "bin", "obj":
		return true
	default:
		return false
	}
}
