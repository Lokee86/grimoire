package repostate

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGitRepositoryStatusUsesOneCleanWorkingTreeSignal(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGitTest(t, root, "init")
	runGitTest(t, root, "config", "user.email", "test@example.com")
	runGitTest(t, root, "config", "user.name", "Test")
	runGitTest(t, root, "add", "main.go")
	runGitTest(t, root, "commit", "-m", "initial")

	head, dirty, available := gitRepositoryStatus(context.Background(), root)
	if !available || dirty || head == "" {
		t.Fatalf("clean status = head %q dirty %v available %v", head, dirty, available)
	}

	if err := os.MkdirAll(filepath.Join(root, ".grimoire"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".grimoire", "generated.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, dirty, _ = gitRepositoryStatus(context.Background(), root)
	if dirty {
		t.Fatal("generated Grimoire state made repository dirty")
	}

	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, dirty, _ = gitRepositoryStatus(context.Background(), root)
	if !dirty {
		t.Fatal("tracked source edit was not detected")
	}
}

func TestGitRepositoryStatusDetectsUntrackedSource(t *testing.T) {
	root := t.TempDir()
	runGitTest(t, root, "init")
	if err := os.WriteFile(filepath.Join(root, "new.go"), []byte("package new\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, dirty, available := gitRepositoryStatus(context.Background(), root)
	if !available || !dirty {
		t.Fatalf("untracked status = dirty %v available %v", dirty, available)
	}
}

func runGitTest(t *testing.T, root string, args ...string) {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", root}, args...)...)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, output)
	}
}
