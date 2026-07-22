package adapters

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Analyzer interface {
	Run(context.Context, string, string, string) error
}

type Runner struct {
	Root string
}

func (r Runner) Run(ctx context.Context, language, repository, output string) error {
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return err
	}
	command, err := r.command(ctx, language, repository, output)
	if err != nil {
		return err
	}
	data, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s adapter failed: %w\n%s", language, err, strings.TrimSpace(string(data)))
	}
	if info, err := os.Stat(output); err != nil || info.Size() == 0 {
		return fmt.Errorf("%s adapter produced no output", language)
	}
	return nil
}

func (r Runner) command(ctx context.Context, language, repository, output string) (*exec.Cmd, error) {
	switch language {
	case "go", "gdscript":
		executable, err := findExecutable("go")
		if err != nil {
			return nil, err
		}
		directory := filepath.Join(r.Root, language)
		command := exec.CommandContext(ctx, executable, "run", ".", "--repo", repository, "--output", output)
		command.Dir = directory
		return command, nil
	case "python":
		executable, err := findExecutable("python", "python3")
		if err != nil {
			return nil, err
		}
		command := exec.CommandContext(ctx, executable, "-m", "lexicon_python", "--repo", repository, "--output", output)
		pythonRoot := filepath.Join(r.Root, "python")
		command.Env = append(os.Environ(), "PYTHONPATH="+joinPathList(pythonRoot, os.Getenv("PYTHONPATH")))
		return command, nil
	case "ruby":
		executable, err := findExecutable("ruby")
		if err != nil {
			return nil, err
		}
		return exec.CommandContext(ctx, executable, filepath.Join(r.Root, "ruby", "lexicon_ruby.rb"), "--repo", repository, "--output", output), nil
	case "rust":
		executable, err := findExecutable("cargo")
		if err != nil {
			return nil, err
		}
		manifest := filepath.Join(r.Root, "rust", "Cargo.toml")
		return exec.CommandContext(ctx, executable, "run", "--quiet", "--manifest-path", manifest, "--", "--repo", repository, "--output", output), nil
	case "typescript":
		if err := r.prepareTypeScript(ctx); err != nil {
			return nil, err
		}
		executable, err := findExecutable("node")
		if err != nil {
			return nil, err
		}
		return exec.CommandContext(ctx, executable, filepath.Join(r.Root, "typescript", "dist", "cli.js"), "--repo", repository, "--output", output), nil
	default:
		return nil, fmt.Errorf("unsupported language %q", language)
	}
}

func (r Runner) prepareTypeScript(ctx context.Context) error {
	directory := filepath.Join(r.Root, "typescript")
	distribution := filepath.Join(directory, "dist", "cli.js")
	if _, err := os.Stat(distribution); err == nil {
		return nil
	}
	npm, err := findExecutable(npmExecutable())
	if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(directory, "node_modules")); os.IsNotExist(err) {
		command := exec.CommandContext(ctx, npm, "ci", "--silent")
		command.Dir = directory
		if data, runErr := command.CombinedOutput(); runErr != nil {
			return fmt.Errorf("prepare TypeScript dependencies: %w\n%s", runErr, strings.TrimSpace(string(data)))
		}
	}
	command := exec.CommandContext(ctx, npm, "run", "build", "--silent")
	command.Dir = directory
	if data, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("build TypeScript adapter: %w\n%s", err, strings.TrimSpace(string(data)))
	}
	return nil
}

func findExecutable(names ...string) (string, error) {
	for _, name := range names {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}
	home, _ := os.UserHomeDir()
	for _, name := range names {
		for _, candidate := range commonExecutablePaths(home, name) {
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
				return candidate, nil
			}
		}
	}
	return "", fmt.Errorf("required executable not found: %s", strings.Join(names, " or "))
}

func commonExecutablePaths(home, name string) []string {
	if home == "" {
		return nil
	}
	executable := name
	if runtime.GOOS == "windows" && filepath.Ext(executable) == "" {
		executable += ".exe"
	}
	return []string{
		filepath.Join(home, ".cargo", "bin", executable),
		filepath.Join(home, "go", "bin", executable),
	}
}

func joinPathList(first, remainder string) string {
	if remainder == "" {
		return first
	}
	return first + string(os.PathListSeparator) + remainder
}

func npmExecutable() string {
	if runtime.GOOS == "windows" {
		return "npm.cmd"
	}
	return "npm"
}
