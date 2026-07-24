package embedding

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRotatingLogWriterRotatesAtConfiguredLimit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "embedding.log")
	writer, err := newRotatingLogWriter(path, 8, 2)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write([]byte("12345678")); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write([]byte("abc")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	current, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	previous, err := os.ReadFile(path + ".1")
	if err != nil {
		t.Fatal(err)
	}
	if string(current) != "abc" || string(previous) != "12345678" {
		t.Fatalf("unexpected rotation: current=%q previous=%q", current, previous)
	}
}
