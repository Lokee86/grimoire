package scan

import "github.com/Lokee86/lexicon/internal/config"

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
	manifest, err := s.Store.BuildManifest(s.StateRoot, head, config.AnalysisID())
	if err != nil {
		return "", err
	}
	return s.Store.Publish(manifest)
}
