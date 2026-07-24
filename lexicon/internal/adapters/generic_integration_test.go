package adapters_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Lokee86/lexicon/internal/adapters"
	"github.com/Lokee86/lexicon/internal/objectstore"
)

func TestGenericRunnerProducesValidIncrementalAnalysis(t *testing.T) {
	repository := t.TempDir()
	if err := os.WriteFile(filepath.Join(repository, "Main.java"), []byte("class Main {\n  static int answer() { return 42; }\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	adapterRoot, err := filepath.Abs("../../adapters")
	if err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "generic-java.jsonl")
	request := adapters.Request{
		Language: "generic-java", Repository: repository, Output: output,
		ChangedFiles: []string{"Main.java"}, RemovedFiles: []string{},
	}
	if err := (adapters.Runner{Root: adapterRoot}).Run(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	analysis, err := objectstore.ReadAnalysis(output, "generic-java")
	if err != nil {
		t.Fatal(err)
	}
	if !analysis.IsIncremental() || analysis.Header.SharedComplete == nil || *analysis.Header.SharedComplete {
		t.Fatalf("generic incremental header = %#v", analysis.Header)
	}
}
