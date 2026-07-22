package main

import (
	"flag"
	"fmt"
	"os"
)

type stringList []string

func (values *stringList) String() string { return fmt.Sprint([]string(*values)) }
func (values *stringList) Set(value string) error {
	*values = append(*values, value)
	return nil
}

func main() {
	repository := flag.String("repo", ".", "repository root to scan")
	output := flag.String("output", "", "Lexicon JSONL output path")
	var changedFiles, removedFiles stringList
	flag.Var(&changedFiles, "changed-file", "repository-relative file to emit")
	flag.Var(&removedFiles, "removed-file", "repository-relative removed file")
	flag.Parse()
	if *output == "" {
		fmt.Fprintln(os.Stderr, "-output is required")
		flag.Usage()
		os.Exit(2)
	}

	facts, summary, err := scanRepository(*repository)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan failed: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(*output, []byte(encodeFactsScoped(facts, changedFiles, removedFiles)), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write %s: %v\n", *output, err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "wrote %s: nodes=%d edges=%d directories=%d files=%d packages=%d imports=%d call_expressions=%d resolved_calls=%d possible_call_targets=%d unresolved_calls=%d builtin_calls=%d conversion_calls=%d external_calls=%d dynamic_calls=%d interface_calls=%d closures=%d captures=%d semantic_errors=%d\n", *output, summary.Nodes, summary.Edges, summary.Directories, summary.Files, summary.Packages, summary.Imports, summary.CallExpressions, summary.DirectCalls, summary.PossibleCallTargets, summary.UnresolvedCalls, summary.BuiltinCalls, summary.ConversionCalls, summary.ExternalCalls, summary.DynamicCalls, summary.InterfaceCalls, summary.Closures, summary.Captures, summary.SemanticErrors)
}
