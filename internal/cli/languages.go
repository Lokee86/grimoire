package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/Lokee86/lexicon/internal/config"
	"github.com/Lokee86/lexicon/internal/scan"
)

func runLanguages(ctx context.Context, arguments []string, stdout, stderr io.Writer) error {
	if len(arguments) > 0 && arguments[0] == "set" {
		return runLanguagesSet(ctx, arguments[1:], stdout, stderr)
	}
	if len(arguments) > 0 && arguments[0] == "list" {
		arguments = arguments[1:]
	}
	flags := flag.NewFlagSet("languages", flag.ContinueOnError)
	flags.SetOutput(stderr)
	repository := flags.String("repo", "", "repository to inspect")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	root, err := resolveRepository(*repository)
	if err != nil {
		return err
	}
	configuration, err := config.Load(root)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "enabled languages: %s\n", displayLanguageSelection(configuration.EnabledLanguages))
	return nil
}

func runLanguagesSet(ctx context.Context, arguments []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("languages set", flag.ContinueOnError)
	flags.SetOutput(stderr)
	repository := flags.String("repo", "", "repository to configure")
	languages := flags.String("languages", "", "comma-separated languages or all")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if !flagWasSet(flags, "languages") {
		return fmt.Errorf("languages set requires --languages")
	}
	root, err := resolveRepository(*repository)
	if err != nil {
		return err
	}
	selection, err := parseLanguageSelection(*languages)
	if err != nil {
		return err
	}
	if err := config.UpdateEnabledLanguages(root, selection); err != nil {
		return err
	}
	scanner, err := scan.Open(root, stdout)
	if err != nil {
		return err
	}
	report, err := scanner.Scan(ctx)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "enabled languages: %s\n", displayLanguageSelection(selection))
	fmt.Fprintf(stdout, "snapshot: %s\n", report.SnapshotID)
	return nil
}

func parseLanguageSelection(value string) ([]string, error) {
	value = strings.TrimSpace(value)
	if value == "" || strings.EqualFold(value, "all") {
		return nil, nil
	}
	parts := strings.Split(value, ",")
	languages := make([]string, 0, len(parts))
	for _, part := range parts {
		language := strings.ToLower(strings.TrimSpace(part))
		if language == "" {
			return nil, fmt.Errorf("language list contains an empty value")
		}
		languages = append(languages, language)
	}
	return config.NormalizeEnabledLanguages(languages)
}

func displayLanguageSelection(languages []string) string {
	if len(languages) == 0 {
		return "all"
	}
	return strings.Join(languages, ", ")
}

func flagWasSet(flags *flag.FlagSet, name string) bool {
	set := false
	flags.Visit(func(current *flag.Flag) {
		if current.Name == name {
			set = true
		}
	})
	return set
}
