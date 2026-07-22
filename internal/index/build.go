package index

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const defaultMaxFileBytes int64 = 2 << 20

type BuildOptions struct {
	MaxFileBytes int64
}

type BuildStats struct {
	Scanned int `json:"scanned"`
	Reused  int `json:"reused"`
	Updated int `json:"updated"`
	Removed int `json:"removed"`
}

func Build(root string, previous *Snapshot, options BuildOptions) (Snapshot, BuildStats, error) {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return Snapshot{}, BuildStats{}, fmt.Errorf("resolve root: %w", err)
	}

	maxBytes := options.MaxFileBytes
	if maxBytes <= 0 {
		maxBytes = defaultMaxFileBytes
	}

	oldFiles := make(map[string]FileRecord)
	baseRoot := ""
	baseShards := make(map[string]string)
	if previous != nil {
		baseRoot = previous.baseRoot
		for name, hash := range previous.baseShards {
			baseShards[name] = hash
		}
		for _, file := range previous.Files {
			oldFiles[file.Path] = file
		}
	}
	dirtyShards := make(map[string]bool)

	seen := make(map[string]struct{})
	files := make([]FileRecord, 0, len(oldFiles))
	stats := BuildStats{}

	err = filepath.WalkDir(absoluteRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path != absoluteRoot && ignoredDirectory(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if !entry.Type().IsRegular() || !indexableFile(entry.Name()) {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Size() > maxBytes {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if bytes.IndexByte(content, 0) >= 0 {
			return nil
		}

		relative, err := filepath.Rel(absoluteRoot, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		seen[relative] = struct{}{}
		stats.Scanned++

		hash := contentHash(content)
		if old, ok := oldFiles[relative]; ok && old.Hash == hash && old.Size == info.Size() {
			files = append(files, old)
			stats.Reused++
			return nil
		}

		files = append(files, FileRecord{
			Path:   relative,
			Hash:   hash,
			Size:   info.Size(),
			Chunks: chunkFile(relative, string(content)),
		})
		dirtyShards[shardName(relative)] = true
		stats.Updated++
		return nil
	})
	if err != nil {
		return Snapshot{}, BuildStats{}, fmt.Errorf("walk repository: %w", err)
	}

	for path := range oldFiles {
		if _, ok := seen[path]; !ok {
			dirtyShards[shardName(path)] = true
			stats.Removed++
		}
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })

	return Snapshot{
		Version: FormatVersion, Files: files,
		baseRoot: baseRoot, baseShards: baseShards, dirtyShards: dirtyShards,
	}, stats, nil
}

func ignoredDirectory(name string) bool {
	switch name {
	case ".git", ".grimoire", ".worktrees", ".workingtrees", ".obsidian",
		"node_modules", "vendor", "target", "dist", "build", "coverage":
		return true
	default:
		return false
	}
}

func indexableFile(name string) bool {
	extension := strings.ToLower(filepath.Ext(name))
	switch extension {
	case ".go", ".rs", ".py", ".rb", ".js", ".jsx", ".ts", ".tsx",
		".java", ".c", ".h", ".cc", ".cpp", ".hpp", ".cs", ".gd",
		".md", ".txt", ".toml", ".yaml", ".yml", ".json", ".xml",
		".html", ".css", ".scss", ".sql", ".sh", ".ps1":
		return true
	}

	switch strings.ToLower(name) {
	case "readme", "license", "makefile", "dockerfile", "gemfile", "rakefile":
		return true
	default:
		return false
	}
}

func contentHash(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}
