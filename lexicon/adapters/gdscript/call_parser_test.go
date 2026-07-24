package main

import "testing"

func TestFindCallsSkipsUnterminatedCall(t *testing.T) {
	tokens, err := lex("var value = factory(")
	if err != nil {
		t.Fatalf("lex: %v", err)
	}
	statements := makeStatements(tokens)
	if len(statements) != 1 {
		t.Fatalf("statements = %d, want 1", len(statements))
	}
	if calls := findCalls(statements[0], "sample.gd"); len(calls) != 0 {
		t.Fatalf("calls = %v, want none", calls)
	}
}
