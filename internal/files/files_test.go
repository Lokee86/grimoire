package files

import (
	"reflect"
	"testing"
)

func TestLanguages(t *testing.T) {
	tests := map[string][]string{
		"main.go":       {"go"},
		"src/app.tsx":   {"typescript"},
		"src/feed.js":   {"typescript"},
		"src/view.jsx":  {"typescript"},
		"config.mjs":    {"typescript"},
		"legacy.cjs":    {"typescript"},
		"jsconfig.json": {"typescript"},
		"Cargo.toml":    {"rust"},
		"project.godot": {"gdscript"},
		"README.md":     nil,
	}
	for path, expected := range tests {
		if actual := Languages(path); !reflect.DeepEqual(actual, expected) {
			t.Fatalf("Languages(%q) = %v, want %v", path, actual, expected)
		}
	}
}

func TestIgnoredDirectory(t *testing.T) {
	for _, name := range []string{".git", ".lexicon", ".astro", "node_modules", "target"} {
		if !IgnoredDirectory(name) {
			t.Fatalf("expected %q to be ignored", name)
		}
	}
}
