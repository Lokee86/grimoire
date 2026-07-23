//go:build windows

package vectorstore

import (
	"path/filepath"
	"slices"
	"testing"
)

func TestLibraryCandidatesSearchExecutableAncestorBuilds(t *testing.T) {
	root := t.TempDir()
	executable := filepath.Join(root, "bin", "grimoire.exe")
	unrelatedRepository := filepath.Join(t.TempDir(), "repo")

	candidates := libraryCandidates(executable, unrelatedRepository)
	for _, expected := range []string{
		filepath.Join(root, "native", "vector-engine", "target", "release", ABIName+".dll"),
		filepath.Join(root, "native", "vector-engine", "target", "debug", ABIName+".dll"),
	} {
		if !slices.Contains(candidates, expected) {
			t.Fatalf("expected executable-relative candidate %q in %v", expected, candidates)
		}
	}
}

func TestLibraryCandidatesPreferPackagedSibling(t *testing.T) {
	executable := filepath.Join(t.TempDir(), "bin", "grimoire.exe")
	candidates := libraryCandidates(executable, "")
	expected := filepath.Join(filepath.Dir(executable), ABIName+".dll")
	if len(candidates) == 0 || candidates[0] != expected {
		t.Fatalf("expected packaged sibling %q first, got %v", expected, candidates)
	}
}
