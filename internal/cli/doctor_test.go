package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/lexicon/internal/config"
	"github.com/Lokee86/lexicon/internal/objectstore"
)

func TestDoctorUsesDiscoveryAndReportsPassingChecks(t *testing.T) {
	repository, _, objectIDs := doctorFixture(t)
	consumerRoot := filepath.Join(config.StateRoot(repository), "consumers")
	if err := os.MkdirAll(consumerRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(consumerRoot, "arcana.json"), []byte(`{"version":1,"command":"go"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(repository, "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(nested)

	installDoctorSeams(t)
	doctorVerifyState = func(string) error { return nil }
	doctorLookPath = func(name string) (string, error) {
		if name == "go" {
			return "/fake/go", nil
		}
		return "", errors.New("not found")
	}
	calls := make(map[string]int)
	originalLoadObject := doctorLoadObject
	doctorLoadObject = func(current objectstore.Store, id string) (objectstore.FactObject, error) {
		calls[id]++
		return originalLoadObject(current, id)
	}

	var stdout, stderr bytes.Buffer
	if code := Run([]string{"doctor"}, &stdout, &stderr); code != 0 {
		t.Fatalf("doctor exit code = %d, stderr = %s\noutput = %s", code, stderr.String(), stdout.String())
	}
	want := []string{
		"PASS configuration loading",
		"PASS private Git state repository",
		"PASS CURRENT snapshot and referenced objects",
		"PASS configured adapter root",
		"PASS adapter directory: go",
		"PASS runtime executable: go",
		"PASS consumer definition: arcana.json",
		"PASS consumer command: arcana.json",
	}
	assertLines(t, stdout.String(), want)
	for _, id := range objectIDs {
		if calls[id] != 1 {
			t.Fatalf("snapshot object %s verified %d times, want once", id, calls[id])
		}
	}
}

func TestDoctorReportsAllFailuresWithoutExecutingConsumers(t *testing.T) {
	repository, store, objectIDs := doctorFixture(t)
	if err := os.RemoveAll(filepath.Join(repository, "adapters", "go")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(store.ObjectPath(objectIDs[0]), []byte("corrupt\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	consumerRoot := filepath.Join(config.StateRoot(repository), "consumers")
	if err := os.MkdirAll(consumerRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(consumerRoot, "bad.json"), []byte(`{"version":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(consumerRoot, "missing.json"), []byte(`{"version":1,"command":"missing-consumer"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	installDoctorSeams(t)
	doctorVerifyState = func(string) error { return errors.New("state repository unavailable") }
	doctorLookPath = func(string) (string, error) { return "", errors.New("not found") }

	var stdout bytes.Buffer
	if err := doctorAt(repository, &stdout); err == nil {
		t.Fatal("doctorAt returned nil for failing checks")
	}
	assertLines(t, stdout.String(), []string{
		"FAIL private Git state repository: state repository unavailable",
		"FAIL CURRENT snapshot and referenced objects:",
		"FAIL adapter directory: go:",
		"FAIL runtime executable: go:",
		"FAIL consumer definition: bad.json:",
		"FAIL consumer command: bad.json: definition unavailable",
		"PASS consumer definition: missing.json",
		"FAIL consumer command: missing.json:",
	})
	if strings.Contains(stdout.String(), "consumer executed") {
		t.Fatal("doctor output suggests consumer execution")
	}

}

func doctorFixture(t *testing.T) (string, objectstore.Store, []string) {
	t.Helper()
	repository := t.TempDir()
	adapterRoot := filepath.Join(repository, "adapters")
	if err := os.MkdirAll(filepath.Join(adapterRoot, "go"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := config.Save(repository, adapterRoot); err != nil {
		t.Fatal(err)
	}
	store := objectstore.Store{Root: config.StateRoot(repository)}
	ids := make([]string, 0, 2)
	for _, owner := range []string{"shared", "main.go"} {
		id, err := store.WriteObject(objectstore.FactObject{
			Language:         "go",
			Owner:            owner,
			AdapterVersion:   "test",
			SchemaVersion:    1,
			AnalysisConfigID: config.AnalysisID(),
			Records:          nil,
		})
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	if _, err := store.Publish(objectstore.Manifest{
		StateCommit: "state",
		Languages: []objectstore.LanguageEntry{{
			Language:       "go",
			SharedObjectID: ids[0],
			Files:          []objectstore.FileEntry{{Language: "go", Path: "main.go", ObjectID: ids[1]}},
		}},
	}); err != nil {
		t.Fatal(err)
	}
	return repository, store, ids
}

func installDoctorSeams(t *testing.T) {
	t.Helper()
	loadConfig := doctorLoadConfig
	verifyState := doctorVerifyState
	currentSnapshot := doctorCurrentSnapshot
	loadObject := doctorLoadObject
	lookPath := doctorLookPath
	validateConsumer := doctorValidateConsumer
	t.Cleanup(func() {
		doctorLoadConfig = loadConfig
		doctorVerifyState = verifyState
		doctorCurrentSnapshot = currentSnapshot
		doctorLoadObject = loadObject
		doctorLookPath = lookPath
		doctorValidateConsumer = validateConsumer
	})
}

func assertLines(t *testing.T, output string, want []string) {
	t.Helper()
	for _, line := range want {
		if !strings.Contains(output, line) {
			t.Fatalf("output %q does not contain %q", output, line)
		}
	}
}
