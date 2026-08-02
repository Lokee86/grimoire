package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func assertCanonicalJSONL(t *testing.T, data []byte) {
	t.Helper()
	idPattern := regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	for index, line := range lines {
		var ordered map[string]any
		if err := json.Unmarshal([]byte(line), &ordered); err != nil {
			t.Fatal(err)
		}
		keys := make([]string, 0, len(ordered))
		for key := range ordered {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		var compact bytes.Buffer
		encoder := json.NewEncoder(&compact)
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(ordered); err != nil {
			t.Fatal(err)
		}
		if compact.String() != line+"\n" {
			t.Fatalf("line %d is not canonically encoded", index+1)
		}
		if id, ok := ordered["id"].(string); ok && !idPattern.MatchString(id) {
			t.Fatalf("invalid id on line %d: %s", index+1, id)
		}
	}

	records := decodeRecords(t, data)[1:]
	ordered := append([]map[string]any(nil), records...)
	sort.Slice(ordered, func(left, right int) bool {
		return testRecordKey(ordered[left]) < testRecordKey(ordered[right])
	})
	if !reflect.DeepEqual(records, ordered) {
		t.Fatal("records are not in canonical facts-v1 order")
	}
}

func testRecordKey(record map[string]any) string {
	switch record["record"] {
	case "node":
		return "0\x00" + record["id"].(string) + "\x00" + record["kind"].(string) + "\x00" + record["path"].(string) + "\x00" + record["qualified_name"].(string)
	case "edge":
		return "1\x00" + record["source"].(string) + "\x00" + record["target"].(string) + "\x00" + record["relation"].(string) + "\x00" + testSpanKey(record)
	default:
		return "2\x00" + record["source"].(string) + "\x00" + record["relation"].(string) + "\x00" + record["expression"].(string) + "\x00" + record["reason"].(string) + "\x00" + testSpanKey(record)
	}
}

func testSpanKey(record map[string]any) string {
	value, _ := record["span"].(map[string]any)
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%s\x00%09.0f\x00%09.0f\x00%09.0f\x00%09.0f", value["path"], value["start_line"], value["start_column"], value["end_line"], value["end_column"])
}
