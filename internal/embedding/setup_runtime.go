package embedding

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type runtimeArchive struct {
	Filename string
	SHA256   string
}

type runtimeInstallManifest struct {
	Version  string           `json:"version"`
	Backend  string           `json:"backend"`
	Archives []runtimeArchive `json:"archives"`
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

func installRuntime(
	ctx context.Context,
	client *http.Client,
	cacheDir, backend string,
	force bool,
	progress io.Writer,
) (string, error) {
	if existing, err := managedRuntimePathForBackend(cacheDir, backend); err == nil && !force {
		if validateErr := validateRuntimeExecutable(ctx, existing); validateErr == nil {
			return existing, nil
		}
		_, _ = fmt.Fprintf(progress, "existing llama.cpp %s runtime failed validation; reinstalling\n", backend)
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
	temporaryExecutable, err := findNamedFile(temporaryDir, runtimeExecutableName())
	if err != nil {
		return "", fmt.Errorf("locate extracted llama.cpp runtime: %w", err)
	}
	if err := validateRuntimeExecutable(ctx, temporaryExecutable); err != nil {
		return "", fmt.Errorf("validate extracted llama.cpp %s runtime: %w", backend, err)
	}
	manifest := runtimeInstallManifest{Version: RuntimeVersion, Backend: backend, Archives: archives}
	if err := writeRuntimeManifest(filepath.Join(temporaryDir, "grimoire-runtime.json"), manifest); err != nil {
		return "", err
	}
	if err := publishRuntimeDirectory(targetDir, temporaryDir); err != nil {
		return "", err
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

func validateRuntimeExecutable(parent context.Context, path string) error {
	ctx, cancel := context.WithTimeout(parent, 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, path, "--version")
	configureManagedChildProcess(command)
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("%w: %s", err, output)
	}
	return nil
}

func writeRuntimeManifest(path string, manifest runtimeInstallManifest) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func publishRuntimeDirectory(targetDir, temporaryDir string) error {
	backupDir := targetDir + ".previous"
	_ = os.RemoveAll(backupDir)
	targetExists := false
	if _, err := os.Stat(targetDir); err == nil {
		targetExists = true
		if err := os.Rename(targetDir, backupDir); err != nil {
			return fmt.Errorf("preserve existing llama.cpp runtime: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.Rename(temporaryDir, targetDir); err != nil {
		if targetExists {
			_ = os.Rename(backupDir, targetDir)
		}
		return fmt.Errorf("publish llama.cpp runtime: %w", err)
	}
	if targetExists {
		if err := os.RemoveAll(backupDir); err != nil {
			return fmt.Errorf("remove previous llama.cpp runtime: %w", err)
		}
	}
	return nil
}
