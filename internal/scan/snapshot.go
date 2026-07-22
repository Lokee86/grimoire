package scan

import (
	"errors"
	"sort"

	"github.com/Lokee86/lexicon/internal/adapters"
	"github.com/Lokee86/lexicon/internal/config"
	"github.com/Lokee86/lexicon/internal/objectstore"
)

func (s *Scanner) ensureSnapshot() (string, error) {
	head, err := s.Git.Head()
	if err != nil {
		return "", err
	}
	id, manifest, err := s.Store.Current()
	if err == nil && manifest.StateCommit == head {
		return id, nil
	}
	return s.publishSnapshot()
}

func (s *Scanner) publishSnapshot() (string, error) {
	head, err := s.Git.Head()
	if err != nil {
		return "", err
	}
	manifest, err := s.Store.BuildManifest(s.StateRoot, head, config.AnalysisID(), s.AdapterRoot)
	if err != nil {
		return "", err
	}
	return s.Store.Publish(manifest)
}

func (s *Scanner) adapterDriftLanguages() ([]string, error) {
	if s.AdapterRoot == "" {
		return nil, nil
	}
	_, manifest, err := s.Store.Current()
	if errors.Is(err, objectstore.ErrNoCurrentSnapshot) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	drift := make([]string, 0)
	for _, language := range manifest.Languages {
		fingerprint, err := adapters.Fingerprint(s.AdapterRoot, language.Language)
		if err != nil {
			return nil, err
		}
		if fingerprint != language.AdapterFingerprint {
			drift = append(drift, language.Language)
		}
	}
	sort.Strings(drift)
	return drift, nil
}
