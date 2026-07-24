package languages

import (
	"reflect"
	"testing"
)

func TestDedicatedAdaptersTakePrecedence(t *testing.T) {
	for path, want := range map[string][]string{
		"include/api.hpp": {"c-family"},
		"src/App.svelte":  {"typescript"},
		"src/main.c":      {"c-family"},
		"src/main.cpp":    {"c-family"},
		"src/main.go":     {"go"},
		"src/main.py":     {"python"},
	} {
		if got := ForPath(path); !reflect.DeepEqual(got, want) {
			t.Fatalf("ForPath(%s) = %v, want %v", path, got, want)
		}
	}
}

func TestGenericFallbackUsesExtensionIdentity(t *testing.T) {
	if got, want := ForPath("src/Main.java"), []string{"generic-java"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("ForPath(Main.java) = %v, want %v", got, want)
	}
	definition, ok := Lookup("generic-java")
	if !ok || definition.Directory != "generic" || !reflect.DeepEqual(definition.Extensions, []string{".java"}) {
		t.Fatalf("generic definition = %#v, %v", definition, ok)
	}
	if !OwnsSource("generic-java", "src/Main.java") || OwnsSource("generic-java", "src/main.c") {
		t.Fatal("generic extension ownership is not exact")
	}
}

func TestGenericFallbackExcludesNonSourceFiles(t *testing.T) {
	for _, path := range []string{"README.md", "config.json", "settings.yaml", "data.csv", "image.png"} {
		if got := ForPath(path); got != nil {
			t.Fatalf("ForPath(%s) = %v, want nil", path, got)
		}
	}
	if IsGeneric("generic-md") || IsGeneric("generic") {
		t.Fatal("unsupported generic identities must be rejected")
	}
}
