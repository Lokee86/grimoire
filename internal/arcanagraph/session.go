package arcanagraph

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/Lokee86/grimoire/internal/structure"
)

// Session keeps one Arcana protocol process and packed snapshot open across
// discovery requests. Protocol operations are serialized; cancelling an
// in-flight operation terminates the session so the next request can reopen it.
type Session struct {
	client   Client
	snapshot string

	process *exec.Cmd
	stdin   io.WriteCloser
	writer  *bufio.Writer
	stdout  io.ReadCloser
	reader  *bufio.Reader
	stderr  bytes.Buffer

	ioMu    sync.Mutex
	stateMu sync.Mutex
	closed  bool
	wait    sync.Once
	waitErr error
}

func (client Client) OpenSession(_ context.Context, snapshot string) (*Session, error) {
	session := &Session{client: client, snapshot: snapshot}
	if client.Run != nil {
		return session, nil
	}
	command := strings.TrimSpace(client.Command)
	if command == "" {
		command = "arcana"
	}
	process := exec.Command(command, "protocol", "--snapshot", snapshot)
	stdin, err := process.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("open Arcana protocol stdin: %w", err)
	}
	stdout, err := process.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return nil, fmt.Errorf("open Arcana protocol stdout: %w", err)
	}
	process.Stderr = &session.stderr
	if err := process.Start(); err != nil {
		_ = stdin.Close()
		_ = stdout.Close()
		return nil, fmt.Errorf("start Arcana protocol: %w", err)
	}
	session.process = process
	session.stdin = stdin
	session.writer = bufio.NewWriter(stdin)
	session.stdout = stdout
	session.reader = bufio.NewReader(stdout)
	return session, nil
}

func (session *Session) Closed() bool {
	if session == nil {
		return true
	}
	session.stateMu.Lock()
	defer session.stateMu.Unlock()
	return session.closed
}

func (session *Session) Close() error {
	if session == nil {
		return nil
	}
	session.stateMu.Lock()
	if !session.closed {
		session.closed = true
		if session.writer != nil {
			_ = session.writer.Flush()
		}
		if session.stdin != nil {
			_ = session.stdin.Close()
		}
	}
	session.stateMu.Unlock()
	return session.waitProcess(false)
}

func (session *Session) abort() {
	if session == nil {
		return
	}
	session.stateMu.Lock()
	if !session.closed {
		session.closed = true
		if session.stdin != nil {
			_ = session.stdin.Close()
		}
		if session.process != nil && session.process.Process != nil {
			_ = session.process.Process.Kill()
		}
	}
	session.stateMu.Unlock()
	_ = session.waitProcess(true)
}

func (session *Session) waitProcess(cancelled bool) error {
	if session.process == nil {
		return nil
	}
	session.wait.Do(func() {
		session.waitErr = session.process.Wait()
		if session.stdout != nil {
			_ = session.stdout.Close()
		}
	})
	if session.waitErr != nil && !cancelled {
		message := strings.TrimSpace(session.stderr.String())
		if message != "" {
			return fmt.Errorf("close Arcana protocol: %w: %s", session.waitErr, message)
		}
		return fmt.Errorf("close Arcana protocol: %w", session.waitErr)
	}
	return nil
}

func (session *Session) run(ctx context.Context, requests []protocolRequest) (map[string]protocolResponse, error) {
	if session == nil {
		return nil, fmt.Errorf("Arcana protocol session is nil")
	}
	if len(requests) == 0 {
		return map[string]protocolResponse{}, nil
	}
	if session.process == nil {
		return session.client.run(ctx, session.snapshot, requests)
	}
	type runResult struct {
		responses map[string]protocolResponse
		err       error
	}
	completed := make(chan runResult, 1)
	go func() {
		session.ioMu.Lock()
		defer session.ioMu.Unlock()
		responses, err := session.runLocked(requests)
		completed <- runResult{responses: responses, err: err}
	}()
	select {
	case result := <-completed:
		return result.responses, result.err
	case <-ctx.Done():
		session.abort()
		<-completed
		return nil, ctx.Err()
	}
}

func (session *Session) runLocked(requests []protocolRequest) (map[string]protocolResponse, error) {
	session.stateMu.Lock()
	closed := session.closed
	session.stateMu.Unlock()
	if closed {
		return nil, fmt.Errorf("Arcana protocol session is closed")
	}
	encoder := json.NewEncoder(session.writer)
	for _, request := range requests {
		if os.Getenv("GRIMOIRE_DEBUG_TIMINGS") != "" {
			_, _ = fmt.Fprintf(os.Stderr, "grimoire arcana request id=%s op=%s node=%v direction=%s relations=%d limit=%d\n", request.ID, request.Op, request.NodeID, request.Direction, len(request.Relations), request.Limit)
		}
		if err := encoder.Encode(request); err != nil {
			return nil, fmt.Errorf("encode Arcana request: %w", err)
		}
	}
	if err := session.writer.Flush(); err != nil {
		return nil, fmt.Errorf("flush Arcana requests: %w", err)
	}

	responses := make(map[string]protocolResponse, len(requests))
	for range requests {
		line, err := session.reader.ReadBytes('\n')
		if err != nil {
			message := strings.TrimSpace(session.stderr.String())
			if message != "" {
				return nil, fmt.Errorf("read Arcana response: %w: %s", err, message)
			}
			return nil, fmt.Errorf("read Arcana response: %w", err)
		}
		var response protocolResponse
		if err := json.Unmarshal(bytes.TrimSpace(line), &response); err != nil {
			return nil, fmt.Errorf("decode Arcana response: %w", err)
		}
		if response.Protocol != protocolID {
			return nil, fmt.Errorf("unexpected Arcana protocol %q", response.Protocol)
		}
		responses[response.ID] = response
		if os.Getenv("GRIMOIRE_DEBUG_TIMINGS") != "" {
			_, _ = fmt.Fprintf(os.Stderr, "grimoire arcana response id=%s ok=%v bytes=%d\n", response.ID, response.OK, len(line))
		}
	}
	if err := validateResponses(requests, responses); err != nil {
		return nil, err
	}
	return responses, nil
}

func (session *Session) Resolve(ctx context.Context, _ string, name, path string, limit int) ([]structure.Node, error) {
	return session.ResolveTyped(ctx, "", name, "", path, limit)
}

func (session *Session) ResolveTyped(ctx context.Context, _ string, name, kind, path string, limit int) ([]structure.Node, error) {
	return resolveWithRun(name, kind, path, limit, func(requests []protocolRequest) (map[string]protocolResponse, error) {
		return session.run(ctx, requests)
	})
}

func (session *Session) Inspect(ctx context.Context, _ string, nodeID uint32) (structure.Node, error) {
	return inspectWithRun(nodeID, func(requests []protocolRequest) (map[string]protocolResponse, error) {
		return session.run(ctx, requests)
	})
}

func (session *Session) NeighborsBatch(ctx context.Context, _ string, nodeIDs []uint32, direction string, relations []string) (map[uint32][]QueryNeighbor, error) {
	return neighborsBatchWithRun(nodeIDs, direction, relations, func(requests []protocolRequest) (map[string]protocolResponse, error) {
		return session.run(ctx, requests)
	})
}

func (session *Session) Paths(ctx context.Context, _ string, fromNodeID, toNodeID uint32, relations []string, maxDepth, limit int) ([]QueryPath, bool, error) {
	return pathsWithRun(fromNodeID, toNodeID, relations, maxDepth, limit, func(requests []protocolRequest) (map[string]protocolResponse, error) {
		return session.run(ctx, requests)
	})
}

func (session *Session) ImpactQuery(ctx context.Context, snapshot string, startNodeID uint32, direction string, relations []string, maxDepth, limit int) ([]QueryImpact, bool, error) {
	return impactWithQuery(ctx, snapshot, startNodeID, direction, relations, maxDepth, limit, session)
}

func (session *Session) Unresolved(ctx context.Context, _ string, nodeID uint32, limit int) ([]structure.Unresolved, bool, error) {
	return unresolvedWithRun(nodeID, limit, func(requests []protocolRequest) (map[string]protocolResponse, error) {
		return session.run(ctx, requests)
	})
}
