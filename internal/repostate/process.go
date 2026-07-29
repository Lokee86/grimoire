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
	process := exec.Command(command.Executable, command.Arguments...)
	configureProcessTree(process)
	if len(command.Environment) > 0 {
		process.Env = command.Environment
	}
	process.Stdout, process.Stderr = &stdout, &stderr
	if err := process.Start(); err != nil {
		return commandFailure(err, stdout.String(), stderr.String())
	}
	wait := make(chan error, 1)
	go func() { wait <- process.Wait() }()
	select {
	case err := <-wait:
		if err != nil {
			return commandFailure(err, stdout.String(), stderr.String())
		}
		return nil
	case <-ctx.Done():
		terminateProcessTree(process)
		<-wait
		return commandFailure(ctx.Err(), stdout.String(), stderr.String())
	}
}

func commandFailure(err error, stdout, stderr string) error {
	message := strings.TrimSpace(stderr)
	if message == "" {
		message = strings.TrimSpace(stdout)
	}
	if message != "" {
		return fmt.Errorf("%w: %s", err, message)
	}
	return err
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
