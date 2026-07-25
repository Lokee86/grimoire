package diffcontext

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type recordedRun struct {
	name string
	args []string
}

type fakeRunner struct {
	diffOutput      []byte
	diffErr         error
	untrackedOutput []byte
	untrackedErr    error
	runs            []recordedRun
}

func (runner *fakeRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	runner.runs = append(runner.runs, recordedRun{name: name, args: append([]string(nil), args...)})
	if contains(args, "ls-files") {
		return runner.untrackedOutput, runner.untrackedErr
	}
	return runner.diffOutput, runner.diffErr
}

func TestCollectWorkingTreeUsesBoundedGitDiff(t *testing.T) {
	runner := &fakeRunner{diffOutput: []byte("diff --git a/a.go b/a.go\n--- a/a.go\n+++ b/a.go\n@@ -1 +1 @@ func A()\n")}
	changes, err := CollectWithRunner(context.Background(), ".", WorkingTree, runner)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 1 || changes[0].Path != "a.go" {
		t.Fatalf("unexpected changes: %#v", changes)
	}
	if len(runner.runs) != 2 || runner.runs[0].name != "git" || !contains(runner.runs[1].args, "ls-files") {
		t.Fatalf("unexpected command sequence: %#v", runner.runs)
	}
	joined := strings.Join(runner.runs[0].args, " ")
	for _, required := range []string{" diff ", " --no-ext-diff ", " --no-textconv ", " --find-renames ", " --unified=0 ", " HEAD "} {
		if !strings.Contains(" "+joined+" ", required) {
			t.Fatalf("git args %q do not contain %q", joined, required)
		}
	}
	if runner.runs[0].args[len(runner.runs[0].args)-1] != "--" {
		t.Fatalf("git args are not pathspec-terminated: %#v", runner.runs[0].args)
	}
}

func TestCollectWorkingTreeIncludesUntrackedFiles(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "internal"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "internal", "new.go"), []byte("one\ntwo\nthree\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runner := &fakeRunner{untrackedOutput: []byte("internal/new.go\x00")}
	changes, err := CollectWithRunner(context.Background(), root, WorkingTree, runner)
	if err != nil {
		t.Fatal(err)
	}
	want := []Change{{Path: "internal/new.go", StartLine: 1, EndLine: 3, Summary: "untracked file"}}
	if !reflect.DeepEqual(changes, want) {
		t.Fatalf("changes = %#v, want %#v", changes, want)
	}
}

func TestGitDiffArgsNamedAndRevisionScopes(t *testing.T) {
	staged, err := gitDiffArgs("repo", Staged)
	if err != nil {
		t.Fatal(err)
	}
	unstaged, err := gitDiffArgs("repo", Unstaged)
	if err != nil {
		t.Fatal(err)
	}
	revision, err := gitDiffArgs("repo", "main...HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if !contains(staged, "--cached") || contains(unstaged, "HEAD") || !contains(revision, "main...HEAD") {
		t.Fatalf("unexpected args:\nstaged %#v\nunstaged %#v\nrevision %#v", staged, unstaged, revision)
	}
	if _, err := gitDiffArgs("repo", "--output=/tmp/file"); err == nil {
		t.Fatal("expected leading-option revision to be rejected")
	}
}

func TestCollectReportsGitFailure(t *testing.T) {
	runner := &fakeRunner{diffOutput: []byte("fatal: not a git repository"), diffErr: errors.New("exit 128")}
	_, err := CollectWithRunner(context.Background(), ".", Staged, runner)
	if err == nil || !strings.Contains(err.Error(), "not a git repository") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCollectRevisionIsOneArgument(t *testing.T) {
	runner := &fakeRunner{}
	_, err := CollectWithRunner(context.Background(), ".", "HEAD~2 --output=bad", runner)
	if err != nil {
		t.Fatal(err)
	}
	wantTail := []string{"HEAD~2 --output=bad", "--"}
	gotTail := runner.runs[0].args[len(runner.runs[0].args)-2:]
	if !reflect.DeepEqual(gotTail, wantTail) {
		t.Fatalf("revision was split into multiple arguments: %#v", runner.runs[0].args)
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
