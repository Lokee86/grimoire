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
	paths  map[string]struct{}
	values map[string][]catalogEntry
}

var declarationPattern = regexp.MustCompile(`(?m)\b(?:func|type|class|struct|enum|interface|message|service|const|var)\s+([A-Za-z_][A-Za-z0-9_]*)`)
var configKeyPattern = regexp.MustCompile(`(?m)^\s*["']?([A-Za-z_][A-Za-z0-9_.-]*)["']?\s*[:=]`)
var literalPattern = regexp.MustCompile(`["']([A-Za-z][A-Za-z0-9_.-]{2,}|/[A-Za-z0-9_./?=&{}${}:-]+)["']`)
var referencePattern = regexp.MustCompile(`(?:[A-Za-z0-9_!.@-]+/)+[A-Za-z0-9_!.@-]+|/[A-Za-z0-9_./?=&{}$:-]+|[A-Za-z_][A-Za-z0-9_.-]{2,}`)

const (
	maxCodeLinks             = 32
	maxLinksPerReference     = 4
	maxAmbiguousContractUses = 4
	maxAmbiguousSymbolUses   = 8
)

func buildCodeCatalog(root, ignoreFile string, excluded []string) (*codeCatalog, error) {
	policy, err := ignorepolicy.Load(root, ignoreFile)
	if err != nil {
		return nil, err
	}
	catalog := &codeCatalog{
		paths:  make(map[string]struct{}),
		values: make(map[string][]catalogEntry),
	}
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
		catalog.paths[relative] = struct{}{}
		if isKnowledgeFile(relative, entry.Name(), true) {
			return nil
		}
		text := string(content)
		for _, match := range declarationPattern.FindAllStringSubmatch(text, -1) {
			catalog.add(catalogEntry{kind: "symbol", value: match[1], path: relative}, seenValues)
		}
		for _, match := range configKeyPattern.FindAllStringSubmatch(text, -1) {
			catalog.add(catalogEntry{kind: "config-contract", value: match[1], path: relative}, seenValues)
		}
		for _, match := range literalPattern.FindAllStringSubmatch(text, -1) {
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
	for value := range catalog.values {
		sort.Slice(catalog.values[value], func(i, j int) bool {
			left, right := catalog.values[value][i], catalog.values[value][j]
			if left.kind != right.kind {
				return left.kind < right.kind
			}
			return left.path < right.path
		})
	}
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
	catalog.values[entry.value] = append(catalog.values[entry.value], entry)
}

func (catalog *codeCatalog) linksFor(text string) []CodeLink {
	references := referenceCandidates(text)
	matchedPaths := make(map[string]struct{})
	matchedEntries := make(map[string]catalogEntry)
	for reference := range references {
		for _, candidate := range referenceVariants(reference) {
			if _, ok := catalog.paths[candidate]; ok {
				matchedPaths[candidate] = struct{}{}
			}
			entries := catalog.values[candidate]
			kindCounts := make(map[string]int)
			for _, entry := range entries {
				kindCounts[entry.kind]++
			}
			addedByKind := make(map[string]int)
			for _, entry := range entries {
				if entry.kind == "contract" && kindCounts[entry.kind] > maxAmbiguousContractUses {
					continue
				}
				if entry.kind == "symbol" && kindCounts[entry.kind] > maxAmbiguousSymbolUses {
					continue
				}
				if addedByKind[entry.kind] >= maxLinksPerReference {
					continue
				}
				addedByKind[entry.kind]++
				key := entry.kind + "\x00" + entry.value + "\x00" + entry.path
				matchedEntries[key] = entry
			}
		}
	}

	paths := make([]string, 0, len(matchedPaths))
	for path := range matchedPaths {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	entries := make([]catalogEntry, 0, len(matchedEntries))
	for _, entry := range matchedEntries {
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].value != entries[j].value {
			return entries[i].value < entries[j].value
		}
		if entries[i].kind != entries[j].kind {
			return entries[i].kind < entries[j].kind
		}
		return entries[i].path < entries[j].path
	})

	links := make([]CodeLink, 0, min(maxCodeLinks, len(paths)+len(entries)))
	for _, path := range paths {
		if len(links) >= maxCodeLinks {
			break
		}
		links = append(links, CodeLink{Kind: "path", Value: path, SourcePath: path, Evidence: "exact repository path match"})
	}
	for _, entry := range entries {
		if len(links) >= maxCodeLinks {
			break
		}
		links = append(links, CodeLink{Kind: entry.kind, Value: entry.value, SourcePath: entry.path, Evidence: "exact repository name match"})
	}
	return links
}

func referenceCandidates(text string) map[string]struct{} {
	result := make(map[string]struct{})
	for _, reference := range referencePattern.FindAllString(text, -1) {
		result[reference] = struct{}{}
	}
	return result
}

func referenceVariants(reference string) []string {
	result := []string{reference}
	trimmed := strings.TrimRight(reference, ".,;:)]}")
	if trimmed != "" && trimmed != reference {
		result = append(result, trimmed)
	}
	if strings.HasPrefix(reference, "/") {
		if query := strings.IndexByte(reference, '?'); query > 0 {
			result = append(result, reference[:query])
		}
	}
	return result
}
