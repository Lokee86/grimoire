package embedding

import (
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultRuntimeConfigDefinesInputContract(t *testing.T) {
	config, err := DefaultRuntimeConfig(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if config.PerSlotContext() != 2048 {
		t.Fatalf("per-slot context = %d, want 2048", config.PerSlotContext())
	}
	if config.MaxInputTokens() != 1920 {
		t.Fatalf("maximum input tokens = %d, want 1920", config.MaxInputTokens())
	}
	if config.EffectiveGPULayers(RuntimeBackendCUDA) != 99 {
		t.Fatalf("CUDA GPU layers = %d, want 99", config.EffectiveGPULayers(RuntimeBackendCUDA))
	}
	if config.EffectiveGPULayers(RuntimeBackendCPU) != 0 {
		t.Fatalf("CPU GPU layers = %d, want 0", config.EffectiveGPULayers(RuntimeBackendCPU))
	}
}

func TestRuntimeConfigRejectsUnsafePerSlotContext(t *testing.T) {
	config, err := DefaultRuntimeConfig(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	config.ContextSize = 512
	config.Parallel = 4
	if err := config.Validate(); err == nil {
		t.Fatal("unsafe per-slot context accepted")
	}
}

func TestRuntimeConfigAndStateRoundTrip(t *testing.T) {
	cache := t.TempDir()
	config, err := DefaultRuntimeConfig(cache)
	if err != nil {
		t.Fatal(err)
	}
	config.Backend = RuntimeBackendCUDA
	config.RestartLimit = 0
	config.LogBackups = 0
	paths := runtimePaths(cache)
	if err := SaveRuntimeConfig(paths.Config, config); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadRuntimeConfig(paths.Config)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.RestartLimit != 0 || loaded.LogBackups != 0 {
		t.Fatalf("zero operational values were not preserved: %+v", loaded)
	}
	if loaded.LogPath != filepath.Join(cache, "logs", "embedding-server.log") {
		t.Fatalf("log path = %q", loaded.LogPath)
	}

	state := RuntimeState{
		Status: RuntimeStatusReady, Endpoint: config.Endpoint(),
		ContextSize: config.ContextSize, Parallel: config.Parallel,
		MaxInputTokens: config.MaxInputTokens(), StartedAt: time.Now().UTC(),
	}
	if err := SaveRuntimeState(paths.State, state); err != nil {
		t.Fatal(err)
	}
	loadedState, err := LoadRuntimeState(paths.State)
	if err != nil {
		t.Fatal(err)
	}
	if loadedState.Version != runtimeStateVersion || loadedState.MaxInputTokens != 1920 {
		t.Fatalf("unexpected state: %+v", loadedState)
	}
}
