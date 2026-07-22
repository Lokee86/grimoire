package main

import "testing"

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
