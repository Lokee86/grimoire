package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Lokee86/grimoire/internal/embedding"
)

func runModel(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return errors.New("expected model command: setup, info, serve, or probe")
	}

	switch args[0] {
	case "setup":
		return runModelSetup(args[1:], stdout, stderr)
	case "info":
		return runModelInfo(args[1:], stdout, stderr)
	case "serve":
		return runModelServe(args[1:], stdout, stderr)
	case "probe":
		return runModelProbe(args[1:], stdout, stderr)
	default:
		return fmt.Errorf("unknown model command %q", args[0])
	}
}

func runModelSetup(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("model setup", flag.ContinueOnError)
	flags.SetOutput(stderr)
	cacheDir := flags.String("cache", "", "managed model and runtime cache directory")
	timeout := flags.Duration("timeout", 45*time.Minute, "complete setup timeout")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *timeout <= 0 {
		return errors.New("--timeout must be positive")
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	result, err := embedding.Setup(ctx, embedding.SetupOptions{
		CacheDir: *cacheDir,
		Progress: stderr,
	})
	if err != nil {
		return err
	}
	return writeJSON(stdout, result)
}

func runModelInfo(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("model info", flag.ContinueOnError)
	flags.SetOutput(stderr)
	runtimeFlag := flags.String("runtime", "", "llama.cpp server executable")
	endpoint := flags.String("endpoint", embedding.DefaultEndpoint, "OpenAI-compatible embeddings endpoint")
	if err := flags.Parse(args); err != nil {
		return err
	}

	runtimePath, runtimeErr := embedding.FindRuntime(*runtimeFlag)
	modelPath, modelErr := embedding.FindModel("")
	response := struct {
		Identity         string `json:"identity"`
		Model            string `json:"model"`
		Endpoint         string `json:"endpoint"`
		Dimensions       int    `json:"dimensions"`
		NativeDimensions int    `json:"native_dimensions"`
		QueryInstruction string `json:"query_instruction"`
		Runtime          string `json:"runtime,omitempty"`
		RuntimeAvailable bool   `json:"runtime_available"`
		RuntimeError     string `json:"runtime_error,omitempty"`
		ModelPath        string `json:"model_path,omitempty"`
		ModelAvailable   bool   `json:"model_available"`
		ModelError       string `json:"model_error,omitempty"`
	}{
		Identity: embedding.Identity(), Model: embedding.ModelReference,
		Endpoint: *endpoint, Dimensions: embedding.Dimensions,
		NativeDimensions: embedding.NativeDimensions,
		QueryInstruction: embedding.QueryInstruction,
		Runtime:          runtimePath, RuntimeAvailable: runtimeErr == nil,
		ModelPath: modelPath, ModelAvailable: modelErr == nil,
	}
	if runtimeErr != nil {
		response.RuntimeError = runtimeErr.Error()
	}
	if modelErr != nil {
		response.ModelError = modelErr.Error()
	}
	return writeJSON(stdout, response)
}

func runModelServe(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("model serve", flag.ContinueOnError)
	flags.SetOutput(stderr)
	runtimePath := flags.String("runtime", "", "llama.cpp server executable")
	modelPath := flags.String("model-file", "", "local GGUF path; defaults to the official Hugging Face Q8_0 artifact")
	host := flags.String("host", "127.0.0.1", "embedding service host")
	port := flags.Int("port", embedding.DefaultPort, "embedding service port")
	contextSize := flags.Int("context-size", 8192, "llama.cpp context size")
	ubatchSize := flags.Int("ubatch-size", 2048, "llama.cpp physical batch size")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *port <= 0 || *port > 65535 {
		return errors.New("--port must be between 1 and 65535")
	}
	if *contextSize <= 0 || *ubatchSize <= 0 {
		return errors.New("--context-size and --ubatch-size must be positive")
	}

	return embedding.Serve(embedding.ServeOptions{
		RuntimePath: *runtimePath, ModelPath: *modelPath,
		Host: *host, Port: *port, ContextSize: *contextSize, UbatchSize: *ubatchSize,
		Stdin: os.Stdin, Stdout: stdout, Stderr: stderr,
	})
}

func runModelProbe(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("model probe", flag.ContinueOnError)
	flags.SetOutput(stderr)
	endpoint := flags.String("endpoint", embedding.DefaultEndpoint, "OpenAI-compatible embeddings endpoint")
	query := flags.String("query", "where is player damage resolved", "sample repository query")
	document := flags.String("document", "func ResolveDamage applies shield and health damage to a player", "sample source or documentation passage")
	timeout := flags.Duration("timeout", 2*time.Minute, "probe request timeout")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *timeout <= 0 {
		return errors.New("--timeout must be positive")
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	result, err := embedding.NewClient(*endpoint).Probe(ctx, *query, *document)
	if err != nil {
		return err
	}
	response := struct {
		Identity   string  `json:"identity"`
		Model      string  `json:"model"`
		Endpoint   string  `json:"endpoint"`
		Dimension  int     `json:"dimensions"`
		Similarity float64 `json:"similarity"`
	}{
		Identity: embedding.Identity(), Model: embedding.ModelReference,
		Endpoint: *endpoint, Dimension: result.Dimensions, Similarity: result.Similarity,
	}
	return writeJSON(stdout, response)
}
