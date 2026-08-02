package main

import "testing"

func TestCompilerSemanticsReplaceHeuristicCallEvidence(t *testing.T) {
	records := decodeRecords(t, analyzeFixture(t, runtimeFixture))
	nodes := nodeIndex(records)
	source := nodes["demo.runtime.RuntimeSlice.exercise(int,ExternalReceiver)"]
	target := nodes["demo.runtime.RuntimeSlice.overloaded(int)"]
	if source == nil || target == nil {
		t.Fatal("compiler fixture endpoints are missing")
	}
	for _, record := range records {
		if record["record"] != "edge" || record["relation"] != "calls" || record["source"] != source["id"] || record["target"] != target["id"] {
			continue
		}
		attributes, _ := record["attributes"].(map[string]any)
		if attributes["resolution"] != "javac" {
			t.Fatalf("resolved call attributes = %#v, want javac resolution", attributes)
		}
		return
	}
	t.Fatal("compiler-resolved overload call is missing")
}

func TestCompilerSemanticsHaveNoIdentityDrift(t *testing.T) {
	records := decodeRecords(t, analyzeFixture(t, runtimeFixture))
	for _, record := range records {
		if record["record"] == "unresolved" && record["reason"] == "compiler-identity-mismatch" {
			t.Fatalf("compiler identity drift: %#v", record)
		}
	}
}

func TestCompilerSemanticsEmitSubstantialRuntimeConnectivity(t *testing.T) {
	records := decodeRecords(t, analyzeFixture(t, runtimeFixture))
	compilerEdges := 0
	for _, record := range records {
		if record["record"] != "edge" {
			continue
		}
		attributes, _ := record["attributes"].(map[string]any)
		if attributes["resolution"] == "javac" {
			compilerEdges++
		}
	}
	if compilerEdges < 50 {
		t.Fatalf("compiler semantic edges = %d, want at least 50", compilerEdges)
	}
}
