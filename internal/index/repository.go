package index

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage"
)

func openOrInit(path string) (*git.Repository, error) {
	path = filepath.Clean(path)
	repository, err := git.PlainOpen(path)
	if err == nil {
		return repository, nil
	}
	if !errors.Is(err, git.ErrRepositoryNotExists) {
		return nil, fmt.Errorf("open prepared index: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create prepared index parent: %w", err)
	}
	repository, err = git.PlainInit(path, true)
	if err != nil {
		return nil, fmt.Errorf("initialize prepared index: %w", err)
	}
	return repository, nil
}

func currentReference(store storage.Storer) (*plumbing.Reference, error) {
	ref, err := store.Reference(stateReference)
	if errors.Is(err, plumbing.ErrReferenceNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read prepared index state: %w", err)
	}
	return ref, nil
}

func matchesBase(current *plumbing.Reference, base string) bool {
	if base == "" {
		return current == nil
	}
	return current != nil && current.Hash() == plumbing.NewHash(base)
}

func finishSave(path string) error {
	legacy := filepath.Join(filepath.Clean(path), "index.json")
	if err := os.Remove(legacy); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove legacy JSON index: %w", err)
	}
	return nil
}
