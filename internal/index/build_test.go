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

func TestBuildExcludesWorktrees(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".worktrees", "branch"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".worktrees", "branch", "hidden.go"), []byte("package hidden"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "visible.go"), []byte("package visible"), 0o644); err != nil {
		t.Fatal(err)
	}

	snapshot, _, err := Build(root, nil, BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Files) != 1 || snapshot.Files[0].Path != "visible.go" {
		t.Fatalf("unexpected files: %+v", snapshot.Files)
	}
}
