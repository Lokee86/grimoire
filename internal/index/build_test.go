package index

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildReusesUnchangedFiles(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "sample.go")
	if err := os.WriteFile(path, []byte("package sample\n\nfunc Value() int { return 1 }\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	first, firstStats, err := Build(root, nil, BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if firstStats.Updated != 1 || firstStats.Reused != 0 {
		t.Fatalf("unexpected first stats: %+v", firstStats)
	}
	if len(first.Files) != 1 || len(first.Files[0].Chunks) != 1 {
		t.Fatalf("unexpected first index: %+v", first)
	}

	second, secondStats, err := Build(root, &first, BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if secondStats.Reused != 1 || secondStats.Updated != 0 {
		t.Fatalf("unexpected second stats: %+v", secondStats)
	}
	if second.Files[0].Chunks[0].ID != first.Files[0].Chunks[0].ID {
		t.Fatal("unchanged chunk identity was not reused")
	}

	if err := os.WriteFile(path, []byte("package sample\n\nfunc Value() int { return 2 }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	third, thirdStats, err := Build(root, &second, BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if thirdStats.Updated != 1 || thirdStats.Reused != 0 {
		t.Fatalf("unexpected third stats: %+v", thirdStats)
	}
	if third.Files[0].Chunks[0].ID == second.Files[0].Chunks[0].ID {
		t.Fatal("changed chunk identity was reused")
	}
}

func TestBuildExcludesToolStateDirectories(t *testing.T) {
	root := t.TempDir()
	for _, path := range []string{
		".worktrees/branch/hidden.go",
		".workingtrees/branch/hidden.go",
		".grimoire/objects/hidden.go",
		".ddocs/state/hidden.go",
		".arcana/index/hidden.go",
		".warlock/runtime/hidden.go",
	} {
		writeBuildFile(t, root, path, "package hidden")
	}
	writeBuildFile(t, root, "visible.go", "package visible")

	snapshot, _, err := Build(root, nil, BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertBuildPaths(t, snapshot, "visible.go")
}

func TestBuildUsesGitignoreHierarchy(t *testing.T) {
	root := t.TempDir()
	writeBuildFile(t, root, ".gitignore", "*.go\n!visible.go\nnested-ignored/\n")
	writeBuildFile(t, root, "ignored.go", "package ignored")
	writeBuildFile(t, root, "visible.go", "package visible")
	writeBuildFile(t, root, "nested-ignored/hidden.go", "package hidden")
	writeBuildFile(t, root, "nested/.gitignore", "hidden.go\n")
	writeBuildFile(t, root, "nested/hidden.go", "package hidden")
	writeBuildFile(t, root, "nested/visible.md", "# Visible")

	snapshot, _, err := Build(root, nil, BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertBuildPaths(t, snapshot, "nested/visible.md", "visible.go")
}

func TestBuildUsesConfiguredIgnoreFileInstead(t *testing.T) {
	root := t.TempDir()
	writeBuildFile(t, root, ".gitignore", "default.go\n")
	writeBuildFile(t, root, ".contextignore", "custom.go\n")
	writeBuildFile(t, root, "default.go", "package defaultpkg")
	writeBuildFile(t, root, "custom.go", "package custom")
	writeBuildFile(t, root, "visible.go", "package visible")

	snapshot, _, err := Build(root, nil, BuildOptions{IgnoreFile: ".contextignore"})
	if err != nil {
		t.Fatal(err)
	}
	assertBuildPaths(t, snapshot, "default.go", "visible.go")
}

func TestBuildDoesNotHardCodeVendor(t *testing.T) {
	root := t.TempDir()
	writeBuildFile(t, root, "vendor/dependency.go", "package dependency")

	snapshot, _, err := Build(root, nil, BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertBuildPaths(t, snapshot, "vendor/dependency.go")
}

func TestBuildExcludesConfiguredStatePath(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, ".cache", "grimoire")
	writeBuildFile(t, root, ".cache/grimoire/objects/generated.go", "package generated")
	writeBuildFile(t, root, "visible.go", "package visible")

	snapshot, _, err := Build(root, nil, BuildOptions{ExcludePaths: []string{state}})
	if err != nil {
		t.Fatal(err)
	}
	assertBuildPaths(t, snapshot, "visible.go")
}

func writeBuildFile(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertBuildPaths(t *testing.T, snapshot Snapshot, expected ...string) {
	t.Helper()
	if len(snapshot.Files) != len(expected) {
		t.Fatalf("expected paths %v, got %+v", expected, snapshot.Files)
	}
	for index, path := range expected {
		if snapshot.Files[index].Path != path {
			t.Fatalf("expected paths %v, got %+v", expected, snapshot.Files)
		}
	}
}
