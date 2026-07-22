package scan

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Lokee86/lexicon/internal/adapters"
	"github.com/Lokee86/lexicon/internal/config"
	"github.com/Lokee86/lexicon/internal/consumer"
	"github.com/Lokee86/lexicon/internal/lock"
	"github.com/Lokee86/lexicon/internal/objectstore"
	"github.com/Lokee86/lexicon/internal/state"
)

type Scanner struct {
	Repository       string
	StateRoot        string
	AdapterRoot      string
	EnabledLanguages []string
	Git              *state.Repository
	Mirror           state.Mirror
	Analyzer         adapters.Analyzer
	Store            objectstore.Store
	Output           io.Writer
}

type Report struct {
	Changed    []state.Change
	Languages  []string
	SnapshotID string
}

func Initialize(ctx context.Context, repository, adapterRoot string, output io.Writer) (*Scanner, Report, error) {
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
	if err := config.Save(absolute, adapterRoot); err != nil {
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
	if err := scanner.Mirror.SyncAll(absolute); err != nil {
		return nil, Report{}, err
	}
	languages, err := languagesInTree(filepath.Join(stateRoot, "source"))
	if err != nil {
		return nil, Report{}, err
	}
	languages = selectedLanguages(languages, scanner.languageEnabled)
	if _, err := scanner.pruneDisabledLibraries(); err != nil {
		return nil, Report{}, err
	}
	if err := scanner.analyzeFull(ctx, languages); err != nil {
		return nil, Report{}, err
	}
	if err := gitRepository.StageAll(); err != nil {
		return nil, Report{}, err
	}
	if err := gitRepository.CommitState(); err != nil {
		return nil, Report{}, err
	}
	snapshotID, err := scanner.publishSnapshot()
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

func (s *Scanner) Scan(ctx context.Context) (Report, error) {
	report, err := s.scan(ctx, func() error { return s.Mirror.SyncAll(s.Repository) })
	return s.notifyConsumers(ctx, report, err)
}

func (s *Scanner) ScanPaths(ctx context.Context, paths []string) (Report, error) {
	report, err := s.scan(ctx, func() error { return s.Mirror.SyncPaths(s.Repository, paths) })
	return s.notifyConsumers(ctx, report, err)
}

func (s *Scanner) notifyConsumers(ctx context.Context, report Report, scanErr error) (Report, error) {
	if scanErr != nil {
		return report, scanErr
	}
	if err := consumer.Run(ctx, s.Repository, s.Store.Root, report.SnapshotID, s.Output); err != nil {
		return report, err
	}
	return report, nil
}

func (s *Scanner) scan(ctx context.Context, synchronize func() error) (Report, error) {
	guard, err := lock.Acquire(s.Store.Root)
	if err != nil {
		return Report{}, err
	}
	defer guard.Close()
	if err := s.Git.ResetIndex(); err != nil {
		return Report{}, err
	}
	if err := s.Git.RestoreLibrary(); err != nil {
		return Report{}, err
	}
	pruned, err := s.pruneDisabledLibraries()
	if err != nil {
		return Report{}, err
	}
	if err := synchronize(); err != nil {
		return Report{}, err
	}
	if err := s.Git.StageSource(); err != nil {
		return Report{}, err
	}
	changes, err := s.Git.SourceChanges()
	if err != nil {
		return Report{}, err
	}
	drift, err := libraryDriftLanguagesFor(s.StateRoot, s.languageEnabled)
	if err != nil {
		return Report{}, err
	}
	adapterDrift, err := s.adapterDriftLanguages()
	if err != nil {
		return Report{}, err
	}
	drift = mergeLanguages(drift, adapterDrift)
	plans, err := s.plansFor(changes, drift)
	if err != nil {
		return Report{}, err
	}
	if len(changes) == 0 && len(plans) == 0 {
		if pruned {
			if err := s.Git.StageAll(); err != nil {
				return Report{}, err
			}
			if err := s.Git.CommitState(); err != nil {
				return Report{}, err
			}
			snapshotID, err := s.publishSnapshot()
			return Report{SnapshotID: snapshotID}, err
		}
		snapshotID, err := s.ensureSnapshot()
		return Report{SnapshotID: snapshotID}, err
	}
	if err := s.analyzePlans(ctx, plans); err != nil {
		return Report{}, err
	}
	languages := planLanguages(plans)
	if err := s.Git.StageAll(); err != nil {
		return Report{}, err
	}
	if err := s.Git.CommitState(); err != nil {
		return Report{}, err
	}
	snapshotID, err := s.publishSnapshot()
	if err != nil {
		return Report{}, err
	}
	return Report{Changed: changes, Languages: languages, SnapshotID: snapshotID}, nil
}

func (s *Scanner) languageEnabled(language string) bool {
	return config.Config{EnabledLanguages: s.EnabledLanguages}.LanguageEnabled(language)
}

func (s *Scanner) pruneDisabledLibraries() (bool, error) {
	libraryRoot := filepath.Join(s.StateRoot, "library")
	entries, err := os.ReadDir(libraryRoot)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	removed := false
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".jsonl" {
			continue
		}
		language := strings.TrimSuffix(entry.Name(), ".jsonl")
		if s.languageEnabled(language) {
			continue
		}
		if err := os.Remove(filepath.Join(libraryRoot, entry.Name())); err != nil && !os.IsNotExist(err) {
			return false, err
		}
		removed = true
	}
	return removed, nil
}
