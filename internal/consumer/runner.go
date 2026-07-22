package consumer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const Version = 1

type Definition struct {
	Version int      `json:"version"`
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
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
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(directory, entry.Name())
		definition, err := load(path)
		if err != nil {
			return err
		}
		if err := invoke(ctx, definition, repository, stateRoot, snapshotID, output); err != nil {
			return fmt.Errorf("Lexicon consumer %s: %w", entry.Name(), err)
		}
	}
	return nil
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
	if definition.Version != Version {
		return Definition{}, fmt.Errorf("unsupported Lexicon consumer version %d", definition.Version)
	}
	if strings.TrimSpace(definition.Command) == "" {
		return Definition{}, fmt.Errorf("Lexicon consumer %s has no command", path)
	}
	return definition, nil
}

func invoke(
	ctx context.Context,
	definition Definition,
	repository, stateRoot, snapshotID string,
	output io.Writer,
) error {
	command := exec.CommandContext(ctx, definition.Command, definition.Args...)
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
