package investigation

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func acquireLock(sessionDir string) (func(), error) {
	lockPath := filepath.Join(sessionDir, ".lock")
	deadline := time.Now().Add(10 * time.Second)
	for {
		err := os.Mkdir(lockPath, 0o700)
		if err == nil {
			return func() { _ = os.Remove(lockPath) }, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("acquire investigation lock: %w", err)
		}
		if time.Now().After(deadline) {
			return nil, ErrSessionLocked
		}
		time.Sleep(10 * time.Millisecond)
	}
}
