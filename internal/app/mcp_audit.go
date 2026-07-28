package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Lokee86/grimoire/internal/agentruntime"
)

const mcpAuditSchema = "grimoire.mcp.audit.v1"

type mcpAuditLogger struct {
	path string
	mu   sync.Mutex
}

type mcpAuditRecord struct {
	Schema     string               `json:"schema"`
	RecordedAt string               `json:"recorded_at"`
	Request    agentruntime.Request `json:"request"`
	Response   any                  `json:"response,omitempty"`
	Error      string               `json:"error,omitempty"`
}

func newMCPAuditLogger(path string) *mcpAuditLogger {
	return &mcpAuditLogger{path: path}
}

func (logger *mcpAuditLogger) Record(request agentruntime.Request, response any, executeErr error) error {
	if logger == nil || logger.path == "" {
		return executeErr
	}
	record := mcpAuditRecord{
		Schema: mcpAuditSchema, RecordedAt: time.Now().UTC().Format(time.RFC3339Nano),
		Request: request, Response: response,
	}
	if executeErr != nil {
		record.Error = executeErr.Error()
	}
	data, err := json.Marshal(record)
	if err != nil {
		return errors.Join(executeErr, fmt.Errorf("encode MCP audit record: %w", err))
	}
	logger.mu.Lock()
	defer logger.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(logger.path), 0o755); err != nil {
		return errors.Join(executeErr, fmt.Errorf("create MCP audit directory: %w", err))
	}
	file, err := os.OpenFile(logger.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return errors.Join(executeErr, fmt.Errorf("open MCP audit log: %w", err))
	}
	_, writeErr := file.Write(append(data, '\n'))
	closeErr := file.Close()
	if writeErr != nil || closeErr != nil {
		return errors.Join(executeErr, writeErr, closeErr)
	}
	return executeErr
}
