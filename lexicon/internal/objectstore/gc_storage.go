package objectstore

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type gcSnapshot struct {
	id      string
	modTime int64
}

type consumerPin struct {
	SnapshotID string `json:"snapshot_id"`
}

func (s Store) listSnapshots() ([]gcSnapshot, error) {
	entries, err := os.ReadDir(filepath.Join(s.Root, "snapshots"))
	if err != nil {
		return nil, fmt.Errorf("read Lexicon snapshots: %w", err)
	}
	result := make([]gcSnapshot, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		id := "sha256:" + strings.TrimSuffix(entry.Name(), ".json")
		if !validID(id) {
			return nil, fmt.Errorf("invalid Lexicon snapshot manifest filename %q", entry.Name())
		}
		info, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("stat Lexicon snapshot %s: %w", id, err)
		}
		result = append(result, gcSnapshot{id: id, modTime: info.ModTime().UnixNano()})
	}
	sort.Slice(result, func(left, right int) bool {
		if result[left].modTime != result[right].modTime {
			return result[left].modTime > result[right].modTime
		}
		return result[left].id < result[right].id
	})
	return result, nil
}

func (s Store) readConsumerPins() ([]string, error) {
	directory := filepath.Join(s.Root, "consumer-state")
	entries, err := os.ReadDir(directory)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read Lexicon consumer state: %w", err)
	}
	pins := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(directory, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read Lexicon consumer pin %s: %w", path, err)
		}
		var pin consumerPin
		if err := json.Unmarshal(data, &pin); err != nil {
			return nil, fmt.Errorf("decode Lexicon consumer pin %s: %w", path, err)
		}
		if !validID(pin.SnapshotID) {
			return nil, fmt.Errorf("Lexicon consumer pin %s has invalid snapshot_id", path)
		}
		pins = append(pins, pin.SnapshotID)
	}
	return pins, nil
}

func (s Store) listObjects() ([]string, error) {
	root := filepath.Join(s.Root, "objects")
	shards, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read Lexicon objects: %w", err)
	}
	var result []string
	for _, shard := range shards {
		if !shard.IsDir() || len(shard.Name()) != 2 || !isHex(shard.Name()) {
			continue
		}
		entries, err := os.ReadDir(filepath.Join(root, shard.Name()))
		if err != nil {
			return nil, fmt.Errorf("read Lexicon object shard %s: %w", shard.Name(), err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !entry.Type().IsRegular() {
				continue
			}
			id := "sha256:" + shard.Name() + entry.Name()
			if validID(id) {
				result = append(result, id)
			}
		}
	}
	sort.Strings(result)
	return result, nil
}

func isHex(value string) bool {
	_, err := hex.DecodeString(value)
	return err == nil
}
