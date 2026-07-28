package agentruntime

import (
	"fmt"
	"strings"

	"github.com/Lokee86/grimoire/internal/agentquery"
	"github.com/Lokee86/grimoire/internal/investigation"
)

func resolveSessionHandles(statePath string, request *Request) error {
	if request == nil || strings.TrimSpace(request.Session) == "" || !hasSessionHandle(*request) {
		return nil
	}
	ledger, err := investigation.Open(statePath, request.Session)
	if err != nil {
		return fmt.Errorf("open investigation session for handle resolution: %w", err)
	}
	if request.Anchor, err = resolveSessionReference(ledger, request.Anchor); err != nil {
		return fmt.Errorf("resolve investigation anchor: %w", err)
	}
	if request.Target, err = resolveSessionReference(ledger, request.Target); err != nil {
		return fmt.Errorf("resolve investigation target: %w", err)
	}
	for index, value := range request.Handles {
		resolved, resolveErr := resolveSessionReference(ledger, value)
		if resolveErr != nil {
			return fmt.Errorf("resolve investigation handle %d: %w", index+1, resolveErr)
		}
		request.Handles[index] = resolved
	}
	return nil
}

func hasSessionHandle(request Request) bool {
	if strings.HasPrefix(strings.TrimSpace(request.Anchor), "g1_") || strings.HasPrefix(strings.TrimSpace(request.Target), "g1_") {
		return true
	}
	for _, value := range request.Handles {
		if strings.HasPrefix(strings.TrimSpace(value), "g1_") {
			return true
		}
	}
	return false
}

func resolveSessionReference(ledger *investigation.Ledger, value string) (string, error) {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "g1_") {
		return value, nil
	}
	kind, err := investigation.HandleKind(value)
	if err != nil {
		return "", err
	}
	switch kind {
	case "node":
		node, err := ledger.ResolveNodeHandle(value)
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(node.ID) == "" {
			return "", fmt.Errorf("resolved node has no query handle")
		}
		return node.ID, nil
	case "source":
		rangeValue, err := ledger.ResolveSourceRangeHandle(value)
		if err != nil {
			return "", err
		}
		return agentquery.NewSourceHandle(
			ledger.Snapshot().Repository,
			rangeValue.Path,
			rangeValue.StartLine,
			rangeValue.EndLine,
		).Value, nil
	case "document":
		document, err := ledger.ResolveDocumentHandle(value)
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(document.URI) != "" {
			return document.URI, nil
		}
		return document.ID, nil
	case "path":
		return "", fmt.Errorf("graph path handles cannot be used as direct query anchors")
	default:
		return "", fmt.Errorf("unsupported investigation handle kind %q", kind)
	}
}
