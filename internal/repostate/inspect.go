package repostate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/knowledge"
	"github.com/Lokee86/grimoire/internal/knowledgevector"
)

type paths struct {
	root, lexicon, arcana, grimoire, knowledge string
}

type stateMarker struct {
	SourceFingerprint string `json:"source_fingerprint"`
	LexiconSnapshot   string `json:"lexicon_snapshot,omitempty"`
}

func normalize(options Options) (paths, error) {
	root := options.Root
	if root == "" {
		root = "."
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return paths{}, fmt.Errorf("resolve repository root: %w", err)
	}
	resolve := func(value, fallback string) string {
		if value == "" {
			return filepath.Join(absolute, fallback)
		}
		if filepath.IsAbs(value) {
			return filepath.Clean(value)
		}
		return filepath.Join(absolute, value)
	}
	grimoire := resolve(options.GrimoireState, ".grimoire")
	return paths{
		root: absolute, lexicon: resolve(options.LexiconState, ".lexicon"),
		arcana: resolve(options.ArcanaState, ".arcana"), grimoire: grimoire,
		knowledge: filepath.Join(grimoire, "knowledge"),
	}, nil
}

func inspect(ctx context.Context, location paths) (Status, error) {
	started := now()
	repository := RepositoryStatus{Root: location.root}
	fingerprint, err := sourceFingerprint(location.root)
	if err != nil {
		return Status{}, fmt.Errorf("fingerprint source: %w", err)
	}
	repository.SourceFingerprint = fingerprint
	repository.GitHead, repository.GitDirty, repository.GitAvailable = gitIdentity(ctx, location.root)
	status := Status{Version: 2, Repository: repository}
	status.Lexicon = inspectLexicon(location, fingerprint, repository.GitDirty)
	status.Arcana = inspectArcana(location, status.Lexicon.Snapshot)
	status.Grimoire = inspectGrimoire(location, fingerprint, status.Lexicon.Snapshot)
	var currentKnowledge knowledge.Index
	var knowledgeAvailable bool
	status.Knowledge, currentKnowledge, knowledgeAvailable = inspectKnowledge(location)
	status.ArcanaVectors = inspectArcanaVectors(location, status.Arcana.Snapshot)
	status.KnowledgeVectors = inspectKnowledgeVectors(location, currentKnowledge, knowledgeAvailable)
	status.DeterministicQueryReady = status.Grimoire.Status == "current"
	status.ElapsedMS = elapsedMS(started)
	return status, nil
}

func inspectLexicon(location paths, fingerprint string, dirty bool) ComponentStatus {
	id, err := readCurrent(filepath.Join(location.lexicon, "CURRENT"))
	if errors.Is(err, os.ErrNotExist) {
		return ComponentStatus{Status: "absent", StaleReasons: []string{"CURRENT is missing"}}
	}
	if err != nil {
		return ComponentStatus{Status: "failed", StaleReasons: []string{err.Error()}}
	}
	if !validSnapshotID(id) {
		return ComponentStatus{Status: "failed", Snapshot: id, StaleReasons: []string{"CURRENT has an invalid snapshot ID"}}
	}
	if _, err := os.Stat(filepath.Join(location.lexicon, "snapshots", strings.TrimPrefix(id, "sha256:")+".json")); err != nil {
		return ComponentStatus{Status: "failed", Snapshot: id, StaleReasons: []string{"current snapshot manifest is missing"}}
	}
	status := ComponentStatus{Status: "current", Snapshot: id, Prepared: true}
	var marker stateMarker
	if err := readJSON(filepath.Join(location.lexicon, ".repostate.json"), &marker); err == nil {
		if marker.SourceFingerprint != "" && marker.SourceFingerprint != fingerprint {
			status.Status = "stale"
			status.StaleReasons = append(status.StaleReasons, "source fingerprint changed since Lexicon preparation")
		}
	} else if dirty {
		status.Status = "stale"
		status.StaleReasons = append(status.StaleReasons, "Git working tree has source changes")
	}
	return status
}

func inspectArcana(location paths, expected string) ComponentStatus {
	id, err := readCurrent(filepath.Join(location.arcana, "CURRENT"))
	if errors.Is(err, os.ErrNotExist) {
		return ComponentStatus{Status: "absent", Expected: expected, StaleReasons: []string{"CURRENT is missing"}}
	}
	if err != nil {
		return ComponentStatus{Status: "failed", Expected: expected, StaleReasons: []string{err.Error()}}
	}
	status := ComponentStatus{Status: "current", Snapshot: id, Expected: expected}
	if !validSnapshotID(id) {
		status.Status = "failed"
		status.StaleReasons = append(status.StaleReasons, "CURRENT has an invalid snapshot ID")
		return status
	}
	if expected == "" {
		status.Status = "stale"
		status.StaleReasons = append(status.StaleReasons, "Lexicon snapshot is unavailable for alignment")
	} else if id != expected {
		status.Status = "stale"
		status.StaleReasons = append(status.StaleReasons, "Arcana snapshot does not match Lexicon CURRENT")
	}
	directory := filepath.Join(location.arcana, "snapshots", strings.TrimPrefix(id, "sha256:"))
	if !fileExists(filepath.Join(directory, "repository.manifest")) || !fileExists(filepath.Join(directory, "lexicon.snapshot")) {
		status.Status = "stale"
		status.StaleReasons = append(status.StaleReasons, "Arcana snapshot is incomplete")
	} else if value, readErr := readCurrent(filepath.Join(directory, "lexicon.snapshot")); readErr != nil || value != id {
		status.Status = "stale"
		status.StaleReasons = append(status.StaleReasons, "Arcana lexicon.snapshot does not match CURRENT")
	}
	status.Prepared = status.Status == "current"
	return status
}

func inspectGrimoire(location paths, fingerprint, lexiconID string) ComponentStatus {
	status := ComponentStatus{Status: "absent"}
	snapshot, err := index.Load(location.grimoire)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			status.Status = "failed"
			status.StaleReasons = []string{err.Error()}
			return status
		}
		status.StaleReasons = []string{"prepared index is missing"}
		return status
	}
	status.Status = "current"
	status.Snapshot = snapshot.Identity()
	status.Prepared = true
	var marker stateMarker
	if err := readJSON(filepath.Join(location.grimoire, ".repostate.json"), &marker); err != nil {
		status.Status = "stale"
		status.Prepared = false
		status.StaleReasons = append(status.StaleReasons, "preparation metadata is missing")
		return status
	}
	if marker.SourceFingerprint != fingerprint {
		status.Status = "stale"
		status.Prepared = false
		status.StaleReasons = append(status.StaleReasons, "source fingerprint changed since Grimoire preparation")
	}
	if marker.LexiconSnapshot != lexiconID {
		status.Status = "stale"
		status.Prepared = false
		status.StaleReasons = append(status.StaleReasons, "prepared index Lexicon snapshot is not current")
	}
	return status
}

func inspectKnowledge(location paths) (ComponentStatus, knowledge.Index, bool) {
	stored, err := knowledge.Load(location.knowledge)
	if errors.Is(err, os.ErrNotExist) {
		current, _, buildErr := knowledge.Build(location.root, nil, knowledge.BuildOptions{})
		if buildErr != nil {
			return ComponentStatus{Status: "failed", StaleReasons: []string{buildErr.Error()}}, knowledge.Index{}, false
		}
		return ComponentStatus{Status: "absent", Expected: knowledge.Identity(current), StaleReasons: []string{"knowledge index is missing"}}, current, true
	}
	if err != nil {
		return ComponentStatus{Status: "failed", StaleReasons: []string{err.Error()}}, knowledge.Index{}, false
	}
	current, _, err := knowledge.Build(location.root, &stored, knowledge.BuildOptions{})
	if err != nil {
		return ComponentStatus{Status: "failed", Snapshot: knowledge.Identity(stored), StaleReasons: []string{err.Error()}}, knowledge.Index{}, false
	}
	actual := knowledge.Identity(stored)
	expected := knowledge.Identity(current)
	status := ComponentStatus{Status: "current", Snapshot: actual, Expected: expected, Prepared: true}
	if actual != expected || stored.SourceFingerprint != expected {
		status.Status = "stale"
		status.Prepared = false
		status.StaleReasons = append(status.StaleReasons, "documentation or source-link evidence changed since knowledge indexing")
	}
	return status, current, true
}

func inspectKnowledgeVectors(location paths, current knowledge.Index, knowledgeAvailable bool) VectorStatus {
	if !knowledgeAvailable || len(current.Documents) == 0 {
		return VectorStatus{Status: "missing", Reason: "no knowledge index is available"}
	}
	info := knowledgevector.Inspect(location.knowledge, current, "")
	status := VectorStatus{
		Snapshot: info.SnapshotIdentity, Expected: info.ExpectedIdentity,
		Model: info.Model, Count: info.Count, Bytes: info.SnapshotBytes,
	}
	if !info.Available {
		status.Status = "missing"
		status.Reason = info.Error
		if status.Reason == "" {
			status.Reason = "documentation vector snapshot is not prepared"
		}
		return status
	}
	if info.Current {
		status.Status = "current"
		return status
	}
	status.Status = "stale"
	status.Reason = info.Error
	if status.Reason == "" {
		status.Reason = "documentation vector snapshot is not current"
	}
	return status
}

func inspectArcanaVectors(location paths, arcanaID string) VectorStatus {
	if arcanaID == "" || !validSnapshotID(arcanaID) {
		return VectorStatus{Status: "missing", Reason: "no current Arcana snapshot"}
	}
	directory := filepath.Join(location.arcana, "vectors", strings.TrimPrefix(arcanaID, "sha256:"))
	matches, _ := filepath.Glob(filepath.Join(directory, "*", "manifest.json"))
	for _, manifest := range matches {
		base := filepath.Dir(manifest)
		var value struct {
			GraphSnapshotID string `json:"graph_snapshot_id"`
			VectorsFile     string `json:"vectors_file"`
			RecordsFile     string `json:"records_file"`
		}
		if readJSON(manifest, &value) == nil && fileExists(filepath.Join(base, value.VectorsFile)) && fileExists(filepath.Join(base, value.RecordsFile)) {
			return VectorStatus{Status: "available", Snapshot: arcanaID}
		}
	}
	return VectorStatus{Status: "missing", Snapshot: arcanaID, Reason: "matching vector index is not prepared"}
}

func readCurrent(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("%w: %s", os.ErrNotExist, path)
		}
		return "", err
	}
	value := strings.TrimSpace(string(data))
	if value == "" {
		return "", fmt.Errorf("%s is empty", path)
	}
	return value, nil
}

func validSnapshotID(value string) bool {
	digest, ok := strings.CutPrefix(value, "sha256:")
	if !ok || len(digest) != 64 {
		return false
	}
	for _, character := range digest {
		if !strings.ContainsRune("0123456789abcdef", character) {
			return false
		}
	}
	return true
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func readJSON(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}
