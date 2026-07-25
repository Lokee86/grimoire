package diffcontext

import (
	"fmt"
	"strings"

	"github.com/Lokee86/grimoire/internal/retrieve"
)

const DefaultQuery = "Review the changed code for regressions, affected callers and dependencies, relevant tests, and contract changes."

const (
	maxQueryAnchors = 16
	maxAnchorBytes  = 160
)

// EffectiveQuery supplies a useful review task when diff mode is requested
// without an explicit query.
func EffectiveQuery(query string) string {
	query = strings.TrimSpace(query)
	if query == "" {
		return DefaultQuery
	}
	return query
}

// RetrievalQuery adds bounded repository-local anchors to the human task. The
// augmented value is used only for retrieval; the emitted package retains the
// original human-facing query.
func RetrievalQuery(query string, changes []Change, candidates []retrieve.Candidate) string {
	query = EffectiveQuery(query)
	anchors := make([]string, 0, maxQueryAnchors)
	seen := make(map[string]struct{}, maxQueryAnchors)
	appendAnchor := func(value string) {
		value = strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
		if value == "" {
			return
		}
		if len(value) > maxAnchorBytes {
			value = value[:maxAnchorBytes]
		}
		if _, exists := seen[value]; exists || len(anchors) >= maxQueryAnchors {
			return
		}
		seen[value] = struct{}{}
		anchors = append(anchors, value)
	}
	for _, change := range normalizedChanges(changes) {
		appendAnchor(change.Path)
		appendAnchor(change.Summary)
	}
	for _, candidate := range candidates {
		appendAnchor(candidate.Chunk.Path)
		appendAnchor(firstMeaningfulLine(candidate.Chunk.Text))
	}
	if len(anchors) == 0 {
		return query
	}
	return fmt.Sprintf("%s\nChanged-code anchors: %s", query, strings.Join(anchors, "; "))
}

func firstMeaningfulLine(text string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") ||
			strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*") {
			continue
		}
		return line
	}
	return ""
}
