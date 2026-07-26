package knowledge

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	ignorepolicy "github.com/Lokee86/grimoire/internal/ignore"
	git "github.com/go-git/go-git/v5"
)

const defaultMaxBytes int64 = 2 << 20

type repoMetadata struct {
	commit string
	time   *time.Time
}

// Discover returns the current repository prose without persisting an index.
func Discover(root string, options BuildOptions) ([]Document, error) {
	index, _, err := Build(root, nil, options)
	if err != nil {
		return nil, err
	}
	return index.Documents, nil
}

func Build(root string, previous *Index, options BuildOptions) (Index, BuildStats, error) {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return Index{}, BuildStats{}, fmt.Errorf("resolve repository root: %w", err)
	}
	policy, err := ignorepolicy.Load(absolute, options.IgnoreFile)
	if err != nil {
		return Index{}, BuildStats{}, err
	}
	excluded, err := normalizeExcludes(absolute, options.ExcludePaths)
	if err != nil {
		return Index{}, BuildStats{}, err
	}
	metadata := readRepoMetadata(absolute)
	old := make(map[string]Document)
	if previous != nil {
		for _, document := range previous.Documents {
			old[document.Path] = document
		}
	}
	maxBytes := options.MaxFileBytes
	if maxBytes <= 0 {
		maxBytes = defaultMaxBytes
	}
	seen := make(map[string]struct{})
	documents := make([]Document, 0)
	stats := BuildStats{}
	err = filepath.WalkDir(absolute, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path != absolute && ignoredDirectory(entry, path, excluded) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if path != absolute {
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
		relative, relErr := filepath.Rel(absolute, path)
		if relErr != nil {
			return relErr
		}
		relative = filepath.ToSlash(relative)
		if !isKnowledgeFile(relative, entry.Name(), options.IncludeConfig) {
			return nil
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			return infoErr
		}
		if info.Size() > maxBytes {
			return nil
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if bytes.IndexByte(content, 0) >= 0 {
			return nil
		}
		seen[relative] = struct{}{}
		stats.Scanned++
		hash := hashBytes(content)
		if prior, ok := old[relative]; ok && prior.Hash == hash && prior.Size == info.Size() {
			prior.CommitID, prior.CommitTime = metadata.commit, metadata.time
			documents = append(documents, prior)
			stats.Reused++
			return nil
		}
		documents = append(documents, Document{
			ID: stableID("document", relative), Path: relative, Kind: classify(relative),
			Hash: hash, Size: info.Size(), CommitID: metadata.commit, CommitTime: metadata.time,
			Sections: extractSections(relative, string(content)),
		})
		stats.Updated++
		return nil
	})
	if err != nil {
		return Index{}, BuildStats{}, fmt.Errorf("walk knowledge documents: %w", err)
	}
	for path := range old {
		if _, ok := seen[path]; !ok {
			stats.Removed++
		}
	}
	catalog, catalogErr := buildCodeCatalog(absolute, options.IgnoreFile, excluded)
	if catalogErr != nil {
		return Index{}, BuildStats{}, catalogErr
	}
	for documentIndex := range documents {
		for sectionIndex := range documents[documentIndex].Sections {
			documents[documentIndex].Sections[sectionIndex].CodeLinks = catalog.linksFor(documents[documentIndex].Sections[sectionIndex].Text)
		}
	}
	sort.Slice(documents, func(i, j int) bool { return documents[i].Path < documents[j].Path })
	result := Index{Version: FormatVersion, Root: absolute, GitCommit: metadata.commit, GitTime: metadata.time, Documents: documents}
	result.SourceFingerprint = Identity(result)
	return result, stats, nil
}

func isKnowledgeFile(path, name string, includeConfig bool) bool {
	lower := strings.ToLower(path)
	ext := strings.ToLower(filepath.Ext(name))
	if ext == ".md" || ext == ".mdx" || ext == ".markdown" || ext == ".adoc" || ext == ".rst" || ext == ".txt" {
		return true
	}
	base := strings.ToLower(filepath.Base(path))
	if ext == "" && (strings.Contains(base, "readme") || strings.Contains(base, "changelog") || strings.Contains(base, "notes") || strings.Contains(base, "design")) {
		return true
	}
	if looksKnowledgePath(lower) && (ext == ".json" || ext == ".yaml" || ext == ".yml" || ext == ".toml" || ext == ".ini" || ext == ".conf") {
		return includeConfig || proseHeavyConfig(path)
	}
	return false
}

func looksKnowledgePath(path string) bool {
	for _, part := range strings.Split(path, "/") {
		switch strings.ToLower(part) {
		case "doc", "docs", "architecture", "planning", "design", "adr", "adrs", "issues", "issue", "notes", "exports", "export", "proposals":
			return true
		}
	}
	return false
}

func proseHeavyConfig(path string) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	text := string(content)
	words := len(strings.Fields(text))
	if words < 12 {
		return false
	}
	structural := strings.Count(text, ":") + strings.Count(text, "{") + strings.Count(text, "}") + strings.Count(text, "[") + strings.Count(text, "]")
	return words > structural*2
}

func classify(path string) Kind {
	lower := strings.ToLower(path)
	if strings.HasSuffix(lower, ".yaml") || strings.HasSuffix(lower, ".yml") || strings.HasSuffix(lower, ".toml") || strings.HasSuffix(lower, ".json") || strings.HasSuffix(lower, ".ini") || strings.HasSuffix(lower, ".conf") {
		return KindConfig
	}
	if strings.Contains(lower, "/adr") || strings.Contains(filepath.Base(lower), "adr-") || strings.Contains(filepath.Base(lower), "decision") {
		return KindADR
	}
	for _, item := range []struct {
		name string
		kind Kind
	}{
		{"architecture", KindArchitecture}, {"planning", KindPlanning}, {"issue", KindIssue}, {"export", KindExport},
	} {
		if strings.Contains(lower, item.name) {
			return item.kind
		}
	}
	if filepath.Ext(lower) == ".txt" || filepath.Ext(lower) == ".rst" || filepath.Ext(lower) == ".adoc" {
		return KindText
	}
	return KindMarkdown
}

func ignoredDirectory(entry fs.DirEntry, path string, excluded []string) bool {
	if entry.IsDir() {
		switch entry.Name() {
		case ".git", ".worktrees", ".lexicon", ".arcana", ".grimoire", ".ddocs", ".warlock":
			return true
		}
		normalized := filepath.ToSlash(filepath.Clean(path))
		for _, generated := range []string{"/evaluation/results", "/evaluation/generated", "/evaluation/output"} {
			if strings.HasSuffix(normalized, generated) || strings.Contains(normalized, generated+"/") {
				return true
			}
		}
	}
	return pathExcluded(path, excluded)
}

func normalizeExcludes(root string, values []string) ([]string, error) {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if !filepath.IsAbs(value) {
			value = filepath.Join(root, value)
		}
		absolute, err := filepath.Abs(value)
		if err != nil {
			return nil, err
		}
		result = append(result, filepath.Clean(absolute))
	}
	return result, nil
}

func pathExcluded(path string, roots []string) bool {
	path = filepath.Clean(path)
	for _, root := range roots {
		relative, err := filepath.Rel(root, path)
		if err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

func readRepoMetadata(root string) repoMetadata {
	repository, err := git.PlainOpen(root)
	if err != nil {
		return repoMetadata{}
	}
	head, err := repository.Head()
	if err != nil {
		return repoMetadata{}
	}
	commit, err := repository.CommitObject(head.Hash())
	if err != nil {
		return repoMetadata{commit: head.Hash().String()}
	}
	commitTime := commit.Author.When.UTC()
	return repoMetadata{commit: head.Hash().String(), time: &commitTime}
}
