package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func processProjectAutoloads(root string, facts *factSet) error {
	path := filepath.Join(root, "project.godot")
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("open project.godot: %w", err)
	}
	defer file.Close()

	facts.autoloadOwnerByName = make(map[string]string)
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
		if owner := facts.scriptOwnerByPath[sourcePath]; owner != "" {
			facts.autoloadOwnerByName[name] = owner
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read project.godot: %w", err)
	}
	return nil
}
