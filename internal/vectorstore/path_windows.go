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

	candidates := make([]string, 0, 20)
	if executable, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(executable), ABIName+".dll"))
	}
	if cwd, err := os.Getwd(); err == nil {
		for directory := filepath.Clean(cwd); ; directory = filepath.Dir(directory) {
			candidates = append(candidates,
				filepath.Join(directory, "native", "vector-engine", "target", "release", ABIName+".dll"),
				filepath.Join(directory, "native", "vector-engine", "target", "debug", ABIName+".dll"),
			)
			parent := filepath.Dir(directory)
			if parent == directory {
				break
			}
		}
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return filepath.Abs(candidate)
		}
	}
	return "", fmt.Errorf("%w: set GRIMOIRE_VECTOR_ENGINE or build native/vector-engine", ErrUnavailable)
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
