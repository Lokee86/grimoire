package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Lokee86/lexicon/internal/config"
	"github.com/Lokee86/lexicon/internal/consumer"
	"github.com/Lokee86/lexicon/internal/objectstore"
)

type doctorCheck struct {
	label string
	err   error
}

var (
	doctorLoadConfig      = config.Load
	doctorVerifyState     = verifyStateRepository
	doctorCurrentSnapshot = func(store objectstore.Store) (string, objectstore.Manifest, error) {
		return store.Current()
	}
	doctorLoadObject       = func(store objectstore.Store, id string) (objectstore.FactObject, error) { return store.LoadObject(id) }
	doctorLookPath         = exec.LookPath
	doctorValidateConsumer = consumer.Validate
)

func runDoctor(arguments []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("doctor", flag.ContinueOnError)
	flags.SetOutput(stderr)
	repository := flags.String("repo", "", "repository to inspect")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	root, err := resolveRepository(*repository)
	if err != nil {
		fmt.Fprintf(stdout, "FAIL repository discovery: %s\n", err)
		return err
	}
	return doctorAt(root, stdout)
}

func doctorAt(repository string, output io.Writer) error {
	checks := make([]doctorCheck, 0)
	report := func(label string, err error) {
		checks = append(checks, doctorCheck{label: label, err: err})
		if err == nil {
			fmt.Fprintf(output, "PASS %s\n", label)
			return
		}
		fmt.Fprintf(output, "FAIL %s: %s\n", label, oneLine(err))
	}

	configuration, configErr := doctorLoadConfig(repository)
	report("configuration loading", configErr)
	report("private Git state repository", doctorVerifyState(filepath.Join(config.StateRoot(repository), "repo")))

	store := objectstore.Store{Root: config.StateRoot(repository)}
	manifest, snapshotErr := verifySnapshot(store)
	report("CURRENT snapshot and referenced objects", snapshotErr)

	adapterRootOK := false
	adapterRootErr := configErr
	if configErr == nil {
		adapterRootOK = directoryExists(configuration.AdapterRoot)
		if !adapterRootOK {
			adapterRootErr = fmt.Errorf("adapter root is not a directory: %s", configuration.AdapterRoot)
		}
	}
	report("configured adapter root", adapterRootErr)

	languages := manifestLanguages(manifest)
	if snapshotErr != nil && len(languages) == 0 {
		report("detected language adapter directories", fmt.Errorf("snapshot unavailable"))
		report("required runtime executables", fmt.Errorf("snapshot unavailable"))
	} else if len(languages) == 0 {
		report("detected language adapter directories", nil)
		report("required runtime executables", nil)
	} else {
		for _, language := range languages {
			if !adapterRootOK {
				report("adapter directory: "+language, fmt.Errorf("configured adapter root is unavailable"))
			} else {
				report("adapter directory: "+language, checkAdapterDirectory(configuration.AdapterRoot, language))
			}
			report("runtime executable: "+language, checkRuntime(configuration.AdapterRoot, language))
		}
	}

	consumerPaths, consumerErr := consumerDefinitionPaths(repository)
	if consumerErr != nil {
		report("registered consumer definitions", consumerErr)
	} else if len(consumerPaths) == 0 {
		report("registered consumer definitions", nil)
	} else {
		for _, path := range consumerPaths {
			name := filepath.Base(path)
			definition, err := doctorValidateConsumer(path)
			report("consumer definition: "+name, err)
			if err != nil {
				report("consumer command: "+name, fmt.Errorf("definition unavailable"))
				continue
			}
			report("consumer command: "+name, checkConsumerCommand(definition.Command))
		}
	}

	failures := make([]error, 0)
	for _, check := range checks {
		if check.err != nil {
			failures = append(failures, fmt.Errorf("%s: %w", check.label, check.err))
		}
	}
	return errors.Join(failures...)
}

func oneLine(err error) string {
	return strings.ReplaceAll(filepath.ToSlash(err.Error()), "\n", "; ")
}
