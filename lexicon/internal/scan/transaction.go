package scan

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Lokee86/lexicon/internal/config"
	"github.com/Lokee86/lexicon/internal/objectstore"
)

func (s *Scanner) recoverPending() error {
	pending, err := s.Store.Pending()
	if errors.Is(err, objectstore.ErrNoPendingPublication) {
		return nil
	}
	if err != nil {
		return err
	}
	head := ""
	if s.Git.HasHead() {
		head, err = s.Git.Head()
		if err != nil {
			return err
		}
	}
	if pending.CommitRequired && (head == "" || head == pending.BaseStateCommit) {
		return s.Store.ClearPending()
	}
	manifest := pending.Manifest
	manifest.StateCommit = head
	if _, err := s.Store.Publish(manifest); err != nil {
		return fmt.Errorf("recover pending Lexicon publication: %w", err)
	}
	return s.Store.ClearPending()
}

func (s *Scanner) loadManifest() (objectstore.Manifest, error) {
	_, manifest, err := s.Store.Current()
	if errors.Is(err, objectstore.ErrNoCurrentSnapshot) {
		if s.Git.HasHead() && legacyLibraryExists(s.StateRoot) {
			head, headErr := s.Git.Head()
			if headErr != nil {
				return objectstore.Manifest{}, headErr
			}
			manifest, err = s.Store.BuildManifest(s.StateRoot, head, config.AnalysisID(), s.AdapterRoot)
			if err == nil {
				if _, err := s.Store.Publish(manifest); err != nil {
					return objectstore.Manifest{}, err
				}
				return manifest, nil
			}
			// A corrupt legacy materialization is not authoritative. The caller
			// removes it and rebuilds the missing languages from source.
			return objectstore.Manifest{Version: objectstore.SnapshotVersion}, nil
		}
		return objectstore.Manifest{Version: objectstore.SnapshotVersion}, nil
	}
	if err != nil {
		return objectstore.Manifest{}, err
	}
	if s.Git.HasHead() {
		head, err := s.Git.Head()
		if err != nil {
			return objectstore.Manifest{}, err
		}
		if manifest.StateCommit != head {
			if legacyLibraryExists(s.StateRoot) {
				migrated, migrateErr := s.Store.BuildManifest(s.StateRoot, head, config.AnalysisID(), s.AdapterRoot)
				if migrateErr == nil {
					if _, publishErr := s.Store.Publish(migrated); publishErr != nil {
						return objectstore.Manifest{}, publishErr
					}
					return migrated, nil
				}
			}
			return objectstore.Manifest{}, fmt.Errorf(
				"Lexicon snapshot state %s does not match private state %s and no recoverable publication exists",
				manifest.StateCommit,
				head,
			)
		}
	}
	return manifest, nil
}

func (s *Scanner) commitManifest(manifest objectstore.Manifest) (string, error) {
	if err := s.Git.StageAll(); err != nil {
		return "", err
	}
	baseHead := ""
	if s.Git.HasHead() {
		var err error
		baseHead, err = s.Git.Head()
		if err != nil {
			return "", err
		}
	}
	commitRequired := !s.Git.HasHead() || s.Git.HasStagedChanges()
	if err := s.Store.WritePending(baseHead, commitRequired, manifest); err != nil {
		return "", err
	}
	if err := s.Git.CommitState(); err != nil {
		return "", err
	}
	head, err := s.Git.Head()
	if err != nil {
		return "", err
	}
	manifest.StateCommit = head
	id, err := s.Store.Publish(manifest)
	if err != nil {
		return "", err
	}
	if err := s.Store.ClearPending(); err != nil {
		return "", err
	}
	return id, nil
}

func (s *Scanner) removeLegacyLibrary() (bool, error) {
	path := filepath.Join(s.StateRoot, "library")
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := os.RemoveAll(path); err != nil {
		return false, err
	}
	return true, nil
}

func legacyLibraryExists(stateRoot string) bool {
	entries, err := os.ReadDir(filepath.Join(stateRoot, "library"))
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".jsonl" {
			return true
		}
	}
	return false
}
