package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type stringList []string

func (values *stringList) String() string { return fmt.Sprint([]string(*values)) }
func (values *stringList) Set(value string) error {
	*values = append(*values, value)
	return nil
}

func main() {
	repo := flag.String("repo", "", "repository root to analyze")
	output := flag.String("output", "", "JSONL output path")
	var changedFiles, removedFiles stringList
	flag.Var(&changedFiles, "changed-file", "repository-relative file to emit")
	flag.Var(&removedFiles, "removed-file", "repository-relative removed file")
	flag.Parse()
	if *repo == "" || *output == "" {
		flag.Usage()
		os.Exit(2)
	}
	if err := writeFacts(*repo, *output, changedFiles, removedFiles); err != nil {
		fmt.Fprintf(os.Stderr, "gdscript adapter: %v\n", err)
		os.Exit(1)
	}
}

func writeFacts(repo, output string, changedFiles, removedFiles []string) error {
	data, err := analyzeRepository(repo, changedFiles, removedFiles)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	if err := os.WriteFile(output, data, 0o644); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}
