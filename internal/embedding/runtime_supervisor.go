package embedding

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"
)

func RunRuntimeSupervisor(configPath string) error {
	config, err := LoadRuntimeConfig(configPath)
	if err != nil {
		return err
	}
	paths := runtimePaths(config.CacheDir)
	writer, err := newRotatingLogWriter(config.LogPath, config.LogMaxBytes, config.LogBackups)
	if err != nil {
		return err
	}
	defer writer.Close()
	_, _ = fmt.Fprintf(writer, "\n[%s] grimoire embedding supervisor starting\n", time.Now().Format(time.RFC3339))

	candidates, err := RuntimeCandidates(config.RuntimePath, config.Backend, config.CacheDir)
	if err != nil {
		return failRuntimeState(paths.State, config, 0, err)
	}
	state := RuntimeState{
		Version: runtimeStateVersion, Status: RuntimeStatusStarting,
		SupervisorPID: os.Getpid(), RuntimeVersion: RuntimeVersion,
		ModelPath: config.ModelPath, Endpoint: config.Endpoint(),
		ContextSize: config.ContextSize, UbatchSize: config.UbatchSize, Parallel: config.Parallel,
		PerSlotContext: config.PerSlotContext(), MaxInputTokens: config.MaxInputTokens(),
		LogPath: config.LogPath, StartedAt: time.Now().UTC(),
	}
	if err := SaveRuntimeState(paths.State, state); err != nil {
		return err
	}

	failures := 0
	for failures <= config.RestartLimit {
		if stopRequested(paths.Stop) {
			state.Status = RuntimeStatusStopped
			state.ProcessPID = 0
			_ = SaveRuntimeState(paths.State, state)
			return nil
		}
		candidate := candidates[failures%len(candidates)]
		state.Status = RuntimeStatusStarting
		if failures > 0 {
			state.Status = RuntimeStatusRestarting
		}
		state.Backend = candidate.Backend
		state.RuntimePath = candidate.Path
		state.GPULayers = config.EffectiveGPULayers(candidate.Backend)
		state.Restarts = failures
		state.BackendVerified = false
		state.LastError = ""

		command := exec.Command(candidate.Path, ServeArgs(candidate.Path, ServeOptions{
			ModelPath: config.ModelPath, Backend: candidate.Backend,
			Host: config.Host, Port: config.Port, ContextSize: config.ContextSize,
			UbatchSize: config.UbatchSize, Parallel: config.Parallel,
			GPULayers: state.GPULayers,
		})...)
		command.Stdin = nil
		command.Stdout = writer
		command.Stderr = writer
		configureManagedChildProcess(command)
		if err := command.Start(); err != nil {
			failures++
			state.LastError = fmt.Sprintf("start %s runtime: %v", candidate.Backend, err)
			_ = SaveRuntimeState(paths.State, state)
			time.Sleep(config.RestartDelay())
			continue
		}
		state.ProcessPID = command.Process.Pid
		_ = SaveRuntimeState(paths.State, state)
		exit := make(chan error, 1)
		go func() { exit <- command.Wait() }()

		readyCtx, cancelReady := context.WithTimeout(context.Background(), config.StartupTimeout())
		ready := make(chan ProbeResult, 1)
		readyErr := make(chan error, 1)
		go func() {
			probe, probeErr := waitForRuntimeReady(readyCtx, config.Endpoint())
			if probeErr != nil {
				readyErr <- probeErr
				return
			}
			ready <- probe
		}()
		select {
		case processErr := <-exit:
			cancelReady()
			failures++
			state.ProcessPID = 0
			state.LastError = fmt.Sprintf("embedding runtime exited during startup: %v", processErr)
			_ = SaveRuntimeState(paths.State, state)
			time.Sleep(config.RestartDelay())
			continue
		case probeErr := <-readyErr:
			cancelReady()
			_ = terminateProcess(state.ProcessPID)
			<-exit
			failures++
			state.ProcessPID = 0
			state.LastError = probeErr.Error()
			_ = SaveRuntimeState(paths.State, state)
			time.Sleep(config.RestartDelay())
			continue
		case <-ready:
			cancelReady()
		}

		if err := verifyBackendWithRetry(candidate.Backend, config.LogPath); err != nil {
			_ = terminateProcess(state.ProcessPID)
			<-exit
			failures++
			state.ProcessPID = 0
			state.LastError = err.Error()
			_ = SaveRuntimeState(paths.State, state)
			time.Sleep(config.RestartDelay())
			continue
		}
		state.Status = RuntimeStatusReady
		state.BackendVerified = candidate.Backend != RuntimeBackendAuto
		state.ReadyAt = time.Now().UTC()
		state.LastHealthAt = state.ReadyAt
		state.LastError = ""
		_ = SaveRuntimeState(paths.State, state)
		_, _ = fmt.Fprintf(writer, "[%s] embedding runtime ready: backend=%s pid=%d endpoint=%s\n", time.Now().Format(time.RFC3339), candidate.Backend, state.ProcessPID, state.Endpoint)

		healthTicker := time.NewTicker(config.HealthInterval())
		stopped := false
		for !stopped {
			select {
			case processErr := <-exit:
				healthTicker.Stop()
				failures++
				state.ProcessPID = 0
				state.Status = RuntimeStatusRestarting
				state.LastError = fmt.Sprintf("embedding runtime exited: %v", processErr)
				_ = SaveRuntimeState(paths.State, state)
				time.Sleep(config.RestartDelay())
				stopped = true
			case <-healthTicker.C:
				if stopRequested(paths.Stop) {
					healthTicker.Stop()
					_ = terminateProcess(state.ProcessPID)
					<-exit
					state.ProcessPID = 0
					state.Status = RuntimeStatusStopped
					state.LastError = ""
					_ = SaveRuntimeState(paths.State, state)
					_ = os.Remove(paths.Stop)
					return nil
				}
				healthCtx, cancelHealth := context.WithTimeout(context.Background(), 10*time.Second)
				_, healthErr := probeEndpoint(healthCtx, config.Endpoint())
				cancelHealth()
				state.LastHealthAt = time.Now().UTC()
				if healthErr != nil {
					state.Status = RuntimeStatusDegraded
					state.LastError = healthErr.Error()
				} else {
					state.Status = RuntimeStatusReady
					state.LastError = ""
				}
				_ = SaveRuntimeState(paths.State, state)
			}
		}
	}
	return failRuntimeState(paths.State, config, state.Restarts, fmt.Errorf("embedding runtime exceeded restart limit %d: %s", config.RestartLimit, state.LastError))
}

func waitForRuntimeReady(ctx context.Context, endpoint string) (ProbeResult, error) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	var lastErr error
	for {
		probeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		probe, err := probeEndpoint(probeCtx, endpoint)
		cancel()
		if err == nil {
			return probe, nil
		}
		lastErr = err
		select {
		case <-ctx.Done():
			if lastErr != nil {
				return ProbeResult{}, fmt.Errorf("embedding runtime readiness failed: %w", lastErr)
			}
			return ProbeResult{}, ctx.Err()
		case <-ticker.C:
		}
	}
}

func verifyBackendWithRetry(backend, logPath string) error {
	var lastErr error
	for range 5 {
		if err := VerifyBackendLog(backend, logPath); err == nil {
			return nil
		} else {
			lastErr = err
		}
		time.Sleep(200 * time.Millisecond)
	}
	return lastErr
}

func stopRequested(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func failRuntimeState(path string, config RuntimeConfig, restarts int, failure error) error {
	state := RuntimeState{
		Version: runtimeStateVersion, Status: RuntimeStatusFailed,
		SupervisorPID: os.Getpid(), RuntimeVersion: RuntimeVersion,
		Endpoint: config.Endpoint(), ContextSize: config.ContextSize,
		UbatchSize: config.UbatchSize, Parallel: config.Parallel,
		PerSlotContext: config.PerSlotContext(), MaxInputTokens: config.MaxInputTokens(),
		LogPath: config.LogPath, Restarts: restarts, LastError: failure.Error(),
	}
	if err := SaveRuntimeState(path, state); err != nil {
		return errors.Join(failure, err)
	}
	return failure
}
