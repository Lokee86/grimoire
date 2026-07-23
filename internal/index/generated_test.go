package index

import (
	"strings"
	"testing"
)

func TestGeneratedFileContentDetectsLargeSingleLineWebAssets(t *testing.T) {
	content := []byte(strings.Repeat("function bundled(){return true;}", 2000))
	if !generatedFileContent("runtime.js", content) {
		t.Fatal("large single-line JavaScript was not detected as minified")
	}
	if generatedFileContent("runtime.go", content) {
		t.Fatal("non-web source was classified as minified solely by line length")
	}
}

func TestGeneratedFileContentPreservesAuthoredWebAssets(t *testing.T) {
	content := []byte("export function publish() {\n  return true\n}\n")
	if generatedFileContent("posts.js", content) {
		t.Fatal("small authored JavaScript was classified as generated")
	}
}
