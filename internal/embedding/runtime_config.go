package embedding

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

const (
	DefaultContextSize          = 8192
	DefaultUbatchSize           = 2048
	DefaultParallelSlots        = 4
	DefaultContextReserve       = 128
	DefaultStartupTimeout       = 2 * time.Minute
	DefaultRestartLimit         = 5
	DefaultRestartDelay         = 2 * time.Second
	DefaultHealthInterval       = 15 * time.Second
	DefaultLogMaxBytes    int64 = 16 << 20
	DefaultLogBackups           = 3
)

type RuntimeConfig struct {
	CacheDir         string `json:"cache_dir"`
	RuntimePath      string `json:"runtime_path,omitempty"`
	ModelPath        string `json:"model_path,omitempty"`
	Backend          string `json:"backend"`
	Host             string `json:"host"`
	Port             int    `json:"port"`
	ContextSize      int    `json:"context_size"`
	UbatchSize       int    `json:"ubatch_size"`
	Parallel         int    `json:"parallel"`
	GPULayers        int    `json:"gpu_layers"`
	StartupTimeoutMS int64  `json:"startup_timeout_ms"`
	RestartLimit     int    `json:"restart_limit"`
	RestartDelayMS   int64  `json:"restart_delay_ms"`
	HealthIntervalMS int64  `json:"health_interval_ms"`
	LogPath          string `json:"log_path"`
	LogMaxBytes      int64  `json:"log_max_bytes"`
	LogBackups       int    `json:"log_backups"`
}

type RuntimePaths struct {
	Root   string
	Config string
	State  string
	Stop   string
	Log    string
}

func DefaultRuntimeConfig(cacheDir string) (RuntimeConfig, error) {
	root, err := resolveCacheDir(cacheDir)
	if err != nil {
		return RuntimeConfig{}, err
	}
	paths := runtimePaths(root)
	return RuntimeConfig{
		CacheDir: root, Backend: RuntimeBackendAuto,
		Host: "127.0.0.1", Port: DefaultPort,
		ContextSize: DefaultContextSize, UbatchSize: DefaultUbatchSize,
		Parallel: DefaultParallelSlots, GPULayers: -1,
		StartupTimeoutMS: DefaultStartupTimeout.Milliseconds(),
		RestartLimit:     DefaultRestartLimit, RestartDelayMS: DefaultRestartDelay.Milliseconds(),
		HealthIntervalMS: DefaultHealthInterval.Milliseconds(),
		LogPath:          paths.Log, LogMaxBytes: DefaultLogMaxBytes, LogBackups: DefaultLogBackups,
	}, nil
}

func (config *RuntimeConfig) ApplyDefaults() error {
	defaults, err := DefaultRuntimeConfig(config.CacheDir)
	if err != nil {
		return err
	}
	if strings.TrimSpace(config.CacheDir) == "" {
		config.CacheDir = defaults.CacheDir
	} else {
		config.CacheDir, err = resolveCacheDir(config.CacheDir)
		if err != nil {
			return err
		}
	}
	if strings.TrimSpace(config.Backend) == "" {
		config.Backend = defaults.Backend
	}
	if strings.TrimSpace(config.Host) == "" {
		config.Host = defaults.Host
	}
	if config.Port == 0 {
		config.Port = defaults.Port
	}
	if config.ContextSize == 0 {
		config.ContextSize = defaults.ContextSize
	}
	if config.UbatchSize == 0 {
		config.UbatchSize = defaults.UbatchSize
	}
	if config.Parallel == 0 {
		config.Parallel = defaults.Parallel
	}
	if config.GPULayers < -1 {
		config.GPULayers = defaults.GPULayers
	}
	if config.StartupTimeoutMS == 0 {
		config.StartupTimeoutMS = defaults.StartupTimeoutMS
	}
	if config.RestartDelayMS == 0 {
		config.RestartDelayMS = defaults.RestartDelayMS
	}
	if config.HealthIntervalMS == 0 {
		config.HealthIntervalMS = defaults.HealthIntervalMS
	}
	if strings.TrimSpace(config.LogPath) == "" {
		config.LogPath = runtimePaths(config.CacheDir).Log
	} else if !filepath.IsAbs(config.LogPath) {
		config.LogPath = filepath.Join(config.CacheDir, config.LogPath)
	}
	if config.LogMaxBytes == 0 {
		config.LogMaxBytes = defaults.LogMaxBytes
	}
	return config.Validate()
}

func (config RuntimeConfig) Validate() error {
	if _, err := normalizeRuntimeBackend(config.Backend); err != nil {
		return err
	}
	if strings.TrimSpace(config.Host) == "" {
		return errors.New("embedding runtime host is required")
	}
	if config.Port <= 0 || config.Port > 65535 {
		return errors.New("embedding runtime port must be between 1 and 65535")
	}
	if config.ContextSize <= 0 || config.UbatchSize <= 0 || config.Parallel <= 0 {
		return errors.New("embedding context, ubatch, and parallel values must be positive")
	}
	if config.ContextSize/config.Parallel <= DefaultContextReserve {
		return fmt.Errorf("embedding per-slot context %d is too small", config.ContextSize/config.Parallel)
	}
	if config.GPULayers < -1 {
		return errors.New("embedding GPU layers must be -1 (automatic) or non-negative")
	}
	if config.StartupTimeoutMS <= 0 || config.RestartDelayMS <= 0 || config.HealthIntervalMS <= 0 {
		return errors.New("embedding runtime timing values must be positive")
	}
	if config.RestartLimit < 0 {
		return errors.New("embedding restart limit cannot be negative")
	}
	if config.LogMaxBytes <= 0 || config.LogBackups < 0 {
		return errors.New("embedding log size must be positive and backups cannot be negative")
	}
	return nil
}

func (config RuntimeConfig) Endpoint() string {
	return fmt.Sprintf("http://%s:%d/v1", config.Host, config.Port)
}

func (config RuntimeConfig) PerSlotContext() int {
	if config.Parallel <= 0 {
		return 0
	}
	return config.ContextSize / config.Parallel
}

func (config RuntimeConfig) MaxInputTokens() int {
	limit := config.PerSlotContext() - DefaultContextReserve
	if limit < 1 {
		return 0
	}
	return limit
}

func (config RuntimeConfig) StartupTimeout() time.Duration {
	return time.Duration(config.StartupTimeoutMS) * time.Millisecond
}

func (config RuntimeConfig) RestartDelay() time.Duration {
	return time.Duration(config.RestartDelayMS) * time.Millisecond
}

func (config RuntimeConfig) HealthInterval() time.Duration {
	return time.Duration(config.HealthIntervalMS) * time.Millisecond
}

func (config RuntimeConfig) EffectiveGPULayers(backend string) int {
	if config.GPULayers >= 0 {
		return config.GPULayers
	}
	if backend == RuntimeBackendCUDA || backend == RuntimeBackendVulkan {
		return 99
	}
	return 0
}

func runtimePaths(cacheDir string) RuntimePaths {
	root := filepath.Join(cacheDir, "run")
	return RuntimePaths{
		Root:   root,
		Config: filepath.Join(root, "embedding-server-config.json"),
		State:  filepath.Join(root, "embedding-server-state.json"),
		Stop:   filepath.Join(root, "embedding-server.stop"),
		Log:    filepath.Join(cacheDir, "logs", "embedding-server.log"),
	}
}

func ManagedRuntimePaths(cacheDir string) (RuntimePaths, error) {
	root, err := resolveCacheDir(cacheDir)
	if err != nil {
		return RuntimePaths{}, err
	}
	return runtimePaths(root), nil
}

func sameEndpoint(left, right string) bool {
	return strings.EqualFold(strings.TrimRight(strings.TrimSpace(left), "/"), strings.TrimRight(strings.TrimSpace(right), "/"))
}
