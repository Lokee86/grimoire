package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIgnorePolicyUsesRepositoryPatternsAndPermanentExclusions(t *testing.T) {
	root := t.TempDir()
	ignoreFile := `*.py
!keep.py
nested/*.go
!nested/keep.go
cache/
locked/
!locked/keep.py
!.git/keep.go
`
	if err := os.WriteFile(filepath.Join(root, IgnoreFileName), []byte(ignoreFile), 0o644); err != nil {
		t.Fatal(err)
	}
	policy, err := LoadIgnorePolicy(root)
	if err != nil {
		t.Fatal(err)
	}

	tests := map[string]bool{
		"main.py":        true,
		"keep.py":        false,
		"nested/drop.go": true,
		"nested/keep.go": false,
		"cache/data.py":  true,
		"locked/keep.py": true,
		".git/keep.go":   true,
		"src/visible.go": false,
		"cache":          true,
		".lexiconignore": false,
	}
	for relative, expected := range tests {
		infoIsDir := relative == "cache"
		if actual := policy.Ignored(filepath.Join(root, relative), infoIsDir); actual != expected {
			t.Errorf("Ignored(%q, %t) = %t, want %t", relative, infoIsDir, actual, expected)
		}
	}
}
