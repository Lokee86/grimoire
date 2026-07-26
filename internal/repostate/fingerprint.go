package repostate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const maxFingerprintFileBytes int64 = 2 << 20

func sourceFingerprint(root string) (string, error) {
	paths, ok := gitSourcePaths(root)
	if !ok {
		var err error
		paths, err = walkedSourcePaths(root)
		if err != nil {
			return "", err
		}
	}
	sort.Strings(paths)
	hash := sha256.New()
	for _, relative := range paths {
		path := filepath.Join(root, filepath.FromSlash(relative))
		info, err := os.Lstat(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return "", err
		}
		if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Size() > maxFingerprintFileBytes {
			continue
		}
		file, err := os.Open(path)
		if err != nil {
			return "", err
		}
		fileHash := sha256.New()
		_, copyErr := io.Copy(fileHash, file)
		closeErr := file.Close()
		if copyErr != nil {
			return "", copyErr
		}
		if closeErr != nil {
			return "", closeErr
		}
		_, _ = hash.Write([]byte(relative))
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write(fileHash.Sum(nil))
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), nil
}

func gitSourcePaths(root string) ([]string, bool) {
	command := exec.Command("git", "-C", root, "ls-files", "-z", "--cached", "--others", "--exclude-standard")
	data, err := command.Output()
	if err != nil {
		return nil, false
	}
	seen := make(map[string]bool)
	paths := make([]string, 0, 256)
	for _, raw := range strings.Split(string(data), "\x00") {
		relative := filepath.ToSlash(strings.TrimSpace(raw))
		if relative == "" || !fingerprintPath(relative) || seen[relative] {
			continue
		}
		seen[relative] = true
		paths = append(paths, relative)
	}
	return paths, true
}

func walkedSourcePaths(root string) ([]string, error) {
	paths := make([]string, 0, 256)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path != root && entry.IsDir() && excludedDirectory(entry.Name()) {
			return filepath.SkipDir
		}
		if path == root || entry.IsDir() || entry.Type()&os.ModeSymlink != 0 || !entry.Type().IsRegular() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if fingerprintPath(relative) {
			paths = append(paths, relative)
		}
		return nil
	})
	return paths, err
}

func fingerprintPath(relative string) bool {
	for _, part := range strings.Split(filepath.ToSlash(relative), "/") {
		if excludedDirectory(part) {
			return false
		}
	}
	name := filepath.Base(relative)
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

func gitIdentity(ctx context.Context, root string) (head string, dirty, available bool) {
	head = gitOutput(ctx, root, "rev-parse", "--verify", "HEAD")
	if head == "" {
		return "", false, false
	}
	available = true
	output := gitOutput(ctx, root, "status", "--porcelain=v1", "--untracked-files=all")
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) == "" || excludedStatusLine(line) {
			continue
		}
		dirty = true
		break
	}
	return head, dirty, available
}

func gitOutput(ctx context.Context, root string, arguments ...string) string {
	command := exec.CommandContext(ctx, "git", append([]string{"-C", root}, arguments...)...)
	data, err := command.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func excludedStatusLine(line string) bool {
	value := strings.TrimSpace(line)
	if len(value) > 3 {
		value = value[3:]
	}
	value = strings.TrimSpace(strings.Trim(value, `"`))
	for _, part := range strings.Split(value, " -> ") {
		clean := filepath.ToSlash(strings.TrimSpace(part))
		for _, prefix := range []string{".git/", ".worktrees/", ".workingtrees/", ".lexicon/", ".arcana/", ".grimoire/", ".ddocs/", ".warlock/", ".obsidian/"} {
			if strings.HasPrefix(clean, prefix) || clean == strings.TrimSuffix(prefix, "/") {
				return true
			}
		}
	}
	return false
}

func excludedDirectory(name string) bool {
	switch strings.ToLower(name) {
	case ".git", ".worktrees", ".workingtrees", ".lexicon", ".arcana", ".grimoire", ".ddocs", ".warlock", ".obsidian", ".godot",
		"node_modules", "vendor", "target", "dist", "build", "coverage", ".next", ".astro", ".cache":
		return true
	default:
		return false
	}
}
