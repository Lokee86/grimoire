package repostate

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func runCommand(ctx context.Context, command ProcessCommand) error {
	var stdout, stderr bytes.Buffer
	process := exec.CommandContext(ctx, command.Executable, command.Arguments...)
	if len(command.Environment) > 0 {
		process.Env = command.Environment
	}
	process.Stdout, process.Stderr = &stdout, &stderr
	if err := process.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = strings.TrimSpace(stdout.String())
		}
		if message != "" {
			return fmt.Errorf("%w: %s", err, message)
		}
		return err
	}
	return nil
}

func commandEnvironment(key, value string) []string {
	environment := os.Environ()
	filtered := environment[:0]
	for _, entry := range environment {
		name, _, found := strings.Cut(entry, "=")
		if found && strings.EqualFold(name, key) {
			continue
		}
		filtered = append(filtered, entry)
	}
	return append(filtered, key+"="+value)
}

func commandFor(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
