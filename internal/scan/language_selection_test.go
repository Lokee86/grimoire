package scan

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Lokee86/lexicon/internal/config"
	"github.com/Lokee86/lexicon/internal/state"
)

func TestSelectedLanguagesDefaultsToAllAndFiltersSubset(t *testing.T) {
	languages := []string{"python", "ruby", "typescript"}
	if got := selectedLanguages(languages, func(string) bool { return true }); !reflect.DeepEqual(got, languages) {
		t.Fatalf("default selection = %v", got)
	}
	if got := selectedLanguages(languages, func(language string) bool { return language == "python" }); !reflect.DeepEqual(got, []string{"python"}) {
		t.Fatalf("selected subset = %v", got)
	}
}

func TestOpenDisablesPreviouslyAnalyzedLanguage(t *testing.T) {
	repository := t.TempDir()
	adapterRoot := t.TempDir()
	for _, language := range []string{"python", "ruby"} {
		path := filepath.Join(adapterRoot, language, "adapter.txt")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(language+" adapter\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for name, contents := range map[string]string{"main.py": "value = 1\n", "main.rb": "value = 1\n"} {
		if err := os.WriteFile(filepath.Join(repository, name), []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := config.SaveWithEnabledLanguages(repository, adapterRoot, nil); err != nil {
		t.Fatal(err)
	}
	stateRoot := filepath.Join(config.StateRoot(repository), "repo")
	gitRepository, err := state.Ensure(stateRoot)
	if err != nil {
		t.Fatal(err)
	}
	initialAnalyzer := &fakeAnalyzer{}
	initial := New(repository, stateRoot, gitRepository, initialAnalyzer, io.Discard)
	initial.AdapterRoot = adapterRoot
	prepareSnapshotWithLanguages(t, initial, gitRepository, []string{"python", "ruby"})

	if err := config.UpdateEnabledLanguages(repository, []string{"python"}); err != nil {
		t.Fatal(err)
	}
	opened, err := Open(repository, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	analyzer := &fakeAnalyzer{}
	opened.Analyzer = analyzer
	if _, err := opened.Scan(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(stateRoot, "library", "python.jsonl")); err != nil {
		t.Fatalf("enabled library: %v", err)
	}
	if _, err := os.Stat(filepath.Join(stateRoot, "library", "ruby.jsonl")); !os.IsNotExist(err) {
		t.Fatalf("disabled library retained: %v", err)
	}
	_, manifest, err := opened.Store.Current()
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Languages) != 1 || manifest.Languages[0].Language != "python" {
		t.Fatalf("snapshot languages = %#v", manifest.Languages)
	}
	if len(analyzer.languages) != 0 {
		t.Fatalf("disabled language was rebuilt: %v", analyzer.languages)
	}
}

func prepareSnapshotWithLanguages(t *testing.T, scanner *Scanner, repository *state.Repository, languages []string) {
	t.Helper()
	if err := scanner.Mirror.SyncAll(scanner.Repository); err != nil {
		t.Fatal(err)
	}
	if err := scanner.analyzeFull(context.Background(), languages); err != nil {
		t.Fatal(err)
	}
	if err := repository.StageAll(); err != nil {
		t.Fatal(err)
	}
	if err := repository.CommitState(); err != nil {
		t.Fatal(err)
	}
	if _, err := scanner.publishSnapshot(); err != nil {
		t.Fatal(err)
	}
}
