package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func writeCompilerSourceList(sources []repositorySource) (string, error) {
	file, err := os.CreateTemp("", "lexicon-java-sources-*.txt")
	if err != nil {
		return "", fmt.Errorf("create javac source list: %w", err)
	}
	path := file.Name()
	for _, source := range sources {
		if !source.valid || !compilerEligibleSource(source.absolute) {
			continue
		}
		if _, err := fmt.Fprintln(file, source.absolute); err != nil {
			_ = file.Close()
			_ = os.Remove(path)
			return "", fmt.Errorf("write javac source list: %w", err)
		}
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("close javac source list: %w", err)
	}
	return path, nil
}

func compilerEligibleSource(path string) bool {
	normalized := "/" + strings.ToLower(filepath.ToSlash(path)) + "/"
	for _, fragment := range []string{
		"/src/test/resources/",
		"/src/test/projects/",
	} {
		if strings.Contains(normalized, fragment) {
			return false
		}
	}
	return true
}
