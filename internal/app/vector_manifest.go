package app

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Lokee86/grimoire/internal/embedding"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/vectorstore"
)

const vectorSnapshotManifestVersion = 1

type vectorSnapshotManifest struct {
	Version          int      `json:"version"`
	PreparedIdentity string   `json:"prepared_identity"`
	SnapshotIdentity string   `json:"snapshot_identity"`
	Model            string   `json:"model"`
	Dimensions       int      `json:"dimensions"`
	Count            int      `json:"count"`
	Sources          []string `json:"sources,omitempty"`
}

func writeVectorSnapshotManifest(path string, manifest vectorSnapshotManifest) error {
	if err := validateVectorSnapshotManifestFields(manifest); err != nil {
		return err
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	temporary := path + ".next"
	defer os.Remove(temporary)
	file, err := os.OpenFile(temporary, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err = file.Write(data); err == nil {
		err = file.Sync()
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err := os.Rename(temporary, path); err == nil {
		return nil
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.Rename(temporary, path); err != nil {
		return fmt.Errorf("publish vector snapshot manifest: %w", err)
	}
	return nil
}

func readVectorSnapshotManifest(path string) (vectorSnapshotManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return vectorSnapshotManifest{}, fmt.Errorf("read vector snapshot manifest: %w", err)
	}
	var manifest vectorSnapshotManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return vectorSnapshotManifest{}, fmt.Errorf("decode vector snapshot manifest: %w", err)
	}
	if err := validateVectorSnapshotManifestFields(manifest); err != nil {
		return vectorSnapshotManifest{}, err
	}
	return manifest, nil
}

func validateVectorSnapshotManifest(
	path string,
	snapshot index.Snapshot,
	expectedCount int,
) (vectorSnapshotManifest, error) {
	manifest, err := readVectorSnapshotManifest(path)
	if err != nil {
		return vectorSnapshotManifest{}, err
	}
	preparedIdentity := snapshot.Identity()
	if preparedIdentity == "" {
		return vectorSnapshotManifest{}, fmt.Errorf("prepared index has no published identity")
	}
	if manifest.PreparedIdentity != preparedIdentity {
		return vectorSnapshotManifest{}, fmt.Errorf(
			"vector snapshot was built from prepared index %s, current prepared index is %s",
			manifest.PreparedIdentity, preparedIdentity,
		)
	}
	if manifest.Model != embedding.Identity() {
		return vectorSnapshotManifest{}, fmt.Errorf(
			"vector snapshot manifest uses model %s, expected %s",
			manifest.Model, embedding.Identity(),
		)
	}
	if manifest.Dimensions != embedding.Dimensions {
		return vectorSnapshotManifest{}, fmt.Errorf(
			"vector snapshot manifest has %d dimensions, expected %d",
			manifest.Dimensions, embedding.Dimensions,
		)
	}
	if manifest.Count != expectedCount {
		return vectorSnapshotManifest{}, fmt.Errorf(
			"vector snapshot manifest has %d chunks, prepared index has %d",
			manifest.Count, expectedCount,
		)
	}
	return manifest, nil
}

func validateVectorEngineInfo(manifest vectorSnapshotManifest, info vectorstore.Info) error {
	if info.Model != manifest.Model || info.Dimensions != manifest.Dimensions || info.Count != manifest.Count {
		return fmt.Errorf(
			"vector snapshot metadata is %s/%dd/%d, manifest expects %s/%dd/%d",
			info.Model, info.Dimensions, info.Count,
			manifest.Model, manifest.Dimensions, manifest.Count,
		)
	}
	return nil
}

func validateVectorSnapshotManifestFields(manifest vectorSnapshotManifest) error {
	if manifest.Version != vectorSnapshotManifestVersion {
		return fmt.Errorf("unsupported vector snapshot manifest version %d", manifest.Version)
	}
	if manifest.PreparedIdentity == "" || manifest.SnapshotIdentity == "" || manifest.Model == "" {
		return fmt.Errorf("vector snapshot manifest identities and model are required")
	}
	if manifest.Dimensions <= 0 || manifest.Count <= 0 {
		return fmt.Errorf("vector snapshot manifest dimensions and count must be positive")
	}
	return nil
}
