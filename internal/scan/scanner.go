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

type Scanner struct {
	Repository string
	StateRoot  string
	Git        *state.Repository
	Mirror     state.Mirror
	Analyzer   adapters.Analyzer
	Store      objectstore.Store
	Output     io.Writer
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
	if err := scanner.Mirror.SyncAll(absolute); err != nil {
		return nil, Report{}, err
	}
	languages, err := languagesInTree(filepath.Join(stateRoot, "source"))
	if err != nil {
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
	return New(absolute, stateRoot, gitRepository, adapters.Runner{Root: configuration.AdapterRoot}, output), nil
}

func New(repository, stateRoot string, gitRepository *state.Repository, analyzer adapters.Analyzer, output io.Writer) *Scanner {
	return &Scanner{
		Repository: repository,
		StateRoot:  stateRoot,
		Git:        gitRepository,
		Mirror:     state.Mirror{Root: filepath.Join(stateRoot, "source")},
		Analyzer:   analyzer,
		Store:      objectstore.Store{Root: config.StateRoot(repository)},
		Output:     output,
	}
}

func (s *Scanner) Scan(ctx context.Context) (Report, error) {
	return s.scan(ctx, func() error { return s.Mirror.SyncAll(s.Repository) })
}

func (s *Scanner) ScanPaths(ctx context.Context, paths []string) (Report, error) {
	return s.scan(ctx, func() error { return s.Mirror.SyncPaths(s.Repository, paths) })
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
	drift, err := libraryDriftLanguages(s.StateRoot)
	if err != nil {
		return Report{}, err
	}
	plans, err := s.plansFor(changes, drift)
	if err != nil {
		return Report{}, err
	}
	if len(changes) == 0 && len(plans) == 0 {
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
