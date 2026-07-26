package agentruntime

import (
	"os"
	"path/filepath"
	"runtime"
)

func adjacentCommand(name string) string {
	executable, err := os.Executable()
	if err != nil {
		return name
	}
	return adjacentCommandFrom(executable, name)
}

func adjacentCommandFrom(executable, name string) string {
	candidates := []string{name}
	if runtime.GOOS == "windows" {
		candidates = append([]string{name + ".exe"}, candidates...)
	}
	for _, candidate := range candidates {
		path := filepath.Join(filepath.Dir(executable), candidate)
		if info, statErr := os.Stat(path); statErr == nil && !info.IsDir() {
			return path
		}
	}
	return name
}
