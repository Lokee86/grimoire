package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func processProjectAutoloads(root string, projectRoots []string, facts *factSet) error {
	for _, projectRoot := range projectRoots {
		if err := processProjectAutoloadFile(root, projectRoot, facts); err != nil {
			return err
		}
	}
	return nil
}

func processProjectAutoloadFile(root, projectRoot string, facts *factSet) error {
	path := filepath.Join(root, filepath.FromSlash(projectRoot), "project.godot")
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("open %s: %w", filepath.ToSlash(filepath.Join(projectRoot, "project.godot")), err)
	}
	defer file.Close()

	section := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.TrimSpace(line[1 : len(line)-1])
			continue
		}
		if section != "autoload" || line == "" || strings.HasPrefix(line, ";") {
			continue
		}
		name, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		name = strings.TrimSpace(name)
		value = strings.TrimSpace(value)
		if name == "" || len(value) < 2 || value[0] != '"' || value[len(value)-1] != '"' {
			continue
		}
		value = strings.TrimPrefix(value[1:len(value)-1], "*")
		sourcePath, ok := normalizeImportPath(value)
		if !ok {
			continue
		}
		sourcePath = projectResourcePath(projectRoot, sourcePath)
		if owner := facts.scriptOwnerByPath[sourcePath]; owner != "" {
			if facts.autoloadOwnerByProjectName[projectRoot] == nil {
				facts.autoloadOwnerByProjectName[projectRoot] = make(map[string]string)
			}
			facts.autoloadOwnerByProjectName[projectRoot][name] = owner
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read project.godot: %w", err)
	}
	return nil
}
