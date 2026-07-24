package embedding

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestRuntimeCandidatesHonorExplicitBackend(t *testing.T) {
	name := "llama-server-cuda"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte("runtime"), 0o755); err != nil {
		t.Fatal(err)
	}
	candidates, err := RuntimeCandidates(path, RuntimeBackendCUDA, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 1 {
		t.Fatalf("candidate count = %d, want 1", len(candidates))
	}
	if candidates[0].Backend != RuntimeBackendCUDA || candidates[0].Source != "explicit" {
		t.Fatalf("unexpected candidate: %+v", candidates[0])
	}
}

func TestInferRuntimeBackend(t *testing.T) {
	cases := map[string]string{
		`C:\\runtime\\llama.cpp-b8121-cuda\\llama-server.exe`:   RuntimeBackendCUDA,
		`C:\\runtime\\llama.cpp-b8121-vulkan\\llama-server.exe`: RuntimeBackendVulkan,
		`/opt/llama.cpp-cpu/llama-server`:                       RuntimeBackendCPU,
		`/usr/local/bin/llama-server`:                           RuntimeBackendAuto,
	}
	for path, want := range cases {
		if got := inferRuntimeBackend(path); got != want {
			t.Fatalf("inferRuntimeBackend(%q) = %q, want %q", path, got, want)
		}
	}
}
