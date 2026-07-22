package consumer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const Version = 1

type Definition struct {
	Version int           `json:"version"`
	Command string        `json:"command"`
	Args    []string      `json:"args,omitempty"`
	Timeout time.Duration `json:"timeout,omitempty"`
}

func Run(ctx context.Context, repository, stateRoot, snapshotID string, output io.Writer) error {
	directory := filepath.Join(stateRoot, "consumers")
	entries, err := os.ReadDir(directory)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read Lexicon consumers: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	var failures []error
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		if err := runOne(ctx, repository, stateRoot, entry.Name(), snapshotID, output); err != nil {
			failures = append(failures, fmt.Errorf("Lexicon consumer %s: %w", entry.Name(), err))
		}
	}
	return errors.Join(failures...)
}

func runOne(
	ctx context.Context,
	repository, stateRoot, name, snapshotID string,
	output io.Writer,
) error {
	definition, err := load(filepath.Join(stateRoot, "consumers", name))
	if err != nil {
		return err
	}
	if err := invoke(ctx, definition, repository, stateRoot, snapshotID, output); err != nil {
		return err
	}
	return saveSnapshot(stateRoot, name, snapshotID)
}

func load(path string) (Definition, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Definition{}, err
	}
	var definition Definition
	if err := json.Unmarshal(data, &definition); err != nil {
		return Definition{}, fmt.Errorf("decode Lexicon consumer %s: %w", path, err)
	}
	if err := validateDefinition(definition); err != nil {
		return Definition{}, fmt.Errorf("Lexicon consumer %s: %w", path, err)
	}
	return definition, nil
}

func invoke(
	ctx context.Context,
	definition Definition,
	repository, stateRoot, snapshotID string,
	output io.Writer,
) error {
	commandContext := ctx
	cancel := func() {}
	if definition.Timeout > 0 {
		commandContext, cancel = context.WithTimeout(ctx, definition.Timeout)
	}
	defer cancel()
	command := exec.CommandContext(commandContext, definition.Command, definition.Args...)
	command.Dir = repository
	command.Env = append(os.Environ(),
		"LEXICON_REPOSITORY="+repository,
		"LEXICON_STATE_ROOT="+stateRoot,
		"LEXICON_SNAPSHOT_ID="+snapshotID,
	)
	var captured bytes.Buffer
	command.Stdout = &captured
	command.Stderr = &captured
	if err := command.Run(); err != nil {
		if commandContext.Err() != nil {
			return fmt.Errorf("%w: %v: %s", commandContext.Err(), err, strings.TrimSpace(captured.String()))
		}
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(captured.String()))
	}
	if output != nil && captured.Len() > 0 {
		_, _ = output.Write(captured.Bytes())
		if !bytes.HasSuffix(captured.Bytes(), []byte("\n")) {
			_, _ = io.WriteString(output, "\n")
		}
	}
	return nil
}
