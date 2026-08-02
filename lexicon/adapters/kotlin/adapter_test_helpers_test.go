package main

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func decodeFacts(t *testing.T, data []byte) []map[string]any {
	t.Helper()
	var records []map[string]any
	decoder := json.NewDecoder(bytes.NewReader(data))
	for decoder.More() {
		var record map[string]any
		if err := decoder.Decode(&record); err != nil {
			t.Fatal(err)
		}
		records = append(records, record)
	}
	if len(records) == 0 {
		t.Fatal("adapter emitted no records")
	}
	return records
}

func factRecords(records []map[string]any, recordType string) []map[string]any {
	var result []map[string]any
	for _, record := range records {
		if record["record"] == recordType {
			result = append(result, record)
		}
	}
	return result
}

func findNode(nodes []map[string]any, kind, name, path string) map[string]any {
	for _, node := range nodes {
		if (kind == "" || node["kind"] == kind) && (name == "" || node["name"] == name) && (path == "" || node["path"] == path) {
			return node
		}
	}
	return nil
}

func findNodes(nodes []map[string]any, kind, name, path string) []map[string]any {
	var result []map[string]any
	for _, node := range nodes {
		if (kind == "" || node["kind"] == kind) && (name == "" || node["name"] == name) && (path == "" || node["path"] == path) {
			result = append(result, node)
		}
	}
	return result
}

func requireNode(t *testing.T, nodes []map[string]any, kind, name, path string) map[string]any {
	t.Helper()
	node := findNode(nodes, kind, name, path)
	if node == nil {
		t.Fatalf("missing %s node %q at %s", kind, name, path)
	}
	return node
}

func requireParameter(t *testing.T, nodes []map[string]any, name, typeName string, hasDefault bool) map[string]any {
	t.Helper()
	for _, node := range findNodes(nodes, "parameter", name, "") {
		attrs := attributes(node)
		if attrs["type"] == typeName && attrs["has_default"] == hasDefault {
			return node
		}
	}
	t.Fatalf("missing parameter %s: %s default=%v", name, typeName, hasDefault)
	return nil
}

func attributes(record map[string]any) map[string]any {
	value, _ := record["attributes"].(map[string]any)
	return value
}

func assertAttribute(t *testing.T, record map[string]any, key string, expected any) {
	t.Helper()
	if actual := attributes(record)[key]; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("attribute %s = %#v, want %#v on %#v", key, actual, expected, record)
	}
}

func anyAttribute(records []map[string]any, key string, expected any) bool {
	for _, record := range records {
		if reflect.DeepEqual(attributes(record)[key], expected) {
			return true
		}
	}
	return false
}

func containsJSONStrings(value any, expected string) bool {
	items, _ := value.([]any)
	for _, item := range items {
		if item == expected {
			return true
		}
	}
	return false
}

func hasRelation(records []map[string]any, relation string) bool {
	for _, record := range records {
		if record["relation"] == relation {
			return true
		}
	}
	return false
}

func hasUnresolved(records []map[string]any, relation, reason string) bool {
	for _, record := range records {
		if record["relation"] == relation && record["reason"] == reason {
			return true
		}
	}
	return false
}

func assertEdgeEndpoints(t *testing.T, nodes, edges []map[string]any) {
	t.Helper()
	ids := make(map[string]struct{}, len(nodes))
	for _, node := range nodes {
		ids[node["id"].(string)] = struct{}{}
	}
	for _, edge := range edges {
		if _, ok := ids[edge["source"].(string)]; !ok {
			t.Fatalf("unknown edge source: %#v", edge)
		}
		if _, ok := ids[edge["target"].(string)]; !ok {
			t.Fatalf("unknown edge target: %#v", edge)
		}
	}
}

func recordSortKey(record map[string]any) string {
	switch record["record"] {
	case "node":
		return "1\x00" + nodeSortKey(record)
	case "edge":
		return "2\x00" + edgeSortKey(record)
	case "unresolved":
		return "3\x00" + unresolvedSortKey(record)
	default:
		return "0"
	}
}
