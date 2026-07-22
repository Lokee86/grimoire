package objectstore

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var ErrNoCurrentSnapshot = errors.New("Lexicon has no current snapshot")

type Store struct {
	Root string
}

func (s Store) Publish(manifest Manifest) (string, error) {
	if manifest.StateCommit == "" {
		return "", errors.New("Lexicon snapshot requires a state commit")
	}
	manifest.Version = SnapshotVersion
	data, err := json.Marshal(manifest)
	if err != nil {
		return "", fmt.Errorf("encode Lexicon snapshot: %w", err)
	}
	id := digest("lexicon:snapshot:v1\x00", data)
	if err := writeImmutable(s.snapshotPath(id), append(data, '\n')); err != nil {
		return "", err
	}
	if err := writeAtomic(filepath.Join(s.Root, "CURRENT"), []byte(id+"\n")); err != nil {
		return "", fmt.Errorf("publish Lexicon snapshot: %w", err)
	}
	return id, nil
}

func (s Store) Current() (string, Manifest, error) {
	data, err := os.ReadFile(filepath.Join(s.Root, "CURRENT"))
	if os.IsNotExist(err) {
		return "", Manifest{}, ErrNoCurrentSnapshot
	}
	if err != nil {
		return "", Manifest{}, fmt.Errorf("read current Lexicon snapshot: %w", err)
	}
	id := strings.TrimSpace(string(data))
	if id == "" {
		return "", Manifest{}, ErrNoCurrentSnapshot
	}
	manifest, err := s.Load(id)
	return id, manifest, err
}

func (s Store) Load(id string) (Manifest, error) {
	if !validID(id) {
		return Manifest{}, fmt.Errorf("invalid Lexicon snapshot ID %q", id)
	}
	data, err := os.ReadFile(s.snapshotPath(id))
	if err != nil {
		return Manifest{}, fmt.Errorf("read Lexicon snapshot %s: %w", id, err)
	}
	canonical := bytes.TrimSpace(data)
	if actual := digest("lexicon:snapshot:v1\x00", canonical); actual != id {
		return Manifest{}, fmt.Errorf("Lexicon snapshot %s failed content verification", id)
	}
	var manifest Manifest
	if err := json.Unmarshal(canonical, &manifest); err != nil {
		return Manifest{}, fmt.Errorf("decode Lexicon snapshot %s: %w", id, err)
	}
	if manifest.Version != SnapshotVersion {
		return Manifest{}, fmt.Errorf("unsupported Lexicon snapshot version %d", manifest.Version)
	}
	return manifest, nil
}

func (s Store) WriteObject(object FactObject) (string, error) {
	if object.Language == "" || object.AdapterVersion == "" || object.SchemaVersion < 1 || object.AnalysisConfigID == "" {
		return "", errors.New("incomplete Lexicon fact object metadata")
	}
	if object.Records == nil {
		object.Records = []json.RawMessage{}
	}
	object.Version = ObjectVersion
	data, err := json.Marshal(object)
	if err != nil {
		return "", fmt.Errorf("encode Lexicon fact object: %w", err)
	}
	id := digest("lexicon:fact-object:v1\x00", data)
	if err := writeImmutable(s.objectPath(id), append(data, '\n')); err != nil {
		return "", err
	}
	return id, nil
}

func (s Store) LoadObject(id string) (FactObject, error) {
	if !validID(id) {
		return FactObject{}, fmt.Errorf("invalid Lexicon object ID %q", id)
	}
	data, err := os.ReadFile(s.objectPath(id))
	if err != nil {
		return FactObject{}, fmt.Errorf("read Lexicon object %s: %w", id, err)
	}
	canonical := bytes.TrimSpace(data)
	if actual := digest("lexicon:fact-object:v1\x00", canonical); actual != id {
		return FactObject{}, fmt.Errorf("Lexicon object %s failed content verification", id)
	}
	var object FactObject
	if err := json.Unmarshal(canonical, &object); err != nil {
		return FactObject{}, fmt.Errorf("decode Lexicon object %s: %w", id, err)
	}
	if object.Version != ObjectVersion {
		return FactObject{}, fmt.Errorf("unsupported Lexicon object version %d", object.Version)
	}
	return object, nil
}

func (s Store) ObjectPath(id string) string {
	return s.objectPath(id)
}

func (s Store) snapshotPath(id string) string {
	hexID := trimID(id)
	return filepath.Join(s.Root, "snapshots", hexID+".json")
}

func (s Store) objectPath(id string) string {
	hexID := trimID(id)
	if len(hexID) < 3 {
		return filepath.Join(s.Root, "objects", hexID)
	}
	return filepath.Join(s.Root, "objects", hexID[:2], hexID[2:])
}

func digest(prefix string, data []byte) string {
	hash := sha256.New()
	_, _ = hash.Write([]byte(prefix))
	_, _ = hash.Write(data)
	return "sha256:" + hex.EncodeToString(hash.Sum(nil))
}

func ContentID(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func trimID(id string) string {
	return strings.TrimPrefix(id, "sha256:")
}

func validID(id string) bool {
	hexID := trimID(id)
	if !strings.HasPrefix(id, "sha256:") || len(hexID) != 64 {
		return false
	}
	_, err := hex.DecodeString(hexID)
	return err == nil
}
