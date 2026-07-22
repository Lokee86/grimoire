//go:build !windows

package consumer

import "os"

func replaceAtomic(source, destination string) error {
	return os.Rename(source, destination)
}
