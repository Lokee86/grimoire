package main

import "testing"

func TestGDScriptLocalTargetUsesNearestPriorDeclaration(t *testing.T) {
	function := &declaration{
		kind:           "function",
		name:           "run",
		nodeID:         "function-id",
		key:            "fixture.gd::function::run",
		parameterNames: []string{"value"},
	}
	first := &declaration{
		kind:          "variable",
		name:          "value",
		nodeID:        "first-id",
		ownerFunction: function.nodeID,
		span: sourceSpan{
			"start_line": 3, "start_column": 5,
			"end_line": 3, "end_column": 18,
		},
	}
	second := &declaration{
		kind:          "variable",
		name:          "value",
		nodeID:        "second-id",
		ownerFunction: function.nodeID,
		span: sourceSpan{
			"start_line": 7, "start_column": 5,
			"end_line": 7, "end_column": 18,
		},
	}
	facts := &factSet{declarationByID: map[string]*declaration{
		second.nodeID:   second,
		function.nodeID: function,
		first.nodeID:    first,
	}}

	if got := gdscriptLocalTarget(facts, function.nodeID, "value", token{line: 2, column: 9}); got != nodeID("parameter", function.key+"::parameter::value") {
		t.Fatalf("before locals target = %q, want parameter", got)
	}
	if got := gdscriptLocalTarget(facts, function.nodeID, "value", token{line: 5, column: 9}); got != first.nodeID {
		t.Fatalf("between locals target = %q, want %q", got, first.nodeID)
	}
	if got := gdscriptLocalTarget(facts, function.nodeID, "value", token{line: 9, column: 9}); got != second.nodeID {
		t.Fatalf("after locals target = %q, want %q", got, second.nodeID)
	}
	if got := gdscriptLocalTarget(facts, function.nodeID, "value", token{line: 7, column: 15}); got != first.nodeID {
		t.Fatalf("initializer target = %q, want prior declaration %q", got, first.nodeID)
	}
}

func TestGDScriptMemberTargetRejectsAmbiguousOwners(t *testing.T) {
	first := &declaration{kind: "variable", name: "session", nodeID: "first-session", ownerID: "first-owner"}
	second := &declaration{kind: "variable", name: "session", nodeID: "second-session", ownerID: "second-owner"}
	model := &semanticModel{
		facts: &factSet{declarationByID: map[string]*declaration{
			first.nodeID:  first,
			second.nodeID: second,
		}},
		locals: map[string]map[string]ownerSet{
			"function-id": {
				"controller": {"first-owner": {}, "second-owner": {}},
			},
		},
	}
	context := analysisContext{file: &parsedFile{path: "fixture.gd"}, functionID: "function-id"}
	receiver := []token{{kind: tokenIdentifier, text: "controller", line: 1, column: 1}}

	if got := gdscriptMemberTargetID(model, context, receiver, "session"); got != "" {
		t.Fatalf("ambiguous member target = %q, want no definite target", got)
	}
}
