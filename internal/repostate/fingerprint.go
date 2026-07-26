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

// RepositoryFingerprint returns the source identity used by repository state
// and knowledge-link freshness checks.
func RepositoryFingerprint(root string) (string, error) {
	return sourceFingerprint(root)
}

func sourceFingerprint(root string) (string, error) {
	if fingerprint, ok, err := gitFingerprint(root); ok || err != nil {
		return fingerprint, err
	}
	paths, err := walkedSourcePaths(root)
	if err != nil {
		return "", err
	}
	sort.Strings(paths)
	hash := sha256.New()
	for _, relative := range paths {
		if err := hashWorkingFile(hash, root, relative, "file"); err != nil {
			return "", err
		}
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), nil
}

func gitFingerprint(root string) (string, bool, error) {
	indexData, err := exec.Command("git", "-C", root, "ls-files", "-s", "-z").Output()
	if err != nil {
		return "", false, nil
	}
	hash := sha256.New()
	for _, raw := range strings.Split(string(indexData), "\x00") {
		metadata, relative, ok := strings.Cut(raw, "\t")
		if !ok || !fingerprintPath(relative) {
			continue
		}
		fields := strings.Fields(metadata)
		if len(fields) != 3 || fields[2] != "0" || strings.HasPrefix(fields[0], "120") {
			continue
		}
		path := filepath.Join(root, filepath.FromSlash(relative))
		if info, statErr := os.Lstat(path); statErr == nil && info.Size() > maxFingerprintFileBytes {
			continue
		}
		_, _ = io.WriteString(hash, "index\x00"+filepath.ToSlash(relative)+"\x00"+fields[1]+"\x00")
	}

	changed, err := gitPathList(root, "diff", "--name-only", "-z")
	if err != nil {
		return "", true, err
	}
	untracked, err := gitPathList(root, "ls-files", "--others", "--exclude-standard", "-z")
	if err != nil {
		return "", true, err
	}
	for _, group := range []struct {
		name  string
		paths []string
	}{{"worktree", changed}, {"untracked", untracked}} {
		sort.Strings(group.paths)
		for _, relative := range group.paths {
			if !fingerprintPath(relative) {
				continue
			}
			if err := hashWorkingFile(hash, root, relative, group.name); err != nil {
				return "", true, err
			}
		}
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), true, nil
}

func gitPathList(root string, arguments ...string) ([]string, error) {
	data, err := exec.Command("git", append([]string{"-C", root}, arguments...)...).Output()
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool)
	result := make([]string, 0)
	for _, raw := range strings.Split(string(data), "\x00") {
		relative := filepath.ToSlash(strings.TrimSpace(raw))
		if relative != "" && !seen[relative] {
			seen[relative] = true
			result = append(result, relative)
		}
	}
	return result, nil
}

func hashWorkingFile(hash io.Writer, root, relative, kind string) error {
	path := filepath.Join(root, filepath.FromSlash(relative))
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		_, _ = io.WriteString(hash, kind+"\x00"+filepath.ToSlash(relative)+"\x00deleted\x00")
		return nil
	}
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Size() > maxFingerprintFileBytes {
		return nil
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	fileHash := sha256.New()
	if _, err := io.Copy(fileHash, file); err != nil {
		return err
	}
	_, _ = io.WriteString(hash, kind+"\x00"+filepath.ToSlash(relative)+"\x00")
	_, _ = hash.Write(fileHash.Sum(nil))
	return nil
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
