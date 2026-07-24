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

func TestIncludedCFilesShareTranslationUnit(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"bundle.c": "#include \"helper.c\"\n#include \"user.c\"\n",
		"helper.c": "static int helper(int value) { return value + 1; }\n",
		"user.c":   "int run(int value) { return helper(value); }\n",
	})
	data, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	assertRelationTarget(t, decodeRecords(t, data), "calls", "user.c", "run", "helper.c", "helper")
}

func TestIncludedHeaderCanCallSourceStaticFunction(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"main.c":     "static int helper(void) { return 42; }\n#include \"fragment.h\"\nint run(void) { return from_header(); }\n",
		"fragment.h": "static inline int from_header(void) { return helper(); }\n",
	})
	data, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	assertRelationTarget(t, decodeRecords(t, data), "calls", "fragment.h", "from_header", "main.c", "helper")
}

func TestMacroAliasesAndWrappersResolveCalls(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"api.h":    "#ifndef API_H\n#define API_H\nint target(int value);\n#define WRAP(value) target(value)\n#define ALIAS target\n#endif\n",
		"main.c":   "#include \"api.h\"\nint run(void) { return WRAP(1) + ALIAS(2); }\n",
		"target.c": "int target(int value) { return value; }\n",
	})
	data, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	records := decodeRecords(t, data)
	assertRelationTarget(t, records, "references", "main.c", "run", "api.h", "WRAP")
	assertRelationTarget(t, records, "references", "main.c", "run", "api.h", "ALIAS")
	assertRelationTarget(t, records, "calls", "main.c", "run", "target.c", "target")
}

func TestControlFlowMacroDoesNotFabricateCallTarget(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"macro.h": "#define CHECK(value) do { if (!(value)) fail(); } while (0)\n",
		"main.c":  "#include \"macro.h\"\nvoid run(void) { CHECK(1); }\n",
	})
	data, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	records := decodeRecords(t, data)
	assertRelationTarget(t, records, "references", "main.c", "run", "macro.h", "CHECK")
	for _, record := range records {
		if record["record"] != "edge" || record["relation"] != "calls" && record["relation"] != "possible-calls" {
			continue
		}
		nodes := nodeRecords(records)
		target := nodes[record["target"].(string)]
		if target != nil && target["name"] == "if" {
			t.Fatal("control-flow keyword was emitted as a call target")
		}
	}
}

func TestConditionalMacroTargetRemainsPossible(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"api.h":    "int platform_target(void);\n#ifdef PLATFORM\n#define ALIAS platform_target\n#endif\n",
		"main.c":   "#include \"api.h\"\nint run(void) { return ALIAS(); }\n",
		"target.c": "int platform_target(void) { return 1; }\n",
	})
	data, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	assertRelationTarget(t, decodeRecords(t, data), "possible-calls", "main.c", "run", "target.c", "platform_target")
}

func assertRelationTarget(t *testing.T, records []map[string]any, relation, sourcePath, sourceName, targetPath, targetName string) {
	t.Helper()
	nodes := nodeRecords(records)
	for _, record := range records {
		if record["record"] != "edge" || record["relation"] != relation {
			continue
		}
		source := nodes[record["source"].(string)]
		target := nodes[record["target"].(string)]
		if source != nil && target != nil && source["path"] == sourcePath && source["name"] == sourceName && target["path"] == targetPath && target["name"] == targetName {
			return
		}
	}
	t.Fatalf("missing %s %s:%s -> %s:%s", relation, sourcePath, sourceName, targetPath, targetName)
}

func nodeRecords(records []map[string]any) map[string]map[string]any {
	nodes := map[string]map[string]any{}
	for _, record := range records {
		if record["record"] == "node" {
			nodes[record["id"].(string)] = record
		}
	}
	return nodes
}
