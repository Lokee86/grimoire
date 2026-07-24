package scan

import (
	"fmt"
	"sort"

	"github.com/Lokee86/lexicon/internal/objectstore"
)

func (s *Scanner) ensureSnapshot(manifest objectstore.Manifest) (string, error) {
	id, current, err := s.Store.Current()
	if err != nil {
		return "", err
	}
	if !s.Git.HasHead() {
		return "", fmt.Errorf("Lexicon state repository has no commit")
	}
	head, err := s.Git.Head()
	if err != nil {
		return "", err
	}
	if current.StateCommit != head || manifest.StateCommit != head {
		return "", fmt.Errorf("Lexicon snapshot does not match private state commit %s", head)
	}
	return id, nil
}

func (s *Scanner) adapterDriftLanguages(manifest objectstore.Manifest) ([]string, error) {
	if s.AdapterRoot == "" {
		return nil, nil
	}
	drift := make([]string, 0)
	for _, language := range manifest.Languages {
		fingerprint, err := s.adapterFingerprint(language.Language)
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
