package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func ListDefinitions(stateRoot string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(stateRoot, "consumers"))
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read Lexicon consumers: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	return names, nil
}

func AddDefinition(stateRoot, name string, definition Definition) error {
	name, err := validateName(name)
	if err != nil {
		return err
	}
	if err := validateDefinition(definition); err != nil {
		return err
	}
	data, err := json.MarshalIndent(definition, "", "  ")
	if err != nil {
		return fmt.Errorf("encode Lexicon consumer %s: %w", name, err)
	}
	return writeAtomic(filepath.Join(stateRoot, "consumers", name), append(data, '\n'))
}

func RemoveDefinition(stateRoot, name string) error {
	name, err := validateName(name)
	if err != nil {
		return err
	}
	paths := []string{
		filepath.Join(stateRoot, "consumers", name),
		filepath.Join(stateRoot, "consumer-state", name),
	}
	var failures []error
	for _, path := range paths {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}

func RunOne(
	ctx context.Context,
	repository, stateRoot, name, snapshotID string,
	output io.Writer,
) error {
	name, err := validateName(name)
	if err != nil {
		return err
	}
	return runOne(ctx, repository, stateRoot, name, snapshotID, output)
}

func validateName(name string) (string, error) {
	if name == "" || filepath.Base(name) != name || strings.ContainsAny(name, `/\\`) {
		return "", fmt.Errorf("invalid Lexicon consumer name %q", name)
	}
	if filepath.Ext(name) != ".json" || strings.TrimSuffix(name, ".json") == "" {
		return "", fmt.Errorf("Lexicon consumer name %q must be a .json filename", name)
	}
	return name, nil
}

func validateDefinition(definition Definition) error {
	if definition.Version != Version {
		return fmt.Errorf("unsupported Lexicon consumer version %d", definition.Version)
	}
	if strings.TrimSpace(definition.Command) == "" {
		return fmt.Errorf("Lexicon consumer has no command")
	}
	if definition.Timeout < 0 {
		return fmt.Errorf("Lexicon consumer timeout must not be negative")
	}
	return nil
}
