package index

import (
	"encoding/binary"
	"errors"
	"testing"

	"github.com/Lokee86/grimoire/internal/tokenizer"
	"github.com/go-git/go-git/v5/storage/memory"
)

func TestManifestRecordsTokenizer(t *testing.T) {
	store := memory.NewStorage()
	hash, err := writeManifest(store, tokenizer.Name)
	if err != nil {
		t.Fatal(err)
	}
	name, err := readManifest(store, hash)
	if err != nil {
		t.Fatal(err)
	}
	if name != tokenizer.Name {
		t.Fatalf("unexpected tokenizer %q", name)
	}
}

func TestManifestRejectsVersionOneIndex(t *testing.T) {
	store := memory.NewStorage()
	data := make([]byte, 6)
	copy(data, manifestMagic)
	binary.BigEndian.PutUint16(data[len(manifestMagic):], 1)
	hash, err := writeBlob(store, data)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := readManifest(store, hash); !errors.Is(err, ErrIncompatibleIndex) {
		t.Fatalf("expected incompatible index error, got %v", err)
	}
}
