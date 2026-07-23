package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParallelSemanticResolutionMatchesSerialOutput(t *testing.T) {
	root := t.TempDir()
	writeParallelFixture(t, root)

	serial, _, err := scanRepositoryWithOptions(root, ScanOptions{
		SemanticWorkers: 1, SemanticShards: 1, MergeFanIn: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	parallel, _, err := scanRepositoryWithOptions(root, ScanOptions{
		SemanticWorkers: 4, SemanticShards: 8, MergeFanIn: 4,
	})
	if err != nil {
		t.Fatal(err)
	}
	serialOutput := encodeFacts(serial)
	parallelOutput := encodeFacts(parallel)
	if serialOutput != parallelOutput {
		t.Fatal("parallel semantic resolution changed canonical output")
	}
}

func TestParallelSemanticResolutionIgnoresReductionShape(t *testing.T) {
	root := t.TempDir()
	writeParallelFixture(t, root)

	binary, _, err := scanRepositoryWithOptions(root, ScanOptions{
		SemanticWorkers: 3, SemanticShards: 6, MergeFanIn: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	wide, _, err := scanRepositoryWithOptions(root, ScanOptions{
		SemanticWorkers: 3, SemanticShards: 6, MergeFanIn: 8,
	})
	if err != nil {
		t.Fatal(err)
	}
	if encodeFacts(binary) != encodeFacts(wide) {
		t.Fatal("semantic output changed with reduction fan-in")
	}
}

func writeParallelFixture(t *testing.T, root string) {
	t.Helper()
	files := map[string]string{
		"go.mod": "module example.com/parallel\n\ngo 1.26\n",
		"worker/worker.go": `package worker

type Runner interface { Run() int }

type Fast struct{ Value int }
func (f *Fast) Run() int { f.Value++; return f.Value }

func Execute(r Runner) int { return r.Run() }
`,
		"service/service.go": `package service

import "example.com/parallel/worker"

func Start() int {
	value := &worker.Fast{}
	return worker.Execute(value)
}
`,
		"main.go": `package main

import "example.com/parallel/service"

func main() { _ = service.Start() }
`,
		"extra.go": `package main

func helper(value int) int { return value + 1 }
`,
	}
	for relative, contents := range files {
		path := filepath.Join(root, filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
