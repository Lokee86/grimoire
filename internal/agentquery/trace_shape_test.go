package agentquery

import (
	"testing"

	"github.com/Lokee86/grimoire/internal/arcanagraph"
	"github.com/Lokee86/grimoire/internal/structure"
)

func TestFinalizeTraceResponsePrefersBehaviorAndCompactsByDefault(t *testing.T) {
	handle := func(value string) Handle { return Handle{Value: value, Provider: "arcana", Snapshot: "snapshot"} }
	span := func(value string) Range {
		return Range{Path: value + ".gd", StartLine: 1, EndLine: 2, Handle: handle("source-" + value)}
	}
	structural := Path{
		Nodes: []Node{
			{Handle: handle("type"), Kind: "type", Name: "AsteroidSync"},
			{Handle: handle("method"), Kind: "method", Name: "helper"},
		},
		Steps: []PathStep{{
			From: handle("type"), To: handle("method"), Relation: "contains", Certainty: "definite",
			Evidence: []string{"one", "two", "three"}, Spans: []Range{span("one"), span("two"), span("three")},
		}},
	}
	behavioral := Path{
		Nodes: []Node{
			{Handle: handle("type"), Kind: "type", Name: "AsteroidSync"},
			{Handle: handle("apply"), Kind: "method", Name: "apply_delete"},
			{Handle: handle("remove"), Kind: "method", Name: "remove_asteroid"},
		},
		Steps: []PathStep{
			{From: handle("type"), To: handle("apply"), Relation: "contains", Certainty: "definite"},
			{
				From: handle("apply"), To: handle("remove"), Relation: "calls", Certainty: "definite",
				Evidence: []string{"one", "two", "three"}, Spans: []Range{span("one"), span("two"), span("three")},
			},
		},
	}
	response := Response{Paths: []Path{structural, behavioral}}

	finalizeTraceResponse(Request{Mode: "trace", Query: "asteroid delete tombstone"}, &response, 8)

	if len(response.Paths) != 1 {
		t.Fatalf("expected structural-only path to be suppressed, got %d paths", len(response.Paths))
	}
	if response.Paths[0].Summary == "" || len(response.Paths[0].Relations) != 1 || response.Paths[0].Relations[0] != "calls" {
		t.Fatalf("behavioral path was not ranked first or containment was not collapsed: %+v", response.Paths)
	}
	for _, path := range response.Paths {
		if len(path.Nodes) != 0 || len(path.Steps) != 0 {
			t.Fatalf("summary trace repeated full graph objects: %+v", path)
		}
		if len(path.ContinuationHandles) == 0 {
			t.Fatalf("summary trace lost continuation handles: %+v", path)
		}
	}
	if got := len(response.Paths[0].Evidence); got != 1 {
		t.Fatalf("compact evidence count = %d, want one cited behavior edge", got)
	}
	if response.Paths[0].Evidence[0].Handle == "" {
		t.Fatalf("compact evidence lost exact source handle: %+v", response.Paths[0].Evidence)
	}
}

func TestFinalizeTraceResponseFullDetailPreservesNodes(t *testing.T) {
	path := Path{
		Nodes: []Node{{Kind: "method", Name: "apply_update"}, {Kind: "method", Name: "write_state"}},
		Steps: []PathStep{{Relation: "calls", Evidence: []string{"one", "two", "three"}}},
	}
	response := Response{Paths: []Path{path}}

	finalizeTraceResponse(Request{Mode: "trace", Detail: "full"}, &response, 8)

	if len(response.Paths[0].Nodes) != 2 {
		t.Fatalf("full detail discarded nodes: %+v", response.Paths[0])
	}
	if len(response.Paths[0].Steps[0].Evidence) != 3 {
		t.Fatalf("full detail unexpectedly bounded evidence: %+v", response.Paths[0].Steps[0])
	}
}

func TestRankTraceNeighborsPrefersBehavioralMutation(t *testing.T) {
	id1, id2, id3 := uint32(1), uint32(2), uint32(3)
	neighbors := []arcanagraph.QueryNeighbor{
		{Relation: "contains", Node: structure.Node{NodeID: &id1, Kind: "method", Name: "helper"}},
		{Relation: "references", Node: structure.Node{NodeID: &id2, Kind: "field", Name: "visual_position"}},
		{Relation: "calls", Node: structure.Node{NodeID: &id3, Kind: "method", Name: "apply_asteroid_delete"}},
	}

	ranked := rankTraceNeighbors(neighbors, "asteroid delete lifecycle")

	if ranked[0].Relation != "calls" || ranked[0].Node.Name != "apply_asteroid_delete" {
		t.Fatalf("mutation path was not prioritized: %+v", ranked)
	}
}

func TestTracePathScoreDeprioritizesDiagnosticsAndParameterReads(t *testing.T) {
	lifecycle := Path{
		Nodes: []Node{
			{Kind: "type", Name: "AsteroidSync", Path: "client/asteroid_sync.gd"},
			{Kind: "method", Name: "remove_missing", Path: "client/asteroid_sync.gd"},
			{Kind: "method", Name: "remove_asteroid", Path: "client/asteroid_sync.gd"},
		},
		Steps: []PathStep{{Relation: "contains"}, {Relation: "calls"}},
	}
	diagnostic := Path{
		Nodes: []Node{
			{Kind: "type", Name: "AsteroidSync", Path: "client/asteroid_sync.gd"},
			{Kind: "method", Name: "apply_asteroid_scale", Path: "client/asteroid_sync.gd"},
			{Kind: "method", Name: "emit_canonical", Path: "client/logging/logger.gd"},
		},
		Steps: []PathStep{{Relation: "contains"}, {Relation: "calls"}},
	}
	parameterRead := Path{
		Nodes: []Node{
			{Kind: "type", Name: "AsteroidSync", Path: "client/asteroid_sync.gd"},
			{Kind: "method", Name: "_clear_deleted_asteroid_id", Path: "client/asteroid_sync.gd"},
			{Kind: "parameter", Name: "asteroid_id", Path: "client/asteroid_sync.gd"},
		},
		Steps: []PathStep{{Relation: "contains"}, {Relation: "reads"}},
	}

	lifecycleScore := tracePathScore(lifecycle, "AsteroidSync")
	if lifecycleScore <= tracePathScore(diagnostic, "AsteroidSync") {
		t.Fatalf("diagnostic branch outranked lifecycle: lifecycle=%f diagnostic=%f", lifecycleScore, tracePathScore(diagnostic, "AsteroidSync"))
	}
	if lifecycleScore <= tracePathScore(parameterRead, "AsteroidSync") {
		t.Fatalf("parameter read outranked lifecycle: lifecycle=%f read=%f", lifecycleScore, tracePathScore(parameterRead, "AsteroidSync"))
	}
}

func TestTraceDefaultsToCompactEightPathResponse(t *testing.T) {
	request := normalizeRequest(Request{Mode: "trace"})
	if request.Limit != 8 {
		t.Fatalf("trace default limit = %d, want 8", request.Limit)
	}
	if request.Detail != "summary" {
		t.Fatalf("trace default detail = %q, want summary", request.Detail)
	}
}
