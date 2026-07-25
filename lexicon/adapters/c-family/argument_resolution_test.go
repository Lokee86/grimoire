package main

import "testing"

func TestDirectArgumentPassesParameterToParameter(t *testing.T) {
	records := analyzeArgumentFixture(t, "main.c", `int sink(int target, int other) { return target + other; }
int run(int value, int second) { return sink(value, second); }
`)
	assertArgumentPass(t, records, "main.c", "parameter", "run::value", "parameter", "sink::target", "value", 0, "sink")
	assertArgumentPass(t, records, "main.c", "parameter", "run::second", "parameter", "sink::other", "second", 1, "sink")
}

func TestDirectArgumentPassesLocalToParameter(t *testing.T) {
	records := analyzeArgumentFixture(t, "main.c", `int sink(int target) { return target; }
int run(int input) { int local = input; return sink(local); }
`)
	assertArgumentPass(t, records, "main.c", "variable", "run::local", "parameter", "sink::target", "local", 0, "sink")
}

func TestDirectArgumentPassesOwnedFieldToParameter(t *testing.T) {
	records := analyzeArgumentFixture(t, "main.cpp", `int sink(int target) { return target; }
struct Holder { int input; int run() { return sink(input); } };
`)
	assertArgumentPass(t, records, "main.cpp", "field", "Holder::input", "parameter", "sink::target", "input", 0, "sink")
}

func TestDirectArgumentShadowingPrefersCallerParameter(t *testing.T) {
	records := analyzeArgumentFixture(t, "main.cpp", `int sink(int target) { return target; }
struct Holder { int value; int run(int value) { return sink(value); } };
`)
	assertArgumentPass(t, records, "main.cpp", "parameter", "Holder::run::value", "parameter", "sink::target", "value", 0, "sink")
	assertNoArgumentPass(t, records, "main.cpp", "Holder::value", "sink::target")
}

func TestDirectArgumentOmitsUnsupportedExpression(t *testing.T) {
	records := analyzeArgumentFixture(t, "main.c", `int sink(int target) { return target; }
int run(int value) { return sink(value + 1); }
`)
	assertNoArgumentPass(t, records, "main.c", "run::value", "sink::target")
}

func analyzeArgumentFixture(t *testing.T, path, content string) []map[string]any {
	t.Helper()
	root := t.TempDir()
	writeFixture(t, root, map[string]string{path: content})
	data, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	return decodeRecords(t, data)
}

func assertArgumentPass(t *testing.T, records []map[string]any, path, sourceKind, sourceQualified, targetKind, targetQualified, expression string, argumentIndex int, callable string) {
	t.Helper()
	nodes := nodeRecords(records)
	var targetCallable map[string]any
	for _, node := range nodes {
		if node["path"] == path && node["name"] == callable && node["kind"] != "parameter" {
			targetCallable = node
			break
		}
	}
	if targetCallable == nil {
		t.Fatalf("missing target callable %s", callable)
	}
	for _, record := range records {
		if record["record"] != "edge" || record["relation"] != "passes-to" {
			continue
		}
		source := nodes[record["source"].(string)]
		target := nodes[record["target"].(string)]
		if source == nil || target == nil || source["path"] != path || target["path"] != path ||
			source["kind"] != sourceKind || source["qualified_name"] != sourceQualified ||
			target["kind"] != targetKind || target["qualified_name"] != targetQualified {
			continue
		}
		attributes, _ := record["attributes"].(map[string]any)
		if attributes["argument_index"] != float64(argumentIndex) || attributes["expression"] != expression || attributes["via_call"] != targetCallable["id"] {
			t.Fatalf("passes-to attributes = %#v", attributes)
		}
		return
	}
	t.Fatalf("missing passes-to %s -> %s", sourceQualified, targetQualified)
}

func assertNoArgumentPass(t *testing.T, records []map[string]any, path, sourceQualified, targetQualified string) {
	t.Helper()
	nodes := nodeRecords(records)
	for _, record := range records {
		if record["record"] != "edge" || record["relation"] != "passes-to" {
			continue
		}
		source := nodes[record["source"].(string)]
		target := nodes[record["target"].(string)]
		if source != nil && target != nil && source["path"] == path && target["path"] == path &&
			source["qualified_name"] == sourceQualified && target["qualified_name"] == targetQualified {
			t.Fatalf("unexpected passes-to edge: %#v", record)
		}
	}
}
