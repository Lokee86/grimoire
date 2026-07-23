package index

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

func permanentlyIgnoredDirectory(entry fs.DirEntry) bool {
	if !entry.IsDir() {
		return false
	}
	switch entry.Name() {
	case ".git", ".grimoire", ".ddocs", ".lexicon", ".arcana", ".warlock", ".worktrees", ".workingtrees":
		return true
	default:
		return false
	}
}

func normalizeExcludedPaths(root string, paths []string) ([]string, error) {
	normalized := make([]string, 0, len(paths))
	for _, path := range paths {
		if path == "" {
			continue
		}
		if !filepath.IsAbs(path) {
			path = filepath.Join(root, path)
		}
		absolute, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("resolve excluded path %s: %w", path, err)
		}
		normalized = append(normalized, filepath.Clean(absolute))
	}
	return normalized, nil
}

func pathExcluded(path string, excluded []string) bool {
	path = filepath.Clean(path)
	for _, root := range excluded {
		relative, err := filepath.Rel(root, path)
		if err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return true
		}
	}
	return false
}
