package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/lexicon/internal/config"
	"github.com/Lokee86/lexicon/internal/objectstore"
)

func TestStatusDiscoversRepositoryAndReportsSnapshot(t *testing.T) {
	repository := t.TempDir()
	if err := config.Save(repository, repository); err != nil {
		t.Fatal(err)
	}
	store := objectstore.Store{Root: config.StateRoot(repository)}
	snapshotID, err := store.Publish(objectstore.Manifest{
		StateCommit: "state-commit",
		Languages: []objectstore.LanguageEntry{
			{Language: "python"},
			{Language: "go"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	consumerRoot := filepath.Join(config.StateRoot(repository), "consumers")
	if err := os.MkdirAll(consumerRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"zeta.json", "alpha.json", "ignore.txt"} {
		if err := os.WriteFile(filepath.Join(consumerRoot, name), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	nested := filepath.Join(repository, "nested", "command")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(nested)

	var stdout, stderr bytes.Buffer
	if code := Run([]string{"status"}, &stdout, &stderr); code != 0 {
		t.Fatalf("status exit code = %d, stderr = %s", code, stderr.String())
	}
	output := stdout.String()
	for _, want := range []string{
		"repository root: " + repository,
		"current snapshot ID: " + snapshotID,
		"detected languages: go, python",
		"registered consumer names: alpha, zeta",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("status output %q does not contain %q", output, want)
		}
	}
}

func TestVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"version"}, &stdout, &stderr); code != 0 {
		t.Fatalf("version exit code = %d, stderr = %s", code, stderr.String())
	}
	if got, want := stdout.String(), "lexicon version dev\n"; got != want {
		t.Fatalf("version output = %q, want %q", got, want)
	}
}
