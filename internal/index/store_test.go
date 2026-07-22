package index

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/Lokee86/grimoire/internal/tokenizer"
	git "github.com/go-git/go-git/v5"
)

func TestStoreRoundTrip(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "alpha.go", "package alpha\n\nfunc Value() int { return 1 }\n")

	snapshot, _, err := Build(root, nil, BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	state := filepath.Join(root, ".grimoire")
	if err := Save(state, snapshot); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(state)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Tokenizer != tokenizer.Name {
		t.Fatalf("unexpected tokenizer %q", loaded.Tokenizer)
	}
	if loaded.Identity() == "" || loaded.Identity() != stateHash(t, state) {
		t.Fatalf("unexpected prepared identity %q", loaded.Identity())
	}
	if len(loaded.Files) != 1 || loaded.Files[0].Path != "alpha.go" {
		t.Fatalf("unexpected files: %+v", loaded.Files)
	}
	if len(loaded.Files[0].Chunks) != 1 || loaded.Files[0].Chunks[0].Text == "" {
		t.Fatalf("unexpected chunks: %+v", loaded.Files[0].Chunks)
	}
}

func TestStoreKeepsRootForUnchangedIndex(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "alpha.go", "package alpha\n")
	state := filepath.Join(root, ".grimoire")

	first, _, err := Build(root, nil, BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err := Save(state, first); err != nil {
		t.Fatal(err)
	}
	before := stateHash(t, state)

	loaded, err := Load(state)
	if err != nil {
		t.Fatal(err)
	}
	second, stats, err := Build(root, &loaded, BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if stats.Reused != 1 || stats.Updated != 0 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
	if err := Save(state, second); err != nil {
		t.Fatal(err)
	}
	if after := stateHash(t, state); after != before {
		t.Fatalf("unchanged index moved root: %s -> %s", before, after)
	}
}

func TestStoreRewritesOnlyChangedShard(t *testing.T) {
	root := t.TempDir()
	firstPath, secondPath := pathsInDifferentShards()
	writeTestFile(t, root, firstPath, "package first\n")
	writeTestFile(t, root, secondPath, "package second\n")
	state := filepath.Join(root, ".grimoire")

	initial, _, err := Build(root, nil, BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err := Save(state, initial); err != nil {
		t.Fatal(err)
	}
	before, err := Load(state)
	if err != nil {
		t.Fatal(err)
	}
	unchangedShard := shardName(secondPath)
	unchangedHash := before.baseShards[unchangedShard]

	writeTestFile(t, root, firstPath, "package first\n\nfunc Changed() {}\n")
	updated, _, err := Build(root, &before, BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err := Save(state, updated); err != nil {
		t.Fatal(err)
	}
	after, err := Load(state)
	if err != nil {
		t.Fatal(err)
	}
	if after.baseShards[unchangedShard] != unchangedHash {
		t.Fatal("unrelated shard was rewritten")
	}
	if after.baseShards[shardName(firstPath)] == before.baseShards[shardName(firstPath)] {
		t.Fatal("changed shard was not replaced")
	}
}

func TestStoreRejectsStaleSnapshot(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "alpha.go", "package alpha\n")
	state := filepath.Join(root, ".grimoire")
	initial, _, err := Build(root, nil, BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err := Save(state, initial); err != nil {
		t.Fatal(err)
	}
	fresh, err := Load(state)
	if err != nil {
		t.Fatal(err)
	}
	stale, err := Load(state)
	if err != nil {
		t.Fatal(err)
	}

	writeTestFile(t, root, "alpha.go", "package alpha\n\nfunc New() {}\n")
	updated, _, err := Build(root, &fresh, BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err := Save(state, updated); err != nil {
		t.Fatal(err)
	}
	staleUpdate, _, err := Build(root, &stale, BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err := Save(state, staleUpdate); !errors.Is(err, ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func writeTestFile(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func pathsInDifferentShards() (string, string) {
	first := "file-0.go"
	for index := 1; ; index++ {
		candidate := "file-" + strconv.Itoa(index) + ".go"
		if shardName(candidate) != shardName(first) {
			return first, candidate
		}
	}
}

func stateHash(t *testing.T, state string) string {
	t.Helper()
	repository, err := git.PlainOpen(state)
	if err != nil {
		t.Fatal(err)
	}
	ref, err := repository.Storer.Reference(stateReference)
	if err != nil {
		t.Fatal(err)
	}
	return ref.Hash().String()
}
