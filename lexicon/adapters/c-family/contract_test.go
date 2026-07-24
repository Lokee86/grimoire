package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestAdapterPassesFactsV1Validator(t *testing.T) {
	python, err := exec.LookPath("python")
	if err != nil {
		python, err = exec.LookPath("python3")
	}
	if err != nil {
		t.Skip("Python is unavailable")
	}

	repository := t.TempDir()
	writeFixture(t, repository, map[string]string{
		"api.h":   "int answer(int value);\n",
		"api.cpp": "#include \"api.h\"\nint answer(int value) { return value + 1; }\n",
	})
	data, err := analyzeRepository(repository, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "facts.jsonl")
	if err := os.WriteFile(output, data, 0o644); err != nil {
		t.Fatal(err)
	}
	validator := filepath.Clean(filepath.Join("..", "..", "tools", "validate_jsonl.py"))
	command := exec.Command(python, validator, output)
	if result, err := command.CombinedOutput(); err != nil {
		t.Fatalf("facts-v1 validation failed: %v\n%s", err, result)
	}
}
