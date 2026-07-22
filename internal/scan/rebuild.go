package scan

import (
	"context"
	"fmt"
	"path/filepath"

	languageRegistry "github.com/Lokee86/lexicon/internal/languages"
	"github.com/Lokee86/lexicon/internal/lock"
)

func (s *Scanner) Rebuild(ctx context.Context, languages []string) (Report, error) {
	report, err := s.rebuild(ctx, languages)
	return s.notifyConsumers(ctx, report, err)
}

func (s *Scanner) rebuild(ctx context.Context, languages []string) (Report, error) {
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
	if _, err := s.pruneDisabledLibraries(); err != nil {
		return Report{}, err
	}
	if err := s.Mirror.SyncAll(s.Repository); err != nil {
		return Report{}, err
	}
	if err := s.Git.StageSource(); err != nil {
		return Report{}, err
	}
	changes, err := s.Git.SourceChanges()
	if err != nil {
		return Report{}, err
	}
	if len(languages) == 0 {
		languages, err = languagesInTree(filepath.Join(s.StateRoot, "source"))
		if err != nil {
			return Report{}, err
		}
		languages = selectedLanguages(languages, s.languageEnabled)
	} else {
		languages, err = s.validateRebuildLanguages(languages)
		if err != nil {
			return Report{}, err
		}
	}
	plans := make([]analysisPlan, 0, len(languages))
	for _, language := range languages {
		plans = append(plans, analysisPlan{Language: language, Full: true})
	}
	if err := s.analyzePlans(ctx, plans); err != nil {
		return Report{}, err
	}
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

func (s *Scanner) validateRebuildLanguages(languages []string) ([]string, error) {
	supported := make(map[string]struct{}, len(languageRegistry.Supported()))
	for _, language := range languageRegistry.Supported() {
		supported[language] = struct{}{}
	}
	for _, language := range languages {
		if _, ok := supported[language]; !ok {
			return nil, fmt.Errorf("unsupported Lexicon language %q", language)
		}
		if !s.languageEnabled(language) {
			return nil, fmt.Errorf("Lexicon language %q is disabled", language)
		}
	}
	return uniqueSorted(languages), nil
}
