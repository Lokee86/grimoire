package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestAdapterExtractsSharedCAndCPPSemantics(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"include/point.h": `#define POINT_SCALE 2
typedef struct Point { int x; int y; } Point;
int point_sum(const Point *point);
`,
		"src/point.c": `#include "../include/point.h"
static int helper(int value) { return value * POINT_SCALE; }
int point_sum(const Point *point) {
  int total = helper(point->x);
  return total + point->y;
}
`,
		"include/item.hpp": `namespace demo {
class Base { public: virtual int value() const = 0; };
class Item : public Base {
public:
  Item(int value) : value_(value) {}
  int value() const override;
private:
  int value_;
};
}
`,
		"src/item.cpp": `#include "../include/item.hpp"
namespace demo {
static int helper(int value) { return value + 1; }
int Item::value() const { return helper(value_); }
}
`,
	})

	data, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	records := decodeRecords(t, data)
	if header := records[0]; header["language"] != "c-family" || header["mode"] != "full" {
		t.Fatalf("header = %#v", header)
	}

	assertFileLanguage(t, records, "include/point.h", "c")
	assertFileLanguage(t, records, "src/point.c", "c")
	assertFileLanguage(t, records, "include/item.hpp", "cpp")
	assertFileLanguage(t, records, "src/item.cpp", "cpp")
	for _, name := range []string{"Point", "POINT_SCALE", "point_sum", "helper", "total", "Base", "Item", "value", "value_"} {
		if !hasNode(records, name) {
			t.Errorf("missing node %q", name)
		}
	}
	for _, relation := range []string{"includes", "extends", "calls", "reads", "writes"} {
		if !hasRelation(records, relation) {
			t.Errorf("missing relation %q", relation)
		}
	}
	if hasUnresolved(records, "extends", "Base") {
		t.Fatal("repository-local base class remained unresolved")
	}
	if hasUnresolved(records, "calls", "helper") {
		t.Fatal("repository-local helper call remained unresolved")
	}
}

func TestAdapterIsDeterministicAndSupportsIncrementalOwnership(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"include/api.h": "int helper(int value);\n",
		"src/api.c": `#include "../include/api.h"
int helper(int value) { return value + 1; }
int run(int value) { return helper(value); }
`,
	})
	first, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	second, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("identical repositories produced different facts")
	}

	incremental, err := analyzeRepository(root, []string{"src/api.c"}, []string{"src/removed.c"}, true)
	if err != nil {
		t.Fatal(err)
	}
	records := decodeRecords(t, incremental)
	header := records[0]
	if header["mode"] != "incremental" || header["shared_complete"] != false {
		t.Fatalf("incremental header = %#v", header)
	}
	if !reflect.DeepEqual(header["changed_files"], []any{"src/api.c"}) || !reflect.DeepEqual(header["removed_files"], []any{"src/removed.c"}) {
		t.Fatalf("incremental scope = %#v / %#v", header["changed_files"], header["removed_files"])
	}
	for _, record := range records[1:] {
		if owner, ok := record["owner"].(string); ok && owner != "src/api.c" {
			t.Fatalf("record escaped incremental ownership: %#v", record)
		}
	}
}

func TestHeaderLanguageInferenceUsesCPlusPlusSyntax(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"include/plain.h":    "typedef struct Value { int count; } Value;\n",
		"include/template.h": "namespace demo { template <typename T> class Box {}; }\n",
	})
	data, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	records := decodeRecords(t, data)
	assertFileLanguage(t, records, "include/plain.h", "c")
	assertFileLanguage(t, records, "include/template.h", "cpp")
}

func writeFixture(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for path, content := range files {
		absolute := filepath.Join(root, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(absolute), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(absolute, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func decodeRecords(t *testing.T, data []byte) []map[string]any {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(data))
	var records []map[string]any
	for decoder.More() {
		var record map[string]any
		if err := decoder.Decode(&record); err != nil {
			t.Fatal(err)
		}
		records = append(records, record)
	}
	return records
}

func assertFileLanguage(t *testing.T, records []map[string]any, path, language string) {
	t.Helper()
	for _, record := range records {
		if record["record"] != "node" || record["kind"] != "file" || record["path"] != path {
			continue
		}
		attributes, _ := record["attributes"].(map[string]any)
		if attributes["language"] != language {
			t.Fatalf("file %s language = %#v, want %s", path, attributes["language"], language)
		}
		return
	}
	t.Fatalf("missing file node %s", path)
}

func hasNode(records []map[string]any, name string) bool {
	for _, record := range records {
		if record["record"] == "node" && record["name"] == name {
			return true
		}
	}
	return false
}

func hasRelation(records []map[string]any, relation string) bool {
	for _, record := range records {
		if record["record"] == "edge" && record["relation"] == relation {
			return true
		}
	}
	return false
}

func hasUnresolved(records []map[string]any, relation, expression string) bool {
	for _, record := range records {
		if record["record"] == "unresolved" && record["relation"] == relation && record["expression"] == expression {
			return true
		}
	}
	return false
}
