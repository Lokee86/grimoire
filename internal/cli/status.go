package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Lokee86/lexicon/internal/config"
	"github.com/Lokee86/lexicon/internal/objectstore"
)

func runStatus(arguments []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("status", flag.ContinueOnError)
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

	store := objectstore.Store{Root: config.StateRoot(root)}
	snapshotID, manifest, err := store.Current()
	if errors.Is(err, objectstore.ErrNoCurrentSnapshot) {
		snapshotID = "none"
	} else if err != nil {
		return err
	}
	consumers, err := registeredConsumers(root)
	if err != nil {
		return err
	}
	languages := make([]string, 0, len(manifest.Languages))
	for _, language := range manifest.Languages {
		languages = append(languages, language.Language)
	}
	sort.Strings(languages)
	fmt.Fprintf(stdout, "repository root: %s\n", root)
	fmt.Fprintf(stdout, "current snapshot ID: %s\n", snapshotID)
	fmt.Fprintf(stdout, "detected languages: %s\n", displayList(languages))
	fmt.Fprintf(stdout, "enabled languages: %s\n", displayLanguageSelection(configuration.EnabledLanguages))
	fmt.Fprintf(stdout, "registered consumer names: %s\n", displayList(consumers))
	return nil
}

func registeredConsumers(repository string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(config.StateRoot(repository), "consumers"))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read Lexicon consumers: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		names = append(names, strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())))
	}
	sort.Strings(names)
	return names, nil
}

func displayList(values []string) string {
	if len(values) == 0 {
		return "none"
	}
	return strings.Join(values, ", ")
}
