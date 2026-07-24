package embedding

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func VerifyBackendLog(backend, path string) error {
	if backend == RuntimeBackendCPU || backend == RuntimeBackendAuto || backend == "" {
		return nil
	}
	content, err := readLogTail(path, 4<<20)
	if err != nil {
		return err
	}
	lower := strings.ToLower(string(content))
	switch backend {
	case RuntimeBackendCUDA:
		if !strings.Contains(lower, "cuda") {
			return errors.New("llama.cpp log does not show CUDA initialization")
		}
	case RuntimeBackendVulkan:
		if !strings.Contains(lower, "vulkan") {
			return errors.New("llama.cpp log does not show Vulkan initialization")
		}
	default:
		return fmt.Errorf("cannot verify unknown runtime backend %q", backend)
	}
	if !strings.Contains(lower, "offload") && !strings.Contains(lower, "device 0") {
		return fmt.Errorf("llama.cpp log does not show %s model placement", backend)
	}
	if strings.Contains(lower, "offloaded 0/") {
		return fmt.Errorf("llama.cpp reported zero layers offloaded to %s", backend)
	}
	return nil
}

func readLogTail(path string, maximum int64) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	start := info.Size() - maximum
	if start < 0 {
		start = 0
	}
	if _, err := file.Seek(start, io.SeekStart); err != nil {
		return nil, err
	}
	return io.ReadAll(file)
}
