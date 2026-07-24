package lock

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofrs/flock"
)

var ErrBusy = errors.New("Lexicon repository is already being updated")

type Lock struct {
	file *flock.Flock
}

func Acquire(root string) (*Lock, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("create Lexicon state directory: %w", err)
	}
	file := flock.New(filepath.Join(root, "LOCK"))
	locked, err := file.TryLock()
	if err != nil {
		return nil, fmt.Errorf("acquire Lexicon update lock: %w", err)
	}
	if !locked {
		return nil, ErrBusy
	}
	return &Lock{file: file}, nil
}

func (l *Lock) Close() error {
	if l == nil || l.file == nil {
		return nil
	}
	if err := l.file.Unlock(); err != nil {
		return fmt.Errorf("release Lexicon update lock: %w", err)
	}
	return nil
}
