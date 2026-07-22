package embedding

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type ServeOptions struct {
	RuntimePath string
	ModelPath   string
	Host        string
	Port        int
	ContextSize int
	UbatchSize  int
	Stdin       io.Reader
	Stdout      io.Writer
	Stderr      io.Writer
}

func FindRuntime(explicit string) (string, error) {
	candidates := make([]string, 0, 4)
	if strings.TrimSpace(explicit) != "" {
		candidates = append(candidates, explicit)
	}
	if configured := strings.TrimSpace(os.Getenv("GRIMOIRE_LLAMA_SERVER")); configured != "" {
		candidates = append(candidates, configured)
	}
	if managed, err := ManagedRuntimePath(""); err == nil {
		candidates = append(candidates, managed)
	}
	candidates = append(candidates, "llama-server", "llama")

	for _, candidate := range candidates {
		path, err := exec.LookPath(candidate)
		if err == nil {
			return path, nil
		}
	}
	return "", errors.New("llama.cpp runtime not found; install it with `winget install llama.cpp` or set GRIMOIRE_LLAMA_SERVER")
}

func Serve(options ServeOptions) error {
	runtimePath, err := FindRuntime(options.RuntimePath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(options.ModelPath) == "" {
		if managed, modelErr := FindModel(""); modelErr == nil {
			options.ModelPath = managed
		}
	}
	args := ServeArgs(runtimePath, options)
	command := exec.Command(runtimePath, args...)
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
		contextSize = 8192
	}
	ubatchSize := options.UbatchSize
	if ubatchSize <= 0 {
		ubatchSize = 2048
	}

	args := make([]string, 0, 20)
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
	)
}
