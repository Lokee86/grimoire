package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type goDependency struct {
	name, constraint, source, replacement string
	category                              string
	path                                  bool
}

func (s *scanner) addDependencyFacts() error {
	repository := s.repositoryKey()
	for _, module := range s.modules {
		manifest := filepath.Join(module.Root, "go.mod")
		dependencies, err := parseGoDependencies(manifest)
		if err != nil {
			return err
		}
		for _, dependency := range dependencies {
			targetName := dependency.name
			local := ""
			if dependency.replacement != "" && (strings.HasPrefix(dependency.replacement, "./") || strings.HasPrefix(dependency.replacement, "../")) {
				candidateRoot := filepath.Clean(filepath.Join(module.Root, filepath.FromSlash(dependency.replacement)))
				relative, relErr := filepath.Rel(s.root, candidateRoot)
				if relErr == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
					if relative == "." {
						local = ".lexicon-repository"
					} else if normalized, normalizeErr := normalizePath(relative); normalizeErr == nil {
						local = normalized
					}
					if replacementModule, readErr := readGoModule(candidateRoot); readErr == nil {
						targetName = replacementModule
					}
					dependency.path = true
				}
			}
			target := s.dependencyNode("go", targetName, dependency.path, local)
			attributes := dependencyAttributes(dependency.category, dependency.source, dependency.constraint, dependency.path)
			s.addEdge(repository, target, RelDependsOn, nil, attributes)
		}
	}
	for relative, imports := range s.fileImports {
		source, ok := s.packageByFile[relative]
		if !ok {
			continue
		}
		for _, importPath := range imports {
			if !s.isInternalNamespace(importPath) {
				continue
			}
			target, found := s.packageNodeForNamespace(importPath)
			if found {
				s.addEdge(source, target, RelDependsOn, nil, dependencyAttributes("local", importPath, "", true))
			}
		}
	}
	return nil
}

func (s *scanner) dependencyNode(ecosystem, name string, local bool, localPath string) NodeKey {
	identity := "dependency:" + ecosystem + ":" + name
	path := ".lexicon/dependencies/" + ecosystem + "/" + strings.ReplaceAll(name, "\\", "/")
	if local {
		path = localPath
	}
	key := hashIdentity("package:" + identity)
	kind := KindPackage
	s.addNode(NodeFact{Key: key, Kind: kind, Path: path, Name: name, Attributes: map[string]any{"dependency": true, "ecosystem": ecosystem}})
	return key
}

func dependencyAttributes(category, source, constraint string, local bool) map[string]any {
	return map[string]any{"build": category == "build", "category": category, "dev": category == "development", "optional": category == "optional", "path": local, "peer": category == "peer", "source": source, "constraint": constraint}
}

func parseGoDependencies(filename string) ([]goDependency, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("open go.mod: %w", err)
	}
	defer file.Close()
	var result []goDependency
	section := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(strings.SplitN(scanner.Text(), "//", 2)[0])
		if line == "" {
			continue
		}
		if strings.HasSuffix(line, "(") {
			section = strings.TrimSpace(strings.TrimSuffix(line, "("))
			continue
		}
		if line == ")" {
			section = ""
			continue
		}
		if section == "require" || strings.HasPrefix(line, "require ") {
			fields := strings.Fields(strings.TrimPrefix(line, "require "))
			if len(fields) >= 2 {
				result = append(result, goDependency{name: fields[0], constraint: fields[1], source: "go.mod:require", category: "runtime"})
			}
			continue
		}
		if section == "replace" || strings.HasPrefix(line, "replace ") {
			text := strings.TrimSpace(strings.TrimPrefix(line, "replace "))
			parts := strings.SplitN(text, "=>", 2)
			if len(parts) != 2 {
				continue
			}
			left, right := strings.Fields(strings.TrimSpace(parts[0])), strings.Fields(strings.TrimSpace(parts[1]))
			if len(left) == 0 || len(right) == 0 {
				continue
			}
			constraint := ""
			if len(left) > 1 {
				constraint = left[1]
			}
			replacement := right[0]
			if len(right) > 1 && !strings.HasPrefix(replacement, ".") {
				replacement += "@" + right[1]
			}
			result = append(result, goDependency{name: left[0], constraint: constraint, replacement: replacement, source: "go.mod:replace", category: "runtime"})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read go.mod: %w", err)
	}
	return result, nil
}

func readGoModule(root string) (string, error) {
	file, err := os.Open(filepath.Join(root, "go.mod"))
	if err != nil {
		return "", err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(strings.TrimSpace(scanner.Text()))
		if len(fields) >= 2 && fields[0] == "module" {
			return fields[1], nil
		}
	}
	return "", fmt.Errorf("no module directive")
}
