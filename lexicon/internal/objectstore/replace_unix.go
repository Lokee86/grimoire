//go:build !windows

package objectstore

import "os"

func replaceAtomic(source, destination string) error {
	return os.Rename(source, destination)
}
