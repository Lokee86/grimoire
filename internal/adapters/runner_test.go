package adapters

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestCommandPrefersPackagedExecutables(t *testing.T) {
	for _, language := range []string{"go", "gdscript", "rust"} {
		t.Run(language, func(t *testing.T) {
			root := t.TempDir()
			executable := filepath.Join(root, language, "lexicon-"+language)
			if runtime.GOOS == "windows" {
				executable += ".exe"
			}
			if err := os.MkdirAll(filepath.Dir(executable), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(executable, []byte("packaged"), 0o755); err != nil {
				t.Fatal(err)
			}

			request := Request{Language: language, Repository: "repo", Output: "facts.jsonl", ChangedFiles: []string{"src/main.go"}}
			command, err := (Runner{Root: root}).command(context.Background(), request)
			if err != nil {
				t.Fatal(err)
			}
			if command.Path != executable {
				t.Fatalf("command path = %q, want packaged executable %q", command.Path, executable)
			}
			if command.Dir != "" {
				t.Fatalf("packaged command directory = %q, want empty", command.Dir)
			}
			if got := command.Args[1:]; !reflect.DeepEqual(got, []string{"--repo", "repo", "--output", "facts.jsonl", "--changed-file", "src/main.go"}) {
				t.Fatalf("packaged command arguments = %#v", got)
			}
		})
	}
}

func TestCommandUsesGenericAdapterForExtensionLanguage(t *testing.T) {
	root := t.TempDir()
	executable := filepath.Join(root, "generic", "lexicon-generic")
	if runtime.GOOS == "windows" {
		executable += ".exe"
	}
	if err := os.MkdirAll(filepath.Dir(executable), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(executable, []byte("packaged"), 0o755); err != nil {
		t.Fatal(err)
	}
	request := Request{Language: "generic-c", Repository: "repo", Output: "facts.jsonl", ChangedFiles: []string{"src/main.c"}}
	command, err := (Runner{Root: root}).command(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if command.Path != executable {
		t.Fatalf("command path = %q, want %q", command.Path, executable)
	}
	want := []string{"--repo", "repo", "--output", "facts.jsonl", "--language", "generic-c", "--changed-file", "src/main.c"}
	if got := command.Args[1:]; !reflect.DeepEqual(got, want) {
		t.Fatalf("generic arguments = %#v, want %#v", got, want)
	}
	if _, err := (Runner{Root: root}).command(context.Background(), Request{Language: "generic"}); err == nil {
		t.Fatal("bare generic language was accepted")
	}
}

func TestCommandPrefersPackagedTypeScriptEntrypoint(t *testing.T) {
	root := t.TempDir()
	entrypoint := filepath.Join(root, "typescript", "dist", "cli.js")
	if err := os.MkdirAll(filepath.Dir(entrypoint), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(entrypoint, []byte("packaged"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := exec.LookPath("node"); err != nil {
		t.Skipf("node is unavailable for packaged entrypoint test: %v", err)
	}
	request := Request{Language: "typescript", Repository: "repo", Output: "facts.jsonl"}
	command, err := (Runner{Root: root}).command(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if len(command.Args) < 2 || command.Args[1] != entrypoint {
		t.Fatalf("TypeScript command arguments = %#v, want entrypoint %q", command.Args, entrypoint)
	}
}
