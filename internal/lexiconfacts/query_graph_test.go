package lexiconfacts

import "testing"

func TestTraceUsesBreadthFirstCandidateFairness(t *testing.T) {
	corpus := &Corpus{facts: library{
		nodes: map[string]Node{
			"root":    {ID: "root", Kind: "type", Name: "Root"},
			"methodA": {ID: "methodA", Kind: "method", Name: "MethodA"},
			"methodB": {ID: "methodB", Kind: "method", Name: "MethodB"},
			"helperA": {ID: "helperA", Kind: "method", Name: "HelperA"},
			"helperB": {ID: "helperB", Kind: "method", Name: "HelperB"},
		},
		edges: []Edge{
			{Source: "root", Target: "methodA", Relation: "contains"},
			{Source: "methodA", Target: "helperA", Relation: "calls"},
			{Source: "root", Target: "methodB", Relation: "contains"},
			{Source: "methodB", Target: "helperB", Relation: "calls"},
		},
	}}

	paths := corpus.Trace([]string{"root"}, nil, "outgoing", nil, 3, 2)
	if len(paths) != 2 {
		t.Fatalf("path count = %d, want 2", len(paths))
	}
	if got := paths[0].Nodes[1].Name; got != "MethodA" {
		t.Fatalf("first breadth entry = %q, want MethodA", got)
	}
	if got := paths[1].Nodes[1].Name; got != "MethodB" {
		t.Fatalf("second breadth entry = %q, want MethodB before another MethodA descendant", got)
	}
	if paths[0].Edges[0].Relation != "contains" || paths[0].Edges[1].Relation != "calls" {
		t.Fatalf("containment should remain traversal context without becoming its own result: %+v", paths[0].Edges)
	}
}
