package diffcontext

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	WorkingTree = "working-tree"
	Staged      = "staged"
	Unstaged    = "unstaged"
)

// Runner executes one command without involving a shell. It exists so diff
// collection can be tested independently from a local Git installation.
type Runner interface {
	Run(context.Context, string, ...string) ([]byte, error)
}

type commandRunner struct{}

func (commandRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, name, args...)
	return command.CombinedOutput()
}

// Collect resolves a named diff scope or Git revision expression and converts
// it into current-tree source spans. Supported named scopes are working-tree,
// staged, and unstaged. Other values are passed to Git as one revision/range
// argument, never through a shell. Working-tree also includes untracked files.
func Collect(ctx context.Context, root, spec string) ([]Change, error) {
	return CollectWithRunner(ctx, root, spec, commandRunner{})
}

func CollectWithRunner(ctx context.Context, root, spec string, runner Runner) ([]Change, error) {
	if runner == nil {
		return nil, errors.New("diff runner is required")
	}
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve repository root: %w", err)
	}
	args, err := gitDiffArgs(absoluteRoot, spec)
	if err != nil {
		return nil, err
	}
	output, err := runner.Run(ctx, "git", args...)
	if err != nil {
		return nil, gitCommandError("collect git diff", spec, output, err)
	}
	changes := Parse(string(output))
	if strings.EqualFold(strings.TrimSpace(spec), WorkingTree) {
		untrackedOutput, untrackedErr := runner.Run(
			ctx, "git", "-C", absoluteRoot, "ls-files", "--others", "--exclude-standard", "-z", "--",
		)
		if untrackedErr != nil {
			return nil, gitCommandError("collect untracked files", spec, untrackedOutput, untrackedErr)
		}
		changes = append(changes, untrackedChanges(absoluteRoot, untrackedOutput)...)
	}
	return deduplicateChanges(changes), nil
}

func gitCommandError(operation, spec string, output []byte, err error) error {
	message := strings.TrimSpace(string(output))
	if message == "" {
		message = err.Error()
	}
	return fmt.Errorf("%s %q: %s", operation, strings.TrimSpace(spec), message)
}

func untrackedChanges(root string, output []byte) []Change {
	paths := bytes.Split(output, []byte{0})
	changes := make([]Change, 0, len(paths))
	for _, rawPath := range paths {
		path := normalizePath(string(rawPath))
		if path == "" {
			continue
		}
		absolutePath := filepath.Join(root, filepath.FromSlash(path))
		relative, err := filepath.Rel(root, absolutePath)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			continue
		}
		endLine := 1
		if data, readErr := os.ReadFile(absolutePath); readErr == nil {
			endLine = sourceLineCount(data)
		}
		changes = append(changes, Change{
			Path: path, StartLine: 1, EndLine: endLine, Summary: "untracked file",
		})
	}
	return changes
}

func sourceLineCount(data []byte) int {
	if len(data) == 0 {
		return 1
	}
	count := bytes.Count(data, []byte{'\n'})
	if data[len(data)-1] != '\n' {
		count++
	}
	if count == 0 {
		return 1
	}
	return count
}

func gitDiffArgs(root, spec string) ([]string, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil, errors.New("diff scope must not be empty")
	}
	if strings.ContainsRune(spec, '\x00') || strings.HasPrefix(spec, "-") {
		return nil, fmt.Errorf("unsafe diff scope %q", spec)
	}
	args := []string{
		"-C", root, "diff", "--no-ext-diff", "--no-textconv", "--no-color",
		"--find-renames", "--unified=0",
	}
	switch strings.ToLower(spec) {
	case WorkingTree:
		args = append(args, "HEAD")
	case Staged:
		args = append(args, "--cached")
	case Unstaged:
		// No revision or cache flag means index versus working tree.
	default:
		args = append(args, spec)
	}
	return append(args, "--"), nil
}
