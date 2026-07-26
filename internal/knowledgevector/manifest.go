package knowledgevector

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Lokee86/grimoire/internal/embedding"
	"github.com/Lokee86/grimoire/internal/knowledge"
	"github.com/Lokee86/grimoire/internal/vectorstore"
)

func ReadManifest(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("read knowledge vector manifest: %w", err)
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Manifest{}, fmt.Errorf("decode knowledge vector manifest: %w", err)
	}
	if err := validateManifestFields(manifest); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func writeManifest(path string, manifest Manifest) error {
	if err := validateManifestFields(manifest); err != nil {
		return err
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	temporary := path + ".next"
	if err := os.WriteFile(temporary, data, 0o644); err != nil {
		return err
	}
	defer os.Remove(temporary)
	if err := os.Rename(temporary, path); err == nil {
		return nil
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return os.Rename(temporary, path)
}

func validateManifest(manifest Manifest, index knowledge.Index, info vectorstore.Info) error {
	expected := IndexIdentity(index)
	if manifest.KnowledgeIdentity != expected {
		return fmt.Errorf("knowledge vector snapshot was built from %s, current knowledge index is %s", manifest.KnowledgeIdentity, expected)
	}
	if manifest.Model != embedding.Identity() || manifest.Dimensions != embedding.Dimensions {
		return fmt.Errorf("knowledge vector snapshot model is %s/%dd, expected %s/%dd", manifest.Model, manifest.Dimensions, embedding.Identity(), embedding.Dimensions)
	}
	if info.Model != manifest.Model || info.Dimensions != manifest.Dimensions || info.Count != manifest.Count {
		return fmt.Errorf("knowledge vector engine metadata is %s/%dd/%d, manifest expects %s/%dd/%d", info.Model, info.Dimensions, info.Count, manifest.Model, manifest.Dimensions, manifest.Count)
	}
	return nil
}

func validateManifestFields(manifest Manifest) error {
	if manifest.Version != manifestVersion {
		return fmt.Errorf("unsupported knowledge vector manifest version %d", manifest.Version)
	}
	if manifest.KnowledgeIdentity == "" || manifest.SnapshotIdentity == "" || manifest.Model == "" {
		return fmt.Errorf("knowledge vector manifest identities and model are required")
	}
	if manifest.Dimensions <= 0 || manifest.Count <= 0 {
		return fmt.Errorf("knowledge vector manifest dimensions and count must be positive")
	}
	return nil
}
