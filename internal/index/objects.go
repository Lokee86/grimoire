package index

import (
	"encoding/binary"
	"fmt"
	"io"
	"sort"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage"
)

const manifestName = "manifest"

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

func writeManifest(store storage.Storer) (plumbing.Hash, error) {
	data := make([]byte, 6)
	copy(data, "GRIM")
	binary.BigEndian.PutUint16(data[4:], uint16(FormatVersion))
	return writeBlob(store, data)
}

func readManifest(store storage.Storer, hash plumbing.Hash) error {
	data, err := readBlob(store, hash)
	if err != nil {
		return err
	}
	if len(data) != 6 || string(data[:4]) != "GRIM" {
		return fmt.Errorf("invalid Grimoire manifest")
	}
	version := int(binary.BigEndian.Uint16(data[4:]))
	if version != FormatVersion {
		return fmt.Errorf("unsupported index version %d", version)
	}
	return nil
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
