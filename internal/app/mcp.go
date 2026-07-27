package app

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/Lokee86/grimoire/internal/agentruntime"
	"github.com/Lokee86/grimoire/internal/mcpserver"
	"github.com/Lokee86/grimoire/internal/repostate"
)

func runMCP(args []string, input io.Reader, output, stderr io.Writer) error {
	flags := flag.NewFlagSet("mcp", flag.ContinueOnError)
	flags.SetOutput(stderr)
	root := flags.String("root", ".", "default repository root")
	state := flags.String("state", "", "default Grimoire state directory")
	stateMode := flags.String("state-mode", string(repostate.RefreshIfNeeded), "current-only, refresh-if-needed, or force-refresh")
	maxMessage := flags.Int("max-message", 10<<20, "maximum MCP request bytes")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected mcp arguments: %s", strings.Join(flags.Args(), " "))
	}
	mode := repostate.Mode(*stateMode)
	switch mode {
	case repostate.CurrentOnly, repostate.RefreshIfNeeded, repostate.ForceRefresh:
	default:
		return errors.New("--state-mode must be current-only, refresh-if-needed, or force-refresh")
	}
	if *maxMessage <= 0 {
		return errors.New("positive --max-message is required")
	}

	handler := mcpserver.HandlerFunc(func(ctx context.Context, arguments json.RawMessage) (any, error) {
		var request agentruntime.Request
		if err := json.Unmarshal(arguments, &request); err != nil {
			return nil, fmt.Errorf("decode Grimoire agent request: %w", err)
		}
		return agentruntime.Execute(ctx, request, agentruntime.Options{
			DefaultRoot:      *root,
			DefaultState:     *state,
			DefaultMode:      mode,
			EnsureRepository: ensureDiscoveryRepository,
		})
	})
	server, err := mcpserver.New(mcpserver.Options{
		Name:         "grimoire",
		Version:      Version,
		ToolName:     "grimoire_discover",
		Description:  "Discover repository evidence through independent exact-source, source, documentation, symbol, and relationship lanes, then trace, assess impact, or inspect stable handles.",
		Instructions: "Start with search. Treat source as current behavior and documentation as separately ranked intent or rationale. Follow returned symbol and relationship handles with trace or impact, and inspect only the exact evidence needed. Reuse one session name to deduplicate repeated evidence.",
		InputSchema:  agentToolInputSchema(),
		MaxMessage:   *maxMessage,
	}, handler)
	if err != nil {
		return err
	}
	return server.Serve(context.Background(), input, output)
}

func agentToolInputSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"schema":               map[string]any{"type": "string", "const": "grimoire.discovery.v1"},
			"mode":                 map[string]any{"type": "string", "enum": []string{"orient", "search", "trace", "impact", "inspect"}},
			"root":                 map[string]any{"type": "string"},
			"state":                map[string]any{"type": "string"},
			"state_mode":           map[string]any{"type": "string", "enum": []string{"current-only", "refresh-if-needed", "force-refresh"}},
			"session":              map[string]any{"type": "string"},
			"query":                map[string]any{"type": "string"},
			"anchor":               map[string]any{"type": "string"},
			"target":               map[string]any{"type": "string"},
			"handles":              map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"limit":                map[string]any{"type": "integer", "minimum": 1, "maximum": 200},
			"depth":                map[string]any{"type": "integer", "minimum": 1, "maximum": 16},
			"direction":            map[string]any{"type": "string", "enum": []string{"incoming", "outgoing", "both"}},
			"relations":            map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"adjacent_context":     map[string]any{"type": "integer", "minimum": 0, "maximum": 200},
			"code_only":            map[string]any{"type": "boolean"},
			"detail":               map[string]any{"type": "string", "enum": []string{"summary", "full"}},
			"include_documents":    map[string]any{"type": "boolean"},
			"use_document_vectors": map[string]any{"type": "boolean"},
			"lexicon_facts":        map[string]any{"type": "string"},
			"lexicon_state":        map[string]any{"type": "string"},
			"lexicon_command":      map[string]any{"type": "string"},
			"arcana_state":         map[string]any{"type": "string"},
			"arcana_command":       map[string]any{"type": "string"},
		},
		"required": []string{"mode"},
	}
}
