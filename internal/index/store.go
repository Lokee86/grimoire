package index

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/Lokee86/grimoire/internal/tokenizer"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage"
)

const stateReference = plumbing.ReferenceName("refs/grimoire/state")

var ErrConflict = errors.New("grimoire index changed during publication")

func RebuildBase(path string) (Snapshot, error) {
	repository, err := git.PlainOpen(filepath.Clean(path))
	if err != nil {
		return Snapshot{}, fmt.Errorf("open prepared index: %w", err)
	}
	ref, err := repository.Storer.Reference(stateReference)
	if err != nil {
		return Snapshot{}, fmt.Errorf("read prepared index state: %w", err)
	}
	return Snapshot{
		Version: FormatVersion, Tokenizer: tokenizer.Name,
		baseRoot: ref.Hash().String(), baseShards: make(map[string]string),
	}, nil
}

func Load(path string) (Snapshot, error) {
	repository, err := git.PlainOpen(filepath.Clean(path))
	if errors.Is(err, git.ErrRepositoryNotExists) {
		return Snapshot{}, fmt.Errorf("%w: prepared index does not exist", os.ErrNotExist)
	}
	if err != nil {
		return Snapshot{}, fmt.Errorf("open prepared index: %w", err)
	}
	ref, err := repository.Storer.Reference(stateReference)
	if errors.Is(err, plumbing.ErrReferenceNotFound) {
		return Snapshot{}, fmt.Errorf("%w: prepared index has no state", os.ErrNotExist)
	}
	if err != nil {
		return Snapshot{}, fmt.Errorf("read prepared index state: %w", err)
	}

	root, err := object.GetTree(repository.Storer, ref.Hash())
	if err != nil {
		return Snapshot{}, fmt.Errorf("read prepared index root: %w", err)
	}
	var manifestEntry *object.TreeEntry
	for entryIndex := range root.Entries {
		entry := &root.Entries[entryIndex]
		if entry.Mode != filemode.Regular {
			return Snapshot{}, fmt.Errorf("invalid prepared index entry %q", entry.Name)
		}
		if entry.Name == manifestName {
			if manifestEntry != nil {
				return Snapshot{}, fmt.Errorf("duplicate prepared index manifest")
			}
			manifestEntry = entry
			continue
		}
		if !isShardName(entry.Name) {
			return Snapshot{}, fmt.Errorf("invalid prepared index shard %q", entry.Name)
		}
	}
	if manifestEntry == nil {
		return Snapshot{}, fmt.Errorf("prepared index manifest is missing")
	}
	manifestTokenizer, err := readManifest(repository.Storer, manifestEntry.Hash)
	if err != nil {
		return Snapshot{}, err
	}

	files := make([]FileRecord, 0)
	shards := make(map[string]string)
	seen := make(map[string]struct{})
	for _, entry := range root.Entries {
		if entry.Name == manifestName {
			continue
		}
		data, err := readBlob(repository.Storer, entry.Hash)
		if err != nil {
			return Snapshot{}, fmt.Errorf("read index shard %s: %w", entry.Name, err)
		}
		records, err := decodeShard(data)
		if err != nil {
			return Snapshot{}, fmt.Errorf("decode index shard %s: %w", entry.Name, err)
		}
		for path, value := range records {
			if shardName(path) != entry.Name {
				return Snapshot{}, fmt.Errorf("index path %q is stored in the wrong shard", path)
			}
			if _, exists := seen[path]; exists {
				return Snapshot{}, fmt.Errorf("duplicate index path %q", path)
			}
			file, err := decodeFile(path, value)
			if err != nil {
				return Snapshot{}, err
			}
			seen[path] = struct{}{}
			files = append(files, file)
		}
		shards[entry.Name] = entry.Hash.String()
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return Snapshot{
		Version: FormatVersion, Tokenizer: manifestTokenizer, Files: files,
		baseRoot: ref.Hash().String(), baseShards: shards,
	}, nil
}

func Save(path string, snapshot Snapshot) error {
	if snapshot.Version != FormatVersion {
		return fmt.Errorf("cannot save index version %d", snapshot.Version)
	}
	if snapshot.Tokenizer != tokenizer.Name {
		return fmt.Errorf("cannot save index tokenizer %q", snapshot.Tokenizer)
	}
	seen := make(map[string]struct{}, len(snapshot.Files))
	for _, file := range snapshot.Files {
		if err := validateRecordPath(file.Path); err != nil {
			return err
		}
		if _, exists := seen[file.Path]; exists {
			return fmt.Errorf("duplicate index path %q", file.Path)
		}
		seen[file.Path] = struct{}{}
	}
	repository, err := openOrInit(path)
	if err != nil {
		return err
	}
	current, err := currentReference(repository.Storer)
	if err != nil {
		return err
	}
	if !matchesBase(current, snapshot.baseRoot) {
		return ErrConflict
	}

	shards := make(map[string]plumbing.Hash)
	for name, hash := range snapshot.baseShards {
		shards[name] = plumbing.NewHash(hash)
	}
	dirty := snapshot.dirtyShards
	if snapshot.baseRoot == "" {
		dirty = make(map[string]bool, 256)
		for _, file := range snapshot.Files {
			dirty[shardName(file.Path)] = true
		}
	}
	for name := range dirty {
		records := make(map[string][]byte)
		for _, file := range snapshot.Files {
			if shardName(file.Path) != name {
				continue
			}
			encoded, err := encodeFile(file)
			if err != nil {
				return err
			}
			records[file.Path] = encoded
		}
		if len(records) == 0 {
			delete(shards, name)
			continue
		}
		data, err := encodeShard(records)
		if err != nil {
			return err
		}
		hash, err := writeBlob(repository.Storer, data)
		if err != nil {
			return fmt.Errorf("write index shard %s: %w", name, err)
		}
		shards[name] = hash
	}
	manifest, err := writeManifest(repository.Storer, snapshot.Tokenizer)
	if err != nil {
		return fmt.Errorf("write index manifest: %w", err)
	}
	root, err := writeRoot(repository.Storer, manifest, shards)
	if err != nil {
		return fmt.Errorf("write index root: %w", err)
	}
	if current != nil && current.Hash() == root {
		return finishSave(path)
	}
	if err := repository.Storer.CheckAndSetReference(plumbing.NewHashReference(stateReference, root), current); err != nil {
		if errors.Is(err, storage.ErrReferenceHasChanged) || err.Error() == storage.ErrReferenceHasChanged.Error() {
			return ErrConflict
		}
		return fmt.Errorf("publish prepared index: %w", err)
	}
	return finishSave(path)
}
