package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

type stringList struct {
	values []string
	set    bool
}

func (value *stringList) String() string {
	return strings.Join(value.values, ",")
}

func (value *stringList) Set(item string) error {
	value.set = true
	value.values = append(value.values, item)
	return nil
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(arguments []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("lexicon-c-family", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	repository := flags.String("repo", "", "repository root")
	output := flags.String("output", "-", "facts-v1 JSONL output path or - for stdout")
	var changedFiles stringList
	var removedFiles stringList
	flags.Var(&changedFiles, "changed-file", "changed repository-relative source path")
	flags.Var(&removedFiles, "removed-file", "removed repository-relative source path")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if *repository == "" {
		return fmt.Errorf("--repo is required")
	}
	data, err := analyzeRepository(*repository, changedFiles.values, removedFiles.values, changedFiles.set || removedFiles.set)
	if err != nil {
		return err
	}
	if *output == "-" {
		_, err = stdout.Write(data)
		return err
	}
	return os.WriteFile(*output, data, 0o644)
}
