package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	repository := flag.String("repo", ".", "repository root to scan")
	output := flag.String("output", "", "fact TSV output path")
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
	if err := os.WriteFile(*output, []byte(encodeFacts(facts)), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write %s: %v\n", *output, err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "wrote %s: nodes=%d edges=%d directories=%d files=%d packages=%d imports=%d call_expressions=%d direct_calls=%d unresolved_calls=%d\n", *output, summary.Nodes, summary.Edges, summary.Directories, summary.Files, summary.Packages, summary.Imports, summary.CallExpressions, summary.DirectCalls, summary.UnresolvedCalls)
}
