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
  int unknown_receiver(A *other) { return other->foo(); }
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
	assertCallRemainsAmbiguous(t, records, "A::unknown_receiver", "other->foo", []string{"A::foo", "B::foo"})
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

func assertCallRemainsAmbiguous(t *testing.T, records []map[string]any, sourceQualified, expression string, targetQualified []string) {
	t.Helper()
	nodes := nodeRecords(records)
	found := map[string]bool{}
	for _, record := range records {
		if record["record"] == "edge" && record["relation"] == "possible-calls" {
			source := nodes[record["source"].(string)]
			target := nodes[record["target"].(string)]
			if source != nil && target != nil && source["qualified_name"] == sourceQualified {
				found[target["qualified_name"].(string)] = true
			}
		}
	}
	for _, target := range targetQualified {
		if !found[target] {
			t.Fatalf("missing possible target %s for %s", target, sourceQualified)
		}
	}
	for _, record := range records {
		if record["record"] == "unresolved" && record["relation"] == "calls" && record["expression"] == expression && record["reason"] == "ambiguous-target" {
			return
		}
	}
	t.Fatalf("missing ambiguous unresolved call %s: %s", sourceQualified, expression)
}
