package arcanagraph

import (
	"context"
	"encoding/json"
	"testing"
)

func TestNeighborsBatchUsesOneProtocolRunForFrontier(t *testing.T) {
	runs := 0
	run := func(
		_ context.Context,
		_ string,
		_ string,
		requests []protocolRequest,
	) (map[string]protocolResponse, error) {
		runs++
		if len(requests) != 4 {
			t.Fatalf("request count = %d, want 4 for two nodes and both directions", len(requests))
		}
		responses := make(map[string]protocolResponse, len(requests))
		for _, request := range requests {
			if request.Limit != maxNeighborResults {
				t.Fatalf("neighbor request limit = %d, want %d", request.Limit, maxNeighborResults)
			}
			nodeID := *request.NodeID + 100
			data, err := json.Marshal(neighborResult{
				Direction: request.Direction,
				Relationships: []relatedNode{{
					Relation: "calls",
					Node:     arcanaNode{NodeID: nodeID, Identity: "target", Kind: "function", Name: "target"},
				}},
			})
			if err != nil {
				t.Fatal(err)
			}
			responses[request.ID] = protocolResponse{Protocol: protocolID, ID: request.ID, OK: true, Result: data}
		}
		return responses, nil
	}

	neighbors, err := (Client{Run: run}).NeighborsBatch(
		context.Background(), "snapshot", []uint32{1, 2}, "both", nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if runs != 1 {
		t.Fatalf("protocol runs = %d, want 1", runs)
	}
	if len(neighbors[1]) != 2 || len(neighbors[2]) != 2 {
		t.Fatalf("batched neighbors = %+v", neighbors)
	}
}

func TestNeighborsBatchCollapsesMultipleRelationFilters(t *testing.T) {
	run := func(
		_ context.Context,
		_ string,
		_ string,
		requests []protocolRequest,
	) (map[string]protocolResponse, error) {
		if len(requests) != 2 {
			t.Fatalf("request count = %d, want one unfiltered request per direction", len(requests))
		}
		responses := make(map[string]protocolResponse, len(requests))
		for _, request := range requests {
			if request.Relation != "" || len(request.Relations) != 2 {
				t.Fatalf("protocol filters = relation %q relations %v, want one relation set", request.Relation, request.Relations)
			}
			data, err := json.Marshal(neighborResult{
				Direction: request.Direction,
				Relationships: []relatedNode{
					{Relation: "calls", Node: arcanaNode{NodeID: 2, Identity: "call", Kind: "function", Name: "call"}},
					{Relation: "reads", Node: arcanaNode{NodeID: 3, Identity: "read", Kind: "local", Name: "read"}},
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			responses[request.ID] = protocolResponse{Protocol: protocolID, ID: request.ID, OK: true, Result: data}
		}
		return responses, nil
	}

	neighbors, err := (Client{Run: run}).NeighborsBatch(
		context.Background(), "snapshot", []uint32{1}, "both", []string{"calls", "publishes"},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(neighbors[1]) != 2 {
		t.Fatalf("filtered neighbors = %+v", neighbors[1])
	}
	for _, neighbor := range neighbors[1] {
		if neighbor.Relation != "calls" {
			t.Fatalf("unrequested relation survived local filtering: %+v", neighbor)
		}
	}
}
