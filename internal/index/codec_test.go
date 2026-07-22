package index

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestShardCodecIsDeterministic(t *testing.T) {
	first, err := encodeShard(map[string][]byte{
		"zeta.go":  []byte("zeta"),
		"alpha.go": []byte("alpha"),
	})
	if err != nil {
		t.Fatal(err)
	}
	second, err := encodeShard(map[string][]byte{
		"alpha.go": []byte("alpha"),
		"zeta.go":  []byte("zeta"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("map iteration changed shard bytes")
	}
	records, err := decodeShard(first)
	if err != nil {
		t.Fatal(err)
	}
	if string(records["alpha.go"]) != "alpha" || string(records["zeta.go"]) != "zeta" {
		t.Fatalf("unexpected records: %+v", records)
	}
}

func TestShardCodecRejectsImpossibleCount(t *testing.T) {
	data := make([]byte, shardHeaderSize)
	copy(data, shardMagic)
	data[len(shardMagic)] = shardVersion
	binary.BigEndian.PutUint32(data[len(shardMagic)+1:], 100)
	if _, err := decodeShard(data); err == nil {
		t.Fatal("expected malformed count error")
	}
}

func TestFileCodecRoundTrip(t *testing.T) {
	original := FileRecord{
		Path: "alpha.go",
		Hash: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		Size: 42,
		Chunks: []Chunk{{
			ID: "chunk-1", Path: "alpha.go", StartLine: 2, EndLine: 4,
			TokenCount: 7, Text: "func Alpha() {}",
		}},
	}
	encoded, err := encodeFile(original)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := decodeFile(original.Path, encoded)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Path != original.Path || decoded.Hash != original.Hash || decoded.Size != original.Size {
		t.Fatalf("unexpected file metadata: %+v", decoded)
	}
	if len(decoded.Chunks) != 1 || decoded.Chunks[0].Text != original.Chunks[0].Text {
		t.Fatalf("unexpected chunks: %+v", decoded.Chunks)
	}
}
