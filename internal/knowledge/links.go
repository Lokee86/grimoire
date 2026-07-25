package knowledge

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	ignorepolicy "github.com/Lokee86/grimoire/internal/ignore"
)

type catalogEntry struct {
	kind  string
	value string
	path  string
}

type codeCatalog struct {
	paths  []string
	values []catalogEntry
}

var declarationPattern = regexp.MustCompile(`(?m)\b(?:func|type|class|struct|enum|interface|message|service|const|var)\s+([A-Za-z_][A-Za-z0-9_]*)`)
var configKeyPattern = regexp.MustCompile(`(?m)^\s*["']?([A-Za-z_][A-Za-z0-9_.-]*)["']?\s*[:=]`)
var literalPattern = regexp.MustCompile(`["']([A-Za-z][A-Za-z0-9_.-]{2,}|/[A-Za-z0-9_./?=&{}${}:-]+)["']`)

func buildCodeCatalog(root, ignoreFile string, excluded []string) (*codeCatalog, error) {
	policy, err := ignorepolicy.Load(root, ignoreFile)
	if err != nil {
		return nil, err
	}
	catalog := &codeCatalog{}
	seenPaths := make(map[string]struct{})
	seenValues := make(map[string]struct{})
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path != root && ignoredDirectory(entry, path, excluded) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if path != root {
			ignored, ignoreErr := policy.Ignored(path, entry.IsDir())
			if ignoreErr != nil {
				return ignoreErr
			}
			if ignored {
				if entry.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		if entry.IsDir() {
			return policy.LoadDirectory(path)
		}
		if path == policy.ControlFile() || !entry.Type().IsRegular() {
			return nil
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil || len(content) > 2<<20 || bytes.IndexByte(content, 0) >= 0 {
			return nil
		}
		relative, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		relative = filepath.ToSlash(relative)
		if _, ok := seenPaths[relative]; !ok {
			seenPaths[relative] = struct{}{}
			catalog.paths = append(catalog.paths, relative)
		}
		if isKnowledgeFile(relative, entry.Name(), true) {
			return nil
		}
		for _, match := range declarationPattern.FindAllStringSubmatch(string(content), -1) {
			catalog.add(catalogEntry{kind: "symbol", value: match[1], path: relative}, seenValues)
		}
		for _, match := range configKeyPattern.FindAllStringSubmatch(string(content), -1) {
			catalog.add(catalogEntry{kind: "config-contract", value: match[1], path: relative}, seenValues)
		}
		for _, match := range literalPattern.FindAllStringSubmatch(string(content), -1) {
			value := match[1]
			kind := "contract"
			if strings.HasPrefix(value, "/") {
				kind = "endpoint"
			}
			catalog.add(catalogEntry{kind: kind, value: value, path: relative}, seenValues)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("build code-link catalog: %w", err)
	}
	sort.Strings(catalog.paths)
	sort.Slice(catalog.values, func(i, j int) bool {
		if catalog.values[i].value != catalog.values[j].value {
			return catalog.values[i].value < catalog.values[j].value
		}
		if catalog.values[i].kind != catalog.values[j].kind {
			return catalog.values[i].kind < catalog.values[j].kind
		}
		return catalog.values[i].path < catalog.values[j].path
	})
	return catalog, nil
}

func (catalog *codeCatalog) add(entry catalogEntry, seen map[string]struct{}) {
	if entry.value == "" || len(entry.value) < 3 {
		return
	}
	key := entry.kind + "\x00" + entry.value + "\x00" + entry.path
	if _, exists := seen[key]; exists {
		return
	}
	seen[key] = struct{}{}
	catalog.values = append(catalog.values, entry)
}

func (catalog *codeCatalog) linksFor(text string) []CodeLink {
	links := make([]CodeLink, 0)
	seen := make(map[string]struct{})
	for _, path := range catalog.paths {
		if !strings.Contains(text, path) {
			continue
		}
		key := "path\x00" + path
		if _, ok := seen[key]; !ok {
			seen[key] = struct{}{}
			links = append(links, CodeLink{Kind: "path", Value: path, SourcePath: path, Evidence: "exact repository path match"})
		}
	}
	for _, entry := range catalog.values {
		if !wholeWordContains(text, entry.value) {
			continue
		}
		key := entry.kind + "\x00" + entry.value + "\x00" + entry.path
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		links = append(links, CodeLink{Kind: entry.kind, Value: entry.value, SourcePath: entry.path, Evidence: "exact repository name match"})
	}
	return links
}

func wholeWordContains(text, value string) bool {
	if strings.Contains(value, "/") || strings.Contains(value, ".") || strings.Contains(value, "-") {
		return strings.Contains(text, value)
	}
	pattern := regexp.MustCompile(`(?m)(?:^|[^A-Za-z0-9_])` + regexp.QuoteMeta(value) + `(?:$|[^A-Za-z0-9_])`)
	return pattern.MatchString(text)
}
