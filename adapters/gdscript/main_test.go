package main

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestAnalyzeExtractsGDScriptSlice(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "scripts/base.gd", `class_name Base
extends Node
signal changed(value)
const LIMIT = 3
var title: String = "# is not a comment"
func greet(name):
    return name
`)
	writeFixture(t, root, "scripts/player.gd", `# func ignored()
class_name Player extends Base
@onready var scene = preload("res://scripts/base.gd")
signal spawned
func greet(text):
    return text
func run(
    value: int,
):
    var message = "load(\\\"res://fake.gd\\\") # string"
    greet(message)
    load(get_path())
`)
	writeFixture(t, root, ".worktrees/ignored.gd", "class_name Ignored\n")
	writeFixture(t, root, "vendor/ignored.gd", "class_name IgnoredVendor\n")

	data, err := analyzeRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	records := decodeRecords(t, data)
	if records[0]["language"] != language || records[0]["schema_version"] != float64(1) {
		t.Fatalf("unexpected header: %#v", records[0])
	}
	playerType := findNode(records, "type", "Player", "scripts/player.gd")
	if playerType["id"] != nodeID("type", "scripts/player.gd::type::Player") {
		t.Fatalf("unexpected stable type ID: %v", playerType["id"])
	}

	var kinds []string
	var names []string
	var relations []string
	var unresolvedReasons []string
	for _, record := range records[1:] {
		switch record["record"] {
		case "node":
			kinds = append(kinds, record["kind"].(string))
			names = append(names, record["name"].(string))
		case "edge":
			relations = append(relations, record["relation"].(string))
		case "unresolved":
			unresolvedReasons = append(unresolvedReasons, record["reason"].(string))
		}
	}
	for _, expected := range []string{"repository", "directory", "file", "module", "type", "function", "signal", "constant", "variable", "import"} {
		if !contains(kinds, expected) {
			t.Errorf("missing node kind %q in %v", expected, kinds)
		}
	}
	for _, expected := range []string{"Player", "Base", "greet", "run", "changed", "LIMIT", "title", "spawned"} {
		if !contains(names, expected) {
			t.Errorf("missing node name %q in %v", expected, names)
		}
	}
	for _, expected := range []string{"contains", "defines", "imports", "references", "extends", "calls"} {
		if !contains(relations, expected) {
			t.Errorf("missing edge relation %q in %v", expected, relations)
		}
	}
	if !contains(unresolvedReasons, "dynamic-target") {
		t.Errorf("dynamic load/call was not reported unresolved: %v", unresolvedReasons)
	}
	if contains(names, "Ignored") || contains(names, "IgnoredVendor") {
		t.Fatalf("excluded source was scanned: %v", names)
	}
}

func TestParserHandlesIndentationStringsCommentsAndMultilineDeclarations(t *testing.T) {
	pf, err := parseFile("scene.gd", []byte(`class_name Scene
var text = "func fake() # not a comment"
# signal fake()
func run(
    value: int,
    label = "signal fake()",
):
    var nested = preload(
        "res://other.gd"
    )
`))
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, decl := range pf.declarations {
		got = append(got, decl.kind+":"+decl.name)
	}
	want := []string{"type:Scene", "variable:text", "function:run", "variable:nested"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("declarations = %v, want %v", got, want)
	}
	if len(pf.imports) != 0 || len(pf.calls) != 0 {
		t.Fatalf("parseFile should defer references to the fact pass: imports=%v calls=%v", pf.imports, pf.calls)
	}
}

func TestAnalyzeIsDeterministicAcrossRepeatRuns(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "b.gd", "func z():\n    pass\n")
	writeFixture(t, root, "a.gd", "func a():\n    z()\n")
	first, err := analyzeRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	second, err := analyzeRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("repeat analysis changed the JSONL output")
	}
	for _, line := range strings.Split(strings.TrimSpace(string(first)), "\n") {
		var record map[string]any
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatal(err)
		}
		if record["record"] == "node" && !strings.HasPrefix(record["id"].(string), "sha256:") {
			t.Fatalf("unstable node ID: %v", record["id"])
		}
	}
}

func TestJSONLRecordsUseContractOrder(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "one.gd", "class_name One\n")
	data, err := analyzeRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) < 2 {
		t.Fatal("expected header and facts")
	}
	previous := ""
	for _, line := range lines[1:] {
		var record map[string]any
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatal(err)
		}
		kind := record["record"].(string)
		key := kind
		if kind == "node" {
			key = "0/" + record["id"].(string) + "/" + record["kind"].(string) + "/" + record["path"].(string) + "/" + record["qualified_name"].(string)
		} else if kind == "edge" {
			key = "1/" + record["source"].(string) + "/" + record["target"].(string) + "/" + record["relation"].(string)
		} else {
			key = "2/" + record["source"].(string) + "/" + record["relation"].(string) + "/" + record["expression"].(string) + "/" + record["reason"].(string)
		}
		if previous != "" && key < previous {
			t.Fatalf("records are not ordered: %q before %q", previous, key)
		}
		previous = key
	}
}

func writeFixture(t *testing.T, root, path, content string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func decodeRecords(t *testing.T, data []byte) []map[string]any {
	t.Helper()
	var records []map[string]any
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
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

func findNode(records []map[string]any, kind, name, path string) map[string]any {
	for _, record := range records {
		if record["record"] == "node" && record["kind"] == kind && record["name"] == name && record["path"] == path {
			return record
		}
	}
	return nil
}

func contains(values []string, value string) bool {
	return sort.SearchStrings(appendSorted(values), value) < len(values) && appendSorted(values)[sort.SearchStrings(appendSorted(values), value)] == value
}

func appendSorted(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}
