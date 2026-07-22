package files

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"
)

const IgnoreFileName = ".lexiconignore"

type IgnorePolicy struct {
	root    string
	matcher *ignore.GitIgnore
}

func LoadIgnorePolicy(root string) (IgnorePolicy, error) {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return IgnorePolicy{}, err
	}
	matcher, err := ignore.CompileIgnoreFile(filepath.Join(absolute, IgnoreFileName))
	if errors.Is(err, os.ErrNotExist) {
		return IgnorePolicy{root: absolute}, nil
	}
	if err != nil {
		return IgnorePolicy{}, err
	}
	return IgnorePolicy{root: absolute, matcher: matcher}, nil
}

func (p IgnorePolicy) Ignored(path string, isDir bool) bool {
	relative, err := filepath.Rel(p.root, path)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return true
	}
	relative = filepath.Clean(relative)
	if relative == "." {
		return false
	}
	if containsIgnoredDirectory(relative) {
		return true
	}
	if p.matcher == nil {
		return false
	}
	matchPath := filepath.ToSlash(relative)
	parts := strings.Split(matchPath, "/")
	for index := 1; index < len(parts); index++ {
		parent := strings.Join(parts[:index], "/") + "/"
		if p.matcher.MatchesPath(parent) {
			return true
		}
	}
	if isDir {
		matchPath += "/"
	}
	return p.matcher.MatchesPath(matchPath)
}

func IsIgnoreFile(root, path string) bool {
	relative, err := filepath.Rel(root, path)
	return err == nil && filepath.Clean(relative) == IgnoreFileName
}

func containsIgnoredDirectory(path string) bool {
	for _, part := range strings.Split(filepath.Clean(path), string(filepath.Separator)) {
		if IgnoredDirectory(part) {
			return true
		}
	}
	return false
}
