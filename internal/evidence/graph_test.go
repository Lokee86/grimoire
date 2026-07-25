package evidence

import (
	"slices"
	"testing"
)

func TestMergeGraphSignalsKeepsStrongestIndependentValues(t *testing.T) {
	left := Descriptor{Graph: &GraphSignals{
		Distance: 2, Relations: []string{"incoming:calls"}, ModuleProximity: 0.4,
		SymbolRole: "function", Centrality: 0.2,
	}}
	right := Descriptor{Graph: &GraphSignals{
		Distance: 1, Relations: []string{"outgoing:calls"}, ModuleProximity: 0.9,
		SymbolRole: "class", Centrality: 0.8,
	}}

	got := Merge(left, right)
	if got.Graph == nil {
		t.Fatal("Merge() omitted graph signals")
	}
	if got.Graph.Distance != 1 || got.Graph.ModuleProximity != 0.9 ||
		got.Graph.SymbolRole != "function" || got.Graph.Centrality != 0.8 {
		t.Fatalf("Merge() graph signals = %+v", got.Graph)
	}
	if !slices.Contains(got.Graph.Relations, "incoming:calls") ||
		!slices.Contains(got.Graph.Relations, "outgoing:calls") {
		t.Fatalf("Merge() graph relations = %+v", got.Graph.Relations)
	}

	right.Graph.Relations[0] = "mutated"
	if slices.Contains(got.Graph.Relations, "mutated") {
		t.Fatal("Merge() retained aliased graph relations")
	}
}
