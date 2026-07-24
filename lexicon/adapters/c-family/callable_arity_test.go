package main

import "testing"

func TestCallableArityPrunesFixedOverload(t *testing.T) {
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
	assertSingleResolvedCall(t, decodeRecords(t, data), "run", "select", 1)
}

func TestCallableArityKeepsVariadicOverload(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"main.cpp": `int route(int first, ...) { return first; }
int route(int first, int second) { return first + second; }
int run() { return route(1, 2, 3); }
`,
	})
	data, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	assertSingleResolvedCall(t, decodeRecords(t, data), "run", "route", 1)
}

func TestCallableArityRecordsKnownDefaultRange(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"main.cpp": `int choose(int value, int fallback = 0) { return value + fallback; }
int choose(int value, int second) { return value + second; }
int run() { return choose(1); }
`,
	})
	model, err := buildRepositoryModel(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, declaration := range model.Declarations {
		if declaration.Name != "choose" || declaration.CallableShape == nil {
			continue
		}
		if declaration.CallableShape.Minimum == 1 && declaration.CallableShape.Maximum == 2 {
			return
		}
	}
	t.Fatal("missing known default-argument range")
}

func TestCallableArityUncertaintyPreservesAmbiguity(t *testing.T) {
	known := &declaration{CallableShape: &callableParameterShape{Minimum: 2, Maximum: 2}}
	unknown := &declaration{}
	candidates := []*declaration{known, unknown}
	pruned := pruneCallableCandidates(candidates, 1)
	if len(pruned) != len(candidates) || pruned[0] != known || pruned[1] != unknown {
		t.Fatalf("uncertain candidates were pruned: %#v", pruned)
	}
}

func assertSingleResolvedCall(t *testing.T, records []map[string]any, sourceName, targetName string, parameterCount int) {
	t.Helper()
	nodes := nodeRecords(records)
	calls, possible := 0, 0
	for _, record := range records {
		if record["record"] != "edge" {
			continue
		}
		source := nodes[record["source"].(string)]
		target := nodes[record["target"].(string)]
		if source == nil || target == nil || source["name"] != sourceName || target["name"] != targetName {
			continue
		}
		switch record["relation"] {
		case "calls":
			calls++
			attributes, _ := target["attributes"].(map[string]any)
			if attributes["parameter_count"] != float64(parameterCount) {
				t.Fatalf("resolved target parameter count = %#v", attributes["parameter_count"])
			}
		case "possible-calls":
			possible++
		}
	}
	if calls != 1 || possible != 0 {
		t.Fatalf("call resolution = %d definite, %d possible", calls, possible)
	}
	if hasUnresolved(records, "calls", targetName) {
		t.Fatalf("resolved call %q remained ambiguous", targetName)
	}
}
