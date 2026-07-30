package app

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/Lokee86/grimoire/internal/agentquery"
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
	auditLog := flags.String("audit-log", "", "optional JSONL log of MCP requests and structured responses")
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

	audit := newMCPAuditLogger(*auditLog)
	queryRuntime := agentquery.NewRuntime()
	defer queryRuntime.Close()
	handler := mcpserver.HandlerFunc(func(ctx context.Context, arguments json.RawMessage) (any, error) {
		var request agentruntime.Request
		if err := json.Unmarshal(arguments, &request); err != nil {
			return nil, fmt.Errorf("decode Grimoire agent request: %w", err)
		}
		response, executeErr := agentruntime.Execute(ctx, request, agentruntime.Options{
			DefaultRoot:      *root,
			DefaultState:     *state,
			DefaultMode:      mode,
			EnsureRepository: ensureDiscoveryRepository,
			ExecuteQuery:     queryRuntime.Execute,
		})
		return response, audit.Record(request, response, executeErr)
	})
	server, err := mcpserver.New(mcpserver.Options{
		Name:         "grimoire",
		Version:      Version,
		ToolName:     "grimoire_discover",
		Description:  "Discover repository evidence through independent exact-source, source, documentation, and symbol lanes, then explicitly trace, assess impact, or inspect stable handles.",
		Instructions: "Before calling Grimoire, use direct search and file reads when an exact path or symbol is already known. Use breadth=narrow for localized ownership or impact questions; it returns handle-only discovery by default, so inspect selected handles before expanding. Use balanced only for distributed or unclear context. Stop once the owner, controlling behavior, public boundary, and relevant tests are established. Follow trace or impact only for a named unresolved relationship. Reuse one session name to deduplicate repeated evidence.",
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
			"schema": map[string]any{
				"type": "string", "const": "grimoire.discovery.v1",
			},
			"mode": map[string]any{
				"type": "string", "enum": []string{"orient", "search", "trace", "impact", "inspect"},
				"description": "Operation to perform. Search requires query; inspect requires anchor or handles; trace and impact require anchor or query.",
			},
			"root":       map[string]any{"type": "string"},
			"state":      map[string]any{"type": "string"},
			"state_mode": map[string]any{"type": "string", "enum": []string{"current-only", "refresh-if-needed", "force-refresh"}},
			"session":    map[string]any{"type": "string"},
			"query": map[string]any{
				"type": "string", "minLength": 1,
				"description": "Search text, or a textual starting point for trace or impact.",
			},
			"anchor": map[string]any{
				"type": "string", "minLength": 1,
				"description": "A returned Grimoire handle or resolvable symbol anchor. Use this or handles for inspect.",
			},
			"target": map[string]any{
				"type": "string", "minLength": 1,
				"description": "Optional destination for trace mode only. This is not a file-inspection argument.",
			},
			"handles": map[string]any{
				"type": "array", "minItems": 1,
				"items":       map[string]any{"type": "string", "minLength": 1},
				"description": "Stable handles returned by Grimoire. Inspect exact evidence by passing one or more handles here.",
			},
			"limit":            map[string]any{"type": "integer", "minimum": 1, "maximum": 200},
			"depth":            map[string]any{"type": "integer", "minimum": 1, "maximum": 16},
			"direction":        map[string]any{"type": "string", "enum": []string{"incoming", "outgoing", "both"}},
			"relations":        map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"adjacent_context": map[string]any{"type": "integer", "minimum": 0, "maximum": 200},
			"code_only":        map[string]any{"type": "boolean"},
			"detail":           map[string]any{"type": "string", "enum": []string{"handles", "summary", "full"}},
			"breadth": map[string]any{
				"type": "string", "enum": []string{"narrow", "balanced"},
				"description": "Search result budgeting. narrow caps the combined exact, symbol, and source evidence set; balanced preserves independent per-lane limits.",
			},
			"include_documents":    map[string]any{"type": "boolean"},
			"use_document_vectors": map[string]any{"type": "boolean"},
			"lexicon_facts":        map[string]any{"type": "string"},
			"lexicon_state":        map[string]any{"type": "string"},
			"lexicon_command":      map[string]any{"type": "string"},
			"arcana_state":         map[string]any{"type": "string"},
			"arcana_command":       map[string]any{"type": "string"},
		},
		"required": []string{"mode"},
		"allOf": []any{
			map[string]any{
				"if":   map[string]any{"properties": map[string]any{"mode": map[string]any{"const": "search"}}},
				"then": map[string]any{"required": []string{"query"}},
			},
			map[string]any{
				"if": map[string]any{"properties": map[string]any{"mode": map[string]any{"const": "inspect"}}},
				"then": map[string]any{"anyOf": []any{
					map[string]any{"required": []string{"anchor"}},
					map[string]any{"required": []string{"handles"}},
				}},
			},
			map[string]any{
				"if": map[string]any{"properties": map[string]any{"mode": map[string]any{"enum": []string{"trace", "impact"}}}},
				"then": map[string]any{"anyOf": []any{
					map[string]any{"required": []string{"anchor"}},
					map[string]any{"required": []string{"query"}},
				}},
			},
			map[string]any{
				"if":   map[string]any{"properties": map[string]any{"mode": map[string]any{"enum": []string{"orient", "search", "impact", "inspect"}}}},
				"then": map[string]any{"not": map[string]any{"required": []string{"target"}}},
			},
		},
	}
}
