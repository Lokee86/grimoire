package embedding

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVerifyBackendLogRequiresGPUPlacement(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime.log")
	content := "ggml_cuda_init: found 1 CUDA devices\nload_tensors: offloaded 29/29 layers to GPU\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := VerifyBackendLog(RuntimeBackendCUDA, path); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("using CPU only\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := VerifyBackendLog(RuntimeBackendCUDA, path); err == nil {
		t.Fatal("CUDA backend accepted without CUDA initialization")
	}
}
