package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

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
