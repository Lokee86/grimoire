package embedding

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type rotatingLogWriter struct {
	mu      sync.Mutex
	path    string
	maxSize int64
	backups int
	file    *os.File
	size    int64
}

func newRotatingLogWriter(path string, maxSize int64, backups int) (*rotatingLogWriter, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	writer := &rotatingLogWriter{path: path, maxSize: maxSize, backups: backups}
	if err := writer.open(); err != nil {
		return nil, err
	}
	return writer, nil
}

func (writer *rotatingLogWriter) Write(data []byte) (int, error) {
	writer.mu.Lock()
	defer writer.mu.Unlock()
	if writer.file == nil {
		return 0, os.ErrClosed
	}
	if writer.size > 0 && writer.size+int64(len(data)) > writer.maxSize {
		if err := writer.rotate(); err != nil {
			return 0, err
		}
	}
	count, err := writer.file.Write(data)
	writer.size += int64(count)
	return count, err
}

func (writer *rotatingLogWriter) Close() error {
	writer.mu.Lock()
	defer writer.mu.Unlock()
	if writer.file == nil {
		return nil
	}
	err := writer.file.Close()
	writer.file = nil
	return err
}

func (writer *rotatingLogWriter) open() error {
	file, err := os.OpenFile(writer.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return err
	}
	writer.file = file
	writer.size = info.Size()
	return nil
}

func (writer *rotatingLogWriter) rotate() error {
	if err := writer.file.Close(); err != nil {
		return err
	}
	writer.file = nil
	if writer.backups == 0 {
		if err := os.Remove(writer.path); err != nil && !os.IsNotExist(err) {
			return err
		}
	} else {
		oldest := fmt.Sprintf("%s.%d", writer.path, writer.backups)
		_ = os.Remove(oldest)
		for index := writer.backups - 1; index >= 1; index-- {
			from := fmt.Sprintf("%s.%d", writer.path, index)
			to := fmt.Sprintf("%s.%d", writer.path, index+1)
			if err := os.Rename(from, to); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
		if err := os.Rename(writer.path, writer.path+".1"); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return writer.open()
}
