package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	repo := flag.String("repo", "", "repository root to analyze")
	output := flag.String("output", "", "JSONL output path")
	flag.Parse()
	if *repo == "" || *output == "" {
		flag.Usage()
		os.Exit(2)
	}
	if err := writeFacts(*repo, *output); err != nil {
		fmt.Fprintf(os.Stderr, "gdscript adapter: %v\n", err)
		os.Exit(1)
	}
}

func writeFacts(repo, output string) error {
	data, err := analyzeRepository(repo)
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
