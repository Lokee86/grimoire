package embedding

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type ServeOptions struct {
	RuntimePath string
	ModelPath   string
	Backend     string
	Host        string
	Port        int
	ContextSize int
	UbatchSize  int
	Parallel    int
	GPULayers   int
	Stdin       io.Reader
	Stdout      io.Writer
	Stderr      io.Writer
}

func Serve(options ServeOptions) error {
	candidate, err := FindRuntimeForBackend(options.RuntimePath, options.Backend, "")
	if err != nil {
		return err
	}
	if strings.TrimSpace(options.ModelPath) == "" {
		if managed, modelErr := FindModel(""); modelErr == nil {
			options.ModelPath = managed
		}
	}
	options.RuntimePath = candidate.Path
	if options.Backend == "" || options.Backend == RuntimeBackendAuto {
		options.Backend = candidate.Backend
	}
	args := ServeArgs(candidate.Path, options)
	command := exec.Command(candidate.Path, args...)
	command.Stdin = options.Stdin
	command.Stdout = options.Stdout
	command.Stderr = options.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("run embedding service: %w", err)
	}
	return nil
}

func ServeArgs(runtimePath string, options ServeOptions) []string {
	host := options.Host
	if host == "" {
		host = "127.0.0.1"
	}
	port := options.Port
	if port <= 0 {
		port = DefaultPort
	}
	contextSize := options.ContextSize
	if contextSize <= 0 {
		contextSize = DefaultContextSize
	}
	ubatchSize := options.UbatchSize
	if ubatchSize <= 0 {
		ubatchSize = DefaultUbatchSize
	}
	parallel := options.Parallel
	if parallel <= 0 {
		parallel = DefaultParallelSlots
	}
	backend := options.Backend
	if backend == "" {
		backend = inferRuntimeBackend(runtimePath)
	}
	gpuLayers := options.GPULayers
	if gpuLayers < 0 {
		if backend == RuntimeBackendCUDA || backend == RuntimeBackendVulkan {
			gpuLayers = 99
		} else {
			gpuLayers = 0
		}
	}

	args := make([]string, 0, 24)
	base := strings.TrimSuffix(strings.ToLower(filepath.Base(runtimePath)), ".exe")
	if base == "llama" {
		args = append(args, "serve")
	}
	if strings.TrimSpace(options.ModelPath) == "" {
		args = append(args, "-hf", ModelReference)
	} else {
		args = append(args, "-m", options.ModelPath)
	}
	return append(args,
		"--embedding",
		"--pooling", "last",
		"--host", host,
		"--port", strconv.Itoa(port),
		"--ctx-size", strconv.Itoa(contextSize),
		"--ubatch-size", strconv.Itoa(ubatchSize),
		"--parallel", strconv.Itoa(parallel),
		"--n-gpu-layers", strconv.Itoa(gpuLayers),
	)
}
