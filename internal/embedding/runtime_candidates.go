package embedding

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type RuntimeCandidate struct {
	Path    string `json:"path"`
	Backend string `json:"backend"`
	Managed bool   `json:"managed"`
	Source  string `json:"source"`
}

func RuntimeCandidates(explicit, requestedBackend, cacheDir string) ([]RuntimeCandidate, error) {
	backend, err := normalizeRuntimeBackend(requestedBackend)
	if err != nil {
		return nil, err
	}
	candidates := make([]RuntimeCandidate, 0, 8)
	add := func(candidate RuntimeCandidate) {
		if strings.TrimSpace(candidate.Path) == "" {
			return
		}
		path, lookupErr := exec.LookPath(candidate.Path)
		if lookupErr != nil {
			return
		}
		candidate.Path = path
		if candidate.Backend == "" || candidate.Backend == RuntimeBackendAuto {
			candidate.Backend = inferRuntimeBackend(path)
		}
		for _, existing := range candidates {
			if sameRuntimePath(existing.Path, candidate.Path) {
				return
			}
		}
		candidates = append(candidates, candidate)
	}

	if strings.TrimSpace(explicit) != "" {
		add(RuntimeCandidate{Path: explicit, Backend: backend, Source: "explicit"})
		if len(candidates) == 0 {
			return nil, os.ErrNotExist
		}
		return candidates, nil
	}
	if configured := strings.TrimSpace(os.Getenv("GRIMOIRE_LLAMA_SERVER")); configured != "" {
		add(RuntimeCandidate{Path: configured, Backend: backend, Source: "environment"})
	}

	root, rootErr := resolveCacheDir(cacheDir)
	if rootErr != nil {
		return nil, rootErr
	}
	backends := []string{backend}
	if backend == RuntimeBackendAuto {
		preferred, selectErr := selectRuntimeBackend(RuntimeBackendAuto)
		if selectErr != nil {
			return nil, selectErr
		}
		backends = uniqueRuntimeBackends(preferred, RuntimeBackendCUDA, RuntimeBackendVulkan, RuntimeBackendCPU)
	}
	for _, candidateBackend := range backends {
		path, pathErr := managedRuntimePathForBackend(root, candidateBackend)
		if pathErr == nil {
			add(RuntimeCandidate{Path: path, Backend: candidateBackend, Managed: true, Source: "managed"})
		}
	}

	if backend == RuntimeBackendAuto {
		add(RuntimeCandidate{Path: "llama-server", Backend: RuntimeBackendAuto, Source: "path"})
		add(RuntimeCandidate{Path: "llama", Backend: RuntimeBackendAuto, Source: "path"})
	}
	if len(candidates) == 0 {
		return nil, errors.New("llama.cpp runtime not found; run `grimoire model setup` or set GRIMOIRE_LLAMA_SERVER")
	}
	return candidates, nil
}

func FindRuntimeForBackend(explicit, backend, cacheDir string) (RuntimeCandidate, error) {
	candidates, err := RuntimeCandidates(explicit, backend, cacheDir)
	if err != nil {
		return RuntimeCandidate{}, err
	}
	return candidates[0], nil
}

func FindRuntime(explicit string) (string, error) {
	candidate, err := FindRuntimeForBackend(explicit, RuntimeBackendAuto, "")
	if err != nil {
		return "", err
	}
	return candidate.Path, nil
}

func inferRuntimeBackend(path string) string {
	lower := strings.ToLower(filepath.Clean(path))
	switch {
	case strings.Contains(lower, "cuda"):
		return RuntimeBackendCUDA
	case strings.Contains(lower, "vulkan"):
		return RuntimeBackendVulkan
	case strings.Contains(lower, "cpu"):
		return RuntimeBackendCPU
	default:
		return RuntimeBackendAuto
	}
}

func sameRuntimePath(left, right string) bool {
	left = filepath.Clean(left)
	right = filepath.Clean(right)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(left, right)
	}
	return left == right
}
