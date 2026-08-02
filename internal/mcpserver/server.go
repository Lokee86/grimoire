package mcpserver

import (
	"bytes"
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
	serverBusy     = -32001
)

// Server exposes one progressive repository-query tool over MCP stdio.
type Server struct {
	options Options
	handler Handler

	writeMu sync.Mutex

	requestMu sync.Mutex
	requests  map[string]context.CancelFunc
	inFlight  chan struct{}
	calls     sync.WaitGroup

	failureMu sync.Mutex
	failure   error
}

func New(options Options, handler Handler) (*Server, error) {
	if handler == nil {
		return nil, errors.New("missing MCP query handler")
	}
	options = options.normalized()
	return &Server{
		options:  options,
		handler:  handler,
		requests: make(map[string]context.CancelFunc),
		inFlight: make(chan struct{}, options.MaxInFlight),
	}, nil
}

// Serve reads JSON-RPC requests until EOF or context cancellation. Tool calls
// execute independently so cancellation and ping notifications remain responsive.
func (server *Server) Serve(ctx context.Context, input io.Reader, output io.Writer) error {
	messages := newMessageReader(input, server.options.MaxMessage)
	for {
		if err := ctx.Err(); err != nil {
			server.cancelAll()
			server.calls.Wait()
			return errors.Join(err, server.asyncFailure())
		}
		message, err := messages.next()
		if errors.Is(err, io.EOF) {
			server.calls.Wait()
			return server.asyncFailure()
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
		if call.Method == "tools/call" {
			if err := server.startToolCall(ctx, output, call); err != nil {
				return err
			}
			continue
		}
		if err := server.write(output, server.dispatch(ctx, call)); err != nil {
			return err
		}
	}
}

func (server *Server) startToolCall(ctx context.Context, output io.Writer, call request) error {
	key, err := requestKey(call.ID)
	if err != nil {
		return server.write(output, responseError(call.ID, invalidRequest, err.Error(), nil))
	}
	select {
	case server.inFlight <- struct{}{}:
	default:
		return server.write(output, responseError(call.ID, serverBusy, "too many in-flight tool calls", map[string]any{
			"max_in_flight": server.options.MaxInFlight,
		}))
	}

	callContext, cancel := context.WithCancel(ctx)
	if !server.register(key, cancel) {
		cancel()
		<-server.inFlight
		return server.write(output, responseError(call.ID, invalidRequest, "duplicate active request id", nil))
	}

	server.calls.Add(1)
	go func() {
		defer server.calls.Done()
		defer func() { <-server.inFlight }()
		defer server.unregister(key)
		defer cancel()
		if err := server.write(output, server.dispatch(callContext, call)); err != nil {
			server.recordFailure(err)
		}
	}()
	return nil
}

func (server *Server) dispatch(ctx context.Context, call request) response {
	switch call.Method {
	case "initialize":
		result, err := server.initializeResult(call.Params)
		if err != nil {
			return responseError(call.ID, invalidParams, err.Error(), map[string]any{
				"supported_protocol_version": ProtocolVersion,
			})
		}
		return responseResult(call.ID, result)
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
	if call.Method != "notifications/cancelled" {
		return
	}
	var params struct {
		RequestID json.RawMessage `json:"requestId"`
	}
	if json.Unmarshal(call.Params, &params) != nil {
		return
	}
	key, err := requestKey(params.RequestID)
	if err != nil {
		return
	}
	server.cancel(key)
}

func (server *Server) initializeResult(params json.RawMessage) (map[string]any, error) {
	var request struct {
		ProtocolVersion string `json:"protocolVersion"`
	}
	if err := json.Unmarshal(params, &request); err != nil {
		return nil, fmt.Errorf("decode initialize parameters: %w", err)
	}
	if request.ProtocolVersion == "" {
		return nil, errors.New("initialize requires protocolVersion")
	}
	if request.ProtocolVersion != ProtocolVersion {
		return nil, fmt.Errorf("unsupported MCP protocol version %q", request.ProtocolVersion)
	}
	return map[string]any{
		"protocolVersion": ProtocolVersion,
		"capabilities": map[string]any{
			"tools": map[string]any{"listChanged": false},
		},
		"serverInfo":   map[string]any{"name": server.options.Name, "version": server.options.Version},
		"instructions": server.options.Instructions,
	}, nil
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
		return responseResult(call.ID, toolCallResult{
			Content: []textContent{{Type: "text", Text: err.Error()}},
			IsError: true,
		})
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

func requestKey(id json.RawMessage) (string, error) {
	if len(bytes.TrimSpace(id)) == 0 || bytes.Equal(bytes.TrimSpace(id), []byte("null")) {
		return "", errors.New("request id must not be null")
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, id); err != nil {
		return "", fmt.Errorf("invalid request id: %w", err)
	}
	return compact.String(), nil
}

func (server *Server) register(key string, cancel context.CancelFunc) bool {
	server.requestMu.Lock()
	defer server.requestMu.Unlock()
	if _, exists := server.requests[key]; exists {
		return false
	}
	server.requests[key] = cancel
	return true
}

func (server *Server) unregister(key string) {
	server.requestMu.Lock()
	delete(server.requests, key)
	server.requestMu.Unlock()
}

func (server *Server) cancel(key string) {
	server.requestMu.Lock()
	cancel := server.requests[key]
	server.requestMu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (server *Server) cancelAll() {
	server.requestMu.Lock()
	cancellations := make([]context.CancelFunc, 0, len(server.requests))
	for _, cancel := range server.requests {
		cancellations = append(cancellations, cancel)
	}
	server.requestMu.Unlock()
	for _, cancel := range cancellations {
		cancel()
	}
}

func (server *Server) recordFailure(err error) {
	server.failureMu.Lock()
	if server.failure == nil {
		server.failure = err
	}
	server.failureMu.Unlock()
}

func (server *Server) asyncFailure() error {
	server.failureMu.Lock()
	defer server.failureMu.Unlock()
	return server.failure
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
