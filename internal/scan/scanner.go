package scan

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/Lokee86/lexicon/internal/adapters"
	"github.com/Lokee86/lexicon/internal/config"
	lexfiles "github.com/Lokee86/lexicon/internal/files"
	"github.com/Lokee86/lexicon/internal/state"
)

type Scanner struct {
	Repository string
	StateRoot  string
	Git        *state.Repository
	Mirror     state.Mirror
	Analyzer   adapters.Analyzer
	Output     io.Writer
}

type Report struct {
	Changed   []state.Change
	Languages []string
}

func Initialize(ctx context.Context, repository, adapterRoot string, output io.Writer) (*Scanner, Report, error) {
	absolute, err := filepath.Abs(repository)
	if err != nil {
		return nil, Report{}, err
	}
	if err := config.Save(absolute, adapterRoot); err != nil {
		return nil, Report{}, err
	}
	stateRoot := filepath.Join(config.StateRoot(absolute), "repo")
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
	if err := scanner.analyze(ctx, languages); err != nil {
		return nil, Report{}, err
	}
	if err := gitRepository.StageAll(); err != nil {
		return nil, Report{}, err
	}
	if err := gitRepository.CommitState(); err != nil {
		return nil, Report{}, err
	}
	return scanner, Report{Languages: languages}, nil
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
	if err := s.Git.ResetIndex(); err != nil {
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
	paths := changedPaths(changes)
	languages := lexfiles.CollectLanguages(paths)
	if len(changes) == 0 {
		return Report{}, nil
	}
	if err := s.analyze(ctx, languages); err != nil {
		return Report{}, err
	}
	if err := s.Git.StageAll(); err != nil {
		return Report{}, err
	}
	if err := s.Git.CommitState(); err != nil {
		return Report{}, err
	}
	return Report{Changed: changes, Languages: languages}, nil
}

func (s *Scanner) analyze(ctx context.Context, languages []string) error {
	library := filepath.Join(s.StateRoot, "library")
	temporary := filepath.Join(config.StateRoot(s.Repository), "tmp")
	if err := os.MkdirAll(temporary, 0o755); err != nil {
		return err
	}
	for _, language := range languages {
		tempOutput := filepath.Join(temporary, language+".jsonl")
		_ = os.Remove(tempOutput)
		if s.Output != nil {
			fmt.Fprintf(s.Output, "analyzing %s\n", language)
		}
		if err := s.Analyzer.Run(ctx, language, filepath.Join(s.StateRoot, "source"), tempOutput); err != nil {
			return err
		}
		if err := replace(tempOutput, filepath.Join(library, language+".jsonl")); err != nil {
			return err
		}
	}
	return nil
}

func languagesInTree(root string) ([]string, error) {
	set := make(map[string]struct{})
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if os.IsNotExist(walkErr) {
			return nil
		}
		if walkErr != nil || entry.IsDir() {
			return walkErr
		}
		for _, language := range lexfiles.Languages(path) {
			set[language] = struct{}{}
		}
		return nil
	})
	languages := make([]string, 0, len(set))
	for language := range set {
		languages = append(languages, language)
	}
	sort.Strings(languages)
	return languages, err
}

func changedPaths(changes []state.Change) []string {
	paths := make([]string, 0, len(changes)*2)
	for _, change := range changes {
		if change.Old != "" {
			paths = append(paths, change.Old)
		}
		if change.New != "" {
			paths = append(paths, change.New)
		}
	}
	return paths
}

func replace(source, destination string) error {
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	if err := os.Rename(source, destination); err == nil {
		return nil
	}
	if err := os.Remove(destination); err != nil && !os.IsNotExist(err) {
		return err
	}
	return os.Rename(source, destination)
}
