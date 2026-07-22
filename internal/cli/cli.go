package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/Lokee86/lexicon/internal/config"
	"github.com/Lokee86/lexicon/internal/scan"
	lexwatch "github.com/Lokee86/lexicon/internal/watch"
)

func Run(arguments []string, stdout, stderr io.Writer) int {
	if len(arguments) == 0 {
		usage(stderr)
		return 2
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	var err error
	switch arguments[0] {
	case "init":
		err = runInit(ctx, arguments[1:], stdout, stderr)
	case "scan":
		err = runScan(ctx, arguments[1:], stdout, stderr)
	case "daemon":
		err = runDaemon(ctx, arguments[1:], stdout, stderr)
	case "help", "-h", "--help":
		usage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", arguments[0])
		usage(stderr)
		return 2
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func runInit(ctx context.Context, arguments []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("init", flag.ContinueOnError)
	flags.SetOutput(stderr)
	repository := flags.String("repo", ".", "repository to initialize")
	adapterRoot := flags.String("adapters", "", "Lexicon adapters directory")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	root, err := absolute(*repository)
	if err != nil {
		return err
	}
	adapters, err := config.FindAdapterRoot(root, *adapterRoot)
	if err != nil {
		return err
	}
	_, report, err := scan.Initialize(ctx, root, adapters, stdout)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "initialized Lexicon: %s\n", root)
	if len(report.Languages) > 0 {
		fmt.Fprintf(stdout, "libraries: %s\n", strings.Join(report.Languages, ", "))
	}
	return nil
}

func runScan(ctx context.Context, arguments []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("scan", flag.ContinueOnError)
	flags.SetOutput(stderr)
	repository := flags.String("repo", ".", "repository to scan")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	scanner, err := scan.Open(*repository, stdout)
	if err != nil {
		return err
	}
	report, err := scanner.Scan(ctx)
	if err != nil {
		return err
	}
	if len(report.Changed) == 0 {
		fmt.Fprintln(stdout, "Lexicon is current")
		return nil
	}
	fmt.Fprintf(stdout, "updated %d files: %s\n", len(report.Changed), strings.Join(report.Languages, ", "))
	return nil
}

func runDaemon(ctx context.Context, arguments []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("daemon", flag.ContinueOnError)
	flags.SetOutput(stderr)
	repository := flags.String("repo", ".", "repository to watch")
	debounce := flags.Duration("debounce", 150*time.Millisecond, "quiet period before scanning changes")
	reconcile := flags.Duration("reconcile", 30*time.Second, "full reconciliation interval")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	scanner, err := scan.Open(*repository, stdout)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Lexicon daemon watching %s\n", scanner.Repository)
	return lexwatch.Run(ctx, scanner, lexwatch.Options{Debounce: *debounce, Reconcile: *reconcile, Output: stderr})
}

func absolute(path string) (string, error) {
	value, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(value)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("repository is not a directory: %s", value)
	}
	return value, nil
}

func usage(output io.Writer) {
	fmt.Fprintln(output, "Usage: lexicon <init|scan|daemon> [options]")
}
