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
	if err := s.recoverPending(); err != nil {
		return Report{}, err
	}
	if err := s.Git.ResetIndex(); err != nil {
		return Report{}, err
	}
	manifest, err := s.loadManifest()
	if err != nil {
		return Report{}, err
	}
	if _, err := s.removeLegacyLibrary(); err != nil {
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
	manifest, _ = s.pruneDisabledLanguages(manifest)
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
	manifest, err = s.analyzeFull(ctx, manifest, languages)
	if err != nil {
		return Report{}, err
	}
	snapshotID, err := s.commitManifest(manifest)
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
