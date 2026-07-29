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
	if snapshot == "" || !validDigest(digest) {
		return "", fmt.Errorf("invalid %s handle identity", kind)
	}
	code, ok := compactKindCode(kind)
	if !ok {
		return "", fmt.Errorf("invalid %s handle identity", kind)
	}
	body := code + digest
	checksum := sha256.Sum256([]byte(body))
	return "g2_" + body + "_" + hex.EncodeToString(checksum[:8]), nil
}

func decodeAnyHandle(token string) (handlePayload, error) {
	if strings.HasPrefix(token, "g2_") {
		return decodeCompactHandle(token)
	}
	return decodeLegacyHandle(token)
}

func decodeCompactHandle(token string) (handlePayload, error) {
	parts := strings.Split(token[3:], "_")
	if len(parts) != 2 || len(parts[0]) != 1+sha256.Size*2 {
		return handlePayload{}, fmt.Errorf("invalid investigation handle")
	}
	kind, ok := compactKind(parts[0][:1])
	if !ok || !validDigest(parts[0][1:]) {
		return handlePayload{}, fmt.Errorf("invalid investigation handle identity")
	}
	checksum := sha256.Sum256([]byte(parts[0]))
	if parts[1] != hex.EncodeToString(checksum[:8]) {
		return handlePayload{}, fmt.Errorf("invalid investigation handle checksum")
	}
	return handlePayload{Version: ledgerVersion, Kind: kind, Digest: parts[0][1:]}, nil
}

func decodeLegacyHandle(token string) (handlePayload, error) {
	if !strings.HasPrefix(token, "g1_") {
		return handlePayload{}, fmt.Errorf("invalid investigation handle")
	}
	parts := strings.Split(token[3:], "_")
	if len(parts) != 2 {
		return handlePayload{}, fmt.Errorf("invalid investigation handle")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return handlePayload{}, fmt.Errorf("invalid investigation handle: %w", err)
	}
	checksum := sha256.Sum256(payload)
	if parts[1] != hex.EncodeToString(checksum[:8]) {
		return handlePayload{}, fmt.Errorf("invalid investigation handle checksum")
	}
	var decoded handlePayload
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return handlePayload{}, fmt.Errorf("invalid investigation handle payload: %w", err)
	}
	if decoded.Version != ledgerVersion || !validEvidenceKind(decoded.Kind) || decoded.Snapshot == "" || !validDigest(decoded.Digest) {
		return handlePayload{}, fmt.Errorf("invalid investigation handle identity")
	}
	return decoded, nil
}

func compactKindCode(kind string) (string, bool) {
	switch kind {
	case "node":
		return "n", true
	case "source":
		return "s", true
	case "path":
		return "p", true
	case "document":
		return "d", true
	default:
		return "", false
	}
}

func compactKind(code string) (string, bool) {
	switch code {
	case "n":
		return "node", true
	case "s":
		return "source", true
	case "p":
		return "path", true
	case "d":
		return "document", true
	default:
		return "", false
	}
}

func decodeHandle(token, expectedKind string) (handlePayload, error) {
	decoded, err := decodeAnyHandle(token)
	if err != nil {
		return handlePayload{}, fmt.Errorf("invalid %s handle: %w", expectedKind, err)
	}
	if decoded.Kind != expectedKind {
		return handlePayload{}, fmt.Errorf("invalid %s handle identity", expectedKind)
	}
	return decoded, nil
}

// HandleKind returns the evidence kind encoded by an opaque investigation handle.
func HandleKind(token string) (string, error) {
	decoded, err := decodeAnyHandle(token)
	if err != nil {
		return "", err
	}
	return decoded.Kind, nil
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
