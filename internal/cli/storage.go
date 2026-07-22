package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/Lokee86/lexicon/internal/config"
	"github.com/Lokee86/lexicon/internal/lock"
	"github.com/Lokee86/lexicon/internal/objectstore"
	"github.com/Lokee86/lexicon/internal/scan"
)

func runRebuild(ctx context.Context, arguments []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("rebuild", flag.ContinueOnError)
	flags.SetOutput(stderr)
	repository := flags.String("repo", "", "repository to rebuild")
	languageText := flags.String("languages", "", "comma-separated enabled languages")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	root, err := resolveRepository(*repository)
	if err != nil {
		return err
	}
	languages, err := parseOptionalLanguages(*languageText)
	if err != nil {
		return err
	}
	scanner, err := scan.Open(root, stdout)
	if err != nil {
		return err
	}
	report, err := scanner.Rebuild(ctx, languages)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "rebuilt libraries: %s\n", displayList(report.Languages))
	fmt.Fprintf(stdout, "snapshot: %s\n", report.SnapshotID)
	return nil
}

func runExport(arguments []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("export", flag.ContinueOnError)
	flags.SetOutput(stderr)
	repository := flags.String("repo", "", "repository to export")
	output := flags.String("output", "", "destination directory")
	snapshot := flags.String("snapshot", "CURRENT", "snapshot ID or CURRENT")
	languageText := flags.String("languages", "", "comma-separated languages")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if strings.TrimSpace(*output) == "" {
		return fmt.Errorf("export requires --output")
	}
	root, err := resolveRepository(*repository)
	if err != nil {
		return err
	}
	languages, err := parseOptionalLanguages(*languageText)
	if err != nil {
		return err
	}
	store := objectstore.Store{Root: config.StateRoot(root)}
	if err := store.Export(*snapshot, *output, languages); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "exported snapshot %s to %s\n", *snapshot, *output)
	return nil
}

func runGC(arguments []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("gc", flag.ContinueOnError)
	flags.SetOutput(stderr)
	repository := flags.String("repo", "", "repository to collect")
	retain := flags.Int("retain", 20, "number of newest snapshots to retain")
	dryRun := flags.Bool("dry-run", false, "report deletions without removing files")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	root, err := resolveRepository(*repository)
	if err != nil {
		return err
	}
	stateRoot := config.StateRoot(root)
	guard, err := lock.Acquire(stateRoot)
	if err != nil {
		return err
	}
	defer guard.Close()
	store := objectstore.Store{Root: stateRoot}
	result, err := store.GarbageCollect(objectstore.GCOptions{KeepSnapshots: *retain}, *dryRun)
	if err != nil {
		return err
	}
	mode := "deleted"
	if result.DryRun {
		mode = "would delete"
	}
	fmt.Fprintf(stdout, "%s %d snapshots and %d objects\n", mode, len(result.DeletedSnapshots), len(result.DeletedObjects))
	return nil
}

func parseOptionalLanguages(value string) ([]string, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	return parseLanguageSelection(value)
}
