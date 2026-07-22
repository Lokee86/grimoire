package main

import "testing"

func TestSemanticResolutionCoversInternalFunctionsMethodsAndRecursion(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"go.mod": "module example.com/semantic\n\ngo 1.22\n",
		"main.go": `package semantic

import (
	"fmt"
	"example.com/semantic/internal/sub"
)

type Widget int

func helper() {}
func recursive() { recursive() }
func caller() {
	helper()
	sub.Function()
	var value sub.Thing
	value.Method()
	Widget(0).Local()
	_ = len([]int{})
	fmt.Println("external")
	var dynamic func()
	dynamic()
}
func (Widget) Local() {}
`,
		"internal/sub/sub.go": `package sub

type Thing struct{}
func Function() {}
func (Thing) Method() {}
`,
	})

	facts, summary, err := scanRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	if summary.SemanticErrors != 0 {
		t.Fatalf("semantic errors = %d, want 0", summary.SemanticErrors)
	}
	if summary.CallExpressions != 9 || summary.DirectCalls != 7 || summary.UnresolvedCalls != 1 {
		t.Fatalf(
			"call counts = total %d resolved %d unresolved %d, want 9/7/1",
			summary.CallExpressions,
			summary.DirectCalls,
			summary.UnresolvedCalls,
		)
	}
	if summary.BuiltinCalls != 1 || summary.ConversionCalls != 1 ||
		summary.ExternalCalls != 1 || summary.DynamicCalls != 1 {
		t.Fatalf(
			"classified calls = builtin %d conversion %d external %d dynamic %d, want 1 each",
			summary.BuiltinCalls,
			summary.ConversionCalls,
			summary.ExternalCalls,
			summary.DynamicCalls,
		)
	}

	caller := hashIdentity("function:example.com/semantic:caller")
	helper := hashIdentity("function:example.com/semantic:helper")
	recursive := hashIdentity("function:example.com/semantic:recursive")
	internalFunction := hashIdentity("function:example.com/semantic/internal/sub:Function")
	internalMethod := hashIdentity("method:example.com/semantic/internal/sub:Thing.Method")
	localMethod := hashIdentity("method:example.com/semantic:Widget.Local")
	for _, edge := range [][2]NodeKey{
		{caller, helper},
		{caller, internalFunction},
		{caller, internalMethod},
		{caller, localMethod},
		{recursive, recursive},
	} {
		if !hasEdge(facts, edge[0], edge[1], RelCalls) {
			t.Fatalf("missing resolved call edge %016x -> %016x", edge[0], edge[1])
		}
	}

	builtin := hashIdentity("function:go:builtins:len")
	external := hashIdentity("function:fmt:Println")
	widgetType := hashIdentity("type:example.com/semantic:Widget")
	if !hasEdge(facts, caller, builtin, RelCalls) {
		t.Fatal("missing built-in call edge")
	}
	if !hasEdge(facts, caller, external, RelCalls) {
		t.Fatal("missing external call edge")
	}
	if !hasEdge(facts, caller, widgetType, RelConvertsTo) {
		t.Fatal("missing conversion edge")
	}
	assertUnresolvedReason(t, facts, "dynamic", ReasonDynamicTarget)

	if encoded := encodeFacts(facts); encoded == "" {
		t.Fatal("unexpected empty encoded facts")
	}
}

func assertUnresolvedReason(
	t *testing.T,
	facts RepositoryFacts,
	expression string,
	reason UnresolvedReason,
) {
	t.Helper()
	for _, reference := range facts.Unresolved {
		if reference.Expression == expression && reference.Reason == reason {
			return
		}
	}
	t.Fatalf("missing unresolved %q with reason %q", expression, reason)
}
