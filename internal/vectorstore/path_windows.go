//go:build windows

package vectorstore

import (
	"fmt"
	"os"
	"path/filepath"
	"unsafe"
)

func FindLibrary(explicit string) (string, error) {
	if explicit != "" {
		return requireLibrary(explicit)
	}
	if configured := os.Getenv("GRIMOIRE_VECTOR_ENGINE"); configured != "" {
		return requireLibrary(configured)
	}

	executable, _ := os.Executable()
	cwd, _ := os.Getwd()
	for _, candidate := range libraryCandidates(executable, cwd) {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return filepath.Abs(candidate)
		}
	}
	return "", fmt.Errorf("%w: set GRIMOIRE_VECTOR_ENGINE or build native/vector-engine", ErrUnavailable)
}

func libraryCandidates(executable, cwd string) []string {
	candidates := make([]string, 0, 32)
	seen := map[string]struct{}{}
	appendCandidate := func(path string) {
		if path == "" {
			return
		}
		path = filepath.Clean(path)
		if _, exists := seen[path]; exists {
			return
		}
		seen[path] = struct{}{}
		candidates = append(candidates, path)
	}
	appendDevelopmentCandidates := func(start string) {
		if start == "" {
			return
		}
		for directory := filepath.Clean(start); ; directory = filepath.Dir(directory) {
			appendCandidate(filepath.Join(directory, "native", "vector-engine", "target", "release", ABIName+".dll"))
			appendCandidate(filepath.Join(directory, "native", "vector-engine", "target", "debug", ABIName+".dll"))
			parent := filepath.Dir(directory)
			if parent == directory {
				break
			}
		}
	}

	if executable != "" {
		executableDirectory := filepath.Dir(executable)
		appendCandidate(filepath.Join(executableDirectory, ABIName+".dll"))
		appendDevelopmentCandidates(executableDirectory)
	}
	if cwd != "" {
		appendCandidate(filepath.Join(cwd, ABIName+".dll"))
		appendDevelopmentCandidates(cwd)
	}
	return candidates
}

func requireLibrary(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(absolute)
	if err != nil || info.IsDir() {
		return "", fmt.Errorf("%w: %s", ErrUnavailable, absolute)
	}
	return absolute, nil
}

func bytePointer(data []byte) uintptr {
	if len(data) == 0 {
		return 0
	}
	return uintptr(unsafe.Pointer(unsafe.SliceData(data)))
}
