package agentruntime

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestAdjacentCommandPrefersPackagedProvider(t *testing.T) {
	directory := t.TempDir()
	providerPath := writeProvider(t, directory, "lexicon")
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

func TestResolveProviderCommandPreservesExplicitCommand(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	got := resolveProviderCommand(t.TempDir(), "explicit-provider", "lexicon")
	if got != "explicit-provider" {
		t.Fatalf("explicit provider = %q", got)
	}
}

func TestResolveProviderCommandUsesRepositoryConfigurationWithoutPATH(t *testing.T) {
	root := t.TempDir()
	providerPath := writeProvider(t, t.TempDir(), "lexicon")
	writeProviderConfiguration(t, root, providerConfiguration{
		Version: providerConfigVersion, LexiconCommand: providerPath,
	})
	t.Setenv("PATH", t.TempDir())

	got := resolveProviderCommand(root, "", "lexicon")
	if got != providerPath {
		t.Fatalf("configured provider = %q, want %q", got, providerPath)
	}
}

func TestRepositoryConfigurationResolvesRelativeCommandsFromRepository(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "tools")
	providerPath := writeProvider(t, directory, "arcana")
	relative, err := filepath.Rel(root, providerPath)
	if err != nil {
		t.Fatal(err)
	}
	writeProviderConfiguration(t, root, providerConfiguration{
		Version: providerConfigVersion, ArcanaCommand: relative,
	})

	got := repositoryConfiguredCommand(root, "arcana")
	if got != providerPath {
		t.Fatalf("relative configured provider = %q, want %q", got, providerPath)
	}
}

func TestResolveProviderCommandFindsCheckoutFromLexiconAdapterRoot(t *testing.T) {
	checkout := t.TempDir()
	if err := os.WriteFile(filepath.Join(checkout, "go.mod"), []byte("module github.com/Lokee86/grimoire\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	adapterRoot := filepath.Join(checkout, "lexicon", "adapters")
	if err := os.MkdirAll(adapterRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	providerPath := writeProvider(t, filepath.Join(checkout, "bin"), "lexicon")

	repository := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repository, ".lexicon"), 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(map[string]any{"version": 1, "adapter_root": adapterRoot})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repository, ".lexicon", "config.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", t.TempDir())
	t.Setenv("GRIMOIRE_HOME", "")

	got := resolveProviderCommand(repository, "", "lexicon")
	if got != providerPath {
		t.Fatalf("checkout provider = %q, want %q", got, providerPath)
	}
}

func TestResolveProviderCommandUsesGrimoireHomeBeforePATH(t *testing.T) {
	checkout := t.TempDir()
	providerPath := writeProvider(t, filepath.Join(checkout, "arcana", "target", "debug"), "arcana")
	t.Setenv("GRIMOIRE_HOME", checkout)
	t.Setenv("PATH", t.TempDir())

	got := resolveProviderCommand(t.TempDir(), "", "arcana")
	if got != providerPath {
		t.Fatalf("GRIMOIRE_HOME provider = %q, want %q", got, providerPath)
	}
}

func TestResolveProviderCommandUsesPATHAsFinalFallback(t *testing.T) {
	directory := t.TempDir()
	providerPath := writeProvider(t, directory, "lexicon")
	t.Setenv("GRIMOIRE_HOME", "")
	t.Setenv("PATH", directory)

	got := resolveProviderCommand(t.TempDir(), "", "lexicon")
	if got != providerPath {
		t.Fatalf("PATH provider = %q, want %q", got, providerPath)
	}
}

func writeProviderConfiguration(t *testing.T, root string, configuration providerConfiguration) {
	t.Helper()
	directory := filepath.Join(root, ".grimoire")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(configuration)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "providers.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeProvider(t *testing.T, directory, name string) string {
	t.Helper()
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	filename := name
	if runtime.GOOS == "windows" {
		filename += ".exe"
	}
	path := filepath.Join(directory, filename)
	if err := os.WriteFile(path, []byte("provider"), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}
