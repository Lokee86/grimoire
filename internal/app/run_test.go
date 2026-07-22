package app

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Lokee86/grimoire/internal/compiler"
)

func TestIndexThenCompileContext(t *testing.T) {
	root := t.TempDir()
	content := "package damage\n\nfunc ResolveDamage() int { return 10 }\n"
	if err := os.WriteFile(filepath.Join(root, "damage.go"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	var indexOutput bytes.Buffer
	if err := Run([]string{"index", "--root", root}, &indexOutput, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".grimoire", "index.json")); err != nil {
		t.Fatal(err)
	}

	var contextOutput bytes.Buffer
	if err := Run([]string{
		"context", "--root", root,
		"--query", "resolve damage",
		"--budget", "100",
	}, &contextOutput, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	var result compiler.Package
	if err := json.Unmarshal(contextOutput.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Selections) != 1 {
		t.Fatalf("expected one selection, got %+v", result.Selections)
	}
	if result.Selections[0].Path != "damage.go" {
		t.Fatalf("unexpected selection: %+v", result.Selections[0])
	}
}
