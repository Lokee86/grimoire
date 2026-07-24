package main

import "testing"

func TestFunctionLikeMacrosEmitReferences(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"macro.h": "#define APPLY(value) ((value) + 1)\n",
		"main.c":  "#include \"macro.h\"\nint run(void) { return APPLY(41); }\n",
	})
	data, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	records := decodeRecords(t, data)
	assertRelationTarget(t, records, "references", "main.c", "run", "macro.h", "APPLY")
	if hasUnresolved(records, "calls", "APPLY") {
		t.Fatal("resolved macro expansion remained an unresolved call")
	}
}

func TestFunctionPointerCallsAreDynamic(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"callback.c": "int apply(int (*callback)(int), int value) { return callback(value); }\n",
	})
	data, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range decodeRecords(t, data) {
		if record["record"] == "unresolved" && record["relation"] == "calls" && record["expression"] == "callback" {
			if record["reason"] != "dynamic-target" {
				t.Fatalf("callback reason = %v, want dynamic-target", record["reason"])
			}
			return
		}
	}
	t.Fatal("missing dynamic function-pointer call")
}

func assertRelationTarget(t *testing.T, records []map[string]any, relation, sourcePath, sourceName, targetPath, targetName string) {
	t.Helper()
	nodes := map[string]map[string]any{}
	for _, record := range records {
		if record["record"] == "node" {
			nodes[record["id"].(string)] = record
		}
	}
	for _, record := range records {
		if record["record"] != "edge" || record["relation"] != relation {
			continue
		}
		source := nodes[record["source"].(string)]
		target := nodes[record["target"].(string)]
		if source != nil && target != nil &&
			source["path"] == sourcePath && source["name"] == sourceName &&
			target["path"] == targetPath && target["name"] == targetName {
			return
		}
	}
	t.Fatalf("missing %s %s:%s -> %s:%s", relation, sourcePath, sourceName, targetPath, targetName)
}
