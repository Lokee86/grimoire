package embedding

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type runtimeArchive struct {
	Filename string
	SHA256   string
}

var windowsRuntimeArchives = map[string][]runtimeArchive{
	RuntimeBackendCPU: {
		{Filename: "llama-b8121-bin-win-cpu-x64.zip", SHA256: "e7de7919fc141dd1193e6116dc9f965b872b15f85a7a13b4631f3893a250f5b8"},
	},
	RuntimeBackendVulkan: {
		{Filename: "llama-b8121-bin-win-vulkan-x64.zip", SHA256: "66d7f461ab2e4c2c5570a8613ae31342d33fdc232ab026e4aef65f7992dcdf65"},
	},
	RuntimeBackendCUDA: {
		{Filename: "llama-b8121-bin-win-cuda-12.4-x64.zip", SHA256: "f3b0c3425b5b6f8adb7417f2acae1b16ed6752641e1a8eaf93fb6c132858dac2"},
		{Filename: "cudart-llama-bin-win-cuda-12.4-x64.zip", SHA256: "8c79a9b226de4b3cacfd1f83d24f962d0773be79f1e7b75c6af4ded7e32ae1d6"},
	},
}

func installRuntime(ctx context.Context, client *http.Client, cacheDir, backend string, progress io.Writer) (string, error) {
	if existing, err := managedRuntimePathForBackend(cacheDir, backend); err == nil {
		return existing, nil
	}
	archives, ok := windowsRuntimeArchives[backend]
	if !ok {
		return "", fmt.Errorf("no managed llama.cpp runtime for backend %q", backend)
	}

	downloads := make([]string, 0, len(archives))
	for _, archive := range archives {
		archivePath := filepath.Join(cacheDir, "downloads", archive.Filename)
		url := "https://github.com/ggml-org/llama.cpp/releases/download/" + RuntimeVersion + "/" + archive.Filename
		if err := downloadVerified(ctx, client, url, archive.SHA256, archivePath, "llama.cpp "+backend+" runtime", progress); err != nil {
			return "", err
		}
		downloads = append(downloads, archivePath)
	}

	targetDir := filepath.Join(cacheDir, "runtime", "llama.cpp-"+RuntimeVersion+"-"+backend)
	temporaryDir := targetDir + ".partial"
	if err := os.RemoveAll(temporaryDir); err != nil {
		return "", err
	}
	for _, archivePath := range downloads {
		if err := extractZip(archivePath, temporaryDir); err != nil {
			return "", fmt.Errorf("extract llama.cpp %s runtime: %w", backend, err)
		}
	}
	if err := os.RemoveAll(targetDir); err != nil {
		return "", err
	}
	if err := os.Rename(temporaryDir, targetDir); err != nil {
		return "", fmt.Errorf("publish llama.cpp runtime: %w", err)
	}
	for _, archivePath := range downloads {
		if err := os.Remove(archivePath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
	}

	runtimePath, err := managedRuntimePathForBackend(cacheDir, backend)
	if err != nil {
		return "", fmt.Errorf("locate installed llama.cpp runtime: %w", err)
	}
	return runtimePath, nil
}
