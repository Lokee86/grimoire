package agentruntime

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestAdjacentCommandPrefersPackagedProvider(t *testing.T) {
	directory := t.TempDir()
	provider := "lexicon"
	if runtime.GOOS == "windows" {
		provider += ".exe"
	}
	providerPath := filepath.Join(directory, provider)
	if err := os.WriteFile(providerPath, []byte("provider"), 0o755); err != nil {
		t.Fatal(err)
	}
	got := adjacentCommandFrom(filepath.Join(directory, "grimoire.exe"), "lexicon")
	if got != providerPath {
		t.Fatalf("adjacent provider = %q, want %q", got, providerPath)
	}
}

func TestAdjacentCommandFallsBackToPathName(t *testing.T) {
	got := adjacentCommandFrom(filepath.Join(t.TempDir(), "grimoire.exe"), "arcana")
	if got != "arcana" {
		t.Fatalf("fallback provider = %q", got)
	}
}
