package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestDefinitionWithoutTimeoutRemainsCompatible(t *testing.T) {
	var decoded Definition
	if err := json.Unmarshal([]byte(`{"version":1,"command":"arcana","args":["sync"]}`), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Timeout != 0 || decoded.Command != "arcana" || len(decoded.Args) != 1 {
		t.Fatalf("decoded definition = %#v", decoded)
	}
	if err := json.Unmarshal([]byte(`{"version":1,"command":"arcana","timeout":"2s"}`), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Timeout != 2*time.Second {
		t.Fatalf("decoded timeout = %s", decoded.Timeout)
	}
}

func TestRunExecutesAllConsumersAggregatesFailuresAndPersistsSuccesses(t *testing.T) {
	repository := t.TempDir()
	stateRoot := filepath.Join(repository, ".lexicon")
	consumerRoot := filepath.Join(stateRoot, "consumers")
	if err := os.MkdirAll(consumerRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	log := filepath.Join(repository, "consumer-order")
	t.Setenv("LEXICON_TEST_HELPER", "1")
	t.Setenv("LEXICON_TEST_LOG", log)
	for name, mode := range map[string]string{
		"01-failing.json": "fail",
		"02-success.json": "success-02",
		"03-success.json": "success-03",
	} {
		writeDefinition(t, filepath.Join(consumerRoot, name), []string{"-test.run=TestConsumerHelperProcess", mode})
	}

	const snapshotID = "sha256:1111111111111111111111111111111111111111111111111111111111111111"
	err := Run(context.Background(), repository, stateRoot, snapshotID, nil)
	if err == nil || !strings.Contains(err.Error(), "01-failing.json") {
		t.Fatalf("Run error = %v", err)
	}
	order, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if string(order) != "fail\nsuccess-02\nsuccess-03\n" {
		t.Fatalf("execution order = %q", order)
	}
	for _, name := range []string{"02-success.json", "03-success.json"} {
		data, err := os.ReadFile(filepath.Join(stateRoot, "consumer-state", name))
		if err != nil {
			t.Fatal(err)
		}
		var state SuccessState
		if err := json.Unmarshal(data, &state); err != nil {
			t.Fatal(err)
		}
		if state.Version != StateVersion || state.SnapshotID != snapshotID {
			t.Fatalf("state for %s = %#v", name, state)
		}
		expected := "{\n  \"version\": 1,\n  \"snapshot_id\": \"" + snapshotID + "\"\n}\n"
		if string(data) != expected {
			t.Fatalf("state for %s is not deterministic JSON: %q", name, data)
		}
	}
	if _, err := os.Stat(filepath.Join(stateRoot, "consumer-state", "01-failing.json")); !os.IsNotExist(err) {
		t.Fatalf("failed consumer state error = %v", err)
	}
}

func TestConsumerTimeout(t *testing.T) {
	t.Setenv("LEXICON_TEST_HELPER", "1")
	definition := Definition{
		Version: Version,
		Command: os.Args[0],
		Args:    []string{"-test.run=TestConsumerHelperProcess", "sleep"},
		Timeout: 10 * time.Millisecond,
	}
	err := invoke(context.Background(), definition, t.TempDir(), t.TempDir(), "snapshot", nil)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("timeout error = %v", err)
	}
}

func writeDefinition(t *testing.T, path string, args []string) {
	t.Helper()
	data, err := json.Marshal(Definition{Version: Version, Command: os.Args[0], Args: args})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestConsumerHelperProcess(t *testing.T) {
	if os.Getenv("LEXICON_TEST_HELPER") != "1" {
		return
	}
	mode := "snapshot"
	for _, argument := range os.Args {
		if argument == "fail" || argument == "sleep" || strings.HasPrefix(argument, "success-") {
			mode = argument
		}
	}
	if log := os.Getenv("LEXICON_TEST_LOG"); log != "" {
		file, err := os.OpenFile(log, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			os.Exit(2)
		}
		_, _ = file.WriteString(mode + "\n")
		_ = file.Close()
	}
	if mode == "sleep" {
		time.Sleep(200 * time.Millisecond)
	}
	if mode == "fail" {
		os.Exit(3)
	}
	if marker := os.Getenv("LEXICON_TEST_MARKER"); marker != "" {
		if err := os.WriteFile(marker, []byte(os.Getenv("LEXICON_SNAPSHOT_ID")), 0o644); err != nil {
			os.Exit(2)
		}
	}
	os.Exit(0)
}
