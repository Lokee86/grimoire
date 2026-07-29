package repostate

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
			_, _ = fmt.Fprintf(file, "pid=%d\ncreated=%s\n", os.Getpid(), time.Now().UTC().Format(time.RFC3339Nano))
			_ = file.Close()
			return fileGuard{path: path}, nil
		}
		if !os.IsExist(err) {
			return fileGuard{}, err
		}
		if staleRefreshLock(path) {
			if removeErr := os.Remove(path); removeErr == nil || os.IsNotExist(removeErr) {
				continue
			}
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

func staleRefreshLock(path string) bool {
	pid, ok := refreshLockPID(path)
	if ok {
		return !processAlive(pid)
	}
	info, err := os.Stat(path)
	return err == nil && time.Since(info.ModTime()) > 10*time.Minute
}

func refreshLockPID(path string) (int, bool) {
	file, err := os.Open(path)
	if err != nil {
		return 0, false
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		value := strings.TrimSpace(scanner.Text())
		if raw, ok := strings.CutPrefix(value, "pid="); ok {
			pid, err := strconv.Atoi(raw)
			return pid, err == nil && pid > 0
		}
	}
	return 0, false
}

func (guard fileGuard) Close() error {
	if guard.path == "" {
		return nil
	}
	err := os.Remove(guard.path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func jsonMarshalIndent(value any) ([]byte, error) {
	return json.MarshalIndent(value, "", "  ")
}
