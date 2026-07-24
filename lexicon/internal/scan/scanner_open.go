package scan

import (
	"context"
	"io"
	"path/filepath"

	"github.com/Lokee86/lexicon/internal/adapters"
	"github.com/Lokee86/lexicon/internal/config"
	"github.com/Lokee86/lexicon/internal/lock"
	"github.com/Lokee86/lexicon/internal/objectstore"
	"github.com/Lokee86/lexicon/internal/state"
)

func Initialize(ctx context.Context, repository, adapterRoot string, output io.Writer) (*Scanner, Report, error) {
	return initialize(ctx, repository, adapterRoot, nil, false, output)
}

func InitializeWithLanguages(
	ctx context.Context,
	repository, adapterRoot string,
	enabledLanguages []string,
	output io.Writer,
) (*Scanner, Report, error) {
	return initialize(ctx, repository, adapterRoot, enabledLanguages, true, output)
}

func initialize(
	ctx context.Context,
	repository, adapterRoot string,
	enabledLanguages []string,
	explicitSelection bool,
	output io.Writer,
) (*Scanner, Report, error) {
	absolute, err := filepath.Abs(repository)
	if err != nil {
		return nil, Report{}, err
	}
	lexiconRoot := config.StateRoot(absolute)
	guard, err := lock.Acquire(lexiconRoot)
	if err != nil {
		return nil, Report{}, err
	}
	defer guard.Close()
	if explicitSelection {
		err = config.SaveWithEnabledLanguages(absolute, adapterRoot, enabledLanguages)
	} else {
		err = config.Save(absolute, adapterRoot)
	}
	if err != nil {
		return nil, Report{}, err
	}
	stateRoot := filepath.Join(lexiconRoot, "repo")
	gitRepository, err := state.Ensure(stateRoot)
	if err != nil {
		return nil, Report{}, err
	}
	scanner := New(absolute, stateRoot, gitRepository, adapters.Runner{Root: adapterRoot}, output)
	configuration, err := config.Load(absolute)
	if err != nil {
		return nil, Report{}, err
	}
	scanner.EnabledLanguages = configuration.EnabledLanguages
	if err := scanner.recoverPending(); err != nil {
		return nil, Report{}, err
	}
	if err := scanner.Mirror.SyncAll(absolute); err != nil {
		return nil, Report{}, err
	}
	if _, err := scanner.removeLegacyLibrary(); err != nil {
		return nil, Report{}, err
	}
	languages, err := languagesInTree(filepath.Join(stateRoot, "source"))
	if err != nil {
		return nil, Report{}, err
	}
	languages = selectedLanguages(languages, scanner.languageEnabled)
	manifest, err := scanner.analyzeFull(
		ctx,
		objectstore.Manifest{Version: objectstore.SnapshotVersion},
		languages,
	)
	if err != nil {
		return nil, Report{}, err
	}
	snapshotID, err := scanner.commitManifest(manifest)
	if err != nil {
		return nil, Report{}, err
	}
	return scanner, Report{Languages: languages, SnapshotID: snapshotID}, nil
}

func Open(repository string, output io.Writer) (*Scanner, error) {
	absolute, err := filepath.Abs(repository)
	if err != nil {
		return nil, err
	}
	configuration, err := config.Load(absolute)
	if err != nil {
		return nil, err
	}
	stateRoot := filepath.Join(config.StateRoot(absolute), "repo")
	gitRepository, err := state.Open(stateRoot)
	if err != nil {
		return nil, err
	}
	scanner := New(absolute, stateRoot, gitRepository, adapters.Runner{Root: configuration.AdapterRoot}, output)
	scanner.EnabledLanguages = configuration.EnabledLanguages
	return scanner, nil
}

func New(repository, stateRoot string, gitRepository *state.Repository, analyzer adapters.Analyzer, output io.Writer) *Scanner {
	return &Scanner{
		Repository:  repository,
		StateRoot:   stateRoot,
		AdapterRoot: adapterRoot(analyzer),
		Git:         gitRepository,
		Mirror:      state.Mirror{Root: filepath.Join(stateRoot, "source")},
		Analyzer:    analyzer,
		Store:       objectstore.Store{Root: config.StateRoot(repository)},
		Output:      output,
	}
}

func adapterRoot(analyzer adapters.Analyzer) string {
	switch value := analyzer.(type) {
	case adapters.Runner:
		return value.Root
	case *adapters.Runner:
		if value != nil {
			return value.Root
		}
	}
	return ""
}
