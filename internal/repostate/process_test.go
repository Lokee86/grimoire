package repostate

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestRunCommandCancelsProcessTree(t *testing.T) {
	const helperKey = "GRIMOIRE_PROCESS_TREE_HELPER"
	mode := os.Getenv(helperKey)
	pidFile := os.Getenv("GRIMOIRE_PROCESS_TREE_PID_FILE")
	switch mode {
	case "child":
		_ = os.WriteFile(pidFile, []byte(strconv.Itoa(os.Getpid())), 0o644)
		for {
			time.Sleep(time.Second)
		}
	case "parent":
		child := exec.Command(os.Args[0], "-test.run=^TestRunCommandCancelsProcessTree$")
		child.Env = replaceTestEnvironment(os.Environ(), map[string]string{
			helperKey:                        "child",
			"GRIMOIRE_PROCESS_TREE_PID_FILE": pidFile,
		})
		if err := child.Start(); err != nil {
			os.Exit(2)
		}
		_ = child.Wait()
		os.Exit(0)
	}

	pidFile = filepath.Join(t.TempDir(), "child.pid")
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- runCommand(ctx, ProcessCommand{
			Executable: os.Args[0],
			Arguments:  []string{"-test.run=^TestRunCommandCancelsProcessTree$"},
			Environment: replaceTestEnvironment(os.Environ(), map[string]string{
				helperKey:                        "parent",
				"GRIMOIRE_PROCESS_TREE_PID_FILE": pidFile,
			}),
		})
	}()
	var childPID int
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(pidFile)
		if err == nil {
			childPID, _ = strconv.Atoi(strings.TrimSpace(string(data)))
			if childPID > 0 {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	if childPID == 0 {
		cancel()
		<-done
		t.Fatal("helper child did not start")
	}
	cancel()
	if err := <-done; err == nil || !strings.Contains(err.Error(), context.Canceled.Error()) {
		t.Fatalf("cancelled command error = %v", err)
	}
	deadline = time.Now().Add(5 * time.Second)
	for processAlive(childPID) && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	if processAlive(childPID) {
		t.Fatalf("child process %d survived parent cancellation", childPID)
	}
}

func replaceTestEnvironment(environment []string, replacements map[string]string) []string {
	result := make([]string, 0, len(environment)+len(replacements))
	for _, entry := range environment {
		name, _, found := strings.Cut(entry, "=")
		if found {
			if _, replaced := replacements[strings.ToUpper(name)]; replaced {
				continue
			}
		}
		result = append(result, entry)
	}
	for name, value := range replacements {
		result = append(result, fmt.Sprintf("%s=%s", name, value))
	}
	return result
}
