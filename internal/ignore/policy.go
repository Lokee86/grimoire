package ignore

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

const defaultFileName = ".gitignore"

type Policy struct {
	root       string
	customFile string
	patterns   []gitignore.Pattern
	matcher    gitignore.Matcher
	loaded     map[string]struct{}
}

func Load(root, customFile string) (*Policy, error) {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve ignore root: %w", err)
	}
	policy := &Policy{
		root:    filepath.Clean(absoluteRoot),
		matcher: gitignore.NewMatcher(nil),
		loaded:  make(map[string]struct{}),
	}
	if customFile != "" {
		if filepath.IsAbs(customFile) {
			policy.customFile = filepath.Clean(customFile)
		} else {
			policy.customFile = filepath.Join(policy.root, customFile)
		}
		if err := policy.loadFile(policy.customFile, nil, true); err != nil {
			return nil, err
		}
		return policy, nil
	}
	if err := policy.LoadDirectory(policy.root); err != nil {
		return nil, err
	}
	return policy, nil
}

func (p *Policy) LoadDirectory(dir string) error {
	if p.customFile != "" {
		return nil
	}
	absoluteDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	absoluteDir = filepath.Clean(absoluteDir)
	domain, err := p.relativeParts(absoluteDir)
	if err != nil {
		return err
	}
	if _, exists := p.loaded[absoluteDir]; exists {
		return nil
	}
	p.loaded[absoluteDir] = struct{}{}
	return p.loadFile(filepath.Join(absoluteDir, defaultFileName), domain, false)
}

func (p *Policy) Ignored(path string, isDir bool) (bool, error) {
	parts, err := p.relativeParts(path)
	if err != nil {
		return false, err
	}
	return p.matcher.Match(parts, isDir), nil
}

func (p *Policy) ControlFile() string {
	return p.customFile
}

func (p *Policy) loadFile(path string, domain []string, required bool) error {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) && !required {
		return nil
	}
	if err != nil {
		return fmt.Errorf("open ignore file %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	for scanner.Scan() {
		p.patterns = append(p.patterns, gitignore.ParsePattern(scanner.Text(), domain))
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read ignore file %s: %w", path, err)
	}
	p.matcher = gitignore.NewMatcher(p.patterns)
	return nil
}

func (p *Policy) relativeParts(path string) ([]string, error) {
	relative, err := filepath.Rel(p.root, filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	if relative == "." {
		return nil, nil
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return nil, fmt.Errorf("path %s is outside ignore root %s", path, p.root)
	}
	return strings.Split(filepath.ToSlash(relative), "/"), nil
}
