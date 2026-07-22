package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const Version = 1

type Config struct {
	Version     int    `json:"version"`
	AdapterRoot string `json:"adapter_root"`
}

func StateRoot(repository string) string {
	return filepath.Join(repository, ".lexicon")
}

func AnalysisID() string {
	sum := sha256.Sum256([]byte("lexicon:analysis-config:v1"))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func Path(repository string) string {
	return filepath.Join(StateRoot(repository), "config.json")
}

func Save(repository, adapterRoot string) error {
	absolute, err := filepath.Abs(adapterRoot)
	if err != nil {
		return err
	}
	value := Config{Version: Version, AdapterRoot: absolute}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(StateRoot(repository), 0o755); err != nil {
		return err
	}
	return os.WriteFile(Path(repository), append(data, '\n'), 0o644)
}

func Load(repository string) (Config, error) {
	data, err := os.ReadFile(Path(repository))
	if err != nil {
		return Config{}, fmt.Errorf("read Lexicon configuration: %w", err)
	}
	var value Config
	if err := json.Unmarshal(data, &value); err != nil {
		return Config{}, fmt.Errorf("decode Lexicon configuration: %w", err)
	}
	if value.Version != Version {
		return Config{}, fmt.Errorf("unsupported Lexicon configuration version %d", value.Version)
	}
	return value, nil
}

func FindAdapterRoot(repository, explicit string) (string, error) {
	candidates := []string{explicit, os.Getenv("LEXICON_ADAPTERS")}
	if executable, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(executable), "adapters"))
	}
	candidates = append(candidates, filepath.Join(repository, "adapters"))
	if current, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(current, "adapters"))
	}
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		absolute, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		if info, err := os.Stat(filepath.Join(absolute, "python")); err == nil && info.IsDir() {
			return absolute, nil
		}
	}
	return "", fmt.Errorf("adapter root not found; use --adapters or LEXICON_ADAPTERS")
}
