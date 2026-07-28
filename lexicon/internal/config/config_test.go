package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestStateRootUsesExplicitWarlockStateDirectory(t *testing.T) {
	repository := t.TempDir()
	state := filepath.Join(repository, ".warlock", "tools", "lexicon")
	t.Setenv("LEXICON_STATE_DIR", state)
	if actual := StateRoot(repository); actual != state {
		t.Fatalf("StateRoot() = %q, want %q", actual, state)
	}
	if err := Save(repository, t.TempDir()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(state, "config.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(repository, ".lexicon")); !os.IsNotExist(err) {
		t.Fatalf("standalone state was created: %v", err)
	}
}

func TestEnabledLanguagesDefaultToAllSupported(t *testing.T) {
	value := Config{}
	if !value.LanguageEnabled("python") || !value.LanguageEnabled("typescript") {
		t.Fatal("omitted enabled_languages must enable all detected languages")
	}
}

func TestLoadKeepsLegacyConfigurationAsDefaultAll(t *testing.T) {
	repository := t.TempDir()
	if err := os.MkdirAll(filepath.Dir(Path(repository)), 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(map[string]any{"version": 1, "adapter_root": "adapters"})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(Path(repository), data, 0o644); err != nil {
		t.Fatal(err)
	}
	value, err := Load(repository)
	if err != nil {
		t.Fatal(err)
	}
	if len(value.EnabledLanguages) != 0 || !value.LanguageEnabled("go") {
		t.Fatalf("legacy configuration = %#v", value)
	}
}

func TestSaveAndUpdateEnabledLanguages(t *testing.T) {
	repository := t.TempDir()
	adapterRoot := t.TempDir()
	if err := SaveWithEnabledLanguages(repository, adapterRoot, []string{"typescript", "python", "python"}); err != nil {
		t.Fatal(err)
	}
	value, err := Load(repository)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(value.EnabledLanguages, []string{"python", "typescript"}) {
		t.Fatalf("enabled languages = %v", value.EnabledLanguages)
	}
	if err := Save(repository, filepath.Join(repository, "replacement-adapters")); err != nil {
		t.Fatal(err)
	}
	value, err = Load(repository)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(value.EnabledLanguages, []string{"python", "typescript"}) {
		t.Fatalf("Save dropped selection: %v", value.EnabledLanguages)
	}
	if err := UpdateEnabledLanguages(repository, []string{}); err != nil {
		t.Fatal(err)
	}
	value, err = Load(repository)
	if err != nil {
		t.Fatal(err)
	}
	if len(value.EnabledLanguages) != 0 || !value.LanguageEnabled("ruby") {
		t.Fatalf("empty selection = %#v", value.EnabledLanguages)
	}
	data, err := os.ReadFile(Path(repository))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == "" || filepath.Base(Path(repository)) != "config.json" {
		t.Fatal("configuration was not saved")
	}
}

func TestEnabledLanguagesRejectUnknownValues(t *testing.T) {
	if _, err := NormalizeEnabledLanguages([]string{"python", "klingon"}); err == nil {
		t.Fatal("expected unknown language to be rejected")
	}
}

func TestGenericSelectionEnablesExtensionVariants(t *testing.T) {
	if !(Config{EnabledLanguages: []string{"generic"}}).LanguageEnabled("generic-java") {
		t.Fatal("generic selection did not enable generic-java")
	}
	if (Config{EnabledLanguages: []string{"python"}}).LanguageEnabled("generic-java") {
		t.Fatal("generic-java enabled without generic selection")
	}
}

func TestLoadRejectsUnknownEnabledLanguage(t *testing.T) {
	repository := t.TempDir()
	if err := os.MkdirAll(filepath.Dir(Path(repository)), 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(map[string]any{
		"version": 1, "adapter_root": "adapters", "enabled_languages": []string{"klingon"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(Path(repository), data, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(repository); err == nil {
		t.Fatal("expected invalid configuration to be rejected")
	}
}

func TestFindAdapterRootUsesBundleSiblingAdapters(t *testing.T) {
	t.Setenv("LEXICON_ADAPTERS", "")
	bundle := t.TempDir()
	adapterRoot := filepath.Join(bundle, "adapters")
	if err := os.MkdirAll(filepath.Join(adapterRoot, "python"), 0o755); err != nil {
		t.Fatal(err)
	}

	actual, err := findAdapterRoot(
		t.TempDir(),
		"",
		filepath.Join(bundle, "bin", "lexicon.exe"),
		t.TempDir(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if actual != adapterRoot {
		t.Fatalf("adapter root = %q, want %q", actual, adapterRoot)
	}
}
