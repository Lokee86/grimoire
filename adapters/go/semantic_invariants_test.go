package main

import "testing"

func TestMergeSemanticCallPreservesConcreteTargetsAndRemovesContracts(t *testing.T) {
	contract := NodeKey("contract")
	first := NodeKey("first")
	second := NodeKey("second")
	scanner := scanner{semanticCalls: map[string]semanticCall{}}
	scanner.mergeSemanticCall("call", semanticCall{
		edges: []semanticEdge{
			{target: contract, relation: RelCalls},
			{target: first, relation: RelCalls},
		},
		resolved: true,
		class:    callClassInterface,
		contract: contract,
	})
	scanner.mergeSemanticCall("call", semanticCall{
		edges:    []semanticEdge{{target: second, relation: RelPossibleCalls}},
		resolved: true,
		class:    callClassInterface,
		contract: contract,
	})
	result := scanner.semanticCalls["call"]
	if len(result.edges) != 2 {
		t.Fatalf("merged edges = %#v, want two concrete targets", result.edges)
	}
	for _, edge := range result.edges {
		if edge.target == contract {
			t.Fatal("interface declaration remained a runtime call target")
		}
		if edge.relation != RelPossibleCalls {
			t.Fatalf("concrete edge relation = %s, want possible-calls", edge.relation)
		}
	}
}

func TestSemanticRelationsDoNotEmitImplementsSelfEdges(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"go.mod": "module example.com/implements\n\ngo 1.22\n",
		"main.go": `package implements

type Contract interface { Run() }
type Embedded struct { Contract }
type Direct struct{}
func (Direct) Run() {}
func recursive() { recursive() }
`,
	})

	facts, _, err := scanRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, edge := range facts.Edges {
		if edge.Relation == RelImplements && edge.Source == edge.Target {
			t.Fatalf("implements self-edge emitted for %s", edge.Source)
		}
	}
	recursive := hashIdentity("function:example.com/implements:recursive")
	if !hasEdge(facts, recursive, recursive, RelCalls) {
		t.Fatal("recursive call self-edge should remain valid")
	}
}
