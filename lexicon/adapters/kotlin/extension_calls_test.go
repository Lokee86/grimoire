package main

import (
	"bytes"
	"path/filepath"
	"testing"
)

const extensionFixturePath = "testdata/extensions"

func TestTypedReceiverExtensionsEmitConservativeCalls(t *testing.T) {
	first, err := analyzeRepository(filepath.FromSlash(extensionFixturePath))
	if err != nil {
		t.Fatal(err)
	}
	second, err := analyzeRepository(filepath.FromSlash(extensionFixturePath))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("extension analysis changed byte output")
	}
	records := decodeFacts(t, first)
	nodes := factRecords(records, "node")
	edges := factRecords(records, "edge")
	unresolved := factRecords(records, "unresolved")
	assertEdgeEndpoints(t, nodes, edges)

	direct := requireRuntimeNode(t, nodes, "function", "extensions.api.direct(Int)")
	for _, sourceQN := range []string{
		"extensions.app.Usage.throughParameter(ModelItem)",
		"extensions.app.Usage.throughLocal()",
		"extensions.app.Usage.throughProperty()",
		"extensions.app.Usage.shadowing(ModelItem,Boolean)",
	} {
		requireRuntimeEdge(t, edges,
			requireRuntimeNode(t, nodes, "method", sourceQN), direct, "calls")
	}

	parameter := requireRuntimeNode(t, nodes, "method", "extensions.app.Usage.throughParameter(ModelItem)")
	for _, targetQN := range []string{
		"extensions.api.defaulted(Int)",
		"extensions.api.spread(Int)",
	} {
		requireRuntimeEdge(t, edges, parameter,
			requireRuntimeNode(t, nodes, "function", targetQN), "calls")
	}

	imports := requireRuntimeNode(t, nodes, "method", "extensions.app.Usage.throughImports(ModelItem)")
	for _, target := range []map[string]any{
		requireRuntimeNode(t, nodes, "function", "extensions.api.imported()"),
		requireRuntimeNode(t, nodes, "function", "extensions.wild.wild()"),
		requireRuntimeNode(t, nodes, "method", "extensions.app.Usage.lexical(Int)"),
		requireRuntimeNode(t, nodes, "function", "extensions.app.samePackage()"),
	} {
		requireRuntimeEdge(t, edges, imports, target, "calls")
	}

	ambiguous := requireRuntimeNode(t, nodes, "method", "extensions.app.Usage.ambiguous(ModelItem)")
	for _, targetQN := range []string{
		"extensions.api.ambiguous(Int)",
		"extensions.api.ambiguous(String)",
	} {
		target := requireRuntimeNode(t, nodes, "function", targetQN)
		requireRuntimeEdge(t, edges, ambiguous, target, "possible-calls")
		assertNoRuntimeEdge(t, edges, ambiguous, target, "calls")
	}

	external := requireRuntimeNode(t, nodes, "method", "extensions.app.Usage.external(String,ModelItem)")
	requireRuntimeUnresolved(t, unresolved, external, "text.externalOnly()", "dynamic-target")
	requireRuntimeUnresolved(t, unresolved, external, "item.thirdParty()", "external-target")

	memberWins := requireRuntimeNode(t, nodes, "method", "extensions.app.Usage.memberWins(ModelItem)")
	collision := requireRuntimeNode(t, nodes, "function", "extensions.api.collision()")
	requireRuntimeUnresolved(t, unresolved, memberWins, "item.collision()", "dynamic-target")
	assertNoRuntimeEdge(t, edges, memberWins, collision, "calls")
}

func TestTypedReceiverExtensionsRejectUnsafeReceiverFormsAndShadowing(t *testing.T) {
	data, err := analyzeRepository(filepath.FromSlash(extensionFixturePath))
	if err != nil {
		t.Fatal(err)
	}
	records := decodeFacts(t, data)
	nodes := factRecords(records, "node")
	edges := factRecords(records, "edge")
	unresolved := factRecords(records, "unresolved")
	direct := requireRuntimeNode(t, nodes, "function", "extensions.api.direct(Int)")

	unsupported := requireRuntimeNode(t, nodes, "method", "extensions.app.Usage.unsupported(ModelItem,Any)")
	assertNoRuntimeEdge(t, edges, unsupported, direct, "calls")
	for _, expected := range []struct {
		expression string
		reason     string
	}{
		{expression: "item.direct(1)", reason: "dynamic-target"},
		{expression: "fluent()", reason: "unsupported-form"},
		{expression: "direct(2)", reason: "unsupported-form"},
		{expression: "dynamic.direct(3)", reason: "dynamic-target"},
		{expression: "inferred.direct(4)", reason: "dynamic-target"},
	} {
		requireRuntimeUnresolved(t, unresolved, unsupported, expected.expression, expected.reason)
	}

	shadowing := requireRuntimeNode(t, nodes, "method", "extensions.app.Usage.shadowing(ModelItem,Boolean)")
	requireRuntimeUnresolved(t, unresolved, shadowing, "item.direct(6)", "dynamic-target")
	requireRuntimeUnresolved(t, unresolved, shadowing, "property.direct(7)", "dynamic-target")
	requireRuntimeEdge(t, edges, shadowing, direct, "calls")
	if count := runtimeEdgeCount(edges, shadowing, direct, "calls"); count != 3 {
		t.Fatalf("shadowing direct call count = %d, want 3", count)
	}
}

func TestRuntimeTypeHasOrdinaryMemberUsesAcceptedArityIndex(t *testing.T) {
	file := &parsedKotlinFile{packageName: "members", path: "Members.kt"}
	baseDeclaration := &declaration{kind: "type", name: "Base"}
	childDeclaration := &declaration{
		kind: "type", name: "Child", supertypes: []supertypeDecl{{targetName: "Base"}},
	}
	state := &analysis{runtime: newRuntimeIndex()}
	state.indexRuntimeDeclaration(file, baseDeclaration, "base", "type", "members", "module", "members.Base")
	state.indexRuntimeDeclaration(file, childDeclaration, "child", "type", "members", "module", "members.Child")

	callables := []struct {
		declaration *declaration
		id          string
		owner       string
	}{
		{declaration: &declaration{name: "exact", parameters: []parameterDecl{{name: "value"}}}, id: "exact", owner: "members.Child"},
		{declaration: &declaration{name: "defaulted", parameters: []parameterDecl{
			{name: "required"}, {name: "optional", hasDefault: true},
		}}, id: "defaulted", owner: "members.Child"},
		{declaration: &declaration{name: "spread", parameters: []parameterDecl{
			{name: "values", modifiers: []string{"vararg"}},
		}}, id: "spread", owner: "members.Child"},
		{declaration: &declaration{name: "inherited", parameters: []parameterDecl{{name: "value"}}}, id: "inherited", owner: "members.Base"},
	}
	for _, callable := range callables {
		state.indexRuntimeDeclaration(
			file, callable.declaration, callable.id, "method", callable.owner, "type",
			callable.owner+"."+callable.declaration.name,
		)
	}

	receiver := state.runtime.typesByID["child"]
	checks := []struct {
		name  string
		arity int
		want  bool
	}{
		{name: "exact", arity: 1, want: true},
		{name: "exact", arity: 0, want: false},
		{name: "defaulted", arity: 1, want: true},
		{name: "defaulted", arity: 2, want: true},
		{name: "defaulted", arity: 0, want: false},
		{name: "spread", arity: 0, want: true},
		{name: "spread", arity: 3, want: true},
		{name: "inherited", arity: 1, want: true},
	}
	for _, check := range checks {
		if got := state.runtimeTypeHasOrdinaryMember(receiver, check.name, check.arity); got != check.want {
			t.Errorf("ordinary member %s/%d = %v, want %v", check.name, check.arity, got, check.want)
		}
	}
}

func runtimeEdgeCount(edges []map[string]any, source, target map[string]any, relation string) int {
	count := 0
	for _, edge := range edges {
		if edge["source"] == source["id"] && edge["target"] == target["id"] && edge["relation"] == relation {
			count++
		}
	}
	return count
}
