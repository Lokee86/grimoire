package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestGenericAdapterEmitsConservativeFacts(t *testing.T) {
	repository := t.TempDir()
	writeFixture(t, repository, "src/main.c", `#include <stdio.h>
struct Player {
};
int launch() {
  return 0;
}
`)
	writeFixture(t, repository, "vendor/ignored.c", "int ignored() { return 0; }\n")
	writeFixture(t, repository, "src/generated.c", "// Code generated; DO NOT EDIT.\nint generated() { return 0; }\n")
	writeFixture(t, repository, "README.md", "class NotSource\n")

	output, err := analyzeRepository(repository, "generic-c", nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	records := decodeRecords(t, output)
	if records[0]["language"] != "generic-c" || records[0]["record"] != "lexicon" {
		t.Fatalf("header = %#v", records[0])
	}
	assertNode(t, records, "file", "main.c")
	assertNode(t, records, "module", "main")
	assertNode(t, records, "type", "Player")
	assertNode(t, records, "function", "launch")
	assertNode(t, records, "import", "stdio.h")
	if hasNode(records, "function", "ignored") || hasNode(records, "function", "generated") || hasNode(records, "type", "NotSource") {
		t.Fatal("excluded source produced facts")
	}
	if !hasUnresolved(records, "imports", "stdio.h", "external-target") {
		t.Fatal("static include did not remain explicit unresolved import evidence")
	}
}

func TestGenericAdapterIsDeterministic(t *testing.T) {
	repository := t.TempDir()
	writeFixture(t, repository, "src/example.java", "import java.util.List;\nclass Example {\n  int run() {\n  }\n}\n")
	first, err := analyzeRepository(repository, "generic-java", nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	second, err := analyzeRepository(repository, "generic-java", nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("identical input produced different JSONL")
	}
}

func TestGenericAdapterIncrementalOwnership(t *testing.T) {
	repository := t.TempDir()
	writeFixture(t, repository, "a.lua", "function alpha()\nend\n")
	writeFixture(t, repository, "b.lua", "function beta()\nend\n")
	output, err := analyzeRepository(repository, "generic-lua", []string{"a.lua"}, []string{"removed.lua"}, true)
	if err != nil {
		t.Fatal(err)
	}
	records := decodeRecords(t, output)
	header := records[0]
	if header["mode"] != "incremental" || !reflect.DeepEqual(header["changed_files"], []any{"a.lua"}) || !reflect.DeepEqual(header["removed_files"], []any{"removed.lua"}) {
		t.Fatalf("incremental header = %#v", header)
	}
	for _, record := range records[1:] {
		if owner, ok := record["owner"].(string); !ok || owner != "a.lua" {
			t.Fatalf("incremental record owner = %#v", record)
		}
	}
	if !hasNode(records, "function", "alpha") || hasNode(records, "function", "beta") {
		t.Fatal("incremental selection emitted the wrong source facts")
	}
}

func TestGenericAdapterRejectsInvalidLanguage(t *testing.T) {
	if _, err := analyzeRepository(t.TempDir(), "generic-md", nil, nil, false); err == nil {
		t.Fatal("expected unsupported generic language to fail")
	}
}

func writeFixture(t *testing.T, root, relative, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func decodeRecords(t *testing.T, data []byte) []map[string]any {
	t.Helper()
	var records []map[string]any
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		var record map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			t.Fatal(err)
		}
		records = append(records, record)
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	return records
}

func assertNode(t *testing.T, records []map[string]any, kind, name string) {
	t.Helper()
	if !hasNode(records, kind, name) {
		t.Fatalf("missing %s node %q", kind, name)
	}
}

func hasNode(records []map[string]any, kind, name string) bool {
	for _, record := range records {
		if record["record"] == "node" && record["kind"] == kind && record["name"] == name {
			return true
		}
	}
	return false
}

func hasUnresolved(records []map[string]any, relation, expression, reason string) bool {
	for _, record := range records {
		if record["record"] == "unresolved" && record["relation"] == relation && record["expression"] == expression && record["reason"] == reason {
			return true
		}
	}
	return false
}
