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

func TestMemberCallsResolveDirectReceiverTypes(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"receivers.cpp": `class A {
public:
  int run() { return 1; }
};
class B {
public:
  int run() { return 2; }
};
class Base {
public:
  int inherited() { return 3; }
};
class Derived : public Base {};
namespace left { class Thing {}; }
namespace right { class Thing {}; }
class Holder {
  A *member;
public:
  int parameter(A *value) { return value->run(); }
  int local() { A value; return value.run(); }
  int field_receiver() { return member->run(); }
  int unknown(Missing *value) { return value->run(); }
  int ambiguous(Thing *value) { return value->run(); }
  int inherited_receiver(Derived *value) { return value->inherited(); }
  int qualified() { return B::run(); }
};
`,
	})
	data, err := analyzeRepository(root, nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	records := decodeRecords(t, data)

	for _, source := range []string{"Holder::parameter", "Holder::local", "Holder::field_receiver"} {
		assertQualifiedCall(t, records, "calls", source, "A::run")
		assertNoCallRelation(t, records, "possible-calls", source)
	}
	assertAmbiguousCall(t, records, "Holder::unknown", []string{"A::run", "B::run"})
	assertAmbiguousCall(t, records, "Holder::ambiguous", []string{"A::run", "B::run"})
	assertQualifiedCall(t, records, "calls", "Holder::inherited_receiver", "Base::inherited")
	assertQualifiedCall(t, records, "calls", "Holder::qualified", "B::run")
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

func assertAmbiguousCall(t *testing.T, records []map[string]any, sourceQualified string, targetQualified []string) {
	t.Helper()
	nodes := nodeRecords(records)
	found := map[string]bool{}
	for _, record := range records {
		if record["record"] != "edge" || record["relation"] != "possible-calls" {
			continue
		}
		source := nodes[record["source"].(string)]
		target := nodes[record["target"].(string)]
		if source != nil && target != nil && source["qualified_name"] == sourceQualified {
			found[target["qualified_name"].(string)] = true
		}
	}
	for _, target := range targetQualified {
		if !found[target] {
			t.Fatalf("missing possible target %s for %s", target, sourceQualified)
		}
	}
	for _, record := range records {
		if record["record"] != "unresolved" || record["relation"] != "calls" || record["reason"] != "ambiguous-target" {
			continue
		}
		if source := nodes[record["source"].(string)]; source != nil && source["qualified_name"] == sourceQualified {
			return
		}
	}
	t.Fatalf("missing ambiguous call for %s", sourceQualified)
}
