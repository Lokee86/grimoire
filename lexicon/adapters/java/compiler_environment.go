package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func javaExecutable(environment, name string) (string, error) {
	if configured := strings.TrimSpace(os.Getenv(environment)); configured != "" {
		if regularFile(configured) {
			return configured, nil
		}
		return "", fmt.Errorf("%s does not exist: %s", environment, configured)
	}
	for _, home := range []string{
		strings.TrimSpace(os.Getenv("LEXICON_JDK_HOME")),
		strings.TrimSpace(os.Getenv("JAVA_HOME")),
	} {
		if home == "" {
			continue
		}
		candidate := filepath.Join(home, "bin", executableName(name))
		if regularFile(candidate) {
			return candidate, nil
		}
	}
	for _, root := range compilerSearchRoots() {
		for current := root; current != filepath.Dir(current); current = filepath.Dir(current) {
			candidate := filepath.Join(current, ".tools", "jdk", "bin", executableName(name))
			if regularFile(candidate) {
				return candidate, nil
			}
		}
	}
	found, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("%s is required for compiler-backed Java analysis", name)
	}
	return found, nil
}

func compilerHeap() string {
	value := strings.ToLower(strings.TrimSpace(os.Getenv("LEXICON_JAVA_HEAP")))
	if value == "" {
		return "4g"
	}
	digits := 0
	for index, character := range value {
		if character >= '0' && character <= '9' {
			digits++
			continue
		}
		if index == len(value)-1 && digits > 0 && (character == 'k' || character == 'm' || character == 'g') {
			continue
		}
		return "4g"
	}
	if digits == 0 {
		return "4g"
	}
	return value
}

func compilerSearchRoots() []string {
	var roots []string
	if working, err := os.Getwd(); err == nil {
		roots = append(roots, working)
	}
	if executable, err := os.Executable(); err == nil {
		roots = append(roots, filepath.Dir(executable))
	}
	return roots
}

func executableName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func regularFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

func boundedText(value string, limit int) string {
	value = strings.TrimSpace(value)
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "..."
}
