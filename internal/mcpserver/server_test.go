package mcpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"testing"
)

func TestServerHandshakeListsAndCallsQueryTool(t *testing.T) {
	var received json.RawMessage
	server, err := New(Options{Instructions: "Use stable handles for follow-up queries."}, HandlerFunc(
		func(_ context.Context, arguments json.RawMessage) (any, error) {
			received = append(json.RawMessage(nil), arguments...)
			return map[string]any{"version": 1, "handles": []string{"node:test"}}, nil
		},
	))
	if err != nil {
		t.Fatal(err)
	}
	input := strings.Join([]string{
		`{"jsonrpc":"2.0","id":"init","method":"initialize","params":{"protocolVersion":"2025-11-25"}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"grimoire_query","arguments":{"mode":"orient","query":"trace respawn"}}}`,
	}, "\n") + "\n"
	var output bytes.Buffer
	if err := server.Serve(context.Background(), strings.NewReader(input), &output); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("response count = %d: %s", len(lines), output.String())
	}
	var listed struct {
		Result struct {
			Tools []struct {
				Name string `json:"name"`
			} `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(lines[1]), &listed); err != nil {
		t.Fatal(err)
	}
	if len(listed.Result.Tools) != 1 || listed.Result.Tools[0].Name != "grimoire_query" {
		t.Fatalf("unexpected tools response: %s", lines[1])
	}
	if !strings.Contains(string(received), `"trace respawn"`) {
		t.Fatalf("handler arguments = %s", received)
	}
	if !strings.Contains(lines[2], `"structuredContent":{"handles":["node:test"],"version":1}`) {
		t.Fatalf("missing structured tool result: %s", lines[2])
	}
}

func TestServerPreservesStringIDsAndReportsToolErrors(t *testing.T) {
	server, err := New(Options{}, HandlerFunc(func(context.Context, json.RawMessage) (any, error) {
		return nil, context.DeadlineExceeded
	}))
	if err != nil {
		t.Fatal(err)
	}
	input := `{"jsonrpc":"2.0","id":"call-7","method":"tools/call","params":{"name":"grimoire_query","arguments":{}}}` + "\n"
	var output bytes.Buffer
	if err := server.Serve(context.Background(), strings.NewReader(input), &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), `"id":"call-7"`) ||
		!strings.Contains(output.String(), `"isError":true`) ||
		!strings.Contains(output.String(), `context deadline exceeded`) {
		t.Fatalf("unexpected error result: %s", output.String())
	}
}

func TestServerReadsContentLengthFraming(t *testing.T) {
	server, err := New(Options{}, HandlerFunc(func(context.Context, json.RawMessage) (any, error) {
		return map[string]any{"ok": true}, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	body := `{"jsonrpc":"2.0","id":1,"method":"ping"}`
	framed := "Content-Length: " + strconv.Itoa(len(body)) + "\r\n\r\n" + body
	var output bytes.Buffer
	if err := server.Serve(context.Background(), strings.NewReader(framed), &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), `"result":{}`) {
		t.Fatalf("unexpected ping response: %s", output.String())
	}
}

func TestServerRejectsUnknownMethod(t *testing.T) {
	server, err := New(Options{}, HandlerFunc(func(context.Context, json.RawMessage) (any, error) {
		return nil, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := server.Serve(context.Background(), strings.NewReader(
		`{"jsonrpc":"2.0","id":9,"method":"resources/list"}`+"\n",
	), &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), `"code":-32601`) {
		t.Fatalf("unexpected response: %s", output.String())
	}
}
