package main

import (
	"strings"
	"testing"
)

func TestAdvancedResolutionModelsInterfacesClosuresAndFunctionValues(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"go.mod": "module example.com/advanced\n\ngo 1.22\n",
		"main.go": `package advanced

type Runner interface { Run() }
type First struct{}
type Second struct{}

func (First) Run() {}
func (Second) Run() {}
func target() {}
func invoke(r Runner) { r.Run() }

func caller() {
	f := target
	f()
	captured := 1
	closure := func() { _ = captured; target() }
	closure()
	func() { target() }()
	invoke(First{})
	invoke(Second{})
}
`,
	})

	facts, summary, err := scanRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	if summary.SemanticErrors != 0 {
		t.Fatalf("semantic errors = %d, want 0", summary.SemanticErrors)
	}
	if summary.Closures != 2 {
		t.Fatalf("closures = %d, want 2", summary.Closures)
	}
	if summary.Captures == 0 {
		t.Fatal("expected at least one modeled closure capture")
	}
	if summary.UnresolvedCalls != 0 {
		t.Fatalf("unresolved calls = %d, want 0: %#v", summary.UnresolvedCalls, facts.Unresolved)
	}

	invoke := hashIdentity("function:example.com/advanced:invoke")
	caller := hashIdentity("function:example.com/advanced:caller")
	target := hashIdentity("function:example.com/advanced:target")
	contract := hashIdentity(interfaceMethodIdentity("example.com/advanced", "Runner", "Run"))
	firstMethod := hashIdentity("method:example.com/advanced:First.Run")
	secondMethod := hashIdentity("method:example.com/advanced:Second.Run")
	firstType := hashIdentity("type:example.com/advanced:First")
	secondType := hashIdentity("type:example.com/advanced:Second")
	runnerType := hashIdentity("type:example.com/advanced:Runner")

	for _, edge := range []struct {
		source, target NodeKey
		relation       RelationKind
	}{
		{invoke, contract, RelCalls},
		{invoke, firstMethod, RelPossibleCalls},
		{invoke, secondMethod, RelPossibleCalls},
		{firstMethod, contract, RelImplements},
		{secondMethod, contract, RelImplements},
		{firstType, runnerType, RelImplements},
		{secondType, runnerType, RelImplements},
		{caller, target, RelCalls},
	} {
		if !hasEdge(facts, edge.source, edge.target, edge.relation) {
			t.Fatalf("missing %s edge %s -> %s", edge.relation, edge.source, edge.target)
		}
	}

	closures := make([]NodeKey, 0, 2)
	for _, node := range facts.Nodes {
		if node.Kind == KindFunction && strings.HasPrefix(node.Name, "closure@") {
			closures = append(closures, node.Key)
		}
	}
	if len(closures) != 2 {
		t.Fatalf("closure nodes = %d, want 2", len(closures))
	}
	var capturedVariable NodeKey
	for _, node := range facts.Nodes {
		if node.Kind == KindVariable && node.Name == "captured" {
			capturedVariable = node.Key
			break
		}
	}
	if capturedVariable == "" {
		t.Fatal("missing captured variable node")
	}
	captureEdges := 0
	for _, closure := range closures {
		if hasEdge(facts, closure, capturedVariable, RelReferences) {
			captureEdges++
		}
		if !hasEdge(facts, caller, closure, RelCalls) {
			t.Fatalf("caller does not resolve closure %s", closure)
		}
		if !hasEdge(facts, closure, target, RelCalls) {
			t.Fatalf("closure %s does not call target", closure)
		}
	}
	if captureEdges != 1 {
		t.Fatalf("captured variable references = %d, want 1", captureEdges)
	}
}
