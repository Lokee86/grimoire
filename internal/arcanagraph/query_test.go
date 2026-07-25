package arcanagraph

import (
	"context"
	"encoding/json"
	"slices"
	"testing"
)

func TestPathsPreserveInterstackEndpointRelations(t *testing.T) {
	sender := graphFixtureNode(1, "sender", "method", "client/login.gd", "SubmitLogin")
	endpoint := graphFixtureNode(2, "endpoint", "http-endpoint", "@interstack/http/session", "POST /session/start")
	handler := graphFixtureNode(3, "handler", "method", "server/sessions.rb", "create")
	run := func(
		_ context.Context,
		_ string,
		_ string,
		requests []protocolRequest,
	) (map[string]protocolResponse, error) {
		responses := make(map[string]protocolResponse, len(requests))
		for _, request := range requests {
			if request.Op != "paths" {
				t.Fatalf("unexpected operation %q", request.Op)
			}
			data, err := json.Marshal(map[string]any{
				"truncated": false,
				"paths": []any{map[string]any{
					"depth": 2, "nodes": []any{sender, endpoint, handler},
					"relations": []string{"calls-endpoint", "handled-by"},
				}},
			})
			if err != nil {
				t.Fatal(err)
			}
			responses[request.ID] = protocolResponse{
				Protocol: protocolID, ID: request.ID, OK: true, Result: data,
			}
		}
		return responses, nil
	}

	paths, truncated, err := (Client{Run: run}).Paths(
		context.Background(), "snapshot", 1, 3,
		[]string{"calls-endpoint", "handled-by"}, 3, 5,
	)
	if err != nil {
		t.Fatal(err)
	}
	if truncated || len(paths) != 1 ||
		!slices.Equal(paths[0].Relations, []string{"calls-endpoint", "handled-by"}) ||
		paths[0].Nodes[1].Kind != "http-endpoint" {
		t.Fatalf("interstack Arcana path was not preserved: %+v", paths)
	}
}

func graphFixtureNode(id int, identity, kind, path, name string) map[string]any {
	return map[string]any{
		"node_id": id, "identity": identity, "kind": kind, "path": path, "name": name,
		"span": map[string]any{"path": path, "start_line": 3, "end_line": 3},
	}
}
