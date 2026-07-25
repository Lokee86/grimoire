package repostate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func sourceFingerprint(root string) (string, error) {
	hash := sha256.New()
	paths := make([]string, 0, 128)
	contents := make(map[string][]byte)
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
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		paths = append(paths, relative)
		contents[relative] = data
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(paths)
	for _, relative := range paths {
		_, _ = hash.Write([]byte(relative))
		_, _ = hash.Write([]byte{0})
		fileHash := sha256.Sum256(contents[relative])
		_, _ = hash.Write(fileHash[:])
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), nil
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
		for _, prefix := range []string{".git/", ".worktrees/", ".lexicon/", ".arcana/", ".grimoire/"} {
			if strings.HasPrefix(clean, prefix) || clean == strings.TrimSuffix(prefix, "/") {
				return true
			}
		}
	}
	return false
}

func excludedDirectory(name string) bool {
	switch name {
	case ".git", ".worktrees", ".lexicon", ".arcana", ".grimoire":
		return true
	default:
		return false
	}
}
