package app

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/Lokee86/grimoire/internal/investigation"
)

func runInvestigation(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		_, err := fmt.Fprintln(stdout, "Usage: grimoire investigation create|status|close [flags]")
		return err
	}
	switch args[0] {
	case "create":
		return runInvestigationCreate(args[1:], stdout, stderr)
	case "status":
		return runInvestigationStatus(args[1:], stdout, false)
	case "close":
		return runInvestigationStatus(args[1:], stdout, true)
	default:
		return fmt.Errorf("unknown investigation command %q", args[0])
	}
}

func runInvestigationCreate(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("investigation create", flag.ContinueOnError)
	flags.SetOutput(stderr)
	root := flags.String("root", ".", "repository root")
	state := flags.String("state", "", "Grimoire state directory")
	session := flags.String("session", "", "investigation session id")
	repository := flags.String("snapshot", "", "immutable repository snapshot identity")
	var providers stringListFlag
	flags.Var(&providers, "provider", "provider snapshot as name=identity; may be repeated")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*session) == "" || strings.TrimSpace(*repository) == "" {
		return errors.New("--session and --snapshot are required")
	}
	providerSnapshots, err := parseInvestigationProviders(providers)
	if err != nil {
		return err
	}
	statePath, err := resolveState(*root, *state)
	if err != nil {
		return err
	}
	ledger, err := investigation.Create(statePath, *session, investigation.Snapshot{Repository: *repository, Providers: providerSnapshots})
	if err != nil {
		return err
	}
	status, err := ledger.Status()
	if err != nil {
		return err
	}
	return writeJSON(stdout, status)
}

func runInvestigationStatus(args []string, stdout io.Writer, closeSession bool) error {
	flags := flag.NewFlagSet("investigation status", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	root := flags.String("root", ".", "repository root")
	state := flags.String("state", "", "Grimoire state directory")
	session := flags.String("session", "", "investigation session id")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*session) == "" {
		return errors.New("--session is required")
	}
	statePath, err := resolveState(*root, *state)
	if err != nil {
		return err
	}
	ledger, err := investigation.Open(statePath, *session)
	if err != nil {
		return err
	}
	if closeSession {
		if err := ledger.Close(); err != nil {
			return err
		}
	}
	status, err := ledger.Status()
	if err != nil {
		return err
	}
	return writeJSON(stdout, status)
}

func parseInvestigationProviders(values []string) (map[string]string, error) {
	result := make(map[string]string, len(values))
	for _, value := range values {
		parts := strings.SplitN(value, "=", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
			return nil, fmt.Errorf("provider must use name=identity form: %q", value)
		}
		if _, exists := result[strings.TrimSpace(parts[0])]; exists {
			return nil, fmt.Errorf("provider snapshot repeated: %q", parts[0])
		}
		result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return result, nil
}
