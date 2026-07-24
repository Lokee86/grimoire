package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMirrorSyncAllAppliesLexiconIgnore(t *testing.T) {
	source := t.TempDir()
	mirrorRoot := t.TempDir()
	for relative, data := range map[string]string{
		"main.py":           "main = 1\n",
		"keep.py":           "keep = 1\n",
		"ignored/nested.py": "ignored = 1\n",
		"vendor/vendor.py":  "vendor = 1\n",
	} {
		path := filepath.Join(source, filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(source, ".lexiconignore"), []byte("*.py\n!keep.py\nignored/\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := (Mirror{Root: mirrorRoot}).SyncAll(source); err != nil {
		t.Fatal(err)
	}
	assertMirrorFile(t, mirrorRoot, "keep.py", true)
	assertMirrorFile(t, mirrorRoot, "main.py", false)
	assertMirrorFile(t, mirrorRoot, "ignored/nested.py", false)
	assertMirrorFile(t, mirrorRoot, "vendor/vendor.py", false)
}

func TestMirrorSyncPathsAppliesLexiconIgnore(t *testing.T) {
	source := t.TempDir()
	mirrorRoot := t.TempDir()
	for relative, data := range map[string]string{
		"main.py":          "main = 1\n",
		"blocked.py":       "blocked = 1\n",
		"vendor/vendor.py": "vendor = 1\n",
	} {
		path := filepath.Join(source, filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(source, ".lexiconignore"), []byte("blocked.py\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := (Mirror{Root: mirrorRoot}).SyncPaths(source, []string{"main.py", "blocked.py", "vendor/vendor.py"}); err != nil {
		t.Fatal(err)
	}
	assertMirrorFile(t, mirrorRoot, "main.py", true)
	assertMirrorFile(t, mirrorRoot, "blocked.py", false)
	assertMirrorFile(t, mirrorRoot, "vendor/vendor.py", false)
}

func assertMirrorFile(t *testing.T, root, relative string, expected bool) {
	t.Helper()
	_, err := os.Stat(filepath.Join(root, filepath.FromSlash(relative)))
	actual := err == nil
	if actual != expected {
		t.Fatalf("mirror file %q present = %t, want %t (err: %v)", relative, actual, expected, err)
	}
}
