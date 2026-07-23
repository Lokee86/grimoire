package arcanagraph

import (
	"strings"
	"testing"
)

func TestValidateResponsesRejectsProviderOperationFailure(t *testing.T) {
	requests := []protocolRequest{{ID: "impact-0", Op: "impact"}}
	responses := map[string]protocolResponse{
		"impact-0": {
			Protocol: protocolID,
			ID:       "impact-0",
			OK:       false,
			Error: &protocolError{
				Code:    "invalid_node",
				Message: "node 42 does not exist",
			},
		},
	}

	err := validateResponses(requests, responses)
	if err == nil || !strings.Contains(err.Error(), "invalid_node") {
		t.Fatalf("expected Arcana operation error, got %v", err)
	}
}

func TestValidateResponsesRejectsMissingResponse(t *testing.T) {
	err := validateResponses(
		[]protocolRequest{{ID: "role-0", Op: "operational_role"}},
		map[string]protocolResponse{},
	)
	if err == nil || !strings.Contains(err.Error(), "returned 0 responses") {
		t.Fatalf("expected missing-response error, got %v", err)
	}
}
