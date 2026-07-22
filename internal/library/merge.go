package library

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type record map[string]any

func Merge(fullPath, incrementalPath, destination string) error {
	fullHeader, fullRecords, err := readStream(fullPath)
	if err != nil {
		return fmt.Errorf("read full library: %w", err)
	}
	incrementalHeader, incrementalRecords, err := readStream(incrementalPath)
	if err != nil {
		return fmt.Errorf("read incremental library: %w", err)
	}
	changed, changedPresent := stringSet(incrementalHeader, "changed_files")
	removed, removedPresent := stringSet(incrementalHeader, "removed_files")
	if incrementalHeader["mode"] != "incremental" || !changedPresent || !removedPresent {
		return fmt.Errorf("adapter output is not a complete incremental stream")
	}
	sharedComplete, sharedPresent := incrementalHeader["shared_complete"].(bool)
	if !sharedPresent {
		return fmt.Errorf("incremental adapter output must declare shared_complete")
	}
	if fullHeader["language"] != incrementalHeader["language"] {
		return fmt.Errorf("incremental language does not match full library")
	}

	owners := nodeOwners(fullRecords)
	for id, owner := range nodeOwners(incrementalRecords) {
		owners[id] = owner
	}
	merged := make([]record, 0, len(fullRecords)+len(incrementalRecords))
	for _, item := range fullRecords {
		owner := recordOwner(item, owners)
		if owner == "" {
			if !sharedComplete {
				merged = append(merged, item)
			}
			continue
		}
		if !changed[owner] && !removed[owner] {
			merged = append(merged, item)
		}
	}
	for _, item := range incrementalRecords {
		owner := recordOwner(item, owners)
		if owner == "" {
			if sharedComplete {
				merged = append(merged, item)
			}
			continue
		}
		if !changed[owner] {
			return fmt.Errorf("incremental record is owned by undeclared file %q", owner)
		}
		if removed[owner] {
			return fmt.Errorf("incremental record is owned by removed file %q", owner)
		}
		merged = append(merged, item)
	}

	header := cloneRecord(incrementalHeader)
	header["mode"] = "full"
	delete(header, "changed_files")
	delete(header, "removed_files")
	delete(header, "shared_complete")
	sort.Slice(merged, func(left, right int) bool {
		return sortKey(merged[left]) < sortKey(merged[right])
	})
	return writeStream(destination, header, merged)
}

func readStream(path string) (record, []record, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 32*1024*1024)
	if !scanner.Scan() {
		return nil, nil, fmt.Errorf("empty JSONL stream")
	}
	header, err := decodeRecord(scanner.Bytes())
	if err != nil || header["record"] != "lexicon" {
		return nil, nil, fmt.Errorf("invalid Lexicon header")
	}
	records := make([]record, 0)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		item, err := decodeRecord(line)
		if err != nil {
			return nil, nil, err
		}
		records = append(records, item)
	}
	return header, records, scanner.Err()
}

func decodeRecord(data []byte) (record, error) {
	var value record
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func writeStream(path string, header record, records []record) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(file)
	for _, item := range append([]record{header}, records...) {
		if err := encoder.Encode(item); err != nil {
			_ = file.Close()
			return err
		}
	}
	return file.Close()
}

func nodeOwners(records []record) map[string]string {
	owners := make(map[string]string)
	for _, item := range records {
		if item["record"] != "node" {
			continue
		}
		id, _ := item["id"].(string)
		if id != "" {
			owners[id] = directOwner(item)
		}
	}
	return owners
}

func recordOwner(item record, owners map[string]string) string {
	if owner := directOwner(item); owner != "" {
		return owner
	}
	source, _ := item["source"].(string)
	return owners[source]
}

func directOwner(item record) string {
	if owner, _ := item["owner"].(string); owner != "" {
		return normalize(owner)
	}
	if span, ok := item["span"].(map[string]any); ok {
		if path, _ := span["path"].(string); path != "" {
			return normalize(path)
		}
	}
	if item["record"] == "node" && item["kind"] == "file" {
		path, _ := item["path"].(string)
		return normalize(path)
	}
	return ""
}

func normalize(path string) string {
	path = filepath.ToSlash(filepath.Clean(filepath.FromSlash(path)))
	if path == "." || path == "" || strings.HasPrefix(path, "../") {
		return ""
	}
	return path
}

func stringSet(header record, key string) (map[string]bool, bool) {
	value, present := header[key]
	items, ok := value.([]any)
	if !present || !ok {
		return nil, present
	}
	result := make(map[string]bool, len(items))
	for _, item := range items {
		path, ok := item.(string)
		if !ok {
			return nil, false
		}
		result[normalize(path)] = true
	}
	return result, true
}

func sortKey(item record) string {
	kind, _ := item["record"].(string)
	span := spanKey(item)
	switch kind {
	case "node":
		return "0\x00" + fields(item, "id", "kind", "path", "qualified_name")
	case "edge":
		return "1\x00" + fields(item, "source", "target", "relation") + "\x00" + span
	default:
		return "2\x00" + fields(item, "source", "relation", "expression", "reason") + "\x00" + span
	}
}

func fields(item record, names ...string) string {
	values := make([]string, len(names))
	for index, name := range names {
		values[index], _ = item[name].(string)
	}
	return strings.Join(values, "\x00")
}

func spanKey(item record) string {
	span, _ := item["span"].(map[string]any)
	if span == nil {
		return ""
	}
	data, _ := json.Marshal(span)
	return string(data)
}

func cloneRecord(source record) record {
	result := make(record, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}
