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
}

func (value *stringList) String() string {
	return strings.Join(value.values, ",")
}

func (value *stringList) Set(item string) error {
	value.values = append(value.values, item)
	return nil
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "lotusscript adapter:", err)
		os.Exit(1)
	}
}

func run(arguments []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("lexicon-lotusscript", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	repository := flags.String("repo", "", "repository root")
	output := flags.String("output", "-", "facts-v1 JSONL destination or - for stdout")
	var changedFiles stringList
	var removedFiles stringList
	flags.Var(&changedFiles, "changed-file", "changed repository-relative source path")
	flags.Var(&removedFiles, "removed-file", "removed repository-relative source path")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", flags.Args())
	}
	if *repository == "" {
		return fmt.Errorf("--repo is required")
	}
	data, err := analyzeRepository(*repository)
	if err != nil {
		return err
	}
	if *output == "-" {
		_, err = stdout.Write(data)
		return err
	}
	return os.WriteFile(*output, data, 0o644)
}
