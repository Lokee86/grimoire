package index

import (
	"encoding/binary"
	"fmt"
	"io"
	"sort"

	"github.com/Lokee86/grimoire/internal/tokenizer"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage"
)

const (
	manifestName       = "manifest"
	manifestMagic      = "GRIM"
	manifestHeaderSize = len(manifestMagic) + 2 + 2
)

func writeBlob(store storage.Storer, data []byte) (plumbing.Hash, error) {
	encoded := store.NewEncodedObject()
	encoded.SetType(plumbing.BlobObject)
	encoded.SetSize(int64(len(data)))
	writer, err := encoded.Writer()
	if err != nil {
		return plumbing.ZeroHash, err
	}
	if _, err := writer.Write(data); err != nil {
		_ = writer.Close()
		return plumbing.ZeroHash, err
	}
	if err := writer.Close(); err != nil {
		return plumbing.ZeroHash, err
	}
	return store.SetEncodedObject(encoded)
}

func readBlob(store storage.Storer, hash plumbing.Hash) ([]byte, error) {
	encoded, err := store.EncodedObject(plumbing.BlobObject, hash)
	if err != nil {
		return nil, err
	}
	reader, err := encoded.Reader()
	if err != nil {
		return nil, err
	}
	data, readErr := io.ReadAll(reader)
	closeErr := reader.Close()
	if readErr != nil {
		return nil, readErr
	}
	return data, closeErr
}

func writeManifest(store storage.Storer, tokenizerName string) (plumbing.Hash, error) {
	if tokenizerName == "" || len(tokenizerName) > int(^uint16(0)) {
		return plumbing.ZeroHash, fmt.Errorf("invalid tokenizer name %q", tokenizerName)
	}
	data := make([]byte, manifestHeaderSize+len(tokenizerName))
	copy(data, manifestMagic)
	binary.BigEndian.PutUint16(data[len(manifestMagic):], uint16(FormatVersion))
	binary.BigEndian.PutUint16(data[len(manifestMagic)+2:], uint16(len(tokenizerName)))
	copy(data[manifestHeaderSize:], tokenizerName)
	return writeBlob(store, data)
}

func readManifest(store storage.Storer, hash plumbing.Hash) (string, error) {
	data, err := readBlob(store, hash)
	if err != nil {
		return "", err
	}
	if len(data) < len(manifestMagic)+2 || string(data[:len(manifestMagic)]) != manifestMagic {
		return "", fmt.Errorf("invalid Grimoire manifest")
	}
	version := int(binary.BigEndian.Uint16(data[len(manifestMagic):]))
	if version != FormatVersion {
		return "", fmt.Errorf("%w: index version %d", ErrIncompatibleIndex, version)
	}
	if len(data) < manifestHeaderSize {
		return "", fmt.Errorf("invalid Grimoire manifest")
	}
	nameLength := int(binary.BigEndian.Uint16(data[len(manifestMagic)+2:]))
	if len(data) != manifestHeaderSize+nameLength {
		return "", fmt.Errorf("invalid Grimoire manifest tokenizer")
	}
	tokenizerName := string(data[manifestHeaderSize:])
	if tokenizerName != tokenizer.Name {
		return "", fmt.Errorf("%w: tokenizer %q", ErrIncompatibleIndex, tokenizerName)
	}
	return tokenizerName, nil
}

func writeRoot(store storage.Storer, manifest plumbing.Hash, shards map[string]plumbing.Hash) (plumbing.Hash, error) {
	entries := make([]object.TreeEntry, 0, len(shards)+1)
	entries = append(entries, object.TreeEntry{Name: manifestName, Mode: filemode.Regular, Hash: manifest})
	for name, hash := range shards {
		if !isShardName(name) {
			return plumbing.ZeroHash, fmt.Errorf("invalid index shard %q", name)
		}
		entries = append(entries, object.TreeEntry{Name: name, Mode: filemode.Regular, Hash: hash})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	encoded := store.NewEncodedObject()
	if err := (&object.Tree{Entries: entries}).Encode(encoded); err != nil {
		return plumbing.ZeroHash, err
	}
	return store.SetEncodedObject(encoded)
}
