package main

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestRunWritesStdout(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{"main.c": "int main(void) { return 0; }\n"})
	var output bytes.Buffer
	if err := run([]string{"--repo", root, "--output", "-"}, &output); err != nil {
		t.Fatal(err)
	}
	if output.Len() == 0 {
		t.Fatal("stdout output is empty")
	}
}

func TestRunWritesFile(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{"main.cpp": "int main() { return 0; }\n"})
	output := filepath.Join(t.TempDir(), "facts.jsonl")
	if err := run([]string{"--repo", root, "--output", output}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if info, err := filepath.Glob(output); err != nil || len(info) != 1 {
		t.Fatalf("output file missing: %v, %v", info, err)
	}
}
