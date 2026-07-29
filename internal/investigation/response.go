package investigation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type evidenceIdentity struct {
	Kind      string          `json:"kind"`
	Snapshot  string          `json:"snapshot"`
	Canonical json.RawMessage `json:"canonical"`
}

type legacyEvidenceIdentity struct {
	Kind     string          `json:"kind"`
	Snapshot string          `json:"snapshot"`
	Payload  json.RawMessage `json:"payload"`
}

type nodeIdentity struct {
	ID string `json:"id"`
}

type sourceRangeIdentity struct {
	Path        string `json:"path"`
	StartLine   int    `json:"start_line"`
	StartColumn int    `json:"start_column,omitempty"`
	EndLine     int    `json:"end_line"`
	EndColumn   int    `json:"end_column,omitempty"`
}

type graphPathIdentity struct {
	ID    string   `json:"id,omitempty"`
	Nodes []string `json:"nodes"`
	Edges []string `json:"edges,omitempty"`
}

type documentIdentity struct {
	ID string `json:"id"`
}

func evidenceKey(kind, snapshot string, value any) (string, json.RawMessage, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return "", nil, err
	}
	canonical, err := canonicalEvidence(kind, value)
	if err != nil {
		return "", nil, err
	}
	identity, err := json.Marshal(evidenceIdentity{Kind: kind, Snapshot: snapshot, Canonical: canonical})
	if err != nil {
		return "", nil, err
	}
	digest := sha256.Sum256(identity)
	return hex.EncodeToString(digest[:]), payload, nil
}

func legacyEvidenceKey(kind, snapshot string, payload json.RawMessage) string {
	identity, _ := json.Marshal(legacyEvidenceIdentity{Kind: kind, Snapshot: snapshot, Payload: payload})
	digest := sha256.Sum256(identity)
	return hex.EncodeToString(digest[:])
}

func canonicalEvidence(kind string, value any) (json.RawMessage, error) {
	if payload, ok := value.(json.RawMessage); ok {
		decoded, err := decodeEvidencePayload(kind, payload)
		if err != nil {
			return nil, err
		}
		return canonicalEvidence(kind, decoded)
	}

	var canonical any
	switch kind {
	case "node":
		node, ok := value.(Node)
		if !ok {
			return nil, fmt.Errorf("canonical node identity: unexpected %T", value)
		}
		canonical = nodeIdentity{ID: strings.TrimSpace(node.ID)}
	case "source":
		source, ok := value.(SourceRange)
		if !ok {
			return nil, fmt.Errorf("canonical source identity: unexpected %T", value)
		}
		canonical = sourceRangeIdentity{
			Path:        filepath.ToSlash(strings.TrimSpace(source.Path)),
			StartLine:   source.StartLine,
			StartColumn: source.StartColumn,
			EndLine:     source.EndLine,
			EndColumn:   source.EndColumn,
		}
	case "path":
		path, ok := value.(GraphPath)
		if !ok {
			return nil, fmt.Errorf("canonical graph path identity: unexpected %T", value)
		}
		canonical = graphPathIdentity{
			ID:    strings.TrimSpace(path.ID),
			Nodes: trimmedStrings(path.Nodes),
			Edges: trimmedStrings(path.Edges),
		}
	case "document":
		document, ok := value.(Document)
		if !ok {
			return nil, fmt.Errorf("canonical document identity: unexpected %T", value)
		}
		canonical = documentIdentity{ID: strings.TrimSpace(document.ID)}
	default:
		canonical = value
	}
	return json.Marshal(canonical)
}

func decodeEvidencePayload(kind string, payload json.RawMessage) (any, error) {
	var target any
	switch kind {
	case "node":
		target = &Node{}
	case "source":
		target = &SourceRange{}
	case "path":
		target = &GraphPath{}
	case "document":
		target = &Document{}
	case "question":
		target = &UnresolvedQuestion{}
	case "rejected_branch", "accepted_branch":
		target = &Branch{}
	default:
		return nil, fmt.Errorf("decode canonical evidence: unknown kind %q", kind)
	}
	if err := json.Unmarshal(payload, target); err != nil {
		return nil, fmt.Errorf("decode canonical %s evidence: %w", kind, err)
	}
	switch value := target.(type) {
	case *Node:
		return *value, nil
	case *SourceRange:
		return *value, nil
	case *GraphPath:
		return *value, nil
	case *Document:
		return *value, nil
	case *UnresolvedQuestion:
		return *value, nil
	case *Branch:
		return *value, nil
	default:
		return nil, fmt.Errorf("decode canonical evidence: unexpected %T", target)
	}
}

func trimmedStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	trimmed := make([]string, len(values))
	for index, value := range values {
		trimmed[index] = strings.TrimSpace(value)
	}
	return trimmed
}

func recordEvidence(dir string, current manifest, response Response) (Delta, storedResponse, error) {
	delta := Delta{ResponseID: responseDigest(response)}
	event := storedResponse{Version: ledgerVersion, ResponseID: delta.ResponseID, Snapshot: current.Snapshot.Digest()}
	seen := make(map[string]bool)
	add := func(kind, handleKind string, value any, newBody func(string) error, existing func(string) error) error {
		key, payload, err := evidenceKey(kind, current.Snapshot.Digest(), value)
		if err != nil {
			return err
		}
		if seen[key] {
			return nil
		}
		seen[key] = true
		meta, exists := current.Evidence[key]
		if !exists {
			meta, err = persistEvidence(dir, current, key, kind, payload, handleKind)
			if err != nil {
				return err
			}
			current.Evidence[key] = meta
			return newBody(key)
		}
		return existing(key)
	}
	for _, value := range response.Nodes {
		key := mustKey(current.Snapshot.Digest(), "node", value)
		if err := add("node", "node", value, func(key string) error {
			handle, err := makeNodeHandle(current.Snapshot.Digest(), key)
			if err != nil {
				return err
			}
			delta.NewNodes = append(delta.NewNodes, NodeRecord{Handle: handle, Evidence: value})
			return nil
		}, func(key string) error {
			handle, err := indexedHandle(current.Evidence[key], "node")
			if err != nil {
				return err
			}
			delta.PriorNodeHandles = append(delta.PriorNodeHandles, NodeHandle{token: handle})
			return nil
		}); err != nil {
			return delta, event, err
		}
		event.NodeKeys = append(event.NodeKeys, key)
	}
	for _, value := range response.SourceRanges {
		key := mustKey(current.Snapshot.Digest(), "source", value)
		if err := add("source", "source", value, func(key string) error {
			handle, err := makeSourceRangeHandle(current.Snapshot.Digest(), key)
			if err != nil {
				return err
			}
			delta.NewSourceRanges = append(delta.NewSourceRanges, SourceRangeRecord{Handle: handle, Evidence: value})
			return nil
		}, func(key string) error {
			handle, err := indexedHandle(current.Evidence[key], "source")
			if err != nil {
				return err
			}
			delta.PriorSourceRanges = append(delta.PriorSourceRanges, SourceRangeHandle{token: handle})
			return nil
		}); err != nil {
			return delta, event, err
		}
		event.SourceRangeKeys = append(event.SourceRangeKeys, key)
	}
	for _, value := range response.GraphPaths {
		key := mustKey(current.Snapshot.Digest(), "path", value)
		if err := add("path", "path", value, func(key string) error {
			handle, err := makeGraphPathHandle(current.Snapshot.Digest(), key)
			if err != nil {
				return err
			}
			delta.NewGraphPaths = append(delta.NewGraphPaths, GraphPathRecord{Handle: handle, Evidence: value})
			return nil
		}, func(key string) error {
			handle, err := indexedHandle(current.Evidence[key], "path")
			if err != nil {
				return err
			}
			delta.PriorGraphPaths = append(delta.PriorGraphPaths, GraphPathHandle{token: handle})
			return nil
		}); err != nil {
			return delta, event, err
		}
		event.GraphPathKeys = append(event.GraphPathKeys, key)
	}
	for _, value := range response.Documents {
		key := mustKey(current.Snapshot.Digest(), "document", value)
		if err := add("document", "document", value, func(key string) error {
			handle, err := makeDocumentHandle(current.Snapshot.Digest(), key)
			if err != nil {
				return err
			}
			delta.NewDocuments = append(delta.NewDocuments, DocumentRecord{Handle: handle, Evidence: value})
			return nil
		}, func(key string) error {
			handle, err := indexedHandle(current.Evidence[key], "document")
			if err != nil {
				return err
			}
			delta.PriorDocuments = append(delta.PriorDocuments, DocumentHandle{token: handle})
			return nil
		}); err != nil {
			return delta, event, err
		}
		event.DocumentKeys = append(event.DocumentKeys, key)
	}
	for _, value := range response.UnresolvedQuestions {
		key := mustKey(current.Snapshot.Digest(), "question", value)
		if err := add("question", "", value, func(string) error { delta.NewQuestions = append(delta.NewQuestions, value); return nil }, func(key string) error { delta.PriorQuestionIDs = append(delta.PriorQuestionIDs, key); return nil }); err != nil {
			return delta, event, err
		}
		event.QuestionKeys = append(event.QuestionKeys, key)
	}
	for _, value := range response.RejectedBranches {
		key := mustKey(current.Snapshot.Digest(), "rejected_branch", value)
		if err := add("rejected_branch", "", value, func(string) error { delta.NewRejectedBranches = append(delta.NewRejectedBranches, value); return nil }, func(key string) error { delta.PriorRejectedIDs = append(delta.PriorRejectedIDs, key); return nil }); err != nil {
			return delta, event, err
		}
		event.RejectedBranchKeys = append(event.RejectedBranchKeys, key)
	}
	for _, value := range response.AcceptedBranches {
		key := mustKey(current.Snapshot.Digest(), "accepted_branch", value)
		if err := add("accepted_branch", "", value, func(string) error { delta.NewAcceptedBranches = append(delta.NewAcceptedBranches, value); return nil }, func(key string) error { delta.PriorAcceptedIDs = append(delta.PriorAcceptedIDs, key); return nil }); err != nil {
			return delta, event, err
		}
		event.AcceptedBranchKeys = append(event.AcceptedBranchKeys, key)
	}
	retrievalHits, err := resolveRetrievalHits(current, response)
	if err != nil {
		return delta, event, err
	}
	delta.RetrievalHits = retrievalHits
	event.RetrievalHits = retrievalHits
	finalizeDelta(&delta)
	return delta, event, nil
}

func buildDelta(current manifest, response Response, responseID string) (Delta, error) {
	delta := Delta{ResponseID: responseID}
	seen := make(map[string]bool)
	add := func(kind string, value any, newBody func(string) error, prior func(evidenceMeta) error) error {
		key, _, err := evidenceKey(kind, current.Snapshot.Digest(), value)
		if err != nil {
			return err
		}
		if seen[key] {
			return nil
		}
		seen[key] = true
		if meta, ok := current.Evidence[key]; ok {
			return prior(meta)
		}
		return newBody(key)
	}
	for _, value := range response.Nodes {
		if err := add("node", value, func(key string) error {
			handle, err := makeNodeHandle(current.Snapshot.Digest(), key)
			if err != nil {
				return err
			}
			delta.NewNodes = append(delta.NewNodes, NodeRecord{Handle: handle, Evidence: value})
			return nil
		}, func(meta evidenceMeta) error {
			handle, err := indexedHandle(meta, "node")
			if err != nil {
				return err
			}
			delta.PriorNodeHandles = append(delta.PriorNodeHandles, NodeHandle{token: handle})
			return nil
		}); err != nil {
			return delta, err
		}
	}
	for _, value := range response.SourceRanges {
		if err := add("source", value, func(key string) error {
			handle, err := makeSourceRangeHandle(current.Snapshot.Digest(), key)
			if err != nil {
				return err
			}
			delta.NewSourceRanges = append(delta.NewSourceRanges, SourceRangeRecord{Handle: handle, Evidence: value})
			return nil
		}, func(meta evidenceMeta) error {
			handle, err := indexedHandle(meta, "source")
			if err != nil {
				return err
			}
			delta.PriorSourceRanges = append(delta.PriorSourceRanges, SourceRangeHandle{token: handle})
			return nil
		}); err != nil {
			return delta, err
		}
	}
	for _, value := range response.GraphPaths {
		if err := add("path", value, func(key string) error {
			handle, err := makeGraphPathHandle(current.Snapshot.Digest(), key)
			if err != nil {
				return err
			}
			delta.NewGraphPaths = append(delta.NewGraphPaths, GraphPathRecord{Handle: handle, Evidence: value})
			return nil
		}, func(meta evidenceMeta) error {
			handle, err := indexedHandle(meta, "path")
			if err != nil {
				return err
			}
			delta.PriorGraphPaths = append(delta.PriorGraphPaths, GraphPathHandle{token: handle})
			return nil
		}); err != nil {
			return delta, err
		}
	}
	for _, value := range response.Documents {
		if err := add("document", value, func(key string) error {
			handle, err := makeDocumentHandle(current.Snapshot.Digest(), key)
			if err != nil {
				return err
			}
			delta.NewDocuments = append(delta.NewDocuments, DocumentRecord{Handle: handle, Evidence: value})
			return nil
		}, func(meta evidenceMeta) error {
			handle, err := indexedHandle(meta, "document")
			if err != nil {
				return err
			}
			delta.PriorDocuments = append(delta.PriorDocuments, DocumentHandle{token: handle})
			return nil
		}); err != nil {
			return delta, err
		}
	}
	for _, value := range response.UnresolvedQuestions {
		if err := add("question", value, func(string) error { delta.NewQuestions = append(delta.NewQuestions, value); return nil }, func(meta evidenceMeta) error {
			delta.PriorQuestionIDs = append(delta.PriorQuestionIDs, evidenceMetaKey(meta))
			return nil
		}); err != nil {
			return delta, err
		}
	}
	for _, value := range response.RejectedBranches {
		if err := add("rejected_branch", value, func(string) error { delta.NewRejectedBranches = append(delta.NewRejectedBranches, value); return nil }, func(meta evidenceMeta) error {
			delta.PriorRejectedIDs = append(delta.PriorRejectedIDs, evidenceMetaKey(meta))
			return nil
		}); err != nil {
			return delta, err
		}
	}
	for _, value := range response.AcceptedBranches {
		if err := add("accepted_branch", value, func(string) error { delta.NewAcceptedBranches = append(delta.NewAcceptedBranches, value); return nil }, func(meta evidenceMeta) error {
			delta.PriorAcceptedIDs = append(delta.PriorAcceptedIDs, evidenceMetaKey(meta))
			return nil
		}); err != nil {
			return delta, err
		}
	}
	retrievalHits, err := resolveRetrievalHits(current, response)
	if err != nil {
		return delta, err
	}
	delta.RetrievalHits = retrievalHits
	finalizeDelta(&delta)
	return delta, nil
}

func finalizeDelta(delta *Delta) {
	if delta == nil {
		return
	}
	for index := range delta.NewNodes {
		delta.NewNodes[index].Evidence.ID = ""
		delta.NewNodes[index].Evidence.Metadata = nil
	}
	for index := range delta.NewDocuments {
		delta.NewDocuments[index].Evidence.ID = ""
		delta.NewDocuments[index].Evidence.Metadata = nil
	}
	summary := PriorEvidenceSummary{
		Nodes:        len(delta.PriorNodeHandles),
		SourceRanges: len(delta.PriorSourceRanges),
		GraphPaths:   len(delta.PriorGraphPaths),
		Documents:    len(delta.PriorDocuments),
	}
	if summary.Nodes > 0 || summary.SourceRanges > 0 || summary.GraphPaths > 0 || summary.Documents > 0 {
		delta.PriorEvidence = &summary
	}
	prior := make(map[string]bool, len(delta.PriorNodeHandles)+len(delta.PriorSourceRanges)+len(delta.PriorGraphPaths)+len(delta.PriorDocuments))
	for _, handle := range delta.PriorNodeHandles {
		prior[handle.String()] = true
	}
	for _, handle := range delta.PriorSourceRanges {
		prior[handle.String()] = true
	}
	for _, handle := range delta.PriorGraphPaths {
		prior[handle.String()] = true
	}
	for _, handle := range delta.PriorDocuments {
		prior[handle.String()] = true
	}
	for index := range delta.RetrievalHits {
		hit := &delta.RetrievalHits[index]
		hit.Reasons = compactOccurrenceReasons(hit.Reasons)
		if hit.Rank > 1 {
			hit.Reasons = nil
			hit.Support = nil
		}
		if prior[hit.EvidenceHandle] {
			hit.Reasons = nil
			hit.Support = nil
		}
		if hit.Seed != nil && prior[hit.Seed.EvidenceHandle] {
			hit.Seed.Reasons = nil
		}
	}
}

func compactOccurrenceReasons(reasons []string) []string {
	if len(reasons) <= 1 {
		return reasons
	}
	result := make([]string, 0, len(reasons))
	for _, reason := range reasons {
		if reason == "prepared source BM25 match" {
			continue
		}
		result = append(result, reason)
	}
	if len(result) == 0 {
		return reasons[:1]
	}
	return result
}

func persistEvidence(dir string, current manifest, key, kind string, payload json.RawMessage, handleKind string) (evidenceMeta, error) {
	stored := storedEvidence{Version: ledgerVersion, Kind: kind, Snapshot: current.Snapshot.Digest(), Key: key, Payload: payload}
	stored.Digest = recordDigest(stored)
	file := filepath.ToSlash(filepath.Join("records", kind+"-"+key+".json"))
	path := filepath.Join(dir, filepath.FromSlash(file))
	if err := writeImmutable(path, stored); err != nil {
		return evidenceMeta{}, err
	}
	meta := evidenceMeta{Kind: kind, File: file, Snapshot: stored.Snapshot}
	if handleKind != "" {
		meta.Handle, _ = handleFor(handleKind, stored.Snapshot, key)
	}
	return meta, nil
}

func writeImmutable(path string, value any) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return writeJSONAtomic(path, value)
}

func indexedHandle(meta evidenceMeta, kind string) (string, error) {
	if meta.Kind != kind || meta.Handle == "" {
		return "", fmt.Errorf("%w: missing %s handle", ErrCorrupt, kind)
	}
	if _, err := decodeHandle(meta.Handle, kind); err != nil {
		return "", fmt.Errorf("%w: %v", ErrCorrupt, err)
	}
	return meta.Handle, nil
}

func evidenceMetaKey(meta evidenceMeta) string {
	base := filepath.Base(filepath.FromSlash(meta.File))
	for _, kind := range []string{"node", "source", "path", "document", "question", "rejected_branch", "accepted_branch"} {
		prefix := kind + "-"
		if len(base) > len(prefix) && base[:len(prefix)] == prefix {
			return base[len(prefix) : len(base)-len(filepath.Ext(base))]
		}
	}
	return ""
}

func mustKey(snapshot, kind string, value any) string {
	key, _, _ := evidenceKey(kind, snapshot, value)
	return key
}
func responseKeys(response Response) []string {
	keys := make([]string, 0)
	for _, value := range response.Nodes {
		keys = append(keys, mustKey(response.Snapshot.Digest(), "node", value))
	}
	for _, value := range response.SourceRanges {
		keys = append(keys, mustKey(response.Snapshot.Digest(), "source", value))
	}
	for _, value := range response.GraphPaths {
		keys = append(keys, mustKey(response.Snapshot.Digest(), "path", value))
	}
	for _, value := range response.Documents {
		keys = append(keys, mustKey(response.Snapshot.Digest(), "document", value))
	}
	for _, value := range response.UnresolvedQuestions {
		keys = append(keys, mustKey(response.Snapshot.Digest(), "question", value))
	}
	for _, value := range response.RejectedBranches {
		keys = append(keys, mustKey(response.Snapshot.Digest(), "rejected_branch", value))
	}
	for _, value := range response.AcceptedBranches {
		keys = append(keys, mustKey(response.Snapshot.Digest(), "accepted_branch", value))
	}
	return keys
}
func responseKeysStored(response storedResponse) []string {
	return append(append(append(append(append(append(append([]string{}, response.NodeKeys...), response.SourceRangeKeys...), response.GraphPathKeys...), response.DocumentKeys...), response.QuestionKeys...), response.RejectedBranchKeys...), response.AcceptedBranchKeys...)
}
