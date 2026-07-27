package investigation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Ledger struct {
	stateDir  string
	sessionID string
	dir       string
	snapshot  Snapshot
}

type manifest struct {
	Version   int                     `json:"version"`
	SessionID string                  `json:"session_id"`
	Snapshot  Snapshot                `json:"snapshot"`
	CreatedAt string                  `json:"created_at"`
	UpdatedAt string                  `json:"updated_at"`
	ClosedAt  string                  `json:"closed_at,omitempty"`
	Evidence  map[string]evidenceMeta `json:"evidence"`
	Responses map[string]responseMeta `json:"responses"`
}
type evidenceMeta struct {
	Kind     string `json:"kind"`
	File     string `json:"file"`
	Handle   string `json:"handle,omitempty"`
	Snapshot string `json:"snapshot"`
}
type responseMeta struct {
	File     string `json:"file"`
	Snapshot string `json:"snapshot"`
}
type storedEvidence struct {
	Version  int             `json:"version"`
	Kind     string          `json:"kind"`
	Snapshot string          `json:"snapshot"`
	Key      string          `json:"key"`
	Payload  json.RawMessage `json:"payload"`
	Digest   string          `json:"digest"`
}
type storedResponse struct {
	Version            int      `json:"version"`
	ResponseID         string   `json:"response_id"`
	Snapshot           string   `json:"snapshot"`
	Digest             string   `json:"digest"`
	NodeKeys           []string `json:"node_keys,omitempty"`
	SourceRangeKeys    []string `json:"source_range_keys,omitempty"`
	GraphPathKeys      []string `json:"graph_path_keys,omitempty"`
	DocumentKeys       []string `json:"document_keys,omitempty"`
	QuestionKeys       []string `json:"question_keys,omitempty"`
	RejectedBranchKeys []string `json:"rejected_branch_keys,omitempty"`
	AcceptedBranchKeys []string `json:"accepted_branch_keys,omitempty"`
}

func Create(stateDir, sessionID string, snapshot Snapshot) (*Ledger, error) {
	if err := snapshot.Validate(); err != nil {
		return nil, err
	}
	if err := validateSessionID(sessionID); err != nil {
		return nil, err
	}
	stateDir, err := filepath.Abs(stateDir)
	if err != nil {
		return nil, fmt.Errorf("resolve investigation state: %w", err)
	}
	parent := filepath.Join(stateDir, "investigations")
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return nil, fmt.Errorf("create investigations directory: %w", err)
	}
	finalDir := filepath.Join(parent, sessionID)
	tempDir, err := os.MkdirTemp(parent, ".session-")
	if err != nil {
		return nil, fmt.Errorf("create session staging directory: %w", err)
	}
	defer os.RemoveAll(tempDir)
	if _, err := os.Stat(finalDir); err == nil {
		return nil, ErrSessionExists
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	current := manifest{Version: ledgerVersion, SessionID: sessionID, Snapshot: snapshot.normalized(), CreatedAt: now, UpdatedAt: now, Evidence: map[string]evidenceMeta{}, Responses: map[string]responseMeta{}}
	if err := writeJSONAtomic(filepath.Join(tempDir, "manifest.json"), current); err != nil {
		return nil, err
	}
	if err := os.Mkdir(filepath.Join(tempDir, "records"), 0o755); err != nil {
		return nil, err
	}
	if err := os.Mkdir(filepath.Join(tempDir, "responses"), 0o755); err != nil {
		return nil, err
	}
	if err := os.Rename(tempDir, finalDir); err != nil {
		if _, statErr := os.Stat(finalDir); statErr == nil {
			return nil, ErrSessionExists
		}
		return nil, fmt.Errorf("publish investigation session: %w", err)
	}
	return &Ledger{stateDir: stateDir, sessionID: sessionID, dir: finalDir, snapshot: current.Snapshot}, nil
}

func Open(stateDir, sessionID string, expected ...Snapshot) (*Ledger, error) {
	if err := validateSessionID(sessionID); err != nil {
		return nil, err
	}
	stateDir, err := filepath.Abs(stateDir)
	if err != nil {
		return nil, fmt.Errorf("resolve investigation state: %w", err)
	}
	dir := filepath.Join(stateDir, "investigations", sessionID)
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		return nil, ErrSessionNotFound
	} else if err != nil {
		return nil, fmt.Errorf("inspect investigation session: %w", err)
	}
	unlock, err := acquireLock(dir)
	if err != nil {
		return nil, err
	}
	defer unlock()
	current, err := loadManifest(dir)
	if err != nil {
		return nil, err
	}
	if current.SessionID != sessionID {
		return nil, fmt.Errorf("%w: session id does not match its directory", ErrCorrupt)
	}
	if len(expected) > 1 {
		return nil, errors.New("at most one expected snapshot may be supplied")
	}
	if len(expected) == 1 && !current.Snapshot.Equal(expected[0]) {
		return nil, ErrSnapshotMismatch
	}
	return &Ledger{stateDir: stateDir, sessionID: sessionID, dir: dir, snapshot: current.Snapshot}, nil
}

func CreateSession(stateDir, sessionID string, snapshot Snapshot) (*Ledger, error) {
	return Create(stateDir, sessionID, snapshot)
}
func OpenSession(stateDir, sessionID string, expected ...Snapshot) (*Ledger, error) {
	return Open(stateDir, sessionID, expected...)
}

func (l *Ledger) SessionID() string  { return l.sessionID }
func (l *Ledger) Snapshot() Snapshot { return l.snapshot.normalized() }
func (l *Ledger) Path() string       { return l.dir }

func (l *Ledger) Status() (Status, error) {
	unlock, err := acquireLock(l.dir)
	if err != nil {
		return Status{}, err
	}
	defer unlock()
	current, err := loadManifest(l.dir)
	if err != nil {
		return Status{}, err
	}
	if err := l.checkOwner(current); err != nil {
		return Status{}, err
	}
	return statusFromManifest(current), nil
}

func (l *Ledger) Close() error {
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
	if current.ClosedAt != "" {
		return nil
	}
	current.ClosedAt = time.Now().UTC().Format(time.RFC3339Nano)
	current.UpdatedAt = current.ClosedAt
	return writeManifest(l.dir, current)
}

func (l *Ledger) RecordResponse(response Response) (Delta, error) {
	if err := responseValidate(response); err != nil {
		return Delta{}, err
	}
	if !l.snapshot.Equal(response.Snapshot) {
		return Delta{}, ErrSnapshotMismatch
	}
	unlock, err := acquireLock(l.dir)
	if err != nil {
		return Delta{}, err
	}
	defer unlock()
	current, err := loadManifest(l.dir)
	if err != nil {
		return Delta{}, err
	}
	if err := l.checkOwner(current); err != nil {
		return Delta{}, err
	}
	if current.ClosedAt != "" {
		return Delta{}, ErrSessionClosed
	}
	if !current.Snapshot.Equal(response.Snapshot) {
		return Delta{}, ErrSnapshotMismatch
	}
	responseID := responseDigest(response)
	if _, exists := current.Responses[responseID]; exists {
		return buildDelta(current, response, responseID)
	}
	result, event, err := recordEvidence(l.dir, current, response)
	if err != nil {
		return Delta{}, err
	}
	event.Digest = responseRecordDigest(event)
	responsePath := filepath.Join(l.dir, "responses", responseID+".json")
	if err := writeJSONAtomic(responsePath, event); err != nil {
		return Delta{}, err
	}
	current.Responses[responseID] = responseMeta{File: filepath.ToSlash(filepath.Join("responses", responseID+".json")), Snapshot: current.Snapshot.Digest()}
	current.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	if err := writeManifest(l.dir, current); err != nil {
		return Delta{}, err
	}
	result.ResponseID = responseID
	return result, nil
}

func (l *Ledger) DeltaFor(response Response) (Delta, error) {
	if err := responseValidate(response); err != nil {
		return Delta{}, err
	}
	if !l.snapshot.Equal(response.Snapshot) {
		return Delta{}, ErrSnapshotMismatch
	}
	unlock, err := acquireLock(l.dir)
	if err != nil {
		return Delta{}, err
	}
	defer unlock()
	current, err := loadManifest(l.dir)
	if err != nil {
		return Delta{}, err
	}
	if err := l.checkOwner(current); err != nil {
		return Delta{}, err
	}
	return buildDelta(current, response, responseDigest(response))
}

func (l *Ledger) EvidenceAlreadyReturned(response Response) (bool, error) {
	if err := responseValidate(response); err != nil {
		return false, err
	}
	if !l.snapshot.Equal(response.Snapshot) {
		return false, ErrSnapshotMismatch
	}
	unlock, err := acquireLock(l.dir)
	if err != nil {
		return false, err
	}
	defer unlock()
	current, err := loadManifest(l.dir)
	if err != nil {
		return false, err
	}
	if err := l.checkOwner(current); err != nil {
		return false, err
	}
	if _, exists := current.Responses[responseDigest(response)]; exists {
		return true, nil
	}
	keys := responseKeys(response)
	if len(keys) == 0 {
		return false, nil
	}
	for _, key := range keys {
		if _, exists := current.Evidence[key]; !exists {
			return false, nil
		}
	}
	return true, nil
}

func (l *Ledger) AlreadyReturned(response Response) (bool, error) {
	return l.EvidenceAlreadyReturned(response)
}

func (l *Ledger) checkOwner(current manifest) error {
	if current.SessionID != l.sessionID {
		return fmt.Errorf("%w: session id does not match its directory", ErrCorrupt)
	}
	if !current.Snapshot.Equal(l.snapshot) {
		return ErrSnapshotMismatch
	}
	return nil
}

func responseDigest(response Response) string {
	data, _ := json.Marshal(response)
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func validateSessionID(value string) error {
	if strings.TrimSpace(value) == "" || value == "." || value == ".." || filepath.Base(value) != value || strings.ContainsAny(value, `/\\`) {
		return errors.New("invalid investigation session id")
	}
	return nil
}

func loadManifest(dir string) (manifest, error) {
	data, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if errors.Is(err, os.ErrNotExist) {
		return manifest{}, ErrSessionNotFound
	}
	if err != nil {
		return manifest{}, fmt.Errorf("read investigation manifest: %w", err)
	}
	var current manifest
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&current); err != nil {
		return manifest{}, fmt.Errorf("%w: decode manifest: %v", ErrCorrupt, err)
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return manifest{}, fmt.Errorf("%w: trailing manifest data", ErrCorrupt)
	}
	if current.Version != ledgerVersion || current.SessionID == "" || current.Evidence == nil || current.Responses == nil {
		return manifest{}, fmt.Errorf("%w: invalid manifest identity", ErrCorrupt)
	}
	if err := current.Snapshot.Validate(); err != nil {
		return manifest{}, fmt.Errorf("%w: invalid snapshot: %v", ErrCorrupt, err)
	}
	if err := validateManifestFiles(dir, current); err != nil {
		return manifest{}, err
	}
	return current, nil
}

func validateManifestFiles(dir string, current manifest) error {
	for key, meta := range current.Evidence {
		if !validEvidenceKind(meta.Kind) || !validDigest(key) {
			return fmt.Errorf("%w: invalid evidence key %q", ErrCorrupt, key)
		}
		expectedFile := filepath.ToSlash(filepath.Join("records", meta.Kind+"-"+key+".json"))
		if meta.Snapshot != current.Snapshot.Digest() || meta.File != expectedFile {
			return fmt.Errorf("%w: invalid evidence index entry %q", ErrCorrupt, key)
		}
		data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(meta.File)))
		if err != nil {
			return fmt.Errorf("%w: read evidence %q: %v", ErrCorrupt, key, err)
		}
		var stored storedEvidence
		if err := json.Unmarshal(data, &stored); err != nil {
			return fmt.Errorf("%w: decode evidence %q: %v", ErrCorrupt, key, err)
		}
		if stored.Version != ledgerVersion || stored.Key != key || stored.Snapshot != current.Snapshot.Digest() || stored.Kind != meta.Kind || stored.Digest != recordDigest(stored) {
			return fmt.Errorf("%w: evidence record %q does not match its index", ErrCorrupt, key)
		}
		derivedKey, _, err := evidenceKey(stored.Kind, stored.Snapshot, stored.Payload)
		if err != nil || derivedKey != key {
			return fmt.Errorf("%w: evidence record %q has the wrong content identity", ErrCorrupt, key)
		}
		if meta.Handle != "" {
			decoded, err := decodeHandle(meta.Handle, meta.Kind)
			if err != nil || decoded.Snapshot != stored.Snapshot || decoded.Digest != key {
				return fmt.Errorf("%w: invalid handle for evidence %q", ErrCorrupt, key)
			}
		}
	}
	for responseID, meta := range current.Responses {
		if !validDigest(responseID) {
			return fmt.Errorf("%w: invalid response id %q", ErrCorrupt, responseID)
		}
		expectedFile := filepath.ToSlash(filepath.Join("responses", responseID+".json"))
		if meta.File != expectedFile {
			return fmt.Errorf("%w: invalid response index entry %q", ErrCorrupt, responseID)
		}
		data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(meta.File)))
		if err != nil {
			return fmt.Errorf("%w: read response %q: %v", ErrCorrupt, responseID, err)
		}
		var response storedResponse
		if err := json.Unmarshal(data, &response); err != nil {
			return fmt.Errorf("%w: decode response %q: %v", ErrCorrupt, responseID, err)
		}
		if response.Version != ledgerVersion || response.ResponseID != responseID || response.Snapshot != current.Snapshot.Digest() || meta.Snapshot != response.Snapshot || response.Digest != responseRecordDigest(response) {
			return fmt.Errorf("%w: response %q does not match its index", ErrCorrupt, responseID)
		}
		for _, key := range responseKeysStored(response) {
			if _, ok := current.Evidence[key]; !ok {
				return fmt.Errorf("%w: response %q references missing evidence", ErrCorrupt, responseID)
			}
		}
	}
	return nil
}

func validEvidenceKind(kind string) bool {
	switch kind {
	case "node", "source", "path", "document", "question", "rejected_branch", "accepted_branch":
		return true
	default:
		return false
	}
}

func validDigest(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func recordDigest(value storedEvidence) string {
	value.Digest = ""
	data, _ := json.Marshal(value)
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func responseRecordDigest(value storedResponse) string {
	value.Digest = ""
	data, _ := json.Marshal(value)
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func writeManifest(dir string, current manifest) error {
	return writeJSONAtomic(filepath.Join(dir, "manifest.json"), current)
}

func writeJSONAtomic(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	dir := filepath.Dir(path)
	temp, err := os.CreateTemp(dir, ".atomic-")
	if err != nil {
		return fmt.Errorf("stage %s: %w", filepath.Base(path), err)
	}
	tempName := temp.Name()
	defer os.Remove(tempName)
	if err := temp.Chmod(0o644); err != nil {
		_ = temp.Close()
		return err
	}
	if _, err := temp.Write(data); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := replaceFile(tempName, path); err != nil {
		return fmt.Errorf("publish %s: %w", filepath.Base(path), err)
	}
	return nil
}
