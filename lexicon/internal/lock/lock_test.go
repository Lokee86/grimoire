package lock

import (
	"errors"
	"testing"
)

func TestSingleWriter(t *testing.T) {
	root := t.TempDir()
	first, err := Acquire(root)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	if _, err := Acquire(root); !errors.Is(err, ErrBusy) {
		t.Fatalf("second lock error = %v", err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := Acquire(root)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
}
