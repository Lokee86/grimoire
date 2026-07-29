package investigation

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
)

func TestCompactHandleRoundTrip(t *testing.T) {
	snapshot := testSnapshot().Digest()
	digest := strings.Repeat("a", sha256.Size*2)
	token, err := makeHandle("node", snapshot, digest)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(token, "g2_n") || len(token) >= 100 {
		t.Fatalf("compact handle = %q bytes=%d", token, len(token))
	}
	decoded, err := decodeHandle(token, "node")
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Digest != digest || decoded.Snapshot != "" || decoded.Kind != "node" {
		t.Fatalf("decoded compact handle = %+v", decoded)
	}
}

func TestLegacyHandleRemainsReadable(t *testing.T) {
	payload, err := json.Marshal(handlePayload{
		Version:  ledgerVersion,
		Kind:     "source",
		Snapshot: testSnapshot().Digest(),
		Digest:   strings.Repeat("b", sha256.Size*2),
	})
	if err != nil {
		t.Fatal(err)
	}
	checksum := sha256.Sum256(payload)
	token := "g1_" + base64.RawURLEncoding.EncodeToString(payload) + "_" + hex.EncodeToString(checksum[:8])
	decoded, err := decodeHandle(token, "source")
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Snapshot == "" || decoded.Kind != "source" {
		t.Fatalf("decoded legacy handle = %+v", decoded)
	}
}
