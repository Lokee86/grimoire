package repostate

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var repositoryLocks sync.Map

func now() time.Time { return time.Now() }

func elapsedMS(start time.Time) int64 { return time.Since(start).Milliseconds() }

func Ensure(ctx context.Context, options Options) (Status, error) {
	started := now()
	location, err := normalize(options)
	if err != nil {
		return Status{}, err
	}
	mode := options.Mode
	if mode == "" {
		mode = CurrentOnly
	}
	if mode != CurrentOnly && mode != RefreshIfNeeded && mode != ForceRefresh {
		return Status{}, fmt.Errorf("unsupported repository state mode %q", mode)
	}
	status, err := inspect(ctx, location)
	if err != nil {
		return Status{}, err
	}
	status.Mode = mode
	status.ElapsedMS = elapsedMS(started)
	if mode == CurrentOnly {
		return status, nil
	}

	guard := lockFor(location.grimoire)
	guard.Lock()
	defer guard.Unlock()
	fileGuard, err := acquireFileLock(ctx, filepath.Join(location.grimoire, "repostate.lock"))
	if err != nil {
		return failStatus(status, fmt.Errorf("acquire repository refresh lock: %w", err))
	}
	defer fileGuard.Close()

	// Another caller may have completed the work while this caller waited.
	status, err = inspect(ctx, location)
	if err != nil {
		return Status{}, err
	}
	status.Mode = mode
	lexiconChanged := false
	arcanaChanged := false
	if mode == ForceRefresh || status.Lexicon.Status != "current" {
		command := "scan"
		if mode == ForceRefresh {
			command = "rebuild"
		}
		if status.Lexicon.Status == "absent" || !fileExists(filepath.Join(location.lexicon, "config.json")) {
			command = "init"
		}
		if err := perform(ctx, &status, "refresh-lexicon", options.Run, commandFor(options.LexiconCommand, "lexicon"), command, "--repo", location.root); err != nil {
			return failStatus(status, err)
		}
		lexiconChanged = true
		if err := markLexiconPrepared(location); err != nil {
			return failStatus(status, err)
		}
		status, err = reinspect(ctx, location, status, mode)
		if err != nil {
			return Status{}, err
		}
		status.Mode = mode
		if status.Lexicon.Status != "current" {
			return failStatus(status, fmt.Errorf("Lexicon refresh did not produce a current snapshot: %s", strings.Join(status.Lexicon.StaleReasons, "; ")))
		}
	}
	if status.Lexicon.Snapshot != "" && (mode == ForceRefresh || status.Arcana.Status != "current" || lexiconChanged) {
		if err := perform(ctx, &status, "synchronize-arcana", options.Run, commandFor(options.ArcanaCommand, "arcana"), "sync", "--lexicon", location.lexicon, "--state", location.arcana); err != nil {
			return failStatus(status, err)
		}
		arcanaChanged = true
		status, err = reinspect(ctx, location, status, mode)
		if err != nil {
			return Status{}, err
		}
		status.Mode = mode
		if status.Arcana.Status != "current" {
			return failStatus(status, fmt.Errorf("Arcana refresh did not align with Lexicon snapshot %s: %s", status.Lexicon.Snapshot, strings.Join(status.Arcana.StaleReasons, "; ")))
		}
	}
	if mode == ForceRefresh || status.Grimoire.Status != "current" || lexiconChanged || arcanaChanged {
		arguments := []string{"index", "--root", location.root, "--state", location.grimoire}
		if status.Lexicon.Snapshot != "" {
			arguments = append(arguments, "--lexicon-state", location.lexicon, "--lexicon-command", commandFor(options.LexiconCommand, "lexicon"))
		}
		if err := perform(ctx, &status, "prepare-grimoire", options.Run, commandFor(options.GrimoireCommand, "grimoire"), arguments...); err != nil {
			return failStatus(status, err)
		}
		fingerprint, fingerprintErr := sourceFingerprint(location.root)
		if fingerprintErr != nil {
			return failStatus(status, fmt.Errorf("fingerprint source after Grimoire preparation: %w", fingerprintErr))
		}
		if err := writeMarkers(location, fingerprint, status.Lexicon.Snapshot); err != nil {
			return failStatus(status, err)
		}
	}
	status, err = reinspect(ctx, location, status, mode)
	if err != nil {
		return Status{}, err
	}
	status.Mode = mode
	if status.Grimoire.Status == "current" {
		if err := writeMarkers(location, status.Repository.SourceFingerprint, status.Lexicon.Snapshot); err != nil {
			return failStatus(status, err)
		}
		// Marker writes are deliberately excluded from the source fingerprint.
	}
	status.DeterministicQueryReady = status.Grimoire.Status == "current"
	status.ElapsedMS = elapsedMS(started)
	return status, nil
}

func perform(ctx context.Context, status *Status, name string, runner CommandRunner, command string, arguments ...string) error {
	started := now()
	action := Action{Name: name, Status: "running"}
	status.Actions = append(status.Actions, action)
	if runner == nil {
		runner = runCommand
	}
	err := runner(ctx, command, arguments...)
	last := &status.Actions[len(status.Actions)-1]
	last.ElapsedMS = elapsedMS(started)
	if err != nil {
		last.Status = "failed"
		last.Error = err.Error()
		return fmt.Errorf("%s: %w", name, err)
	}
	last.Status = "completed"
	return nil
}

func failStatus(status Status, err error) (Status, error) {
	status.Error = err.Error()
	status.Warnings = append(status.Warnings, err.Error())
	status.DeterministicQueryReady = false
	return status, err
}

func writeMarkers(location paths, fingerprint, lexiconID string) error {
	marker := stateMarker{SourceFingerprint: fingerprint, LexiconSnapshot: lexiconID}
	data, err := marshalJSON(marker)
	if err != nil {
		return err
	}
	if err := atomicWrite(filepath.Join(location.lexicon, ".repostate.json"), data); err != nil {
		return fmt.Errorf("write Lexicon preparation metadata: %w", err)
	}
	if err := atomicWrite(filepath.Join(location.grimoire, ".repostate.json"), data); err != nil {
		return fmt.Errorf("write Grimoire preparation metadata: %w", err)
	}
	return nil
}

func markLexiconPrepared(location paths) error {
	fingerprint, err := sourceFingerprint(location.root)
	if err != nil {
		return fmt.Errorf("fingerprint source after Lexicon refresh: %w", err)
	}
	id, err := readCurrent(filepath.Join(location.lexicon, "CURRENT"))
	if err != nil {
		return fmt.Errorf("read Lexicon CURRENT after refresh: %w", err)
	}
	data, err := marshalJSON(stateMarker{SourceFingerprint: fingerprint, LexiconSnapshot: id})
	if err != nil {
		return err
	}
	if err := atomicWrite(filepath.Join(location.lexicon, ".repostate.json"), data); err != nil {
		return fmt.Errorf("write Lexicon preparation metadata: %w", err)
	}
	return nil
}

func reinspect(ctx context.Context, location paths, previous Status, mode Mode) (Status, error) {
	status, err := inspect(ctx, location)
	if err != nil {
		return Status{}, err
	}
	status.Mode = mode
	status.Actions = previous.Actions
	status.Warnings = previous.Warnings
	return status, nil
}

func marshalJSON(value any) ([]byte, error) {
	data, err := jsonMarshalIndent(value)
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func atomicWrite(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temporary := fmt.Sprintf("%s.tmp-%d", path, os.Getpid())
	if err := os.WriteFile(temporary, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(temporary, path); err != nil {
		_ = os.Remove(temporary)
		return err
	}
	return nil
}

func lockFor(path string) *sync.Mutex {
	value, _ := repositoryLocks.LoadOrStore(path, &sync.Mutex{})
	return value.(*sync.Mutex)
}

func runCommand(ctx context.Context, command string, arguments ...string) error {
	var stdout, stderr bytes.Buffer
	process := exec.CommandContext(ctx, command, arguments...)
	process.Stdout, process.Stderr = &stdout, &stderr
	if err := process.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = strings.TrimSpace(stdout.String())
		}
		if message != "" {
			return fmt.Errorf("%w: %s", err, message)
		}
		return err
	}
	return nil
}

func commandFor(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
