package embedding

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const runtimeStateVersion = 1

const (
	RuntimeStatusStarting   = "starting"
	RuntimeStatusReady      = "ready"
	RuntimeStatusDegraded   = "degraded"
	RuntimeStatusRestarting = "restarting"
	RuntimeStatusStopped    = "stopped"
	RuntimeStatusFailed     = "failed"
)

type RuntimeState struct {
	Version         int       `json:"version"`
	Status          string    `json:"status"`
	SupervisorPID   int       `json:"supervisor_pid,omitempty"`
	ProcessPID      int       `json:"process_pid,omitempty"`
	Backend         string    `json:"backend,omitempty"`
	BackendVerified bool      `json:"backend_verified"`
	RuntimePath     string    `json:"runtime_path,omitempty"`
	RuntimeVersion  string    `json:"runtime_version"`
	ModelPath       string    `json:"model_path,omitempty"`
	Endpoint        string    `json:"endpoint"`
	ContextSize     int       `json:"context_size"`
	UbatchSize      int       `json:"ubatch_size"`
	Parallel        int       `json:"parallel"`
	PerSlotContext  int       `json:"per_slot_context"`
	MaxInputTokens  int       `json:"max_input_tokens"`
	GPULayers       int       `json:"gpu_layers"`
	LogPath         string    `json:"log_path"`
	Restarts        int       `json:"restarts"`
	StartedAt       time.Time `json:"started_at,omitempty"`
	ReadyAt         time.Time `json:"ready_at,omitempty"`
	UpdatedAt       time.Time `json:"updated_at"`
	LastHealthAt    time.Time `json:"last_health_at,omitempty"`
	LastError       string    `json:"last_error,omitempty"`
}

type ManagedRuntimeStatus struct {
	Configured      bool         `json:"configured"`
	StateAvailable  bool         `json:"state_available"`
	SupervisorAlive bool         `json:"supervisor_alive"`
	ProcessAlive    bool         `json:"process_alive"`
	Healthy         bool         `json:"healthy"`
	ProbeDimensions int          `json:"probe_dimensions,omitempty"`
	ProbeError      string       `json:"probe_error,omitempty"`
	State           RuntimeState `json:"state,omitempty"`
	GPU             *GPUStats    `json:"gpu,omitempty"`
}

func SaveRuntimeConfig(path string, config RuntimeConfig) error {
	return writeRuntimeJSON(path, config)
}

func LoadRuntimeConfig(path string) (RuntimeConfig, error) {
	var config RuntimeConfig
	if err := readRuntimeJSON(path, &config); err != nil {
		return RuntimeConfig{}, err
	}
	if err := config.ApplyDefaults(); err != nil {
		return RuntimeConfig{}, err
	}
	return config, nil
}

func SaveRuntimeState(path string, state RuntimeState) error {
	state.Version = runtimeStateVersion
	state.UpdatedAt = time.Now().UTC()
	return writeRuntimeJSON(path, state)
}

func LoadRuntimeState(path string) (RuntimeState, error) {
	var state RuntimeState
	if err := readRuntimeJSON(path, &state); err != nil {
		return RuntimeState{}, err
	}
	if state.Version != runtimeStateVersion {
		return RuntimeState{}, fmt.Errorf("unsupported embedding runtime state version %d", state.Version)
	}
	return state, nil
}

func writeRuntimeJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	temporary := fmt.Sprintf("%s.tmp-%d", path, os.Getpid())
	if err := os.WriteFile(temporary, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(temporary, path); err != nil {
		_ = os.Remove(temporary)
		return err
	}
	return nil
}

func readRuntimeJSON(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}

func ManagedInputLimitForEndpoint(endpoint string) int {
	if value := strings.TrimSpace(os.Getenv("GRIMOIRE_EMBEDDING_MAX_TOKENS")); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed >= 0 {
			return parsed
		}
	}
	paths, err := ManagedRuntimePaths("")
	if err == nil {
		if state, stateErr := LoadRuntimeState(paths.State); stateErr == nil && sameEndpoint(endpoint, state.Endpoint) {
			return state.MaxInputTokens
		}
		if config, configErr := LoadRuntimeConfig(paths.Config); configErr == nil && sameEndpoint(endpoint, config.Endpoint()) {
			return config.MaxInputTokens()
		}
	}
	defaults, err := DefaultRuntimeConfig("")
	if err == nil && sameEndpoint(endpoint, defaults.Endpoint()) {
		return defaults.MaxInputTokens()
	}
	return 0
}

func removeRuntimeState(paths RuntimePaths) error {
	for _, path := range []string{paths.State, paths.Stop} {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}
