package main

import (
	"bytes"
	"path/filepath"
	"reflect"
	"testing"
)

const runtimeFixture = "testdata/runtime"

func TestRuntimeCallsResolveDefinitePossibleAndExternalTargets(t *testing.T) {
	records := decodeRecords(t, analyzeFixture(t, runtimeFixture))
	nodes := nodeIndex(records)
	caller := "demo.runtime.RuntimeSlice.exercise(int,ExternalReceiver)"

	for _, target := range []string{
		"demo.runtime.RuntimeSlice.definite(int)",
		"demo.runtime.RuntimeSlice.overloaded(int)",
		"demo.runtime.Helper.staticCall(int)",
		"demo.runtime.Helper.<init>(int)",
		"demo.runtime.Choice.<init>(int)",
		"demo.runtime.RuntimeSlice.<init>(int)",
	} {
		assertRuntimeEdge(t, records, nodes, caller, target, "calls")
	}
	for _, target := range []string{
		"demo.runtime.RuntimeSlice.overloaded(String)",
		"demo.runtime.Choice.<init>(String)",
	} {
		assertNoRuntimeEdge(t, records, nodes, caller, target, "possible-calls")
	}

	assertRuntimeUnresolved(t, records, nodes, caller, "calls", "external.run()", "external-target")
	assertRuntimeUnresolved(t, records, nodes, caller, "calls", "missing(value)", "dynamic-target")
	assertRuntimeUnresolved(t, records, nodes, caller, "calls", "vendor.External.staticCall()", "external-target")
	assertRuntimeUnresolved(t, records, nodes, caller, "calls", "new vendor.External(value)", "external-target")
}

func TestTypedIdentifierReceiverCallsStayBoundedAndScoped(t *testing.T) {
	records := decodeRecords(t, analyzeFixture(t, runtimeFixture))
	nodes := nodeIndex(records)
	target := "demo.receivers.ReceiverTarget.unique(int)"

	parameterCaller := "demo.receivers.ReceiverCalls.parameter(ReceiverTarget)"
	assertRuntimeEdge(t, records, nodes, parameterCaller, target, "calls")
	assertRuntimeEdge(t, records, nodes, parameterCaller, "demo.receivers.ReceiverBase.inherited()", "calls")
	assertNoRuntimeUnresolved(t, records, nodes, parameterCaller, "calls", "receiver.inherited()")
	assertRuntimeEdge(t, records, nodes, "demo.receivers.ReceiverCalls.local()", target, "calls")
	overloadedCaller := "demo.receivers.ReceiverCalls.overloaded(ReceiverTarget)"
	assertRuntimeEdge(t, records, nodes, overloadedCaller, "demo.receivers.ReceiverTarget.overloaded(int)", "calls")
	assertNoRuntimeEdge(t, records, nodes, overloadedCaller, "demo.receivers.ReceiverTarget.overloaded(String)", "possible-calls")

	assertRuntimeUnresolved(t, records, nodes, "demo.receivers.ReceiverCalls.external(vendor.External)", "calls", "receiver.run()", "external-target")
	assertRuntimeUnresolved(t, records, nodes, "demo.receivers.ReceiverCalls.ambiguous(SharedReceiver)", "calls", "receiver.run()", "ambiguous-target")

	scope := "demo.receivers.ReceiverCalls.scope()"
	assertRuntimeEdge(t, records, nodes, scope, target, "calls")
	assertRuntimeUnresolved(t, records, nodes, scope, "calls", "future.unique(1)", "dynamic-target")
	assertRuntimeUnresolved(t, records, nodes, scope, "calls", "inner.unique(1)", "dynamic-target")
	assertRuntimeEdge(t, records, nodes, scope, "demo.receivers.ReceiverTarget.staticOnly()", "calls")
	assertNoRuntimeUnresolved(t, records, nodes, scope, "calls", "ReceiverTarget.staticOnly()")
	assertRuntimeUnresolved(t, records, nodes, scope, "calls", "unique(1)", "dynamic-target")
	assertRuntimeUnresolved(t, records, nodes, scope, "calls", "unknown.unique(1)", "dynamic-target")
	assertAllEndpoints(t, records, nodes)
}

func TestTypedReceiverOverloadsUseConservativeArgumentEvidence(t *testing.T) {
	records := decodeRecords(t, analyzeFixture(t, runtimeFixture))
	nodes := nodeIndex(records)
	definite := map[string]string{
		"demo.receivers.OverloadCalls.booleanLiteral(JsonWriterLike)":                     "boolean",
		"demo.receivers.OverloadCalls.charLiteral(JsonWriterLike)":                        "char",
		"demo.receivers.OverloadCalls.floatLiteral(JsonWriterLike)":                       "float",
		"demo.receivers.OverloadCalls.floatingLiteral(JsonWriterLike)":                    "double",
		"demo.receivers.OverloadCalls.integralLiteral(JsonWriterLike)":                    "long",
		"demo.receivers.OverloadCalls.stringLiteral(JsonWriterLike)":                      "String",
		"demo.receivers.OverloadCalls.typedBoolean(JsonWriterLike,Boolean)":               "Boolean",
		"demo.receivers.OverloadCalls.typedLocal(JsonWriterLike)":                         "long",
		"demo.receivers.OverloadCalls.typedRepositoryType(JsonWriterLike,ReceiverTarget)": "ReceiverTarget",
		"demo.receivers.OverloadCalls.typedString(JsonWriterLike,String)":                 "String",
	}
	for caller, parameterType := range definite {
		assertOnlyValueCall(t, records, nodes, caller, parameterType)
	}

	nullCaller := "demo.receivers.OverloadCalls.nullLiteral(JsonWriterLike)"
	for _, parameterType := range []string{"Boolean", "Number", "ReceiverTarget", "String"} {
		assertRuntimeEdge(t, records, nodes, nullCaller, valueTarget(parameterType), "possible-calls")
	}
	for _, parameterType := range []string{"boolean", "char", "double", "float", "long"} {
		assertNoRuntimeEdge(t, records, nodes, nullCaller, valueTarget(parameterType), "possible-calls")
	}

	unknownCaller := "demo.receivers.OverloadCalls.unknownExpression(JsonWriterLike)"
	for _, parameterType := range valueParameterTypes() {
		assertRuntimeEdge(t, records, nodes, unknownCaller, valueTarget(parameterType), "possible-calls")
	}

	ambiguousCaller := "demo.receivers.OverloadCalls.ambiguousArguments(PairTarget,int,int)"
	for _, target := range []string{
		"demo.receivers.PairTarget.pair(int,long)",
		"demo.receivers.PairTarget.pair(long,int)",
	} {
		assertRuntimeEdge(t, records, nodes, ambiguousCaller, target, "possible-calls")
	}
	assertAllEndpoints(t, records, nodes)
}

func TestRuntimeConstructorCallsAndOverrides(t *testing.T) {
	records := decodeRecords(t, analyzeFixture(t, runtimeFixture))
	nodes := nodeIndex(records)

	assertRuntimeEdge(t, records, nodes, "demo.runtime.RuntimeSlice.<init>()", "demo.runtime.RuntimeSlice.<init>(int)", "calls")
	assertRuntimeEdge(t, records, nodes, "demo.runtime.RuntimeSlice.<init>(int)", "demo.runtime.Base.<init>(int)", "calls")
	assertRuntimeEdge(t, records, nodes, "demo.runtime.RuntimeSlice.match(String)", "demo.runtime.Base.match(String)", "overrides")
	assertRuntimeEdge(t, records, nodes, "demo.runtime.RuntimeSlice.match(String)", "demo.runtime.Contract.match(String)", "overrides")
}

func TestRuntimeReadsWritesStayOnModeledFieldsAndParameters(t *testing.T) {
	records := decodeRecords(t, analyzeFixture(t, runtimeFixture))
	nodes := nodeIndex(records)
	caller := "demo.runtime.RuntimeSlice.exercise(int,ExternalReceiver)"
	parameter := caller + "#parameter:0:value"

	for _, edge := range []struct {
		relation string
		target   string
	}{
		{"reads", "demo.runtime.RuntimeSlice.field"},
		{"writes", "demo.runtime.RuntimeSlice.field"},
		{"reads", parameter},
		{"writes", parameter},
		{"reads", "demo.runtime.Helper.shared"},
		{"writes", "demo.runtime.Helper.shared"},
	} {
		assertRuntimeEdge(t, records, nodes, caller, edge.target, edge.relation)
	}

	shadow := "demo.runtime.RuntimeSlice.shadow(int)"
	assertNoRuntimeEdge(t, records, nodes, shadow, "demo.runtime.RuntimeSlice.field", "reads")
	assertNoRuntimeEdge(t, records, nodes, shadow, "demo.runtime.RuntimeSlice.field", "writes")
	assertNoRuntimeUnresolved(t, records, nodes, shadow, "writes", "field")
}

func TestRuntimeSemanticsDoNotClaimNestedCallableBodies(t *testing.T) {
	records := decodeRecords(t, analyzeFixture(t, runtimeFixture))
	nodes := nodeIndex(records)
	assertNoRuntimeEdge(
		t, records, nodes, "demo.runtime.RuntimeSlice.nestedBodies()",
		"demo.runtime.RuntimeSlice.definite(int)", "calls",
	)
}

func TestCallableBodyTokenRangesAreRetained(t *testing.T) {
	facts := newFactSet()
	state := &analysisState{
		declarations: make(map[string][]string), facts: facts,
		fields: make(map[string]map[string][]fieldDeclaration), namespaces: make(map[string]string),
		types: make(map[string][]typeDeclaration),
	}
	parseJavaSource(state, "file", "Sample.java", `class Sample {
    Sample() { this(1); }
    Sample(int value) { }
    void run(int value) { value++; }
}`)
	if len(state.callables) != 3 {
		t.Fatalf("retained callables = %d, want 3", len(state.callables))
	}
	for _, callable := range state.callables {
		if callable.bodyStart < 1 || callable.bodyEnd < callable.bodyStart || callable.tokens[callable.bodyStart-1].text != "{" || callable.tokens[callable.bodyEnd].text != "}" {
			t.Fatalf("invalid body range for %s: [%d,%d)", callable.name, callable.bodyStart, callable.bodyEnd)
		}
	}
}

func TestDirectParentIndexPreservesSortedRelationshipSets(t *testing.T) {
	state := &analysisState{facts: newFactSet()}
	state.facts.addEdge("child", "interface", "implements", "", nil, nil)
	state.facts.addEdge("child", "base", "extends", "", nil, nil)
	state.facts.addEdge("child", "base", "extends", "", nil, nil)
	state.facts.addEdge("other", "unrelated", "calls", "", nil, nil)

	state.indexDirectParents()
	if got, want := state.directParents("child"), []string{"base", "interface"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("direct parents = %v, want %v", got, want)
	}
	if got, want := state.directSuperclasses("child"), []string{"base"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("direct superclasses = %v, want %v", got, want)
	}
}

func TestOverrideIndexesScopeCandidatesAndTypeKinds(t *testing.T) {
	state := &analysisState{
		callables: []callableDeclaration{
			{id: "matching", ownerID: "parent", name: "run", signature: "String"},
			{id: "wrong-name", ownerID: "parent", name: "stop", signature: "String"},
			{id: "wrong-owner", ownerID: "sibling", name: "run", signature: "String"},
			{id: "wrong-signature", ownerID: "parent", name: "run", signature: "int"},
		},
		declarations: make(map[string][]string),
		types:        make(map[string][]typeDeclaration),
	}
	state.indexCallables()
	candidates := state.overrideCandidates("parent", "run", "String")
	if len(candidates) != 1 || candidates[0].id != "matching" {
		t.Fatalf("override candidates = %#v, want only matching", candidates)
	}

	state.registerType("demo.Contract", "contract", "interface")
	state.registerType("demo.Marker", "marker", "annotation")
	state.registerType("demo.Base", "base", "class")
	if !state.interfaceType("contract") || !state.interfaceType("marker") {
		t.Fatal("interface type index did not retain interface-like declarations")
	}
	if state.interfaceType("base") || state.interfaceType("missing") {
		t.Fatal("interface type index classified a class or missing declaration as interface-like")
	}
}

func TestCallableCandidateIndexPreservesLookupBehavior(t *testing.T) {
	state := &analysisState{callables: []callableDeclaration{
		{id: "method-z-varargs", ownerID: "target", name: "run", arity: 2, parameterTypes: []string{"String", "int..."}},
		{id: "wrong-owner", ownerID: "other", name: "run", arity: 2},
		{id: "method-a-fixed", ownerID: "target", name: "run", arity: 2, modifiers: []string{"static"}},
		{id: "constructor-b", ownerID: "target", name: "Target", arity: 1, constructor: true},
		{id: "wrong-name", ownerID: "target", name: "stop", arity: 2},
		{id: "method-b-one", ownerID: "target", name: "run", arity: 1},
		{id: "method-a-fixed", ownerID: "target", name: "run", arity: 2, modifiers: []string{"static"}},
	}}
	state.indexCallables()
	assertIDs := func(label string, candidates []callableDeclaration, want []string) {
		t.Helper()
		got := make([]string, 0, len(candidates))
		for _, candidate := range candidates {
			got = append(got, candidate.id)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s candidates = %v, want %v", label, got, want)
		}
	}

	assertIDs("overloads", state.callableCandidates("target", "run", 2, false, false), []string{"method-a-fixed", "method-z-varargs"})
	assertIDs("one-argument overloads", state.callableCandidates("target", "run", 1, false, false), []string{"method-b-one", "method-z-varargs"})
	assertIDs("varargs", state.callableCandidates("target", "run", 4, false, false), []string{"method-z-varargs"})
	assertIDs("static", state.callableCandidates("target", "run", 2, false, true), []string{"method-a-fixed"})
	assertIDs("instance", state.instanceCallableCandidates("target", "run", 2), []string{"method-z-varargs"})
	assertIDs("constructors", state.callableCandidates("target", "ignored", 1, true, false), []string{"constructor-b"})
}

func TestRuntimeSemanticsAreCanonicalDeterministicAndCheckoutIndependent(t *testing.T) {
	first := analyzeFixture(t, runtimeFixture)
	if second := analyzeFixture(t, runtimeFixture); !bytes.Equal(first, second) {
		t.Fatal("repeat runtime analysis was not byte-identical")
	}
	assertCanonicalJSONL(t, first)

	temporary := t.TempDir()
	left := filepath.Join(temporary, "left", "runtime")
	right := filepath.Join(temporary, "right", "runtime")
	copyTree(t, runtimeFixture, left)
	copyTree(t, runtimeFixture, right)
	if leftData, rightData := analyzeFixture(t, left), analyzeFixture(t, right); !bytes.Equal(leftData, rightData) {
		t.Fatal("absolute checkout path affected runtime output")
	}
}

func assertRuntimeEdge(t *testing.T, records []map[string]any, nodes map[string]map[string]any, sourceName, targetName, relation string) {
	t.Helper()
	source, target := nodes[sourceName], nodes[targetName]
	if source == nil || target == nil {
		t.Fatalf("missing runtime endpoint %s -> %s", sourceName, targetName)
	}
	for _, record := range records {
		if record["record"] == "edge" && record["relation"] == relation && record["source"] == source["id"] && record["target"] == target["id"] {
			return
		}
	}
	t.Fatalf("missing %s edge %s -> %s", relation, sourceName, targetName)
}

func assertOnlyValueCall(t *testing.T, records []map[string]any, nodes map[string]map[string]any, caller, parameterType string) {
	t.Helper()
	for _, candidateType := range valueParameterTypes() {
		target := valueTarget(candidateType)
		if candidateType == parameterType {
			assertRuntimeEdge(t, records, nodes, caller, target, "calls")
			assertNoRuntimeEdge(t, records, nodes, caller, target, "possible-calls")
			continue
		}
		assertNoRuntimeEdge(t, records, nodes, caller, target, "calls")
		assertNoRuntimeEdge(t, records, nodes, caller, target, "possible-calls")
	}
}

func valueParameterTypes() []string {
	return []string{"Boolean", "Number", "ReceiverTarget", "String", "boolean", "char", "double", "float", "long"}
}

func valueTarget(parameterType string) string {
	return "demo.receivers.JsonWriterLike.value(" + parameterType + ")"
}

func assertNoRuntimeEdge(t *testing.T, records []map[string]any, nodes map[string]map[string]any, sourceName, targetName, relation string) {
	t.Helper()
	source, target := nodes[sourceName], nodes[targetName]
	for _, record := range records {
		if record["record"] == "edge" && record["relation"] == relation && record["source"] == source["id"] && record["target"] == target["id"] {
			t.Fatalf("unexpected %s edge %s -> %s", relation, sourceName, targetName)
		}
	}
}

func assertNoRuntimeUnresolved(t *testing.T, records []map[string]any, nodes map[string]map[string]any, sourceName, relation, expression string) {
	t.Helper()
	source := nodes[sourceName]
	for _, record := range records {
		if record["record"] == "unresolved" && record["source"] == source["id"] && record["relation"] == relation && record["expression"] == expression {
			t.Fatalf("unexpected unresolved %s %s %q", sourceName, relation, expression)
		}
	}
}

func assertRuntimeUnresolved(t *testing.T, records []map[string]any, nodes map[string]map[string]any, sourceName, relation, expression, reason string) {
	t.Helper()
	source := nodes[sourceName]
	for _, record := range records {
		if record["record"] == "unresolved" && record["source"] == source["id"] && record["relation"] == relation && record["expression"] == expression && record["reason"] == reason {
			return
		}
	}
	t.Fatalf("missing unresolved %s %s %q (%s)", sourceName, relation, expression, reason)
}
