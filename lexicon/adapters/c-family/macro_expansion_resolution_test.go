package main

import "testing"

func TestMacroBodyCallCarriesProvenanceAndArgumentFlow(t *testing.T) {
	records := analyzeMacroExpansionFixture(t, `int sink(int target) { return target; }
#define FORWARD(value) sink(value)
int run(int input) { return FORWARD(input); }
`)
	edge := findNamedEdge(t, records, "calls", "run", "sink")
	attributes := edge["attributes"].(map[string]any)
	assertStringEvidence(t, attributes, "macro-body")
	assertStringEvidence(t, attributes, "argument-substitution")
	if attributes["indirect"] != "macro" || attributes["expansion_depth"] != float64(0) {
		t.Fatalf("macro attributes = %#v", attributes)
	}
	substitutions := attributes["substitutions"].(map[string]any)
	if substitutions["value"] != "input" {
		t.Fatalf("substitutions = %#v", substitutions)
	}
	findQualifiedEdge(t, records, "passes-to", "run::input", "sink::target")
}

func TestNestedMacroExpansionRecordsCompleteChain(t *testing.T) {
	records := analyzeMacroExpansionFixture(t, `int sink(int target) { return target; }
#define INNER(value) sink(value)
#define OUTER(value) INNER(value)
int run(int input) { return OUTER(input); }
`)
	edge := findNamedEdge(t, records, "calls", "run", "sink")
	attributes := edge["attributes"].(map[string]any)
	if attributes["expansion_depth"] != float64(1) {
		t.Fatalf("expansion depth = %#v", attributes["expansion_depth"])
	}
	via := attributes["via"].([]any)
	if len(via) != 2 {
		t.Fatalf("expansion chain = %#v", via)
	}
	findQualifiedEdge(t, records, "passes-to", "run::input", "sink::target")
}

func TestMacroExpansionEmitsEveryBodyCall(t *testing.T) {
	records := analyzeMacroExpansionFixture(t, `int first(int target) { return target; }
int second(int target) { return target; }
#define BOTH(value) first(value); second(value)
int run(int input) { BOTH(input); return input; }
`)
	findNamedEdge(t, records, "calls", "run", "first")
	findNamedEdge(t, records, "calls", "run", "second")
}

func TestUnsupportedMacroExpansionRemainsExplicitlyUnresolved(t *testing.T) {
	records := analyzeMacroExpansionFixture(t, `#define PASTE(name) invoke_##name()
int run(void) { return PASTE(task); }
`)
	findUnresolvedReason(t, records, "unsupported-macro-expansion")
}

func TestRecursiveMacroExpansionStopsAtCycle(t *testing.T) {
	records := analyzeMacroExpansionFixture(t, `#define FIRST(value) SECOND(value)
#define SECOND(value) FIRST(value)
int run(int input) { return FIRST(input); }
`)
	findUnresolvedReason(t, records, "macro-expansion-cycle")
}

func analyzeMacroExpansionFixture(t *testing.T, content string) []map[string]any {
	t.Helper()
	root := t.TempDir()
	writeFixture(t, root, map[string]string{"main.c": content})
	data, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	return decodeRecords(t, data)
}

func findNamedEdge(t *testing.T, records []map[string]any, relation, sourceName, targetName string) map[string]any {
	t.Helper()
	nodes := nodeRecords(records)
	for _, record := range records {
		if record["record"] != "edge" || record["relation"] != relation {
			continue
		}
		source := nodes[record["source"].(string)]
		target := nodes[record["target"].(string)]
		if source != nil && target != nil && source["name"] == sourceName && target["name"] == targetName {
			return record
		}
	}
	t.Fatalf("missing %s edge %s -> %s", relation, sourceName, targetName)
	return nil
}

func findQualifiedEdge(t *testing.T, records []map[string]any, relation, sourceName, targetName string) map[string]any {
	t.Helper()
	nodes := nodeRecords(records)
	for _, record := range records {
		if record["record"] != "edge" || record["relation"] != relation {
			continue
		}
		source := nodes[record["source"].(string)]
		target := nodes[record["target"].(string)]
		if source != nil && target != nil && source["qualified_name"] == sourceName && target["qualified_name"] == targetName {
			return record
		}
	}
	t.Fatalf("missing %s edge %s -> %s", relation, sourceName, targetName)
	return nil
}

func findUnresolvedReason(t *testing.T, records []map[string]any, reason string) map[string]any {
	t.Helper()
	for _, record := range records {
		if record["record"] == "unresolved" && record["reason"] == reason {
			return record
		}
	}
	t.Fatalf("missing unresolved reason %s", reason)
	return nil
}

func assertStringEvidence(t *testing.T, attributes map[string]any, wanted string) {
	t.Helper()
	for _, value := range attributes["evidence"].([]any) {
		if value == wanted {
			return
		}
	}
	t.Fatalf("evidence = %#v, missing %s", attributes["evidence"], wanted)
}
