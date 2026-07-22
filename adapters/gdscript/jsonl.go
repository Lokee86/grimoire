package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

func (f *factSet) render(repositoryName string) []byte {
	sort.Slice(f.nodes, func(i, j int) bool { return nodeSortKey(f.nodes[i]) < nodeSortKey(f.nodes[j]) })
	sort.Slice(f.edges, func(i, j int) bool { return edgeSortKey(f.edges[i]) < edgeSortKey(f.edges[j]) })
	sort.Slice(f.unresolved, func(i, j int) bool { return unresolvedSortKey(f.unresolved[i]) < unresolvedSortKey(f.unresolved[j]) })
	records := make([]map[string]any, 0, 1+len(f.nodes)+len(f.edges)+len(f.unresolved))
	records = append(records, map[string]any{"adapter_version": adapterVersion, "language": language, "record": "lexicon", "repository": repositoryName, "schema_version": 1})
	records = append(records, f.nodes...)
	records = append(records, f.edges...)
	records = append(records, f.unresolved...)
	var output strings.Builder
	for _, record := range records {
		data, _ := json.Marshal(record)
		output.Write(data)
		output.WriteByte('\n')
	}
	return []byte(output.String())
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
