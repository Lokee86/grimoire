package investigation

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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
		if !lockContention(err) {
			return nil, fmt.Errorf("acquire investigation lock: %w", err)
		}
		if time.Now().After(deadline) {
			return nil, ErrSessionLocked
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func lockContention(err error) bool {
	if errors.Is(err, os.ErrExist) {
		return true
	}
	// Windows can report ERROR_ACCESS_DENIED while another writer removes or
	// recreates the lock directory. Treat that transient directory race as
	// contention and keep using the existing bounded retry deadline.
	return runtime.GOOS == "windows" && errors.Is(err, os.ErrPermission)
}
