package investigation

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

type handlePayload struct {
	Version  int    `json:"v"`
	Kind     string `json:"k"`
	Snapshot string `json:"s"`
	Digest   string `json:"d"`
}

func makeHandle(kind, snapshot, digest string) (string, error) {
	if snapshot == "" || digest == "" {
		return "", fmt.Errorf("invalid %s handle identity", kind)
	}
	payload, err := json.Marshal(handlePayload{Version: ledgerVersion, Kind: kind, Snapshot: snapshot, Digest: digest})
	if err != nil {
		return "", err
	}
	checksum := sha256.Sum256(payload)
	return "g1_" + base64.RawURLEncoding.EncodeToString(payload) + "_" + hex.EncodeToString(checksum[:8]), nil
}

func decodeHandle(token, expectedKind string) (handlePayload, error) {
	if !strings.HasPrefix(token, "g1_") {
		return handlePayload{}, fmt.Errorf("invalid %s handle", expectedKind)
	}
	parts := strings.Split(token[3:], "_")
	if len(parts) != 2 {
		return handlePayload{}, fmt.Errorf("invalid %s handle", expectedKind)
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return handlePayload{}, fmt.Errorf("invalid %s handle: %w", expectedKind, err)
	}
	checksum := sha256.Sum256(payload)
	if parts[1] != hex.EncodeToString(checksum[:8]) {
		return handlePayload{}, fmt.Errorf("invalid %s handle checksum", expectedKind)
	}
	var decoded handlePayload
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return handlePayload{}, fmt.Errorf("invalid %s handle payload: %w", expectedKind, err)
	}
	if decoded.Version != ledgerVersion || decoded.Kind != expectedKind || decoded.Snapshot == "" || decoded.Digest == "" {
		return handlePayload{}, fmt.Errorf("invalid %s handle identity", expectedKind)
	}
	return decoded, nil
}

func makeNodeHandle(snapshot, digest string) (NodeHandle, error) {
	token, err := makeHandle("node", snapshot, digest)
	return NodeHandle{token: token}, err
}
func makeSourceRangeHandle(snapshot, digest string) (SourceRangeHandle, error) {
	token, err := makeHandle("source", snapshot, digest)
	return SourceRangeHandle{token: token}, err
}
func makeGraphPathHandle(snapshot, digest string) (GraphPathHandle, error) {
	token, err := makeHandle("path", snapshot, digest)
	return GraphPathHandle{token: token}, err
}
func makeDocumentHandle(snapshot, digest string) (DocumentHandle, error) {
	token, err := makeHandle("document", snapshot, digest)
	return DocumentHandle{token: token}, err
}

func ParseNodeHandle(token string) (NodeHandle, error) {
	if _, err := decodeHandle(token, "node"); err != nil {
		return NodeHandle{}, err
	}
	return NodeHandle{token: token}, nil
}

func ParseSourceRangeHandle(token string) (SourceRangeHandle, error) {
	if _, err := decodeHandle(token, "source"); err != nil {
		return SourceRangeHandle{}, err
	}
	return SourceRangeHandle{token: token}, nil
}

func ParseGraphPathHandle(token string) (GraphPathHandle, error) {
	if _, err := decodeHandle(token, "path"); err != nil {
		return GraphPathHandle{}, err
	}
	return GraphPathHandle{token: token}, nil
}

func ParseDocumentHandle(token string) (DocumentHandle, error) {
	if _, err := decodeHandle(token, "document"); err != nil {
		return DocumentHandle{}, err
	}
	return DocumentHandle{token: token}, nil
}

func handleFor(kind, snapshot, digest string) (string, error) {
	return makeHandle(kind, snapshot, digest)
}
