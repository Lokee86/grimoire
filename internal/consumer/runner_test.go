package consumer

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRunInvokesRegisteredConsumer(t *testing.T) {
	repository := t.TempDir()
	stateRoot := filepath.Join(repository, ".lexicon")
	consumerRoot := filepath.Join(stateRoot, "consumers")
	if err := os.MkdirAll(consumerRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(repository, "consumer-result")
	t.Setenv("LEXICON_TEST_HELPER", "1")
	t.Setenv("LEXICON_TEST_MARKER", marker)
	definition := Definition{
		Version: Version,
		Command: os.Args[0],
		Args:    []string{"-test.run=TestConsumerHelperProcess"},
	}
	data, err := json.Marshal(definition)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(consumerRoot, "arcana.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	const snapshotID = "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	if err := Run(context.Background(), repository, stateRoot, snapshotID, nil); err != nil {
		t.Fatal(err)
	}
	result, err := os.ReadFile(marker)
	if err != nil {
		t.Fatal(err)
	}
	if string(result) != snapshotID {
		t.Fatalf("consumer received %q", result)
	}
}

func TestConsumerHelperProcess(t *testing.T) {
	if os.Getenv("LEXICON_TEST_HELPER") != "1" {
		return
	}
	if err := os.WriteFile(
		os.Getenv("LEXICON_TEST_MARKER"),
		[]byte(os.Getenv("LEXICON_SNAPSHOT_ID")),
		0o644,
	); err != nil {
		os.Exit(2)
	}
	os.Exit(0)
}
