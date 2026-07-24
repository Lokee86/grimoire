package main

import "testing"

func TestMemberCallsPreferDirectlyKnownEnclosingType(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"members.cpp": `class A {
public:
  int unqualified() { return foo(); }
  int explicit_this() { return this->foo(); }
  int explicit_self() { return self.foo(); }
  int foo() { return 1; }
};
class B {
public:
  int foo() { return 2; }
};
`,
	})
	data, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	records := decodeRecords(t, data)

	for _, source := range []string{"A::unqualified", "A::explicit_this", "A::explicit_self"} {
		assertQualifiedCall(t, records, "calls", source, "A::foo")
		assertNoCallRelation(t, records, "possible-calls", source)
	}
}

func assertQualifiedCall(t *testing.T, records []map[string]any, relation, sourceQualified, targetQualified string) {
	t.Helper()
	nodes := nodeRecords(records)
	for _, record := range records {
		if record["record"] != "edge" || record["relation"] != relation {
			continue
		}
		source := nodes[record["source"].(string)]
		target := nodes[record["target"].(string)]
		if source != nil && target != nil && source["qualified_name"] == sourceQualified && target["qualified_name"] == targetQualified {
			return
		}
	}
	t.Fatalf("missing %s %s -> %s", relation, sourceQualified, targetQualified)
}

func assertNoCallRelation(t *testing.T, records []map[string]any, relation, sourceQualified string) {
	t.Helper()
	nodes := nodeRecords(records)
	for _, record := range records {
		if record["record"] != "edge" || record["relation"] != relation {
			continue
		}
		if source := nodes[record["source"].(string)]; source != nil && source["qualified_name"] == sourceQualified {
			t.Fatalf("unexpected %s for %s", relation, sourceQualified)
		}
	}
}
