package main

import "testing"

func TestStaticFunctionsResolveWithinTheirTranslationUnit(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"a.c": `static int helper(int value) { return value + 1; }
int run_a(int value) { return helper(value); }
`,
		"b.c": `static int helper(int value) { return value + 2; }
int run_b(int value) { return helper(value); }
`,
	})

	data, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	records := decodeRecords(t, data)
	assertCallTarget(t, records, "a.c", "run_a", "a.c", "helper")
	assertCallTarget(t, records, "b.c", "run_b", "b.c", "helper")
}

func TestFunctionPointerMembersRemainFields(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"callbacks.h": `struct callbacks {
  int (*run)(int value);
};
`,
	})

	data, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	records := decodeRecords(t, data)
	for _, record := range records {
		if record["record"] != "node" || record["name"] != "run" {
			continue
		}
		if record["kind"] == "method" {
			t.Fatal("C function-pointer member was emitted as a method")
		}
		if record["kind"] == "field" {
			attributes, _ := record["attributes"].(map[string]any)
			if attributes["function_pointer"] != true {
				t.Fatalf("function-pointer field attributes = %#v", attributes)
			}
			return
		}
	}
	t.Fatal("missing function-pointer field")
}

func TestAmbiguousHeadersFollowCIncluders(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"compat.h": `/* namespace conflict is an error message, not C++ syntax. */
static inline int answer(void) { return 42; }
`,
		"main.c": "#include \"compat.h\"\nint run(void) { return answer(); }\n",
	})

	data, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	records := decodeRecords(t, data)
	assertFileLanguage(t, records, "compat.h", "c")
	assertCallTarget(t, records, "main.c", "run", "compat.h", "answer")
}

func assertCallTarget(t *testing.T, records []map[string]any, sourcePath, sourceName, targetPath, targetName string) {
	t.Helper()
	nodes := map[string]map[string]any{}
	for _, record := range records {
		if record["record"] == "node" {
			nodes[record["id"].(string)] = record
		}
	}
	for _, record := range records {
		if record["record"] != "edge" || record["relation"] != "calls" {
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
	t.Fatalf("missing call %s:%s -> %s:%s", sourcePath, sourceName, targetPath, targetName)
}
