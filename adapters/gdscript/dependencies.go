package main

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var godotResourcePath = regexp.MustCompile(`res://([A-Za-z0-9_./-]+)`)

func dependencyAttributes(category, source string, path bool) map[string]any {
	return map[string]any{"build": false, "category": category, "constraint": "", "dev": false, "optional": false, "path": path, "peer": false, "source": source}
}

func addGodotDependencyTarget(facts *factSet, targetPath string) string {
	targetPath = normalizeSourcePath(targetPath)
	if id := facts.moduleByPath[targetPath]; id != "" {
		return id
	}
	id := nodeID("module", "dependency:gdscript:"+targetPath)
	facts.addNode(node("module", filepath.Base(filepath.FromSlash(targetPath)), ".lexicon/dependencies/godot/"+targetPath, "dependency:gdscript:"+targetPath, id, nil, "", map[string]any{"dependency": true, "ecosystem": "godot"}))
	return id
}

func processProjectDependencies(root string, projectRoots []string, repositoryName string, facts *factSet) error {
	repositoryID := nodeID("repository", repositoryName)
	for _, projectRoot := range projectRoots {
		filename := filepath.Join(root, filepath.FromSlash(projectRoot), "project.godot")
		file, err := os.Open(filename)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		section := ""
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(strings.SplitN(scanner.Text(), ";", 2)[0])
			if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
				section = strings.Trim(line, "[]")
				continue
			}
			if line == "" {
				continue
			}
			category := "local"
			source := "project.godot:" + section
			if section == "editor_plugins" {
				category = "plugin"
			}
			if section == "autoload" {
				category = "autoload"
			}
			for _, match := range godotResourcePath.FindAllStringSubmatch(line, -1) {
				resource := projectResourcePath(projectRoot, match[1])
				target := addGodotDependencyTarget(facts, resource)
				facts.addEdge(edgeWithAttributes(repositoryID, target, "depends-on", nil, dependencyAttributes(category, source, true)))
			}
		}
		if err := scanner.Err(); err != nil {
			file.Close()
			return err
		}
		file.Close()
	}
	return nil
}
