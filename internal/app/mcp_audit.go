package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Lokee86/grimoire/internal/agentruntime"
)

const mcpAuditSchema = "grimoire.mcp.audit.v2"

type mcpAuditLogger struct {
	path           string
	includeContent bool
	mu             sync.Mutex
}

type mcpAuditRecord struct {
	Schema          string               `json:"schema"`
	RecordedAt      string               `json:"recorded_at"`
	ContentIncluded bool                 `json:"content_included"`
	Request         agentruntime.Request `json:"request"`
	Response        any                  `json:"response,omitempty"`
	Error           string               `json:"error,omitempty"`
}

func newMCPAuditLogger(path string, includeContent bool) *mcpAuditLogger {
	return &mcpAuditLogger{path: path, includeContent: includeContent}
}

func (logger *mcpAuditLogger) Record(request agentruntime.Request, response any, executeErr error) error {
	if logger == nil || logger.path == "" {
		return executeErr
	}
	if !logger.includeContent {
		response = redactMCPAuditContent(response)
	}
	record := mcpAuditRecord{
		Schema: mcpAuditSchema, RecordedAt: time.Now().UTC().Format(time.RFC3339Nano),
		ContentIncluded: logger.includeContent, Request: request, Response: response,
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
	if err := os.MkdirAll(filepath.Dir(logger.path), 0o700); err != nil {
		return errors.Join(executeErr, fmt.Errorf("create MCP audit directory: %w", err))
	}
	file, err := os.OpenFile(logger.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
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

func redactMCPAuditContent(value any) any {
	data, err := json.Marshal(value)
	if err != nil {
		return map[string]any{"redaction_error": err.Error()}
	}
	var decoded any
	if err := json.Unmarshal(data, &decoded); err != nil {
		return map[string]any{"redaction_error": err.Error()}
	}
	return redactMCPAuditValue(decoded)
}

func redactMCPAuditValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			switch strings.ToLower(key) {
			case "source", "excerpt":
				typed[key] = "<redacted>"
			default:
				typed[key] = redactMCPAuditValue(child)
			}
		}
		return typed
	case []any:
		for index, child := range typed {
			typed[index] = redactMCPAuditValue(child)
		}
		return typed
	default:
		return value
	}
}
