package main

import "testing"

func TestCompilerFactsReplaceOnlyIndexedRuntimeEvidenceAtTheSameSite(t *testing.T) {
	facts := newFactSet()
	sourceIdentity := "demo.Source#run()"
	targetIdentity := "demo.Target#call()"
	otherIdentity := "demo.Other#call()"
	source := facts.addNode("method", "run", "Source.java", sourceIdentity, sourceIdentity, "", nil, nil, "")
	target := facts.addNode("method", "call", "Target.java", targetIdentity, targetIdentity, "", nil, nil, "")
	other := facts.addNode("method", "call", "Other.java", otherIdentity, otherIdentity, "", nil, nil, "")
	exact := &span{Path: "Source.java", StartLine: 10, StartColumn: 5, EndLine: 10, EndColumn: 12}
	later := &span{Path: "Source.java", StartLine: 11, StartColumn: 5, EndLine: 11, EndColumn: 12}
	facts.addEdge(source, target, "calls", "Source.java", exact, map[string]any{"resolution": "heuristic"})
	facts.addEdge(source, other, "possible-calls", "Source.java", exact, nil)
	facts.addUnresolved(source, "calls", "call()", "ambiguous-target", "Source.java", exact, nil)
	facts.addEdge(source, other, "calls", "Source.java", later, nil)

	evidence := buildCompilerEvidenceIndex(facts)
	state := analysisState{facts: facts}
	if !state.applyCompilerFact(compilerFact{
		Record: "edge", SourceKind: "method", SourceIdentity: sourceIdentity,
		TargetKind: "method", TargetIdentity: targetIdentity, Relation: "calls",
		Path: "Source.java", StartLine: 10, StartColumn: 5, EndLine: 10, EndColumn: 12,
		Engine: "javac",
	}, &evidence) {
		t.Fatal("compiler fact was not applied")
	}
	if len(facts.edges) != 2 {
		t.Fatalf("edge count = %d, want compiler edge plus unrelated later edge", len(facts.edges))
	}
	if len(facts.unresolved) != 0 {
		t.Fatalf("unresolved count = %d, want exact-site runtime evidence removed", len(facts.unresolved))
	}
}

func TestCompilerSuppressionRemovesNearbyIndexedRuntimeEvidence(t *testing.T) {
	facts := newFactSet()
	sourceIdentity := "demo.Source#read()"
	targetIdentity := "demo.Source#field"
	source := facts.addNode("method", "read", "Source.java", sourceIdentity, sourceIdentity, "", nil, nil, "")
	target := facts.addNode("field", "field", "Source.java", targetIdentity, targetIdentity, "", nil, nil, "")
	for _, line := range []int{9, 10, 11, 12} {
		facts.addEdge(source, target, "reads", "Source.java", &span{
			Path: "Source.java", StartLine: line, StartColumn: 1, EndLine: line, EndColumn: 6,
		}, nil)
	}

	evidence := buildCompilerEvidenceIndex(facts)
	state := analysisState{facts: facts}
	if !state.applyCompilerFact(compilerFact{
		Record: "suppression", SourceKind: "method", SourceIdentity: sourceIdentity,
		Relation: "reads", Path: "Source.java", StartLine: 10, StartColumn: 1,
	}, &evidence) {
		t.Fatal("compiler suppression was not applied")
	}
	if len(facts.edges) != 1 {
		t.Fatalf("edge count = %d, want only evidence outside the suppression window", len(facts.edges))
	}
}
