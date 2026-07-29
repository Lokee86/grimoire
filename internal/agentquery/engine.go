package agentquery

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Lokee86/grimoire/internal/arcanagraph"
	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/lexiconfacts"
)

type Engine struct {
	root             string
	source           index.Snapshot
	lexicon          *lexiconfacts.Corpus
	lexiconSnapshot  string
	arcanaSnapshot   string
	arcanaSnapshotID string
	arcana           arcanagraph.Client
	residentArcana   *arcanagraph.Session
	warnings         []string
}

func Execute(ctx context.Context, request Request) (Response, error) {
	request = normalizeRequest(request)
	if err := validateRequest(request); err != nil {
		return Response{}, err
	}
	engine, err := openEngine(ctx, request)
	if err != nil {
		return Response{}, err
	}
	return executeWithEngine(ctx, request, engine)
}

func executeWithEngine(ctx context.Context, request Request, engine *Engine) (Response, error) {
	response := Response{
		Schema: SchemaVersion, Mode: request.Mode,
		Snapshot: Snapshot{
			Source:    engine.source.Identity(),
			Providers: make(map[string]string),
		},
		Warnings: append([]string(nil), engine.warnings...),
	}
	if engine.lexiconSnapshot != "" {
		response.Snapshot.Providers["lexicon"] = engine.lexiconSnapshot
	}
	if engine.arcanaSnapshotID != "" {
		response.Snapshot.Providers["arcana"] = engine.arcanaSnapshotID
	}
	if len(response.Snapshot.Providers) == 0 {
		response.Snapshot.Providers = nil
	}
	var err error
	switch request.Mode {
	case "orient":
		err = engine.orient(request, &response)
	case "search":
		err = engine.search(ctx, request, &response)
	case "trace":
		err = engine.trace(ctx, request, &response)
	case "impact":
		err = engine.impact(ctx, request, &response)
	case "inspect":
		err = engine.inspect(ctx, request, &response)
	}
	return response, err
}

func openEngine(ctx context.Context, request Request) (*Engine, error) {
	root, err := filepath.Abs(request.Root)
	if err != nil {
		return nil, fmt.Errorf("resolve root: %w", err)
	}
	state := request.State
	if state == "" {
		state = filepath.Join(root, ".grimoire")
	} else if !filepath.IsAbs(state) {
		state = filepath.Join(root, state)
	}
	source, err := index.Load(state)
	if err != nil {
		return nil, fmt.Errorf("load prepared index: %w", err)
	}
	engine := &Engine{root: root, source: source, arcana: arcanagraph.Client{Command: request.ArcanaCmd}}
	export, lexiconSnapshot, err := lexiconfacts.ResolveExport(ctx, lexiconfacts.ExportOptions{
		Root: root, GrimoireState: state, ExplicitDirectory: request.LexiconFacts,
		LexiconState: request.LexiconState, Command: request.LexiconCmd,
	})
	if err != nil {
		engine.warnings = append(engine.warnings, "Lexicon unavailable: "+err.Error())
	} else if export != "" {
		engine.lexicon, err = lexiconfacts.Load(export)
		if err != nil {
			engine.warnings = append(engine.warnings, "Lexicon unavailable: "+err.Error())
			engine.lexicon = nil
		}
		engine.lexiconSnapshot = lexiconSnapshot
	}
	arcanaSnapshot, arcanaID, err := arcanagraph.ResolveSnapshot(ctx, arcanagraph.StateOptions{
		Root: root, State: request.ArcanaState, LexiconState: request.LexiconState,
		ExpectedLexiconSnapshot: lexiconSnapshot, Command: request.ArcanaCmd,
	})
	if err != nil {
		engine.warnings = append(engine.warnings, "Arcana unavailable: "+err.Error())
	} else {
		engine.arcanaSnapshot, engine.arcanaSnapshotID = arcanaSnapshot, arcanaID
	}
	return engine, nil
}

func defaultTraceRelations() []string {
	return []string{
		"calls", "possible-calls", "calls-endpoint", "handled-by",
		"publishes", "consumes", "produces-message", "consumes-message",
		"invokes-process", "reads-config", "implements", "extends",
		"overrides", "uses-trait", "includes", "depends-on",
	}
}

func normalizeRequest(request Request) Request {
	if request.Schema == "" {
		request.Schema = SchemaVersion
	}
	request.Mode = strings.ToLower(strings.TrimSpace(request.Mode))
	if request.Root == "" {
		request.Root = "."
	}
	if request.Limit == 0 {
		switch request.Mode {
		case "search", "orient":
			request.Limit = 6
		default:
			request.Limit = 8
		}
	}
	if request.Depth == 0 {
		request.Depth = 3
	}
	if request.Direction == "" {
		switch request.Mode {
		case "impact":
			request.Direction = "incoming"
		case "trace":
			request.Direction = "both"
		default:
			request.Direction = "outgoing"
		}
	}
	if request.Mode == "trace" && len(request.Relations) == 0 {
		request.Relations = defaultTraceRelations()
	}
	request.Direction = strings.ToLower(strings.TrimSpace(request.Direction))
	if request.LexiconCmd == "" {
		request.LexiconCmd = "lexicon"
	}
	if request.ArcanaCmd == "" {
		request.ArcanaCmd = "arcana"
	}
	request.Detail = strings.ToLower(strings.TrimSpace(request.Detail))
	if request.Mode == "trace" && request.Detail == "" {
		request.Detail = "summary"
	}
	return request
}

func validateRequest(request Request) error {
	if request.Schema != SchemaVersion {
		return fmt.Errorf("unsupported query schema %q", request.Schema)
	}
	switch request.Mode {
	case "orient":
	case "search":
		if strings.TrimSpace(request.Query) == "" {
			return errors.New("search requires query")
		}
	case "trace", "impact":
		if strings.TrimSpace(request.Anchor) == "" && strings.TrimSpace(request.Query) == "" {
			return fmt.Errorf("%s requires anchor or query", request.Mode)
		}
	case "inspect":
		if request.Anchor == "" && len(request.Handles) == 0 {
			return errors.New("inspect requires anchor or handles")
		}
	default:
		return fmt.Errorf("unsupported query mode %q", request.Mode)
	}
	if request.Limit <= 0 || request.Limit > 200 {
		return errors.New("limit must be between 1 and 200")
	}
	if request.Depth <= 0 || request.Depth > 16 {
		return errors.New("depth must be between 1 and 16")
	}
	if request.Adjacent < 0 || request.Adjacent > 200 {
		return errors.New("adjacent_context must be between 0 and 200")
	}
	if request.Detail != "" && request.Detail != "summary" && request.Detail != "full" {
		return errors.New("detail must be summary or full")
	}
	switch request.Direction {
	case "incoming", "outgoing", "both":
	default:
		return fmt.Errorf("unsupported direction %q", request.Direction)
	}
	return nil
}

func (engine *Engine) validateSnapshot(handle Handle) error {
	var expected string
	switch handle.Provider {
	case "source":
		expected = engine.source.Identity()
	case "lexicon":
		expected = engine.lexiconSnapshot
	case "arcana":
		expected = engine.arcanaSnapshotID
	default:
		return fmt.Errorf("unknown handle provider %q", handle.Provider)
	}
	if handle.Snapshot != expected {
		return fmt.Errorf("handle snapshot %s is not active %s snapshot %s", handle.Snapshot, handle.Provider, expected)
	}
	if handle.Provider == "lexicon" && engine.lexicon == nil {
		return errors.New("Lexicon handle cannot be inspected without Lexicon state")
	}
	if handle.Provider == "arcana" && engine.arcanaSnapshot == "" {
		return errors.New("Arcana handle cannot be inspected without Arcana state")
	}
	return nil
}
