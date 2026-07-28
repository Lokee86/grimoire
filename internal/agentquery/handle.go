package agentquery

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Lokee86/grimoire/internal/structure"
)

type handlePayload struct {
	Version      int     `json:"v"`
	Provider     string  `json:"p"`
	Snapshot     string  `json:"s,omitempty"`
	NodeIdentity string  `json:"n,omitempty"`
	NodeID       *uint32 `json:"i,omitempty"`
	Path         string  `json:"f,omitempty"`
	StartLine    int     `json:"a,omitempty"`
	EndLine      int     `json:"b,omitempty"`
}

func newHandle(payload handlePayload) Handle {
	payload.Version = 1
	payload.Path = normalizePath(payload.Path)
	data, _ := json.Marshal(payload)
	return Handle{
		Value:    "grimoire:v1:" + base64.RawURLEncoding.EncodeToString(data),
		Provider: payload.Provider, Snapshot: payload.Snapshot,
		NodeIdentity: payload.NodeIdentity, NodeID: payload.NodeID,
		Path: payload.Path, StartLine: payload.StartLine, EndLine: payload.EndLine,
	}
}

func parseHandle(value string) (Handle, error) {
	encoded, found := strings.CutPrefix(strings.TrimSpace(value), "grimoire:v1:")
	if !found {
		return Handle{}, fmt.Errorf("invalid query handle")
	}
	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return Handle{}, fmt.Errorf("decode query handle: %w", err)
	}
	var payload handlePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return Handle{}, fmt.Errorf("decode query handle: %w", err)
	}
	if payload.Version != 1 || payload.Provider == "" {
		return Handle{}, fmt.Errorf("unsupported query handle")
	}
	payload.Path = normalizePath(payload.Path)
	if payload.Path != "" && (filepath.IsAbs(payload.Path) || payload.Path == ".." || strings.HasPrefix(payload.Path, "../")) {
		return Handle{}, fmt.Errorf("query handle contains invalid source path")
	}
	switch payload.Provider {
	case "source":
		if payload.Path == "" || payload.StartLine <= 0 || payload.EndLine < payload.StartLine {
			return Handle{}, fmt.Errorf("source handle has no normalized range")
		}
	case "lexicon":
		if payload.NodeIdentity == "" {
			return Handle{}, fmt.Errorf("Lexicon handle has no durable node identity")
		}
	case "arcana":
		if payload.NodeIdentity == "" || payload.NodeID == nil {
			return Handle{}, fmt.Errorf("Arcana handle has no durable and snapshot-local node identity")
		}
	default:
		return Handle{}, fmt.Errorf("unknown query handle provider %q", payload.Provider)
	}
	return newHandle(payload), nil
}

func sourceHandle(snapshot, path string, start, end int) Handle {
	return newHandle(handlePayload{
		Provider: "source", Snapshot: snapshot, Path: path,
		StartLine: start, EndLine: end,
	})
}

// NewSourceHandle creates an exact source-range handle for runtime adapters.
func NewSourceHandle(snapshot, path string, start, end int) Handle {
	return sourceHandle(snapshot, path, start, end)
}

func nodeHandle(provider, snapshot string, node structure.Node) Handle {
	payload := handlePayload{
		Provider: provider, Snapshot: snapshot,
		NodeIdentity: node.Identity, NodeID: node.NodeID,
	}
	if node.Span != nil {
		payload.Path = node.Span.Path
		payload.StartLine = node.Span.StartLine
		payload.EndLine = node.Span.EndLine
	} else {
		payload.Path = node.Path
	}
	return newHandle(payload)
}

func normalizePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	return filepath.ToSlash(filepath.Clean(path))
}
