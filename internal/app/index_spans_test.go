package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/grimoire/internal/index"
)

func TestRunIndexUsesExplicitLexiconSourceSpans(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, ".prepared")
	facts := filepath.Join(root, "facts")
	if err := os.MkdirAll(facts, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(root, "sample.go"),
		[]byte("package sample\n\nfunc Value() int {\n\treturn 1\n}\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	record := `{"record":"node","id":"value","kind":"function","name":"Value","path":"sample.go","qualified_name":"sample.Value","span":{"path":"sample.go","start_line":3,"end_line":5}}` + "\n"
	if err := os.WriteFile(filepath.Join(facts, "facts.jsonl"), []byte(record), 0o644); err != nil {
		t.Fatal(err)
	}

	var output, errors bytes.Buffer
	if err := Run([]string{
		"index", "--root", root, "--state", state, "--lexicon-facts", facts,
	}, &output, &errors); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), `"chunking": "lexicon"`) {
		t.Fatalf("index response did not report Lexicon chunking: %s", output.String())
	}

	snapshot, err := index.Load(state)
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, chunk := range snapshot.Files[0].Chunks {
		if chunk.SemanticName == "Value" {
			found = true
			if chunk.StartLine != 3 || chunk.EndLine != 5 || chunk.SemanticKind != "function" {
				t.Fatalf("unexpected semantic chunk: %+v", chunk)
			}
		}
	}
	if !found {
		t.Fatalf("prepared index omitted Lexicon semantic chunk: %+v", snapshot.Files[0].Chunks)
	}
}
