package objectstore

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

func writeImmutable(path string, data []byte) error {
	if existing, err := os.ReadFile(path); err == nil {
		if bytes.Equal(existing, data) {
			return nil
		}
		return fmt.Errorf("content-addressed object collision at %s", path)
	} else if !os.IsNotExist(err) {
		return err
	}
	temporary, err := writeTemporary(path, data)
	if err != nil {
		return err
	}
	if err := os.Rename(temporary, path); err != nil {
		if existing, readErr := os.ReadFile(path); readErr == nil && bytes.Equal(existing, data) {
			_ = os.Remove(temporary)
			return nil
		}
		_ = os.Remove(temporary)
		return err
	}
	return syncParent(path)
}

func writeAtomic(path string, data []byte) error {
	temporary, err := writeTemporary(path, data)
	if err != nil {
		return err
	}
	if err := replaceAtomic(temporary, path); err != nil {
		_ = os.Remove(temporary)
		return err
	}
	return syncParent(path)
}

func writeTemporary(destination string, data []byte) (string, error) {
	directory := filepath.Dir(destination)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return "", err
	}
	file, err := os.CreateTemp(directory, ".lexicon-tmp-*")
	if err != nil {
		return "", err
	}
	path := file.Name()
	cleanup := func() {
		_ = file.Close()
		_ = os.Remove(path)
	}
	if err := file.Chmod(0o644); err != nil {
		cleanup()
		return "", err
	}
	if _, err := file.Write(data); err != nil {
		cleanup()
		return "", err
	}
	if err := file.Sync(); err != nil {
		cleanup()
		return "", err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return "", err
	}
	return path, nil
}
