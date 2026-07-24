package main

import "testing"

func TestUnrelatedFunctionPointerFieldDoesNotCaptureDirectCall(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"callback.h": "struct callbacks { int (*close)(int value); };\n",
		"main.c":     "#include \"callback.h\"\nint run(void) { return close(1); }\n",
	})
	data, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range decodeRecords(t, data) {
		if record["record"] == "unresolved" && record["relation"] == "calls" && record["expression"] == "close" {
			if record["reason"] != "external-target" {
				t.Fatalf("close reason = %v, want external-target", record["reason"])
			}
			return
		}
	}
	t.Fatal("missing unresolved external call")
}

func TestCallbackArgumentsEmitPossibleCalls(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"callback.c": `int increment(int value) { return value + 1; }
int apply(int (*callback)(int), int value) { return callback(value); }
int run(void) { return apply(increment, 41); }
`,
	})
	data, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	assertRelationTarget(t, decodeRecords(t, data), "possible-calls", "callback.c", "apply", "callback.c", "increment")
}

func TestInitializedFunctionPointerEmitsPossibleCall(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"callback.c": `int increment(int value) { return value + 1; }
int run(void) { int (*callback)(int) = increment; return callback(41); }
`,
	})
	data, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	assertRelationTarget(t, decodeRecords(t, data), "possible-calls", "callback.c", "run", "callback.c", "increment")
}

func TestFunctionPointerTypedefParametersEmitPossibleCalls(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"callback.c": `typedef int (*callback_fn)(int value);
int increment(int value) { return value + 1; }
int apply(callback_fn callback, int value) { return callback(value); }
int run(void) { return apply(increment, 41); }
`,
	})
	data, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	assertRelationTarget(t, decodeRecords(t, data), "possible-calls", "callback.c", "apply", "callback.c", "increment")
}

func TestDesignatedFunctionPointerInitializerEmitsPossibleCall(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"callback.c": `typedef int callback_fn(int value);
struct callbacks { callback_fn *run; };
int increment(int value) { return value + 1; }
static struct callbacks callbacks = { .run = increment };
int apply(void) { return callbacks.run(41); }
`,
	})
	data, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	assertRelationTarget(t, decodeRecords(t, data), "possible-calls", "callback.c", "apply", "callback.c", "increment")
}

func TestAssignedFunctionPointerFieldEmitsPossibleCall(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"callback.c": `struct callbacks { int (*run)(int value); };
int increment(int value) { return value + 1; }
int apply(struct callbacks *callbacks) { callbacks->run = increment; return callbacks->run(41); }
`,
	})
	data, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	assertRelationTarget(t, decodeRecords(t, data), "possible-calls", "callback.c", "apply", "callback.c", "increment")
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
