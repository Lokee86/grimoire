package main

import "testing"

func TestNamespaceQualifiedCallKeepsMacroNamespaceFreeFunction(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"main.cpp": `#define API_BEGIN namespace api {
#define API_END }
API_BEGIN
int run(int value) { return value + 1; }
API_END
class Worker {
public:
  int run(int value) { return value + 2; }
};
int caller() { return api::run(1); }
`,
	})
	data, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	assertResolvedCallTarget(t, decodeRecords(t, data), "caller", "run")
}

func TestTypeQualifiedStaticCallKeepsTypeOwnedMethod(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"main.cpp": `class A {
public:
  static int run(int value) { return value + 1; }
};
class B {
public:
  static int run(int value) { return value + 2; }
};
int run(int value) { return value + 3; }
int caller() { return A::run(1); }
`,
	})
	data, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	assertResolvedCallTarget(t, decodeRecords(t, data), "caller", "A::run")
}

func TestTemplateTypeQualifiedCallKeepsFinalMemberName(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"main.cpp": `template <typename T>
class Holder {
public:
  static int get() { return 1; }
};
class Other {
public:
  static int get() { return 2; }
};
int caller() { return Holder<int>::get(); }
`,
	})
	data, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	assertResolvedCallTarget(t, decodeRecords(t, data), "caller", "Holder::get")
}

func assertResolvedCallTarget(t *testing.T, records []map[string]any, sourceName, targetQualified string) {
	t.Helper()
	nodes := nodeRecords(records)
	calls, possible := 0, 0
	for _, record := range records {
		if record["record"] != "edge" || record["relation"] != "calls" && record["relation"] != "possible-calls" {
			continue
		}
		source := nodes[record["source"].(string)]
		target := nodes[record["target"].(string)]
		if source == nil || target == nil || source["name"] != sourceName {
			continue
		}
		switch record["relation"] {
		case "calls":
			if target["qualified_name"] != targetQualified {
				t.Fatalf("resolved %s target = %v, want %s", sourceName, target["qualified_name"], targetQualified)
			}
			calls++
		case "possible-calls":
			possible++
		}
	}
	if calls != 1 || possible != 0 {
		t.Fatalf("call resolution for %s = %d definite, %d possible", sourceName, calls, possible)
	}
}
