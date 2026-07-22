package state

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRepositoryTracksOneReplaceableStateCommit(t *testing.T) {
	root := t.TempDir()
	repository, err := Ensure(root)
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(root, "source")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(source, "main.py")
	if err := os.WriteFile(path, []byte("value = 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := repository.StageAll(); err != nil {
		t.Fatal(err)
	}
	if err := repository.CommitState(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("value = 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := repository.StageSource(); err != nil {
		t.Fatal(err)
	}
	changes, err := repository.SourceChanges()
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 1 || changes[0].Status != "M" || changes[0].New != "main.py" {
		t.Fatalf("unexpected changes: %#v", changes)
	}
	if err := repository.StageAll(); err != nil {
		t.Fatal(err)
	}
	if err := repository.CommitState(); err != nil {
		t.Fatal(err)
	}
	command := exec.Command("git", "rev-list", "--count", "HEAD")
	command.Dir = root
	data, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(data)) != "1" {
		t.Fatalf("expected one reachable commit, got %q", data)
	}
}

func TestParseRename(t *testing.T) {
	data := []byte("R100\x00source/old.py\x00source/new.py\x00")
	changes := parseChanges(data)
	if len(changes) != 1 || changes[0].Old != "old.py" || changes[0].New != "new.py" {
		t.Fatalf("unexpected rename: %#v", changes)
	}
}
