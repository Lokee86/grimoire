package embedding

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeRuntimeBackend(t *testing.T) {
	for _, value := range []string{"", "auto", "CUDA", "vulkan", "cpu"} {
		if _, err := normalizeRuntimeBackend(value); err != nil {
			t.Fatalf("normalize %q: %v", value, err)
		}
	}
	if _, err := normalizeRuntimeBackend("metal"); err == nil {
		t.Fatal("unsupported backend accepted")
	}
}

func TestVersionAtLeast(t *testing.T) {
	cases := []struct {
		actual  string
		minimum string
		want    bool
	}{
		{"528.33", "528.33", true},
		{"551.61", "528.33", true},
		{"528.32", "528.33", false},
		{"462.62", "528.33", false},
		{"13.0.1", "13.0", true},
	}
	for _, test := range cases {
		if got := versionAtLeast(test.actual, test.minimum); got != test.want {
			t.Fatalf("versionAtLeast(%q, %q) = %v, want %v", test.actual, test.minimum, got, test.want)
		}
	}
}

func TestManagedRuntimePathForBackendUsesBackendDirectory(t *testing.T) {
	cache := t.TempDir()
	runtimeDir := filepath.Join(cache, "runtime", "llama.cpp-"+RuntimeVersion+"-"+RuntimeBackendVulkan, "bin")
	if err := os.MkdirAll(runtimeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	executable := filepath.Join(runtimeDir, runtimeExecutableName())
	if err := os.WriteFile(executable, []byte("runtime"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := managedRuntimePathForBackend(cache, RuntimeBackendVulkan)
	if err != nil {
		t.Fatal(err)
	}
	if got != executable {
		t.Fatalf("runtime path = %q, want %q", got, executable)
	}
}

func TestSelectRuntimeBackendHonorsExplicitSelection(t *testing.T) {
	for _, backend := range []string{RuntimeBackendCUDA, RuntimeBackendVulkan, RuntimeBackendCPU} {
		got, err := selectRuntimeBackend(backend)
		if err != nil {
			t.Fatal(err)
		}
		if got != backend {
			t.Fatalf("selected backend = %q, want %q", got, backend)
		}
	}
}
