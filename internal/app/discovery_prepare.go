package app

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Lokee86/grimoire/internal/repostate"
)

func ensureDiscoveryRepository(ctx context.Context, options repostate.Options) (repostate.Status, error) {
	grimoireCommand := filepath.Clean(strings.TrimSpace(options.GrimoireCommand))
	options.Run = func(commandContext context.Context, command string, arguments ...string) error {
		if sameDiscoveryCommand(command, grimoireCommand) && len(arguments) > 0 {
			switch arguments[0] {
			case "index":
				return runIndex(arguments[1:], io.Discard, io.Discard)
			case "knowledge":
				if len(arguments) > 1 && arguments[1] == "index" {
					return runKnowledge(arguments[1:], io.Discard, io.Discard)
				}
			}
		}
		return runDiscoveryCommand(commandContext, command, arguments...)
	}
	return repostate.Ensure(ctx, options)
}

func sameDiscoveryCommand(command, expected string) bool {
	command = filepath.Clean(strings.TrimSpace(command))
	if command == "" || expected == "" {
		return false
	}
	return strings.EqualFold(command, expected)
}

func runDiscoveryCommand(ctx context.Context, command string, arguments ...string) error {
	var stdout, stderr bytes.Buffer
	process := exec.CommandContext(ctx, command, arguments...)
	process.Stdout = &stdout
	process.Stderr = &stderr
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
