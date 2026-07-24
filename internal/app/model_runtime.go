package app

import (
	"context"
	"errors"
	"flag"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Lokee86/grimoire/internal/embedding"
)

func runModelStart(args []string, stdout, stderr io.Writer) error {
	config, err := parseManagedRuntimeConfig("model start", args, stderr)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), config.StartupTimeout()+15*time.Second)
	defer cancel()
	state, err := embedding.StartManagedRuntime(ctx, config)
	if err != nil {
		return err
	}
	return writeJSON(stdout, state)
}

func runModelStop(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("model stop", flag.ContinueOnError)
	flags.SetOutput(stderr)
	cacheDir := flags.String("cache", "", "managed model and runtime cache directory")
	timeout := flags.Duration("timeout", 15*time.Second, "runtime stop timeout")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *timeout <= 0 {
		return errors.New("--timeout must be positive")
	}
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	state, err := embedding.StopManagedRuntime(ctx, *cacheDir)
	if err != nil {
		return err
	}
	return writeJSON(stdout, state)
}

func runModelRestart(args []string, stdout, stderr io.Writer) error {
	config, err := parseManagedRuntimeConfig("model restart", args, stderr)
	if err != nil {
		return err
	}
	stopCtx, cancelStop := context.WithTimeout(context.Background(), 15*time.Second)
	_, stopErr := embedding.StopManagedRuntime(stopCtx, config.CacheDir)
	cancelStop()
	if stopErr != nil {
		return stopErr
	}
	startCtx, cancelStart := context.WithTimeout(context.Background(), config.StartupTimeout()+15*time.Second)
	defer cancelStart()
	state, err := embedding.StartManagedRuntime(startCtx, config)
	if err != nil {
		return err
	}
	return writeJSON(stdout, state)
}

func runModelStatus(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("model status", flag.ContinueOnError)
	flags.SetOutput(stderr)
	cacheDir := flags.String("cache", "", "managed model and runtime cache directory")
	timeout := flags.Duration("timeout", 15*time.Second, "status probe timeout")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *timeout <= 0 {
		return errors.New("--timeout must be positive")
	}
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	status, err := embedding.InspectManagedRuntime(ctx, *cacheDir)
	if err != nil {
		return err
	}
	return writeJSON(stdout, status)
}

func runModelSupervise(args []string, stderr io.Writer) error {
	flags := flag.NewFlagSet("model supervise", flag.ContinueOnError)
	flags.SetOutput(stderr)
	configPath := flags.String("config", "", "managed runtime configuration path")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*configPath) == "" {
		return errors.New("--config is required")
	}
	return embedding.RunRuntimeSupervisor(*configPath)
}

func parseManagedRuntimeConfig(commandName string, args []string, stderr io.Writer) (embedding.RuntimeConfig, error) {
	cacheDir := runtimeCacheArgument(args)
	config, err := embedding.DefaultRuntimeConfig(cacheDir)
	if err != nil {
		return embedding.RuntimeConfig{}, err
	}
	paths, err := embedding.ManagedRuntimePaths(cacheDir)
	if err == nil {
		if stored, loadErr := embedding.LoadRuntimeConfig(paths.Config); loadErr == nil {
			config = stored
		}
	}

	flags := flag.NewFlagSet(commandName, flag.ContinueOnError)
	flags.SetOutput(stderr)
	cache := flags.String("cache", config.CacheDir, "managed model and runtime cache directory")
	runtimePath := flags.String("runtime", config.RuntimePath, "llama.cpp server executable")
	modelPath := flags.String("model-file", config.ModelPath, "local GGUF path")
	backend := flags.String("backend", config.Backend, "llama.cpp backend: auto, cuda, vulkan, or cpu")
	host := flags.String("host", config.Host, "embedding service host")
	port := flags.Int("port", config.Port, "embedding service port")
	contextSize := flags.Int("context-size", config.ContextSize, "llama.cpp total context size")
	ubatchSize := flags.Int("ubatch-size", config.UbatchSize, "llama.cpp physical batch size")
	parallel := flags.Int("parallel", config.Parallel, "llama.cpp server slots")
	gpuLayers := flags.Int("gpu-layers", config.GPULayers, "GPU layers; -1 selects all for GPU backends and zero for CPU")
	startupTimeout := flags.Duration("startup-timeout", config.StartupTimeout(), "runtime startup and verification timeout")
	restartLimit := flags.Int("restart-limit", config.RestartLimit, "maximum automatic restart attempts")
	restartDelay := flags.Duration("restart-delay", config.RestartDelay(), "delay between restart attempts")
	healthInterval := flags.Duration("health-interval", config.HealthInterval(), "runtime health probe interval")
	logPath := flags.String("log", config.LogPath, "embedding runtime log path")
	logMaxBytes := flags.Int64("log-max-bytes", config.LogMaxBytes, "maximum active runtime log size")
	logBackups := flags.Int("log-backups", config.LogBackups, "rotated runtime log files retained")
	if err := flags.Parse(args); err != nil {
		return embedding.RuntimeConfig{}, err
	}

	config = embedding.RuntimeConfig{
		CacheDir: *cache, RuntimePath: *runtimePath, ModelPath: *modelPath, Backend: *backend,
		Host: *host, Port: *port, ContextSize: *contextSize, UbatchSize: *ubatchSize,
		Parallel: *parallel, GPULayers: *gpuLayers,
		StartupTimeoutMS: startupTimeout.Milliseconds(), RestartLimit: *restartLimit,
		RestartDelayMS: restartDelay.Milliseconds(), HealthIntervalMS: healthInterval.Milliseconds(),
		LogPath: *logPath, LogMaxBytes: *logMaxBytes, LogBackups: *logBackups,
	}
	if err := config.ApplyDefaults(); err != nil {
		return embedding.RuntimeConfig{}, err
	}
	return config, nil
}

func runtimeCacheArgument(args []string) string {
	for index, arg := range args {
		if arg == "--cache" && index+1 < len(args) {
			return args[index+1]
		}
		if strings.HasPrefix(arg, "--cache=") {
			return strings.TrimPrefix(arg, "--cache=")
		}
	}
	if value := strings.TrimSpace(os.Getenv("GRIMOIRE_CACHE_DIR")); value != "" {
		return value
	}
	return ""
}
