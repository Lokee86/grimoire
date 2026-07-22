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

	windowsAMD64RuntimeURL = "https://github.com/ggml-org/llama.cpp/releases/download/b8121/llama-b8121-bin-win-cpu-x64.zip"
	windowsAMD64RuntimeSHA = "e7de7919fc141dd1193e6116dc9f965b872b15f85a7a13b4631f3893a250f5b8"

	modelFilename = "Qwen3-Embedding-0.6B-Q8_0.gguf"
	modelURL      = "https://huggingface.co/Qwen/Qwen3-Embedding-0.6B-GGUF/resolve/main/Qwen3-Embedding-0.6B-Q8_0.gguf?download=true"
	modelSHA      = "06507c7b42688469c4e7298b0a1e16deff06caf291cf0a5b278c308249c3e439"
)

type SetupOptions struct {
	CacheDir   string
	HTTPClient *http.Client
	Progress   io.Writer
}

type SetupResult struct {
	CacheDir    string `json:"cache_dir"`
	RuntimePath string `json:"runtime_path"`
	ModelPath   string `json:"model_path"`
	Runtime     string `json:"runtime_version"`
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
	runtimeDir := filepath.Join(root, "runtime", "llama.cpp-"+RuntimeVersion)
	name := "llama-server"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return findNamedFile(runtimeDir, name)
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
	client := options.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 45 * time.Minute}
	}
	progress := options.Progress
	if progress == nil {
		progress = io.Discard
	}

	runtimePath, err := installRuntime(ctx, client, cacheDir, progress)
	if err != nil {
		return SetupResult{}, err
	}
	modelPath := filepath.Join(cacheDir, "models", modelFilename)
	if err := downloadVerified(ctx, client, modelURL, modelSHA, modelPath, "embedding model", progress); err != nil {
		return SetupResult{}, err
	}

	return SetupResult{
		CacheDir: cacheDir, RuntimePath: runtimePath, ModelPath: modelPath,
		Runtime: RuntimeVersion, Model: ModelReference, Dimensions: Dimensions,
	}, nil
}

func installRuntime(ctx context.Context, client *http.Client, cacheDir string, progress io.Writer) (string, error) {
	if existing, err := ManagedRuntimePath(cacheDir); err == nil {
		return existing, nil
	}

	archivePath := filepath.Join(cacheDir, "downloads", "llama.cpp-"+RuntimeVersion+"-windows-amd64.zip")
	if err := downloadVerified(
		ctx, client, windowsAMD64RuntimeURL, windowsAMD64RuntimeSHA,
		archivePath, "llama.cpp runtime", progress,
	); err != nil {
		return "", err
	}

	targetDir := filepath.Join(cacheDir, "runtime", "llama.cpp-"+RuntimeVersion)
	temporaryDir := targetDir + ".partial"
	if err := os.RemoveAll(temporaryDir); err != nil {
		return "", err
	}
	if err := extractZip(archivePath, temporaryDir); err != nil {
		return "", fmt.Errorf("extract llama.cpp runtime: %w", err)
	}
	if err := os.RemoveAll(targetDir); err != nil {
		return "", err
	}
	if err := os.Rename(temporaryDir, targetDir); err != nil {
		return "", fmt.Errorf("publish llama.cpp runtime: %w", err)
	}
	if err := os.Remove(archivePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	runtimePath, err := ManagedRuntimePath(cacheDir)
	if err != nil {
		return "", fmt.Errorf("locate installed llama.cpp runtime: %w", err)
	}
	return runtimePath, nil
}

func resolveCacheDir(cacheDir string) (string, error) {
	if strings.TrimSpace(cacheDir) != "" {
		return filepath.Abs(cacheDir)
	}
	return DefaultCacheDir()
}
