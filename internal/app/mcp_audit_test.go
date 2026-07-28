package app

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Lokee86/grimoire/internal/agentruntime"
)

func TestMCPAuditRecordsStructuredResponseAndExecutionError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit", "calls.jsonl")
	logger := newMCPAuditLogger(path)
	request := agentruntime.Request{}
	request.Mode = "inspect"
	response := map[string]any{
		"delta": map[string]any{
			"new_source_ranges": []any{map[string]any{
				"handle": "g1_source", "evidence": map[string]any{
					"path": "target.go", "start_line": 3, "end_line": 5,
				},
			}},
		},
	}
	executeErr := errors.New("example failure")
	if err := logger.Record(request, response, executeErr); !errors.Is(err, executeErr) {
		t.Fatalf("record error = %v, want execution error", err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		t.Fatalf("audit log is empty: %v", scanner.Err())
	}
	var record mcpAuditRecord
	if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
		t.Fatal(err)
	}
	if record.Schema != mcpAuditSchema || record.Request.Mode != "inspect" || record.Error != executeErr.Error() {
		t.Fatalf("unexpected audit record: %+v", record)
	}
	if scanner.Scan() {
		t.Fatal("audit logger wrote more than one record")
	}
}
