package mcpserver

import (
	"context"
	"encoding/json"
)

const ProtocolVersion = "2025-11-25"

// Handler executes one structured Grimoire query request.
type Handler interface {
	Query(context.Context, json.RawMessage) (any, error)
}

// HandlerFunc adapts a function to Handler.
type HandlerFunc func(context.Context, json.RawMessage) (any, error)

func (function HandlerFunc) Query(ctx context.Context, arguments json.RawMessage) (any, error) {
	return function(ctx, arguments)
}

// Options describes the MCP server and its single query tool.
type Options struct {
	Name         string
	Version      string
	Instructions string
	ToolName     string
	Description  string
	InputSchema  map[string]any
	MaxMessage   int
}

func (options Options) normalized() Options {
	if options.Name == "" {
		options.Name = "grimoire"
	}
	if options.Version == "" {
		options.Version = "0.1.0-dev"
	}
	if options.ToolName == "" {
		options.ToolName = "grimoire_query"
	}
	if options.Description == "" {
		options.Description = "Query deterministic repository structure, source, and knowledge with progressive stable handles."
	}
	if options.InputSchema == nil {
		options.InputSchema = map[string]any{
			"type":                 "object",
			"additionalProperties": true,
			"properties": map[string]any{
				"mode": map[string]any{
					"type": "string",
					"enum": []string{"orient", "search", "trace", "impact", "inspect"},
				},
				"root":    map[string]any{"type": "string"},
				"query":   map[string]any{"type": "string"},
				"handles": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				"session": map[string]any{"type": "string"},
			},
			"required": []string{"mode"},
		}
	}
	if options.MaxMessage <= 0 {
		options.MaxMessage = 10 << 20
	}
	return options
}

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type toolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type textContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type toolCallResult struct {
	Content           []textContent `json:"content"`
	StructuredContent any           `json:"structuredContent,omitempty"`
	IsError           bool          `json:"isError,omitempty"`
}
