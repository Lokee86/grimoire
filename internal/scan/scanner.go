package scan

import (
	"context"
	"io"

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
	legacyRemoved, err := s.removeLegacyLibrary()
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
	manifest, pruned := s.pruneDisabledLanguages(manifest)
	drift, err := snapshotDriftLanguages(s.StateRoot, manifest, s.languageEnabled)
	if err != nil {
		return Report{}, err
	}
	adapterDrift, err := s.adapterDriftLanguages(manifest)
	if err != nil {
		return Report{}, err
	}
	drift = mergeLanguages(drift, adapterDrift)
	plans, err := s.plansFor(changes, drift)
	if err != nil {
		return Report{}, err
	}
	if len(changes) == 0 && len(plans) == 0 && !pruned && !legacyRemoved {
		snapshotID, err := s.ensureSnapshot(manifest)
		return Report{SnapshotID: snapshotID}, err
	}
	manifest, err = s.analyzePlans(ctx, manifest, plans)
	if err != nil {
		return Report{}, err
	}
	snapshotID, err := s.commitManifest(manifest)
	if err != nil {
		return Report{}, err
	}
	return Report{Changed: changes, Languages: planLanguages(plans), SnapshotID: snapshotID}, nil
}

func (s *Scanner) languageEnabled(language string) bool {
	return config.Config{EnabledLanguages: s.EnabledLanguages}.LanguageEnabled(language)
}

func (s *Scanner) pruneDisabledLanguages(manifest objectstore.Manifest) (objectstore.Manifest, bool) {
	pruned := false
	for _, entry := range append([]objectstore.LanguageEntry(nil), manifest.Languages...) {
		if s.languageEnabled(entry.Language) {
			continue
		}
		manifest = manifest.WithoutLanguage(entry.Language)
		pruned = true
	}
	return manifest, pruned
}
