package app

import (
	"io"
	"path/filepath"
	"testing"

	"github.com/Lokee86/grimoire/internal/embedding"
)

func TestParseManagedRuntimeConfigExposesRuntimeControls(t *testing.T) {
	cache := t.TempDir()
	config, err := parseManagedRuntimeConfig("model start", []string{
		"--cache", cache,
		"--backend", "cuda",
		"--context-size", "16384",
		"--parallel", "4",
		"--batch-size-does-not-exist",
	}, io.Discard)
	if err == nil {
		t.Fatal("unknown runtime flag accepted")
	}

	config, err = parseManagedRuntimeConfig("model start", []string{
		"--cache", cache,
		"--backend", "cuda",
		"--context-size", "16384",
		"--parallel", "4",
		"--restart-limit", "0",
		"--log-backups", "0",
	}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if config.Backend != embedding.RuntimeBackendCUDA || config.MaxInputTokens() != 3968 {
		t.Fatalf("unexpected runtime config: %+v", config)
	}
	if config.RestartLimit != 0 || config.LogBackups != 0 {
		t.Fatalf("explicit zero values were not preserved: %+v", config)
	}
	if config.LogPath != filepath.Join(cache, "logs", "embedding-server.log") {
		t.Fatalf("log path = %q", config.LogPath)
	}
}
