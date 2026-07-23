package assembly

import (
	"path/filepath"
	"strings"
)

type orderedSet struct {
	seen   map[string]struct{}
	values []string
}

func newOrderedSet() *orderedSet {
	return &orderedSet{seen: make(map[string]struct{})}
}

func (set *orderedSet) Add(value string) {
	if value == "" {
		return
	}
	if _, exists := set.seen[value]; exists {
		return
	}
	set.seen[value] = struct{}{}
	set.values = append(set.values, value)
}

func (set *orderedSet) Len() int {
	return len(set.values)
}

func (set *orderedSet) Values() []string {
	return append([]string(nil), set.values...)
}

func candidateRole(path string) string {
	normalized := strings.ToLower(filepath.ToSlash(path))
	base := filepath.Base(normalized)
	if strings.Contains(base, "_test.") || strings.Contains(base, ".test.") ||
		containsPathPart(normalized, "test") || containsPathPart(normalized, "tests") ||
		containsPathPart(normalized, "spec") || containsPathPart(normalized, "specs") {
		return "verification"
	}
	if containsPathPart(normalized, "docs") || strings.HasSuffix(base, ".md") ||
		strings.HasSuffix(base, ".rst") || strings.HasSuffix(base, ".txt") {
		return "documentation"
	}
	switch filepath.Ext(base) {
	case ".json", ".toml", ".yaml", ".yml", ".ini", ".conf":
		return "configuration"
	default:
		return "implementation"
	}
}

func containsPathPart(path, target string) bool {
	for _, part := range strings.Split(strings.Trim(path, "/"), "/") {
		if part == target {
			return true
		}
	}
	return false
}
