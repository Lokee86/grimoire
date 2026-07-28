package app

import (
	"strings"
	"testing"
)

func TestAgentToolInputSchemaHasModeSpecificRequirements(t *testing.T) {
	schema := agentToolInputSchema()
	rules, ok := schema["allOf"].([]any)
	if !ok || len(rules) != 4 {
		t.Fatalf("allOf rules = %#v", schema["allOf"])
	}

	assertSchemaRule(t, rules, []string{"search"}, []string{"query"})
	assertSchemaRule(t, rules, []string{"inspect"}, []string{"anchor", "handles"})
	assertSchemaRule(t, rules, []string{"trace", "impact"}, []string{"anchor", "query"})
	assertTargetForbiddenOutsideTrace(t, rules)
}

func TestAgentToolInputSchemaExplainsInspectAndTraceArguments(t *testing.T) {
	properties := agentToolInputSchema()["properties"].(map[string]any)
	target := properties["target"].(map[string]any)
	if description, _ := target["description"].(string); !strings.Contains(description, "trace mode only") {
		t.Fatalf("target description = %q", description)
	}
	handles := properties["handles"].(map[string]any)
	if handles["minItems"] != 1 {
		t.Fatalf("handles minItems = %#v", handles["minItems"])
	}
}

func assertSchemaRule(t *testing.T, rules []any, modes, alternatives []string) {
	t.Helper()
	for _, raw := range rules {
		rule := raw.(map[string]any)
		condition := rule["if"].(map[string]any)["properties"].(map[string]any)["mode"].(map[string]any)
		if !sameModes(condition, modes) {
			continue
		}
		then := rule["then"].(map[string]any)
		if required, ok := then["required"].([]string); ok {
			if len(alternatives) == 1 && len(required) == 1 && required[0] == alternatives[0] {
				return
			}
			t.Fatalf("required fields for %v = %v", modes, required)
		}
		anyOf, ok := then["anyOf"].([]any)
		if !ok || len(anyOf) != len(alternatives) {
			t.Fatalf("alternatives for %v = %#v", modes, then["anyOf"])
		}
		for index, field := range alternatives {
			required := anyOf[index].(map[string]any)["required"].([]string)
			if len(required) != 1 || required[0] != field {
				t.Fatalf("alternative %d for %v = %v", index, modes, required)
			}
		}
		return
	}
	t.Fatalf("schema rule for modes %v not found", modes)
}

func assertTargetForbiddenOutsideTrace(t *testing.T, rules []any) {
	t.Helper()
	expectedModes := []string{"orient", "search", "impact", "inspect"}
	for _, raw := range rules {
		rule := raw.(map[string]any)
		condition := rule["if"].(map[string]any)["properties"].(map[string]any)["mode"].(map[string]any)
		if !sameModes(condition, expectedModes) {
			continue
		}
		not := rule["then"].(map[string]any)["not"].(map[string]any)
		required := not["required"].([]string)
		if len(required) != 1 || required[0] != "target" {
			t.Fatalf("target exclusion = %v", required)
		}
		return
	}
	t.Fatal("target exclusion rule not found")
}

func sameModes(condition map[string]any, expected []string) bool {
	if value, ok := condition["const"].(string); ok {
		return len(expected) == 1 && value == expected[0]
	}
	values, ok := condition["enum"].([]string)
	if !ok || len(values) != len(expected) {
		return false
	}
	for index := range values {
		if values[index] != expected[index] {
			return false
		}
	}
	return true
}
