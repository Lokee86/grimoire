package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

func (f *factSet) render(repositoryName string) []byte {
	sortRecords(f.nodes, nodeSortKey)
	sortRecordsByKeys(f.edges, f.edgeOrderKeys)
	sortRecordsByKeys(f.unresolved, f.unresolvedOrderKeys)
	records := make([]map[string]any, 0, 1+len(f.nodes)+len(f.edges)+len(f.unresolved))
	records = append(records, map[string]any{"adapter_version": adapterVersion, "language": language, "record": "lexicon", "repository": repositoryName, "schema_version": 1})
	records = append(records, f.nodes...)
	records = append(records, f.edges...)
	records = append(records, f.unresolved...)
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

func spanSortKey(record map[string]any) string {
	span, _ := record["span"].(map[string]any)
	if span == nil {
		return ""
	}
	return fmt.Sprintf("%s\x00%08d\x00%08d\x00%08d\x00%08d", spanString(span, "path"), spanInt(span, "start_line"), spanInt(span, "start_column"), spanInt(span, "end_line"), spanInt(span, "end_column"))
}
