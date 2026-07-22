package objectstore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type exportRecord struct {
	raw []byte
	key string
}

func exportRecords(records []json.RawMessage) ([]exportRecord, error) {
	result := make([]exportRecord, 0, len(records))
	for _, raw := range records {
		compact := bytes.TrimSpace(raw)
		var value map[string]any
		if err := json.Unmarshal(compact, &value); err != nil {
			return nil, err
		}
		if value == nil {
			return nil, fmt.Errorf("record is not an object")
		}
		record, _ := value["record"].(string)
		if record != "node" && record != "edge" && record != "unresolved" {
			return nil, fmt.Errorf("unsupported Lexicon record %q", record)
		}
		result = append(result, exportRecord{
			raw: append([]byte(nil), compact...),
			key: exportSortKey(value),
		})
	}
	return result, nil
}

func exportSortKey(record map[string]any) string {
	switch record["record"] {
	case "node":
		return "0\x00" + exportFields(record, "id", "kind", "path", "qualified_name")
	case "edge":
		return "1\x00" + exportFields(record, "source", "target", "relation") + "\x00" + exportSpanKey(record)
	default:
		return "2\x00" + exportFields(record, "source", "relation", "expression", "reason") + "\x00" + exportSpanKey(record)
	}
}

func exportFields(record map[string]any, names ...string) string {
	values := make([]string, len(names))
	for index, name := range names {
		values[index], _ = record[name].(string)
	}
	return strings.Join(values, "\x00")
}

func exportSpanKey(record map[string]any) string {
	span, _ := record["span"]
	if span == nil {
		return ""
	}
	data, err := json.Marshal(span)
	if err != nil {
		return ""
	}
	return string(data)
}

func encodeExport(entry LanguageEntry, records []exportRecord) ([]byte, error) {
	header := map[string]any{
		"adapter_version": entry.AdapterVersion,
		"language":        entry.Language,
		"mode":            "full",
		"record":          "lexicon",
		"repository":      entry.Repository,
		"schema_version":  entry.SchemaVersion,
	}
	headerData, err := json.Marshal(header)
	if err != nil {
		return nil, err
	}
	var output bytes.Buffer
	output.Write(headerData)
	output.WriteByte('\n')
	for _, record := range records {
		output.Write(record.raw)
		output.WriteByte('\n')
	}
	return output.Bytes(), nil
}
