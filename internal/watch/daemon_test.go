package watch

import (
	"os"
	"path/filepath"
	"testing"

	lexfiles "github.com/Lokee86/lexicon/internal/files"
)

func TestDaemonIgnoreFilterAppliesRepositoryPolicy(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, lexfiles.IgnoreFileName), []byte("ignored/\n*.py\n!keep.py\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	policy, err := lexfiles.LoadIgnorePolicy(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "ignored"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, relative := range []string{"ignored/code.py", "blocked.py", "keep.py", "main.go"} {
		path := filepath.Join(root, filepath.FromSlash(relative))
		if err := os.WriteFile(path, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	for relative, expected := range map[string]bool{
		"ignored":         true,
		"ignored/code.py": true,
		"blocked.py":      true,
		"keep.py":         false,
		"main.go":         false,
		".git/config":     true,
		".lexiconignore":  false,
	} {
		if actual := ignored(policy, filepath.Join(root, filepath.FromSlash(relative))); actual != expected {
			t.Errorf("ignored(%q) = %t, want %t", relative, actual, expected)
		}
	}
}

func TestDaemonFilterUsesLoadedPolicyUntilReload(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "main.go")
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, lexfiles.IgnoreFileName), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	policy, err := lexfiles.LoadIgnorePolicy(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, lexfiles.IgnoreFileName), []byte("main.go\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if ignored(policy, path) {
		t.Fatal("event filter re-read the changed ignore file")
	}

	reloaded, err := lexfiles.LoadIgnorePolicy(root)
	if err != nil {
		t.Fatal(err)
	}
	if !ignored(reloaded, path) {
		t.Fatal("reloaded ignore policy did not exclude main.go")
	}
}
