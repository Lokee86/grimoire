package evaluation

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCorpusRequiresExplicitEvidence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cases.json")
	data := `{"version":1,"repository":"test","cases":[{"id":"missing","query":"where","category":"direct-location","budget":100,"required":[]}]}`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCorpus(path); err == nil {
		t.Fatal("expected missing required evidence error")
	}
}

func TestLoadCorpusAcceptsStructuralOnlyRequiredEvidence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cases.json")
	data := `{"version":1,"repository":"test","cases":[{"id":"structural","query":"trace Target","category":"call-chain-investigation","budget":100,"required_structural":[{"provider":"arcana","kind":"call_chain","chain":["Start","Target"]}]}]}`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	corpus, err := LoadCorpus(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(corpus.Cases[0].RequiredStructural) != 1 {
		t.Fatalf("unexpected corpus: %+v", corpus)
	}
}

func TestLoadCorpusRejectsInvalidStructuralExpectation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cases.json")
	data := `{"version":1,"repository":"test","cases":[{"id":"bad","query":"trace Target","category":"call-chain-investigation","budget":100,"required_structural":[{"provider":"arcana","kind":"call_chain","chain":["Target"]}]}]}`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCorpus(path); err == nil {
		t.Fatal("expected invalid structural expectation error")
	}
}

func TestLoadCorpusAcceptsRepositoryOwnedFormat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cases.json")
	data := `{"version":1,"repository":"test","cases":[{"id":"valid","query":"where","category":"direct-location","budget":100,"required":[{"path":"internal/example.go","symbols":["Example"]}]}]}`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	corpus, err := LoadCorpus(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(corpus.Cases) != 1 || corpus.Cases[0].ID != "valid" {
		t.Fatalf("unexpected corpus: %+v", corpus)
	}
}
