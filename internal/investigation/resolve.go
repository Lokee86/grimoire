package investigation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func (l *Ledger) ResolveNodeHandle(token string) (Node, error) {
	var value Node
	err := l.resolveHandle(token, "node", &value)
	return value, err
}

func (l *Ledger) ResolveSourceRangeHandle(token string) (SourceRange, error) {
	var value SourceRange
	err := l.resolveHandle(token, "source", &value)
	return value, err
}

func (l *Ledger) ResolveGraphPathHandle(token string) (GraphPath, error) {
	var value GraphPath
	err := l.resolveHandle(token, "path", &value)
	return value, err
}

func (l *Ledger) ResolveDocumentHandle(token string) (Document, error) {
	var value Document
	err := l.resolveHandle(token, "document", &value)
	return value, err
}

func (l *Ledger) resolveHandle(token, kind string, target any) error {
	decoded, err := decodeHandle(token, kind)
	if err != nil {
		return err
	}
	unlock, err := acquireLock(l.dir)
	if err != nil {
		return err
	}
	defer unlock()
	current, err := loadManifest(l.dir)
	if err != nil {
		return err
	}
	if err := l.checkOwner(current); err != nil {
		return err
	}
	if decoded.Snapshot != "" && decoded.Snapshot != current.Snapshot.Digest() {
		return ErrSnapshotMismatch
	}
	meta, exists := current.Evidence[decoded.Digest]
	if !exists || meta.Kind != kind {
		return fmt.Errorf("%w: %s handle evidence is unavailable", ErrCorrupt, kind)
	}
	data, err := os.ReadFile(filepath.Join(l.dir, filepath.FromSlash(meta.File)))
	if err != nil {
		return fmt.Errorf("read %s handle evidence: %w", kind, err)
	}
	var stored storedEvidence
	if err := json.Unmarshal(data, &stored); err != nil {
		return fmt.Errorf("%w: decode %s handle evidence: %v", ErrCorrupt, kind, err)
	}
	if stored.Kind != kind || stored.Key != decoded.Digest || stored.Snapshot != current.Snapshot.Digest() {
		return fmt.Errorf("%w: %s handle evidence identity does not match", ErrCorrupt, kind)
	}
	if err := json.Unmarshal(stored.Payload, target); err != nil {
		return fmt.Errorf("%w: decode %s handle payload: %v", ErrCorrupt, kind, err)
	}
	return nil
}
