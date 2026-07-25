package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Lokee86/grimoire/evaluation/agent_discovery"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "agent-discovery:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	flags := flag.NewFlagSet("agent-discovery", flag.ContinueOnError)
	cases := flags.String("cases", "", "versioned discovery case corpus")
	adapterName := flags.String("adapter", "", "raw, grimoire-context, progressive-jsonl, or registered adapter")
	input := flags.String("input", "", "recorded transcript or grimoire context JSON")
	caseID := flags.String("case", "", "case id for a context output without case_id")
	outputDir := flags.String("output-dir", "evaluation/results", "comparison report directory")
	name := flags.String("name", "agent-discovery-report", "report filename without extension")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *cases == "" || *adapterName == "" || *input == "" {
		return errors.New("--cases, --adapter, and --input are required")
	}
	corpus, err := agentdiscovery.LoadCorpus(*cases)
	if err != nil {
		return err
	}
	adapter, ok := agentdiscovery.AdapterFor(*adapterName)
	if !ok {
		return fmt.Errorf("unknown adapter %q; CBM exporters can register one with agentdiscovery.RegisterAdapter", *adapterName)
	}
	transcripts, err := adapter(*input)
	if err != nil {
		return err
	}
	for i := range transcripts {
		if transcripts[i].CaseID == "" {
			transcripts[i].CaseID = *caseID
		}
		if transcripts[i].Adapter == "" {
			transcripts[i].Adapter = *adapterName
		}
		if transcripts[i].RunID == "" {
			transcripts[i].RunID = filepath.Base(*input)
		}
		if transcripts[i].CaseID == "" {
			return errors.New("--case is required when transcript records omit case_id")
		}
	}
	report, err := agentdiscovery.BuildReport(corpus, transcripts)
	if err != nil {
		return err
	}
	jsonPath, markdownPath, err := agentdiscovery.WriteReport(report, *outputDir, strings.TrimSpace(*name))
	if err != nil {
		return err
	}
	fmt.Printf("wrote %s\nwrote %s\n", jsonPath, markdownPath)
	return nil
}
