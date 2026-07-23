package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type goModule struct {
	Root     string
	Relative string
	Path     string
}

func discoverModules(root string, files []string) ([]goModule, error) {
	modules := make([]goModule, 0)
	for _, file := range files {
		if filepath.Base(file) != "go.mod" {
			continue
		}
		moduleRoot := filepath.Dir(file)
		modulePath, err := readModule(moduleRoot)
		if err != nil {
			return nil, err
		}
		relative, err := filepath.Rel(root, moduleRoot)
		if err != nil {
			return nil, fmt.Errorf("resolve module root %s: %w", moduleRoot, err)
		}
		if relative == "." {
			relative = ""
		} else {
			relative, err = normalizePath(relative)
			if err != nil {
				return nil, err
			}
		}
		modules = append(modules, goModule{Root: moduleRoot, Relative: relative, Path: modulePath})
	}
	if len(modules) == 0 {
		return nil, fmt.Errorf("no go.mod files found under %s", root)
	}
	sort.Slice(modules, func(i, j int) bool {
		if modules[i].Relative != modules[j].Relative {
			return modules[i].Relative < modules[j].Relative
		}
		return modules[i].Path < modules[j].Path
	})
	return modules, nil
}

func repositoryIdentity(root string, modules []goModule) string {
	for _, module := range modules {
		if module.Relative == "" {
			return module.Path
		}
	}
	return filepath.Base(root)
}

func (s *scanner) moduleForAbsolute(path string) (goModule, bool) {
	var selected goModule
	selectedLength := -1
	for _, module := range s.modules {
		relative, err := filepath.Rel(module.Root, path)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			continue
		}
		if len(module.Root) > selectedLength {
			selected = module
			selectedLength = len(module.Root)
		}
	}
	return selected, selectedLength >= 0
}

func (s *scanner) moduleForNamespace(namespace string) (goModule, bool) {
	var selected goModule
	selectedLength := -1
	for _, module := range s.modules {
		if namespace != module.Path && !strings.HasPrefix(namespace, module.Path+"/") {
			continue
		}
		if len(module.Path) > selectedLength {
			selected = module
			selectedLength = len(module.Path)
		}
	}
	return selected, selectedLength >= 0
}

func (s *scanner) moduleImportPath(absolute string) string {
	module, ok := s.moduleForAbsolute(absolute)
	if !ok {
		relative, err := s.relative(absolute)
		if err != nil {
			return s.repository
		}
		dir := filepath.ToSlash(filepath.Dir(relative))
		if dir == "." || dir == "" {
			return s.repository
		}
		return s.repository + "/" + dir
	}
	relative, err := filepath.Rel(module.Root, absolute)
	if err != nil {
		return module.Path
	}
	dir := filepath.ToSlash(filepath.Dir(relative))
	if dir == "." || dir == "" {
		return module.Path
	}
	return module.Path + "/" + dir
}
