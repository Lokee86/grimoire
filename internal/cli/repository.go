package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Lokee86/lexicon/internal/config"
)

var errRepositoryNotFound = errors.New("Lexicon repository not found")

func resolveRepository(explicit string) (string, error) {
	if explicit != "" {
		return absolute(explicit)
	}
	current, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return discoverRepository(current)
}

func initRepository(explicit string) (string, error) {
	if explicit != "" {
		return absolute(explicit)
	}
	current, err := os.Getwd()
	if err != nil {
		return "", err
	}
	root, err := discoverRepository(current)
	if errors.Is(err, errRepositoryNotFound) {
		return absolute(current)
	}
	return root, err
}

func discoverRepository(start string) (string, error) {
	root, err := absolute(start)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(config.Path(root)); err == nil {
			return root, nil
		} else if !os.IsNotExist(err) {
			return "", fmt.Errorf("find Lexicon configuration: %w", err)
		}
		parent := filepath.Dir(root)
		if parent == root {
			break
		}
		root = parent
	}
	return "", fmt.Errorf("%w from %s; use --repo or run lexicon init", errRepositoryNotFound, start)
}

func absolute(path string) (string, error) {
	value, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(value)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("repository is not a directory: %s", value)
	}
	return value, nil
}
