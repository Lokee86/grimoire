package consumer

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDefinitionRegistry(t *testing.T) {
	stateRoot := filepath.Join(t.TempDir(), ".lexicon")
	definition := Definition{Version: Version, Command: "arcana", Args: []string{"sync"}}
	if err := AddDefinition(stateRoot, "zeta.json", definition); err != nil {
		t.Fatal(err)
	}
	if err := AddDefinition(stateRoot, "alpha.json", definition); err != nil {
		t.Fatal(err)
	}
	names, err := ListDefinitions(stateRoot)
	if err != nil {
		t.Fatal(err)
	}
	if expected := []string{"alpha.json", "zeta.json"}; !reflect.DeepEqual(names, expected) {
		t.Fatalf("definition names = %#v", names)
	}

	replacement := Definition{Version: Version, Command: "replacement", Args: []string{"refresh"}}
	if err := AddDefinition(stateRoot, "alpha.json", replacement); err != nil {
		t.Fatal(err)
	}
	loaded, err := load(filepath.Join(stateRoot, "consumers", "alpha.json"))
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Command != replacement.Command || !reflect.DeepEqual(loaded.Args, replacement.Args) {
		t.Fatalf("replacement = %#v", loaded)
	}
	if err := saveSnapshot(stateRoot, "alpha.json", "sha256:old"); err != nil {
		t.Fatal(err)
	}
	if err := RemoveDefinition(stateRoot, "alpha.json"); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		filepath.Join(stateRoot, "consumers", "alpha.json"),
		filepath.Join(stateRoot, "consumer-state", "alpha.json"),
	} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("removed path %s has error %v", path, err)
		}
	}
	for _, name := range []string{"", "alpha", "../alpha.json", `nested\alpha.json`, "alpha.txt", ".json"} {
		if err := AddDefinition(stateRoot, name, definition); err == nil {
			t.Fatalf("accepted invalid consumer name %q", name)
		}
	}
}

func TestRunOneUpdatesNamedConsumerState(t *testing.T) {
	repository := t.TempDir()
	stateRoot := filepath.Join(repository, ".lexicon")
	consumerRoot := filepath.Join(stateRoot, "consumers")
	if err := os.MkdirAll(consumerRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	log := filepath.Join(repository, "consumer-order")
	t.Setenv("LEXICON_TEST_HELPER", "1")
	t.Setenv("LEXICON_TEST_LOG", log)
	writeDefinition(t, filepath.Join(consumerRoot, "01-first.json"), []string{"-test.run=TestConsumerHelperProcess", "success-01"})
	writeDefinition(t, filepath.Join(consumerRoot, "02-second.json"), []string{"-test.run=TestConsumerHelperProcess", "success-02"})

	const snapshotID = "sha256:2222222222222222222222222222222222222222222222222222222222222222"
	if err := RunOne(context.Background(), repository, stateRoot, "02-second.json", snapshotID, nil); err != nil {
		t.Fatal(err)
	}
	order, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if string(order) != "success-02\n" {
		t.Fatalf("named execution = %q", order)
	}
	state, err := os.ReadFile(filepath.Join(stateRoot, "consumer-state", "02-second.json"))
	if err != nil {
		t.Fatal(err)
	}
	var decoded SuccessState
	if err := json.Unmarshal(state, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Version != StateVersion || decoded.SnapshotID != snapshotID {
		t.Fatalf("named state = %#v", decoded)
	}
	const replacementSnapshot = "sha256:3333333333333333333333333333333333333333333333333333333333333333"
	if err := RunOne(context.Background(), repository, stateRoot, "02-second.json", replacementSnapshot, nil); err != nil {
		t.Fatal(err)
	}
	replaced, err := os.ReadFile(filepath.Join(stateRoot, "consumer-state", "02-second.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(replaced), replacementSnapshot) {
		t.Fatalf("replaced state = %q", replaced)
	}
	entries, err := os.ReadDir(filepath.Join(stateRoot, "consumer-state"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "02-second.json" {
		t.Fatalf("consumer state files = %#v", entries)
	}
	if _, err := os.Stat(filepath.Join(stateRoot, "consumer-state", "01-first.json")); !os.IsNotExist(err) {
		t.Fatalf("unexpected state for unselected consumer: %v", err)
	}
	if err := RunOne(context.Background(), repository, stateRoot, "../02-second.json", snapshotID, nil); err == nil || !strings.Contains(err.Error(), "invalid") {
		t.Fatalf("path traversal result = %v", err)
	}
}
