package investigation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type evidenceIdentity struct {
	Kind     string          `json:"kind"`
	Snapshot string          `json:"snapshot"`
	Payload  json.RawMessage `json:"payload"`
}

func evidenceKey(kind, snapshot string, value any) (string, json.RawMessage, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return "", nil, err
	}
	identity, err := json.Marshal(evidenceIdentity{Kind: kind, Snapshot: snapshot, Payload: payload})
	if err != nil {
		return "", nil, err
	}
	digest := sha256.Sum256(identity)
	return hex.EncodeToString(digest[:]), payload, nil
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
	return delta, nil
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
