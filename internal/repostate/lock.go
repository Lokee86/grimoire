package repostate

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type fileGuard struct{ path string }

func acquireFileLock(ctx context.Context, path string) (fileGuard, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fileGuard{}, err
	}
	for {
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err == nil {
			_, _ = fmt.Fprintf(file, "pid=%d\n", os.Getpid())
			_ = file.Close()
			return fileGuard{path: path}, nil
		}
		if !os.IsExist(err) {
			return fileGuard{}, err
		}
		if info, statErr := os.Stat(path); statErr == nil && time.Since(info.ModTime()) > 10*time.Minute {
			_ = os.Remove(path)
			continue
		}
		timer := time.NewTimer(25 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fileGuard{}, ctx.Err()
		case <-timer.C:
		}
	}
}

func (guard fileGuard) Close() error {
	if guard.path == "" {
		return nil
	}
	return os.Remove(guard.path)
}

func jsonMarshalIndent(value any) ([]byte, error) {
	return json.MarshalIndent(value, "", "  ")
}
