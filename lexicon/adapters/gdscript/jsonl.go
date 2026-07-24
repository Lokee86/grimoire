package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
)

func (f *factSet) render(repositoryName string, changedFiles, removedFiles []string) []byte {
	sortRecords(f.nodes, nodeSortKey)
	sortRecordsByKeys(f.edges, f.edgeOrderKeys)
	sortRecordsByKeys(f.unresolved, f.unresolvedOrderKeys)
	incremental := changedFiles != nil || removedFiles != nil
	selected := selectedPaths(changedFiles)
	owners := gdscriptNodeOwners(f.nodes)
	header := map[string]any{"adapter_version": adapterVersion, "language": language, "record": "lexicon", "repository": repositoryName, "schema_version": 1}
	if incremental {
		header["mode"] = "incremental"
		header["changed_files"] = sortedPathList(changedFiles)
		header["removed_files"] = sortedPathList(removedFiles)
		header["shared_complete"] = true
	}
	records := make([]map[string]any, 0, 1+len(f.nodes)+len(f.edges)+len(f.unresolved))
	records = append(records, header)
	for _, record := range f.nodes {
		if !incremental || includePath(gdscriptRecordOwner(record, owners), selected) {
			records = append(records, record)
		}
	}
	for _, record := range f.edges {
		if !incremental || includePath(gdscriptRecordOwner(record, owners), selected) {
			records = append(records, record)
		}
	}
	for _, record := range f.unresolved {
		if !incremental || includePath(gdscriptRecordOwner(record, owners), selected) {
			records = append(records, record)
		}
	}
	var output bytes.Buffer
	encoder := json.NewEncoder(&output)
	for _, record := range records {
		_ = encoder.Encode(record)
	}
	return output.Bytes()
}

type keyedRecord struct {
	key    string
	record map[string]any
}

func sortRecords(records []map[string]any, keyFor func(map[string]any) string) {
	keys := make([]string, len(records))
	for index, record := range records {
		keys[index] = keyFor(record)
	}
	sortRecordsByKeys(records, keys)
}

func sortRecordsByKeys(records []map[string]any, keys []string) {
	keyed := make([]keyedRecord, len(records))
	for index, record := range records {
		keyed[index] = keyedRecord{key: keys[index], record: record}
	}
	sort.Slice(keyed, func(i, j int) bool { return keyed[i].key < keyed[j].key })
	for index := range keyed {
		records[index] = keyed[index].record
	}
}

func nodeSortKey(record map[string]any) string {
	return fmt.Sprintf("%s\x00%s\x00%s\x00%s", record["id"], record["kind"], record["path"], record["qualified_name"])
}

func edgeSortKey(record map[string]any) string {
	return fmt.Sprintf("%s\x00%s\x00%s\x00%s", record["source"], record["target"], record["relation"], spanSortKey(record))
}

func unresolvedSortKey(record map[string]any) string {
	return fmt.Sprintf("%s\x00%s\x00%s\x00%s\x00%s", record["source"], record["relation"], record["expression"], record["reason"], spanSortKey(record))
}

func gdscriptNodeOwners(records []map[string]any) map[string]string {
	owners := make(map[string]string, len(records))
	for _, record := range records {
		id, _ := record["id"].(string)
		if id != "" {
			owners[id] = gdscriptDirectOwner(record)
		}
	}
	return owners
}

func gdscriptRecordOwner(record map[string]any, owners map[string]string) string {
	if owner := gdscriptDirectOwner(record); owner != "" {
		return owner
	}
	source, _ := record["source"].(string)
	return owners[source]
}

func gdscriptDirectOwner(record map[string]any) string {
	if span, ok := record["span"].(map[string]any); ok {
		if path, _ := span["path"].(string); path != "" {
			return filepath.ToSlash(path)
		}
	}
	if record["record"] == "node" && record["kind"] == "file" {
		path, _ := record["path"].(string)
		return filepath.ToSlash(path)
	}
	return ""
}

func selectedPaths(paths []string) map[string]struct{} {
	result := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		result[filepath.ToSlash(path)] = struct{}{}
	}
	return result
}

func includePath(path string, selected map[string]struct{}) bool {
	if path == "" {
		return true
	}
	_, ok := selected[filepath.ToSlash(path)]
	return ok
}

func sortedPathList(paths []string) []string {
	result := make([]string, len(paths))
	copy(result, paths)
	for index := range result {
		result[index] = filepath.ToSlash(result[index])
	}
	sort.Strings(result)
	return result
}

func spanSortKey(record map[string]any) string {
	span, _ := record["span"].(map[string]any)
	if span == nil {
		return ""
	}
	return fmt.Sprintf("%s\x00%08d\x00%08d\x00%08d\x00%08d", spanString(span, "path"), spanInt(span, "start_line"), spanInt(span, "start_column"), spanInt(span, "end_line"), spanInt(span, "end_column"))
}
