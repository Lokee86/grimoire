//go:build !windows

package objectstore

import (
	"os"
	"path/filepath"
)

func syncParent(path string) error {
	directory, err := os.Open(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}
