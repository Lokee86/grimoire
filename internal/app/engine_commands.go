package app

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/Lokee86/grimoire/internal/agentruntime"
)

// ExitError preserves the exit status of a delegated engine command. The
// top-level executable suppresses an additional Grimoire error line because
// the child command already owns its stdout and stderr.
type ExitError struct {
	Code int
}

func (err *ExitError) Error() string {
	return fmt.Sprintf("engine command exited with status %d", err.Code)
}

type engineSpec struct {
	name        string
	displayName string
	commandEnv  string
	versionArgs []string
	help        string
}

type engineDependencies struct {
	resolve func(root, requested, name string) string
	run     func(command string, arguments []string, stdin io.Reader, stdout, stderr io.Writer) error
}

func runEngineNamespace(spec engineSpec, arguments []string, stdout, stderr io.Writer) error {
	return runEngineNamespaceWith(spec, arguments, stdout, stderr, engineDependencies{
		resolve: agentruntime.ResolveProviderCommand,
		run:     runEngineProcess,
	})
}

func runEngineNamespaceWith(
	spec engineSpec,
	arguments []string,
	stdout, stderr io.Writer,
	dependencies engineDependencies,
) error {
	if len(arguments) == 0 || isHelpArgument(arguments) {
		_, err := io.WriteString(stdout, spec.help)
		return err
	}

	requested := strings.TrimSpace(os.Getenv(spec.commandEnv))
	command := dependencies.resolve(engineRepositoryRoot(spec, arguments), requested, spec.name)
	if strings.TrimSpace(command) == "" {
		return fmt.Errorf("resolve %s command", spec.displayName)
	}

	if len(arguments) == 1 && arguments[0] == "check" {
		return checkEngine(spec, command, stdout, dependencies.run)
	}
	if isVersionArgument(arguments) {
		arguments = append([]string(nil), spec.versionArgs...)
	}
	return dependencies.run(command, arguments, os.Stdin, stdout, stderr)
}

func isHelpArgument(arguments []string) bool {
	return len(arguments) == 1 &&
		(arguments[0] == "help" || arguments[0] == "-h" || arguments[0] == "--help")
}

func isVersionArgument(arguments []string) bool {
	return len(arguments) == 1 &&
		(arguments[0] == "version" || arguments[0] == "-V" || arguments[0] == "--version")
}

func engineRepositoryRoot(spec engineSpec, arguments []string) string {
	if spec.name != "lexicon" {
		return "."
	}
	for index, argument := range arguments {
		if value, found := strings.CutPrefix(argument, "--repo="); found && strings.TrimSpace(value) != "" {
			return value
		}
		if argument == "--repo" && index+1 < len(arguments) && strings.TrimSpace(arguments[index+1]) != "" {
			return arguments[index+1]
		}
	}
	return "."
}

func checkEngine(
	spec engineSpec,
	command string,
	stdout io.Writer,
	run func(command string, arguments []string, stdin io.Reader, stdout, stderr io.Writer) error,
) error {
	var versionOutput, versionErrors bytes.Buffer
	if err := run(command, spec.versionArgs, nil, &versionOutput, &versionErrors); err != nil {
		message := strings.TrimSpace(versionErrors.String())
		if message == "" {
			message = strings.TrimSpace(versionOutput.String())
		}
		if message == "" {
			message = err.Error()
		}
		return fmt.Errorf("%s check failed: %s", spec.displayName, message)
	}
	version := strings.TrimSpace(versionOutput.String())
	if version == "" {
		version = strings.TrimSpace(versionErrors.String())
	}
	return writeJSON(stdout, struct {
		Engine  string `json:"engine"`
		Command string `json:"command"`
		Version string `json:"version"`
	}{
		Engine: spec.name, Command: command, Version: version,
	})
}

func runEngineProcess(
	command string,
	arguments []string,
	stdin io.Reader,
	stdout, stderr io.Writer,
) error {
	process := exec.Command(command, arguments...)
	process.Stdin = stdin
	process.Stdout = stdout
	process.Stderr = stderr
	if err := process.Run(); err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			code := exitError.ExitCode()
			if code < 1 {
				code = 1
			}
			return &ExitError{Code: code}
		}
		if errors.Is(err, exec.ErrNotFound) || errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%s engine command was not found; install the bundled engine or configure its command: %w", command, err)
		}
		return fmt.Errorf("start engine command %s: %w", command, err)
	}
	return nil
}
