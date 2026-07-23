package arcanagraph

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	sharedevidence "github.com/Lokee86/grimoire/internal/evidence"
	"github.com/Lokee86/grimoire/internal/structure"
)

func TestClientReturnsOperationalImpactUnresolvedAndChainEvidence(t *testing.T) {
	alpha := map[string]any{
		"node_id": 1, "identity": "alpha-id", "kind": "function",
		"path": "internal/alpha.go", "name": "Alpha",
		"span": map[string]any{"path": "internal/alpha.go", "start_line": 10, "end_line": 20},
	}
	beta := map[string]any{
		"node_id": 2, "identity": "beta-id", "kind": "function",
		"path": "internal/beta.go", "name": "Beta",
		"span": map[string]any{"path": "internal/beta.go", "start_line": 30, "end_line": 40},
	}
	calls := 0
	run := func(
		_ context.Context,
		command string,
		snapshot string,
		requests []protocolRequest,
	) (map[string]protocolResponse, error) {
		calls++
		if command != "arcana-test" || snapshot != "snapshot" {
			t.Fatalf("unexpected invocation command=%q snapshot=%q", command, snapshot)
		}
		responses := make(map[string]protocolResponse, len(requests))
		for _, request := range requests {
			switch request.Op {
			case "resolve_symbol":
				node := alpha
				if request.Name == "Beta" {
					node = beta
				}
				responses[request.ID] = successfulResponse(t, request.ID, map[string]any{
					"count": 1, "returned": 1, "truncated": false, "nodes": []any{node},
				})
			case "operational_role":
				node, callers, callees := alpha, []any{}, []any{map[string]any{"relation": "calls", "node": beta}}
				summary := "Alpha has one callee."
				if request.NodeID != nil && *request.NodeID == 2 {
					node, callers, callees = beta, []any{map[string]any{"relation": "calls", "node": alpha}}, []any{}
					summary = "Beta has one caller."
				}
				responses[request.ID] = successfulResponse(t, request.ID, map[string]any{
					"node": node, "summary": summary, "callers": callers, "callees": callees,
				})
			case "impact":
				dependents := []any{}
				if request.NodeID != nil && *request.NodeID == 2 {
					dependents = []any{map[string]any{"depth": 1, "node": alpha}}
				}
				responses[request.ID] = successfulResponse(t, request.ID, map[string]any{
					"node_id": request.NodeID, "truncated": false, "dependents": dependents,
				})
			case "unresolved":
				items := []any{}
				if request.NodeID != nil && *request.NodeID == 1 {
					items = []any{map[string]any{
						"relation": "possible-calls", "expression": "Unknown()",
						"candidate_name": "Unknown", "reason": "no-candidate",
						"span": map[string]any{"path": "internal/alpha.go", "start_line": 18, "end_line": 18},
					}}
				}
				responses[request.ID] = successfulResponse(t, request.ID, map[string]any{
					"truncated": false, "unresolved": items,
				})
			case "shortest_call_chain":
				found := request.FromNodeID != nil && request.ToNodeID != nil &&
					*request.FromNodeID == 1 && *request.ToNodeID == 2
				var chain any
				if found {
					chain = map[string]any{
						"depth": 1, "nodes": []any{alpha, beta}, "relations": []string{"calls"},
					}
				}
				responses[request.ID] = successfulResponse(t, request.ID, map[string]any{
					"found": found, "chain": chain,
				})
			default:
				t.Fatalf("unexpected Arcana operation %q", request.Op)
			}
		}
		return responses, nil
	}

	evidence, err := (Client{Command: "arcana-test", Run: run}).Search(
		context.Background(), "snapshot",
		[]structure.Node{
			{Name: "Alpha", Path: "internal/alpha.go", Span: &structure.Span{Path: "internal/alpha.go", StartLine: 10, EndLine: 20}},
			{Name: "Beta", Path: "internal/beta.go", Span: &structure.Span{Path: "internal/beta.go", StartLine: 30, EndLine: 40}},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("expected resolve and detail protocol batches, got %d", calls)
	}
	kinds := make(map[string]int)
	for index, item := range evidence {
		if item.Provider != "arcana" || item.Rank != index+1 {
			t.Fatalf("invalid Arcana provenance: %+v", item)
		}
		kinds[item.Kind]++
		if item.Context == nil || !slices.Contains(item.Context.Roles, sharedevidence.RoleStructural) {
			t.Fatalf("missing structural descriptor: %+v", item)
		}
	}
	for _, kind := range []string{"operational_role", "impact", "unresolved", "call_chain"} {
		if kinds[kind] == 0 {
			t.Fatalf("missing %s evidence: %+v", kind, evidence)
		}
	}
	var chain structure.Evidence
	for _, item := range evidence {
		if item.Kind == "call_chain" {
			chain = item
			break
		}
	}
	if chain.Context == nil || len(chain.Context.GroupIDs) < 2 || len(chain.Context.Links) != 2 {
		t.Fatalf("call-chain descriptor did not retain endpoint groups and links: %+v", chain.Context)
	}
	chainGroup := sharedevidence.StableID(
		"call-chain",
		sharedevidence.RangeIdentity("internal/alpha.go", 10, 20),
		sharedevidence.RangeIdentity("internal/beta.go", 30, 40),
	)
	if !slices.Contains(chain.Context.GroupIDs, chainGroup) {
		t.Fatalf("call-chain group lost ordered chain identity: %+v", chain.Context)
	}
	reversed := sharedevidence.StableID(
		"call-chain",
		sharedevidence.RangeIdentity("internal/beta.go", 30, 40),
		sharedevidence.RangeIdentity("internal/alpha.go", 10, 20),
	)
	if chainGroup == reversed {
		t.Fatal("call-chain group ignored node order")
	}
}

func successfulResponse(t *testing.T, id string, result any) protocolResponse {
	t.Helper()
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	return protocolResponse{Protocol: protocolID, ID: id, OK: true, Result: data}
}
