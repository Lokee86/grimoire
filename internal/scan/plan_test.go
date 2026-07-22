package scan

import "testing"

func TestTypeScriptPlanOwnsJavaScriptSources(t *testing.T) {
	for _, path := range []string{
		"src/app.js",
		"src/view.jsx",
		"src/config.mjs",
		"src/legacy.cjs",
		"src/app.ts",
		"src/view.tsx",
		"src/config.mts",
		"src/legacy.cts",
	} {
		if !languageOwnsSource("typescript", path) {
			t.Fatalf("expected TypeScript adapter to own %q", path)
		}
	}
	if languageOwnsSource("typescript", "src/page.astro") {
		t.Fatal("Astro files must remain outside the JavaScript/TypeScript adapter")
	}
}
