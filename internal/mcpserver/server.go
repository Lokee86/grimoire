package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
)

const (
	parseError     = -32700
	invalidRequest = -32600
	methodMissing  = -32601
	invalidParams  = -32602
	internalError  = -32603
)

// Server exposes one progressive repository-query tool over MCP stdio.
type Server struct {
	options Options
	handler Handler
	writeMu sync.Mutex
}

func New(options Options, handler Handler) (*Server, error) {
	if handler == nil {
		return nil, errors.New("missing MCP query handler")
	}
	return &Server{options: options.normalized(), handler: handler}, nil
}

// Serve reads JSON-RPC requests until EOF or context cancellation.
func (server *Server) Serve(ctx context.Context, input io.Reader, output io.Writer) error {
	messages := newMessageReader(input, server.options.MaxMessage)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		message, err := messages.next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			if writeErr := server.write(output, responseError(nil, parseError, err.Error(), nil)); writeErr != nil {
				return writeErr
			}
			continue
		}
		var call request
		if err := json.Unmarshal(message, &call); err != nil {
			if writeErr := server.write(output, responseError(nil, parseError, "invalid JSON", nil)); writeErr != nil {
				return writeErr
			}
			continue
		}
		if call.JSONRPC != "2.0" || call.Method == "" {
			if len(call.ID) > 0 {
				if err := server.write(output, responseError(call.ID, invalidRequest, "invalid JSON-RPC request", nil)); err != nil {
					return err
				}
			}
			continue
		}
		if len(call.ID) == 0 {
			server.handleNotification(call)
			continue
		}
		result := server.dispatch(ctx, call)
		if err := server.write(output, result); err != nil {
			return err
		}
	}
}

func (server *Server) dispatch(ctx context.Context, call request) response {
	switch call.Method {
	case "initialize":
		return responseResult(call.ID, server.initializeResult(call.Params))
	case "ping":
		return responseResult(call.ID, map[string]any{})
	case "tools/list":
		return responseResult(call.ID, map[string]any{"tools": []any{server.toolDefinition()}})
	case "tools/call":
		return server.callTool(ctx, call)
	default:
		return responseError(call.ID, methodMissing, "method not found", map[string]any{"method": call.Method})
	}
}

func (server *Server) handleNotification(call request) {
	// notifications/initialized and notifications/cancelled require no response.
}

func (server *Server) initializeResult(params json.RawMessage) map[string]any {
	protocol := ProtocolVersion
	var request struct {
		ProtocolVersion string `json:"protocolVersion"`
	}
	if json.Unmarshal(params, &request) == nil && request.ProtocolVersion != "" {
		protocol = request.ProtocolVersion
	}
	return map[string]any{
		"protocolVersion": protocol,
		"capabilities": map[string]any{
			"tools": map[string]any{"listChanged": false},
		},
		"serverInfo":   map[string]any{"name": server.options.Name, "version": server.options.Version},
		"instructions": server.options.Instructions,
	}
}

func (server *Server) toolDefinition() map[string]any {
	return map[string]any{
		"name":        server.options.ToolName,
		"description": server.options.Description,
		"inputSchema": server.options.InputSchema,
		"annotations": map[string]any{
			"readOnlyHint":    false,
			"destructiveHint": false,
			"idempotentHint":  true,
			"openWorldHint":   false,
		},
	}
}

func (server *Server) callTool(ctx context.Context, call request) response {
	var params toolCallParams
	if err := json.Unmarshal(call.Params, &params); err != nil || params.Name == "" {
		return responseError(call.ID, invalidParams, "invalid tools/call parameters", nil)
	}
	if params.Name != server.options.ToolName {
		return responseError(call.ID, methodMissing, "unknown tool", map[string]any{"tool": params.Name})
	}
	if len(params.Arguments) == 0 || string(params.Arguments) == "null" {
		params.Arguments = json.RawMessage(`{}`)
	}
	result, err := server.handler.Query(ctx, params.Arguments)
	if err != nil {
		payload := toolCallResult{
			Content: []textContent{{Type: "text", Text: err.Error()}},
			IsError: true,
		}
		return responseResult(call.ID, payload)
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		return responseError(call.ID, internalError, "encode tool result", err.Error())
	}
	return responseResult(call.ID, toolCallResult{
		Content:           []textContent{{Type: "text", Text: string(encoded)}},
		StructuredContent: result,
	})
}

func (server *Server) write(output io.Writer, value response) error {
	server.writeMu.Lock()
	defer server.writeMu.Unlock()
	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode MCP response: %w", err)
	}
	encoded = append(encoded, '\n')
	_, err = output.Write(encoded)
	return err
}

func responseResult(id json.RawMessage, result any) response {
	return response{JSONRPC: "2.0", ID: cloneID(id), Result: result}
}

func responseError(id json.RawMessage, code int, message string, data any) response {
	if len(id) == 0 {
		id = json.RawMessage("null")
	}
	return response{JSONRPC: "2.0", ID: cloneID(id), Error: &rpcError{Code: code, Message: message, Data: data}}
}

func cloneID(id json.RawMessage) json.RawMessage {
	if len(id) == 0 {
		return json.RawMessage("null")
	}
	return append(json.RawMessage(nil), id...)
}
