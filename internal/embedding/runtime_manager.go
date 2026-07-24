package embedding

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func StartManagedRuntime(ctx context.Context, config RuntimeConfig) (RuntimeState, error) {
	if err := config.ApplyDefaults(); err != nil {
		return RuntimeState{}, err
	}
	paths := runtimePaths(config.CacheDir)
	if state, err := LoadRuntimeState(paths.State); err == nil {
		if processAlive(state.SupervisorPID) && state.Status != RuntimeStatusStopped && state.Status != RuntimeStatusFailed {
			probeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			_, probeErr := probeEndpoint(probeCtx, state.Endpoint)
			cancel()
			if probeErr == nil {
				return state, nil
			}
			return RuntimeState{}, fmt.Errorf("embedding runtime supervisor %d is active but endpoint is unhealthy: %w", state.SupervisorPID, probeErr)
		}
		if err := removeRuntimeState(paths); err != nil {
			return RuntimeState{}, err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return RuntimeState{}, err
	}

	probeCtx, cancelProbe := context.WithTimeout(ctx, 3*time.Second)
	_, endpointErr := probeEndpoint(probeCtx, config.Endpoint())
	cancelProbe()
	if endpointErr == nil {
		return RuntimeState{}, fmt.Errorf("embedding endpoint %s is already active without managed runtime state", config.Endpoint())
	}
	if err := SaveRuntimeConfig(paths.Config, config); err != nil {
		return RuntimeState{}, err
	}
	_ = os.Remove(paths.Stop)
	if err := os.MkdirAll(paths.Root, 0o755); err != nil {
		return RuntimeState{}, err
	}
	if err := os.MkdirAll(filepath.Dir(config.LogPath), 0o755); err != nil {
		return RuntimeState{}, err
	}
	logFile, err := os.OpenFile(config.LogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return RuntimeState{}, err
	}
	defer logFile.Close()
	executable, err := os.Executable()
	if err != nil {
		return RuntimeState{}, err
	}
	command := exec.Command(executable, "model", "supervise", "--config", paths.Config)
	command.Stdin = nil
	command.Stdout = logFile
	command.Stderr = logFile
	configureDetachedProcess(command)
	if err := command.Start(); err != nil {
		return RuntimeState{}, fmt.Errorf("start embedding supervisor: %w", err)
	}
	supervisorPID := command.Process.Pid
	if err := command.Process.Release(); err != nil {
		return RuntimeState{}, fmt.Errorf("release embedding supervisor: %w", err)
	}

	deadline := time.NewTimer(config.StartupTimeout())
	defer deadline.Stop()
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			_ = terminateProcess(supervisorPID)
			return RuntimeState{}, ctx.Err()
		case <-deadline.C:
			_ = terminateProcess(supervisorPID)
			return RuntimeState{}, fmt.Errorf("embedding runtime did not become ready within %s", config.StartupTimeout())
		case <-ticker.C:
			state, stateErr := LoadRuntimeState(paths.State)
			if stateErr != nil {
				if errors.Is(stateErr, os.ErrNotExist) && processAlive(supervisorPID) {
					continue
				}
				return RuntimeState{}, stateErr
			}
			switch state.Status {
			case RuntimeStatusReady, RuntimeStatusDegraded:
				return state, nil
			case RuntimeStatusFailed:
				return state, errors.New(state.LastError)
			}
		}
	}
}

func StopManagedRuntime(ctx context.Context, cacheDir string) (RuntimeState, error) {
	paths, err := ManagedRuntimePaths(cacheDir)
	if err != nil {
		return RuntimeState{}, err
	}
	state, err := LoadRuntimeState(paths.State)
	if errors.Is(err, os.ErrNotExist) {
		return RuntimeState{Version: runtimeStateVersion, Status: RuntimeStatusStopped}, nil
	}
	if err != nil {
		return RuntimeState{}, err
	}
	if err := os.MkdirAll(paths.Root, 0o755); err != nil {
		return RuntimeState{}, err
	}
	if err := os.WriteFile(paths.Stop, []byte("stop\n"), 0o644); err != nil {
		return RuntimeState{}, err
	}
	if state.ProcessPID > 0 && processAlive(state.ProcessPID) {
		_ = terminateProcess(state.ProcessPID)
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	grace := time.NewTimer(5 * time.Second)
	defer grace.Stop()
	for processAlive(state.SupervisorPID) {
		select {
		case <-ctx.Done():
			return RuntimeState{}, ctx.Err()
		case <-grace.C:
			_ = terminateProcess(state.SupervisorPID)
		case <-ticker.C:
		}
	}
	state.Status = RuntimeStatusStopped
	state.ProcessPID = 0
	state.LastError = ""
	_ = SaveRuntimeState(paths.State, state)
	_ = os.Remove(paths.Stop)
	return state, nil
}

func InspectManagedRuntime(ctx context.Context, cacheDir string) (ManagedRuntimeStatus, error) {
	paths, err := ManagedRuntimePaths(cacheDir)
	if err != nil {
		return ManagedRuntimeStatus{}, err
	}
	status := ManagedRuntimeStatus{}
	if _, err := os.Stat(paths.Config); err == nil {
		status.Configured = true
	}
	state, err := LoadRuntimeState(paths.State)
	if errors.Is(err, os.ErrNotExist) {
		return status, nil
	}
	if err != nil {
		return ManagedRuntimeStatus{}, err
	}
	status.StateAvailable = true
	status.State = state
	status.SupervisorAlive = processAlive(state.SupervisorPID)
	status.ProcessAlive = processAlive(state.ProcessPID)
	probeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	probe, probeErr := probeEndpoint(probeCtx, state.Endpoint)
	cancel()
	if probeErr == nil {
		status.Healthy = true
		status.ProbeDimensions = probe.Dimensions
	} else {
		status.ProbeError = probeErr.Error()
	}
	if state.Backend == RuntimeBackendCUDA {
		gpuCtx, gpuCancel := context.WithTimeout(ctx, 5*time.Second)
		if gpu, gpuErr := ReadGPUStats(gpuCtx); gpuErr == nil {
			status.GPU = &gpu
		}
		gpuCancel()
	}
	return status, nil
}

func probeEndpoint(ctx context.Context, endpoint string) (ProbeResult, error) {
	client := NewClient(endpoint)
	client.MaxInputTokens = 0
	return client.Probe(ctx, "runtime readiness", "embedding runtime readiness probe")
}
