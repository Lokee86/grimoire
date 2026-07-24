package embedding

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	RuntimeVersion = "b8121"

	RuntimeBackendAuto   = "auto"
	RuntimeBackendCUDA   = "cuda"
	RuntimeBackendVulkan = "vulkan"
	RuntimeBackendCPU    = "cpu"

	modelFilename = "Qwen3-Embedding-0.6B-Q8_0.gguf"
	modelURL      = "https://huggingface.co/Qwen/Qwen3-Embedding-0.6B-GGUF/resolve/main/Qwen3-Embedding-0.6B-Q8_0.gguf?download=true"
	modelSHA      = "06507c7b42688469c4e7298b0a1e16deff06caf291cf0a5b278c308249c3e439"
)

type SetupOptions struct {
	CacheDir   string
	Backend    string
	Force      bool
	HTTPClient *http.Client
	Progress   io.Writer
}

type SetupResult struct {
	CacheDir    string `json:"cache_dir"`
	RuntimePath string `json:"runtime_path"`
	ModelPath   string `json:"model_path"`
	Runtime     string `json:"runtime_version"`
	Backend     string `json:"backend"`
	Model       string `json:"model"`
	Dimensions  int    `json:"dimensions"`
}

func DefaultCacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("resolve user cache directory: %w", err)
	}
	return filepath.Join(base, "grimoire"), nil
}

func ManagedModelPath(cacheDir string) (string, error) {
	root, err := resolveCacheDir(cacheDir)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "models", modelFilename), nil
}

func ManagedRuntimePath(cacheDir string) (string, error) {
	root, err := resolveCacheDir(cacheDir)
	if err != nil {
		return "", err
	}
	preferred, _ := selectRuntimeBackend("")
	backends := uniqueRuntimeBackends(preferred, RuntimeBackendCUDA, RuntimeBackendVulkan, RuntimeBackendCPU)
	for _, backend := range backends {
		if path, pathErr := managedRuntimePathForBackend(root, backend); pathErr == nil {
			return path, nil
		}
	}
	legacyDir := filepath.Join(root, "runtime", "llama.cpp-"+RuntimeVersion)
	if path, legacyErr := findNamedFile(legacyDir, runtimeExecutableName()); legacyErr == nil {
		return path, nil
	}
	return "", os.ErrNotExist
}

func managedRuntimePathForBackend(cacheDir, backend string) (string, error) {
	root, err := resolveCacheDir(cacheDir)
	if err != nil {
		return "", err
	}
	backend, err = normalizeRuntimeBackend(backend)
	if err != nil || backend == RuntimeBackendAuto {
		return "", fmt.Errorf("specific runtime backend required")
	}
	runtimeDir := filepath.Join(root, "runtime", "llama.cpp-"+RuntimeVersion+"-"+backend)
	return findNamedFile(runtimeDir, runtimeExecutableName())
}

func runtimeExecutableName() string {
	name := "llama-server"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return name
}

func FindModel(explicit string) (string, error) {
	candidates := []string{
		strings.TrimSpace(explicit),
		strings.TrimSpace(os.Getenv("GRIMOIRE_EMBEDDING_MODEL")),
	}
	if managed, err := ManagedModelPath(""); err == nil {
		candidates = append(candidates, managed)
	}
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		info, err := os.Stat(candidate)
		if err == nil && info.Mode().IsRegular() {
			return candidate, nil
		}
	}
	return "", errors.New("managed embedding model not found; run `grimoire model setup` or set GRIMOIRE_EMBEDDING_MODEL")
}

func Setup(ctx context.Context, options SetupOptions) (SetupResult, error) {
	if runtime.GOOS != "windows" || runtime.GOARCH != "amd64" {
		return SetupResult{}, fmt.Errorf(
			"automatic llama.cpp setup is not available for %s/%s; install llama.cpp manually and set GRIMOIRE_LLAMA_SERVER",
			runtime.GOOS, runtime.GOARCH,
		)
	}
	cacheDir, err := resolveCacheDir(options.CacheDir)
	if err != nil {
		return SetupResult{}, err
	}
	backend, err := selectRuntimeBackend(options.Backend)
	if err != nil {
		return SetupResult{}, err
	}
	client := options.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 45 * time.Minute}
	}
	progress := options.Progress
	if progress == nil {
		progress = io.Discard
	}
	_, _ = fmt.Fprintf(progress, "selected llama.cpp backend: %s\n", backend)

	runtimePath, err := installRuntime(ctx, client, cacheDir, backend, options.Force, progress)
	if err != nil {
		return SetupResult{}, err
	}
	modelPath := filepath.Join(cacheDir, "models", modelFilename)
	if err := downloadVerified(ctx, client, modelURL, modelSHA, modelPath, "embedding model", progress); err != nil {
		return SetupResult{}, err
	}

	return SetupResult{
		CacheDir: cacheDir, RuntimePath: runtimePath, ModelPath: modelPath,
		Runtime: RuntimeVersion, Backend: backend, Model: ModelReference, Dimensions: Dimensions,
	}, nil
}

func resolveCacheDir(cacheDir string) (string, error) {
	if strings.TrimSpace(cacheDir) != "" {
		return filepath.Abs(cacheDir)
	}
	return DefaultCacheDir()
}
