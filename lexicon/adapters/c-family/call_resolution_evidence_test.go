package main

import "testing"

func TestCallEdgesCarryDirectScopedEvidence(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"main.c": "int helper(int value) { return value; }\nint run(void) { return helper(1); }\n",
	})
	data, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	assertCallEvidence(t, decodeRecords(t, data), "calls", "run", "helper", "definite", 1, "direct-scoped-name")
}

func TestCallEdgesCarryExplicitAndEnclosingTypeEvidence(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"main.cpp": `namespace api {
int target(int value) { return value; }
}
class Worker {
  int own() { return local(); }
  int local() { return 1; }
};
int caller() { return api::target(1); }
`,
	})
	data, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	records := decodeRecords(t, data)
	assertCallEvidence(t, records, "calls", "caller", "api::target", "definite", 1, "explicit-qualification")
	assertCallEvidence(t, records, "calls", "Worker::own", "Worker::local", "definite", 1, "enclosing-type-ownership")
}

func TestCallEdgesCarryDirectReceiverEvidence(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"main.cpp": `class Receiver {
public:
  int run(int value) { return value; }
};
class Caller {
public:
  int call(Receiver *value) { return value->run(1); }
};
`,
	})
	data, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	assertCallEvidence(t, decodeRecords(t, data), "calls", "Caller::call", "Receiver::run", "definite", 1, "direct-receiver-type")
}

func TestCallEdgesCarryArityPruningEvidence(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"main.cpp": `int select(int value) { return value; }
int select(int first, int second) { return first + second; }
int run() { return select(1); }
`,
	})
	data, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	assertCallEvidence(t, decodeRecords(t, data), "calls", "run", "select", "definite", 1, "arity-pruning")
}

func TestCallEdgesCarryMacroEvidence(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"api.h":  "int target(int value);\n#define ALIAS target\n",
		"main.c": "#include \"api.h\"\nint run(void) { return ALIAS(1); }\n",
	})
	data, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	assertCallEvidence(t, decodeRecords(t, data), "calls", "run", "target", "definite", 1, "macro-mediation")
}

func TestCallEdgesCarryFunctionPointerEvidence(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"callback.c": "int increment(int value) { return value + 1; }\nint run(void) { int (*callback)(int) = increment; return callback(1); }\n",
	})
	data, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	assertCallEvidence(t, decodeRecords(t, data), "possible-calls", "run", "increment", "possible", 1, "function-pointer")
}

func assertCallEvidence(t *testing.T, records []map[string]any, relation, sourceQualified, targetQualified, resolution string, candidateCount int, wantedEvidence string) {
	t.Helper()
	nodes := nodeRecords(records)
	for _, record := range records {
		if record["record"] != "edge" || record["relation"] != relation {
			continue
		}
		source := nodes[record["source"].(string)]
		target := nodes[record["target"].(string)]
		if source == nil || target == nil || source["qualified_name"] != sourceQualified || target["qualified_name"] != targetQualified {
			continue
		}
		attributes, _ := record["attributes"].(map[string]any)
		if attributes["resolution"] != resolution {
			t.Fatalf("%s resolution = %#v, want %s", sourceQualified, attributes["resolution"], resolution)
		}
		if attributes["candidate_count"] != float64(candidateCount) {
			t.Fatalf("%s candidate_count = %#v, want %d", sourceQualified, attributes["candidate_count"], candidateCount)
		}
		for _, evidence := range attributes["evidence"].([]any) {
			if evidence == wantedEvidence {
				return
			}
		}
		t.Fatalf("%s evidence = %#v, missing %s", sourceQualified, attributes["evidence"], wantedEvidence)
	}
	t.Fatalf("missing %s evidence edge %s -> %s", relation, sourceQualified, targetQualified)
}
