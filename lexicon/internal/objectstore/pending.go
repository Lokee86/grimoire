package objectstore

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const pendingVersion = 1

var ErrNoPendingPublication = errors.New("Lexicon has no pending publication")

type PendingPublication struct {
	Version         int      `json:"version"`
	BaseStateCommit string   `json:"base_state_commit,omitempty"`
	CommitRequired  bool     `json:"commit_required"`
	Manifest        Manifest `json:"manifest"`
}

func (s Store) WritePending(baseStateCommit string, commitRequired bool, manifest Manifest) error {
	manifest.StateCommit = ""
	pending := PendingPublication{
		Version: pendingVersion, BaseStateCommit: baseStateCommit,
		CommitRequired: commitRequired, Manifest: manifest,
	}
	data, err := json.Marshal(pending)
	if err != nil {
		return fmt.Errorf("encode pending Lexicon publication: %w", err)
	}
	return writeAtomic(s.pendingPath(), append(data, '\n'))
}

func (s Store) Pending() (PendingPublication, error) {
	data, err := os.ReadFile(s.pendingPath())
	if os.IsNotExist(err) {
		return PendingPublication{}, ErrNoPendingPublication
	}
	if err != nil {
		return PendingPublication{}, err
	}
	var pending PendingPublication
	if err := json.Unmarshal(bytes.TrimSpace(data), &pending); err != nil {
		return PendingPublication{}, fmt.Errorf("decode pending Lexicon publication: %w", err)
	}
	if pending.Version != pendingVersion {
		return PendingPublication{}, fmt.Errorf("unsupported pending Lexicon publication version %d", pending.Version)
	}
	return pending, nil
}

func (s Store) ClearPending() error {
	err := os.Remove(s.pendingPath())
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (s Store) pendingPath() string {
	return filepath.Join(s.Root, "PENDING")
}
