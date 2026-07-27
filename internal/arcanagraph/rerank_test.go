package arcanagraph

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Lokee86/grimoire/internal/structure"
)

func TestSemanticCandidateLimitWidensRecallWithinBound(t *testing.T) {
	if got := SemanticCandidateLimit(0); got != 0 {
		t.Fatalf("zero final limit widened to %d", got)
	}
	if got := SemanticCandidateLimit(6); got != 1536 {
		t.Fatalf("six production seeds widened to %d, want 1536", got)
	}
	if got := SemanticCandidateLimit(100); got != maxHybridSemanticCandidates {
		t.Fatalf("candidate pool exceeded bound: %d", got)
	}
}

func TestRerankSeedsPromotesSpecificConceptualDeclaration(t *testing.T) {
	semantic := []SemanticSeed{
		{
			Node:  structure.Node{Kind: "function", Name: "Search", Path: "internal/knowledge/search.go"},
			Score: 0.95, Rank: 1,
		},
		{
			Node:  structure.Node{Kind: "function", Name: "graph_documents", Path: "arcana/src/vector/documents.rs"},
			Score: 0.90, Rank: 188,
		},
		{
			Node:  structure.Node{Kind: "function", Name: "unrelated_helper", Path: "internal/misc/helper.go"},
			Score: 0.50, Rank: 1000,
		},
	}

	seeds, err := (Client{}).RerankSeeds(
		context.Background(), "",
		"Where is each repository graph entity rendered with bounded neighborhood text for embeddings?",
		nil, semantic, 1,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(seeds) != 1 || seeds[0].Node.Name != "graph_documents" {
		t.Fatalf("specific conceptual declaration was not promoted: %+v", seeds)
	}
}

func TestSeedGraphProximityRewardsCandidateConnections(t *testing.T) {
	candidates := []hybridSeedCandidate{
		{node: structure.Node{Kind: "function", Name: "Isolated", Path: "isolated.go"}, base: 0.8},
		{node: structure.Node{Kind: "function", Name: "Coordinator", Path: "coordinator.go"}, base: 0.7},
		{node: structure.Node{Kind: "function", Name: "Worker", Path: "worker.go"}, base: 0.6},
	}
	nodes := map[string]map[string]any{
		"Isolated": {
			"node_id": 1, "identity": "isolated", "kind": "function",
			"path": "isolated.go", "name": "Isolated",
		},
		"Coordinator": {
			"node_id": 2, "identity": "coordinator", "kind": "function",
			"path": "coordinator.go", "name": "Coordinator",
		},
		"Worker": {
			"node_id": 3, "identity": "worker", "kind": "function",
			"path": "worker.go", "name": "Worker",
		},
	}
	client := Client{Run: func(
		_ context.Context,
		_, _ string,
		requests []protocolRequest,
	) (map[string]protocolResponse, error) {
		responses := make(map[string]protocolResponse, len(requests))
		for _, request := range requests {
			switch request.Op {
			case "resolve_symbol":
				responses[request.ID] = successfulResponse(t, request.ID, map[string]any{
					"nodes": []any{nodes[request.Name]},
				})
			case "neighbors":
				if request.NodeID == nil {
					t.Fatal("neighbor request omitted node id")
				}
				relationships := []any{}
				switch *request.NodeID {
				case 2:
					relationships = append(relationships, map[string]any{
						"relation": "calls", "node": nodes["Worker"],
					})
				case 3:
					relationships = append(relationships, map[string]any{
						"relation": "calls", "node": nodes["Coordinator"],
					})
				}
				responses[request.ID] = successfulResponse(t, request.ID, map[string]any{
					"node":          nodes[nodeNameForID(*request.NodeID)],
					"direction":     request.Direction,
					"relationships": relationships,
				})
			case "unresolved":
				responses[request.ID] = successfulResponse(t, request.ID, map[string]any{
					"truncated": false, "unresolved": []any{},
				})
			default:
				return nil, fmt.Errorf("unexpected operation %q", request.Op)
			}
		}
		return responses, nil
	}}

	scores, err := client.seedGraphProximity(context.Background(), "snapshot", "coordinator worker connection", candidates)
	if err != nil {
		t.Fatal(err)
	}
	if scores[hybridSeedKey(candidates[1].node)] <= scores[hybridSeedKey(candidates[0].node)] {
		t.Fatalf("connected coordinator was not promoted: %+v", scores)
	}
	if scores[hybridSeedKey(candidates[2].node)] <= scores[hybridSeedKey(candidates[0].node)] {
		t.Fatalf("connected worker was not promoted: %+v", scores)
	}
}

func TestRerankSeedsReturnsFallbackWhenGraphLookupFails(t *testing.T) {
	client := Client{Run: func(
		context.Context,
		string,
		string,
		[]protocolRequest,
	) (map[string]protocolResponse, error) {
		return nil, errors.New("graph unavailable")
	}}
	seeds, err := client.RerankSeeds(
		context.Background(), "snapshot", "profile persistence",
		[]structure.Node{
			{Kind: "function", Name: "CreateProfile", Path: "profile.go"},
			{Kind: "function", Name: "PersistSession", Path: "session.go"},
		},
		nil, 1,
	)
	if err == nil || len(seeds) != 1 || seeds[0].Node.Name != "CreateProfile" {
		t.Fatalf("fallback ranking was not retained: seeds=%+v err=%v", seeds, err)
	}
}

func nodeNameForID(nodeID uint32) string {
	switch nodeID {
	case 1:
		return "Isolated"
	case 2:
		return "Coordinator"
	case 3:
		return "Worker"
	default:
		return ""
	}
}
