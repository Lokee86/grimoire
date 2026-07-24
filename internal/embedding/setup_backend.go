package embedding

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const windowsCUDA12MinimumDriver = "528.33"

func selectRuntimeBackend(requested string) (string, error) {
	requested = strings.TrimSpace(requested)
	configured := strings.TrimSpace(os.Getenv("GRIMOIRE_LLAMA_BACKEND"))
	if requested == "" || (strings.EqualFold(requested, RuntimeBackendAuto) && configured != "") {
		requested = configured
	}
	backend, err := normalizeRuntimeBackend(requested)
	if err != nil {
		return "", err
	}
	if backend != RuntimeBackendAuto {
		return backend, nil
	}
	if cuda12Available() {
		return RuntimeBackendCUDA, nil
	}
	if vulkanAvailable() {
		return RuntimeBackendVulkan, nil
	}
	return RuntimeBackendCPU, nil
}

func normalizeRuntimeBackend(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		value = RuntimeBackendAuto
	}
	switch value {
	case RuntimeBackendAuto, RuntimeBackendCUDA, RuntimeBackendVulkan, RuntimeBackendCPU:
		return value, nil
	default:
		return "", fmt.Errorf("unsupported llama.cpp backend %q; expected auto, cuda, vulkan, or cpu", value)
	}
}

func cuda12Available() bool {
	path := findNvidiaSMI()
	if path == "" {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, path, "--query-gpu=driver_version", "--format=csv,noheader").Output()
	if err != nil {
		return false
	}
	line := strings.TrimSpace(strings.Split(string(output), "\n")[0])
	return versionAtLeast(line, windowsCUDA12MinimumDriver)
}

func findNvidiaSMI() string {
	if path, err := exec.LookPath("nvidia-smi"); err == nil {
		return path
	}
	windowsDir := strings.TrimSpace(os.Getenv("WINDIR"))
	programFiles := strings.TrimSpace(os.Getenv("ProgramFiles"))
	candidates := make([]string, 0, 2)
	if windowsDir != "" {
		candidates = append(candidates, filepath.Join(windowsDir, "System32", "nvidia-smi.exe"))
	}
	if programFiles != "" {
		candidates = append(candidates, filepath.Join(programFiles, "NVIDIA Corporation", "NVSMI", "nvidia-smi.exe"))
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.Mode().IsRegular() {
			return candidate
		}
	}
	return ""
}

func vulkanAvailable() bool {
	if _, err := exec.LookPath("vulkaninfo"); err == nil {
		return true
	}
	windowsDir := strings.TrimSpace(os.Getenv("WINDIR"))
	if windowsDir == "" {
		return false
	}
	info, err := os.Stat(filepath.Join(windowsDir, "System32", "vulkan-1.dll"))
	return err == nil && info.Mode().IsRegular()
}

func versionAtLeast(actual, minimum string) bool {
	actualParts := strings.Split(strings.TrimSpace(actual), ".")
	minimumParts := strings.Split(strings.TrimSpace(minimum), ".")
	length := max(len(actualParts), len(minimumParts))
	for index := 0; index < length; index++ {
		actualValue := 0
		minimumValue := 0
		if index < len(actualParts) {
			actualValue, _ = strconv.Atoi(strings.TrimSpace(actualParts[index]))
		}
		if index < len(minimumParts) {
			minimumValue, _ = strconv.Atoi(strings.TrimSpace(minimumParts[index]))
		}
		if actualValue > minimumValue {
			return true
		}
		if actualValue < minimumValue {
			return false
		}
	}
	return true
}

func uniqueRuntimeBackends(values ...string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value == "" || value == RuntimeBackendAuto {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
