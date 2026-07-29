package agentquery

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

// Runtime keeps decoded discovery state and Arcana's packed graph resident for
// repeated requests. It serializes requests because the Arcana protocol is an
// ordered stream and the MCP server currently serves one request at a time.
type Runtime struct {
	mu     sync.Mutex
	key    string
	engine *Engine
	closed bool
}

func NewRuntime() *Runtime { return &Runtime{} }

func (runtime *Runtime) Execute(ctx context.Context, request Request) (Response, error) {
	request = normalizeRequest(request)
	if err := validateRequest(request); err != nil {
		return Response{}, err
	}
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	if runtime.closed {
		return Response{}, fmt.Errorf("agent query runtime is closed")
	}
	key, err := runtimeKey(request)
	if err != nil {
		return Response{}, err
	}
	if runtime.engine == nil || runtime.key != key || residentArcanaClosed(runtime.engine) {
		runtime.closeEngine()
		engine, openErr := openEngine(ctx, request)
		if openErr != nil {
			return Response{}, openErr
		}
		if engine.arcanaSnapshot != "" {
			session, sessionErr := engine.arcana.OpenSession(context.Background(), engine.arcanaSnapshot)
			if sessionErr != nil {
				engine.warnings = append(engine.warnings, "resident Arcana protocol unavailable: "+sessionErr.Error())
			} else {
				engine.residentArcana = session
			}
		}
		runtime.engine = engine
		runtime.key = key
	}
	response, executeErr := executeWithEngine(ctx, request, runtime.engine)
	if ctx.Err() != nil {
		runtime.closeEngine()
	}
	return response, executeErr
}

func (runtime *Runtime) Close() error {
	if runtime == nil {
		return nil
	}
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	if runtime.closed {
		return nil
	}
	runtime.closed = true
	return runtime.closeEngine()
}

func (runtime *Runtime) closeEngine() error {
	var err error
	if runtime.engine != nil && runtime.engine.residentArcana != nil {
		err = runtime.engine.residentArcana.Close()
	}
	runtime.engine = nil
	runtime.key = ""
	return err
}

func residentArcanaClosed(engine *Engine) bool {
	return engine != nil && engine.residentArcana != nil && engine.residentArcana.Closed()
}

func runtimeKey(request Request) (string, error) {
	root, err := filepath.Abs(request.Root)
	if err != nil {
		return "", fmt.Errorf("resolve runtime root: %w", err)
	}
	state := request.State
	if state == "" {
		state = filepath.Join(root, ".grimoire")
	} else if !filepath.IsAbs(state) {
		state = filepath.Join(root, state)
	}
	provider := request.PreparedSnapshot.Providers
	return strings.Join([]string{
		filepath.Clean(root), filepath.Clean(state),
		request.LexiconFacts, request.LexiconState, request.LexiconCmd,
		request.ArcanaState, request.ArcanaCmd,
		request.PreparedSnapshot.Source,
		provider["lexicon"], provider["arcana"],
	}, "\x00"), nil
}
