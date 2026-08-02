package app

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/grimoire/internal/agentruntime"
)

func TestMCPAuditRedactsContentByDefault(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit", "calls.jsonl")
	logger := newMCPAuditLogger(path, false)
	request := agentruntime.Request{}
	request.Mode = "inspect"
	response := map[string]any{
		"delta": map[string]any{
			"new_source_ranges": []any{map[string]any{
				"handle": "g1_source", "source": "private implementation", "evidence": map[string]any{
					"path": "target.go", "start_line": 3, "end_line": 5,
				},
			}},
			"document_matches": []any{map[string]any{"excerpt": "private document"}},
		},
	}
	executeErr := errors.New("example failure")
	if err := logger.Record(request, response, executeErr); !errors.Is(err, executeErr) {
		t.Fatalf("record error = %v, want execution error", err)
	}
	record := readMCPAuditRecord(t, path)
	if record.Schema != mcpAuditSchema || record.Request.Mode != "inspect" || record.Error != executeErr.Error() {
		t.Fatalf("unexpected audit record: %+v", record)
	}
	if record.ContentIncluded {
		t.Fatal("redacted audit record claims content was included")
	}
	encoded, err := json.Marshal(record.Response)
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) == "" || strings.Contains(string(encoded), "private implementation") || strings.Contains(string(encoded), "private document") {
		t.Fatalf("audit content was not redacted: %s", encoded)
	}
}

func TestMCPAuditCanIncludeContentExplicitly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "calls.jsonl")
	logger := newMCPAuditLogger(path, true)
	if err := logger.Record(agentruntime.Request{}, map[string]any{"source": "kept"}, nil); err != nil {
		t.Fatal(err)
	}
	record := readMCPAuditRecord(t, path)
	encoded, err := json.Marshal(record.Response)
	if err != nil {
		t.Fatal(err)
	}
	if !record.ContentIncluded || !strings.Contains(string(encoded), "kept") {
		t.Fatalf("explicit content missing: %+v", record)
	}
}

func readMCPAuditRecord(t *testing.T, path string) mcpAuditRecord {
	t.Helper()
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
	if scanner.Scan() {
		t.Fatal("audit logger wrote more than one record")
	}
	return record
}
