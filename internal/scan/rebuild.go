package scan

import (
	"context"
	"path/filepath"

	"github.com/Lokee86/lexicon/internal/lock"
)

func (s *Scanner) Rebuild(ctx context.Context, languages []string) (Report, error) {
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
	} else {
		languages = uniqueSorted(languages)
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
