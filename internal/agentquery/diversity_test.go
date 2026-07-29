package agentquery

import "testing"

func TestSelectDiverseResultsPreventsOneFileFromMonopolizingLane(t *testing.T) {
	result := func(name, path string) Result {
		return Result{Node: Node{Name: name, Path: path, Handle: Handle{Value: name}}}
	}
	selected := selectDiverseResults([]Result{
		result("one", "same.go"),
		result("two", "same.go"),
		result("three", "same.go"),
		result("other", "other.go"),
		result("third", "third.go"),
	}, 4)
	if len(selected) != 4 {
		t.Fatalf("selected %d results, want 4", len(selected))
	}
	if selected[0].Node.Name != "one" || selected[1].Node.Name != "two" || selected[2].Node.Name != "other" || selected[3].Node.Name != "third" {
		t.Fatalf("one file monopolized the lane: %+v", selected)
	}
	for index, value := range selected {
		if value.Rank != index+1 {
			t.Fatalf("rank %d = %d", index, value.Rank)
		}
	}
}

func TestResultSemanticKeyCollapsesEquivalentContainerNodes(t *testing.T) {
	left := Result{Kind: "directory", Node: Node{Kind: "directory", Name: "realtime", QualifiedName: "client/realtime", Path: "client/realtime"}}
	right := left
	right.Node.Handle.Value = "different-provider-handle"
	if resultSemanticKey(left) != resultSemanticKey(right) {
		t.Fatal("equivalent container nodes did not share a semantic key")
	}
}
