package adapters

import (
	"os"
	"path/filepath"
	"runtime"
)

func packagedExecutable(root, language string) (string, bool) {
	base := filepath.Join(root, language, "lexicon-"+language)
	candidates := []string{base}
	if runtime.GOOS == "windows" {
		candidates = append(candidates, base+".exe")
	}
	for _, candidate := range candidates {
		if fileExists(candidate) {
			return candidate, true
		}
	}
	return "", false
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
