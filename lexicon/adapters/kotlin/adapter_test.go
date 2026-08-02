package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestFoundationFixtureEmitsModeledKotlinSurface(t *testing.T) {
	data, err := analyzeRepository(filepath.FromSlash("testdata/foundation"))
	if err != nil {
		t.Fatal(err)
	}
	records := decodeFacts(t, data)
	if records[0]["record"] != "lexicon" || records[0]["language"] != "kotlin" || records[0]["schema_version"] != float64(1) || records[0]["mode"] != "full" {
		t.Fatalf("unexpected header: %#v", records[0])
	}

	nodes := factRecords(records, "node")
	edges := factRecords(records, "edge")
	unresolved := factRecords(records, "unresolved")
	assertEdgeEndpoints(t, nodes, edges)
	for _, expected := range []string{"repository", "directory", "file", "module", "namespace", "symbol", "type", "interface", "function", "method", "constructor", "field", "parameter", "import"} {
		if findNode(nodes, expected, "", "") == nil {
			t.Errorf("missing node kind %q", expected)
		}
	}

	domainPath := "src/main/kotlin/demo/model/Domain.kt"
	user := requireNode(t, nodes, "type", "User", domainPath)
	assertAttribute(t, user, "declaration_kind", "data_class")
	if user["id"] != stableID("type", "source:"+domainPath+"::type:User") {
		t.Fatalf("User has unexpected canonical identity: %v", user["id"])
	}
	assertAttribute(t, requireNode(t, nodes, "interface", "Result", domainPath), "declaration_kind", "sealed_interface")
	assertAttribute(t, requireNode(t, nodes, "type", "Outcome", domainPath), "declaration_kind", "sealed_class")
	assertAttribute(t, requireNode(t, nodes, "type", "Mode", domainPath), "declaration_kind", "enum_class")
	assertAttribute(t, requireNode(t, nodes, "type", "Registry", domainPath), "declaration_kind", "object")
	assertAttribute(t, requireNode(t, nodes, "type", "Factory", domainPath), "declaration_kind", "companion_object")
	assertAttribute(t, requireNode(t, nodes, "type", "UserId", domainPath), "declaration_kind", "value_class")

	decode := requireNode(t, nodes, "method", "decode", domainPath)
	assertAttribute(t, decode, "suspend", true)
	assertAttribute(t, decode, "extension_receiver", "String?")
	assertAttribute(t, decode, "extension_receiver_nullable", true)
	assertAttribute(t, decode, "return_type", "User?")
	assertAttribute(t, decode, "return_nullable", true)
	load := requireNode(t, nodes, "method", "load", domainPath)
	assertAttribute(t, load, "suspend", true)
	assertAttribute(t, load, "return_nullable", true)
	requireNode(t, nodes, "function", "topLevel", domainPath)

	constructors := findNodes(nodes, "constructor", "User", domainPath)
	if len(constructors) != 2 {
		t.Fatalf("User constructors = %d, want 2: %#v", len(constructors), constructors)
	}
	if !anyAttribute(constructors, "primary", true) || !anyAttribute(constructors, "primary", false) {
		t.Fatalf("missing primary or secondary constructor metadata: %#v", constructors)
	}
	emptyConstructor := requireNode(t, nodes, "constructor", "Empty", domainPath)
	if !containsJSONStrings(attributes(emptyConstructor)["modifiers"], "implicit") {
		t.Fatalf("implicit primary constructor metadata missing: %#v", emptyConstructor)
	}

	idProperty := requireNode(t, nodes, "field", "id", domainPath)
	assertAttribute(t, idProperty, "constructor_parameter", true)
	assertAttribute(t, idProperty, "mutable", false)
	nickname := requireNode(t, nodes, "field", "nickname", domainPath)
	assertAttribute(t, nickname, "constructor_parameter", true)
	assertAttribute(t, nickname, "nullable", true)
	assertAttribute(t, nickname, "mutable", true)
	if findNode(nodes, "field", "age", domainPath) != nil {
		t.Fatal("non-property constructor parameter age was emitted as a field")
	}

	limit := requireParameter(t, nodes, "limit", "Int?", true)
	assertAttribute(t, limit, "nullable", true)
	requireParameter(t, nodes, "value", "String?", false)

	imports := findNodes(nodes, "import", "", domainPath)
	if len(imports) != 2 {
		t.Fatalf("imports = %d, want 2: %#v", len(imports), imports)
	}
	if !anyAttribute(imports, "alias", "SupportHelper") || !anyAttribute(imports, "wildcard", true) {
		t.Fatalf("alias or wildcard import evidence missing: %#v", imports)
	}
	if !hasRelation(edges, "imports") || !hasRelation(edges, "contains") || !hasRelation(edges, "defines") {
		t.Fatalf("required relations missing")
	}

	for _, excluded := range []string{"BuildOutput", "GradleCache", "VendorSource", "LexiconState"} {
		if findNode(nodes, "", excluded, "") != nil {
			t.Fatalf("excluded declaration %q was analyzed", excluded)
		}
	}
	if len(unresolved) == 0 || !hasUnresolved(unresolved, "defines", "unsupported-form") {
		t.Fatalf("malformed fixture did not produce explicit unresolved syntax evidence: %#v", unresolved)
	}
	if findNode(nodes, "function", "broken", "") != nil {
		t.Fatal("unsafe malformed function syntax was guessed as a declaration")
	}
}

func TestAnalysisIsDeterministicCanonicalAndRepositoryRelative(t *testing.T) {
	first, err := analyzeRepository(filepath.FromSlash("testdata/foundation"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := analyzeRepository(filepath.FromSlash("testdata/foundation"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("repeat analysis changed byte output")
	}
	records := decodeFacts(t, first)
	phase := 0
	var previous string
	for lineIndex, line := range strings.Split(strings.TrimSpace(string(first)), "\n") {
		var ordered map[string]any
		if err := json.Unmarshal([]byte(line), &ordered); err != nil {
			t.Fatalf("line %d is not JSON: %v", lineIndex+1, err)
		}
		keys := make([]string, 0, len(ordered))
		for key := range ordered {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		positions := make([]int, len(keys))
		for index, key := range keys {
			positions[index] = strings.Index(line, `"`+key+`"`)
		}
		if !sort.IntsAreSorted(positions) {
			t.Fatalf("line %d keys are not lexicographic: %s", lineIndex+1, line)
		}
	}
	for _, record := range records[1:] {
		recordType := record["record"].(string)
		currentPhase := map[string]int{"node": 1, "edge": 2, "unresolved": 3}[recordType]
		if currentPhase < phase {
			t.Fatalf("record phase regressed at %#v", record)
		}
		phase = currentPhase
		if recordType == "node" {
			id := record["id"].(string)
			if !strings.HasPrefix(id, "sha256:") || len(id) != 71 {
				t.Fatalf("invalid stable ID: %q", id)
			}
			path := record["path"].(string)
			if strings.Contains(path, "\\") || filepath.IsAbs(path) {
				t.Fatalf("non-normalized node path: %q", path)
			}
		}
		key := recordSortKey(record)
		if currentPhase == phase && previous != "" && key < previous && recordType != "unresolved" {
			// The phase transition is handled above; within each phase contract keys must be sorted.
			t.Fatalf("records are not canonically sorted: %q before %q", previous, key)
		}
		previous = key
	}
}

func TestLexerAndParserHandleKotlinLexicalForms(t *testing.T) {
	content := []byte("package sample\n/* outer /* nested */ done */\nclass `Odd Name` {\n val text: String? = \"\"\"fun fake() { }\"\"\"\n}\n")
	parsed := parseKotlinFile("sample.kt", content)
	if len(parsed.diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", parsed.diagnostics)
	}
	if parsed.packageName != "sample" || len(parsed.declarations) != 1 || parsed.declarations[0].name != "Odd Name" {
		t.Fatalf("unexpected parse: %#v", parsed)
	}
	if len(parsed.declarations[0].children) < 2 { // implicit constructor plus property
		t.Fatalf("class body was not structurally parsed: %#v", parsed.declarations[0])
	}
}

func TestCLIWritesFileAndStdout(t *testing.T) {
	repository := filepath.FromSlash("testdata/foundation")
	var stdout bytes.Buffer
	if err := run([]string{"--repo", repository, "--output", "-"}, &stdout); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "facts.jsonl")
	if err := run([]string{"--repo", repository, "--output", output}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	written, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(stdout.Bytes(), written) {
		t.Fatal("stdout and file output differ")
	}
}
