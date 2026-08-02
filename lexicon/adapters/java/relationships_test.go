package main

import (
	"bytes"
	"path/filepath"
	"testing"
)

const relationshipsFixture = "testdata/relationships"

func TestRelationshipsResolveRepositoryLocalTargets(t *testing.T) {
	records := decodeRecords(t, analyzeFixture(t, relationshipsFixture))
	nodes := nodeIndex(records)
	assertAllEndpoints(t, records, nodes)

	assertNamedEdge(t, records, nodes, "demo.app.Positive", "demo.base.ImportedBase", "extends", "clause", "extends")
	assertNamedEdge(t, records, nodes, "demo.app.Positive", "demo.api.Contract", "implements", "clause", "implements")
	assertNamedEdge(t, records, nodes, "demo.app.SameChild", "demo.app.SameBase", "extends", "clause", "extends")
	assertNamedEdge(t, records, nodes, "demo.app.QualifiedChild", "demo.base.QualifiedBase", "extends", "clause", "extends")
	assertNamedEdge(t, records, nodes, "demo.app.Positive.WildChild", "demo.wild.WildBase", "extends", "clause", "extends")
	assertNamedEdge(t, records, nodes, "demo.nested.Outer.Child", "demo.nested.Outer.NestedBase", "extends", "clause", "extends")
	assertNamedEdge(t, records, nodes, "demo.api.Combined", "demo.api.ParentOne", "extends", "clause", "extends")
	assertNamedEdge(t, records, nodes, "demo.api.Combined", "demo.api.ParentTwo", "extends", "clause", "extends")
	assertNamedEdge(t, records, nodes, "demo.sealed.Circle", "demo.sealed.Shape", "implements", "clause", "implements")
	assertNamedEdge(t, records, nodes, "demo.sealed.Shape", "demo.sealed.Circle", "references", "role", "permitted-subtype")

	for _, source := range []string{
		"demo.app.Positive",
		"demo.app.Positive.<init>()",
		"demo.app.Positive.value",
		"demo.app.Positive.apply(String)",
		"demo.app.Positive.apply(String)#parameter:0:input",
		"demo.app.AnnotatedRecord.value",
	} {
		assertNamedEdge(t, records, nodes, source, "demo.api.ImportedAnnotation", "annotates", "expression", "@ImportedAnnotation")
	}
	assertNamedEdge(t, records, nodes, "demo.app.Positive.WildAnnotated", "demo.wild.WildAnnotation", "annotates", "expression", "@WildAnnotation")
	assertNamedEdge(t, records, nodes, "demo.nested.Outer.Child", "demo.nested.Outer.NestedAnnotation", "annotates", "expression", "@NestedAnnotation")
}

func TestRelationshipsPreserveAmbiguousAndExternalTargets(t *testing.T) {
	records := decodeRecords(t, analyzeFixture(t, relationshipsFixture))
	nodes := nodeIndex(records)

	assertNamedUnresolved(t, records, nodes, "demo.ambiguous.Ambiguous", "extends", "Twin", "ambiguous-target", 2)
	assertNamedUnresolved(t, records, nodes, "demo.ambiguous.Ambiguous", "annotates", "@Tag", "ambiguous-target", 2)
	assertNamedUnresolved(t, records, nodes, "demo.external.External", "extends", "vendor.ExternalBase", "external-target", 0)
	assertNamedUnresolved(t, records, nodes, "demo.external.External", "implements", "vendor.ExternalContract", "external-target", 0)
	assertNamedUnresolved(t, records, nodes, "demo.external.External", "annotates", "@vendor.ExternalAnnotation", "external-target", 0)
	assertNamedUnresolved(t, records, nodes, "demo.sealed.Shape", "references", "missing.ExternalShape", "external-target", 0)
}

func TestRelationshipsAreCanonicalDeterministicAndCheckoutIndependent(t *testing.T) {
	first := analyzeFixture(t, relationshipsFixture)
	second := analyzeFixture(t, relationshipsFixture)
	if !bytes.Equal(first, second) {
		t.Fatal("repeat relationship analysis was not byte-identical")
	}
	assertCanonicalJSONL(t, first)

	temporary := t.TempDir()
	left := filepath.Join(temporary, "left", "relationships")
	right := filepath.Join(temporary, "right", "relationships")
	copyTree(t, relationshipsFixture, left)
	copyTree(t, relationshipsFixture, right)
	if leftData, rightData := analyzeFixture(t, left), analyzeFixture(t, right); !bytes.Equal(leftData, rightData) {
		t.Fatal("absolute checkout path affected relationship output")
	}
}

func assertNamedEdge(t *testing.T, records []map[string]any, nodes map[string]map[string]any, sourceName, targetName, relation, attribute, value string) {
	t.Helper()
	source, target := nodes[sourceName], nodes[targetName]
	if source == nil || target == nil {
		t.Fatalf("missing relationship endpoint %s -> %s", sourceName, targetName)
	}
	for _, record := range records {
		attributes, _ := record["attributes"].(map[string]any)
		if record["record"] == "edge" && record["source"] == source["id"] && record["target"] == target["id"] && record["relation"] == relation && attributes[attribute] == value {
			return
		}
	}
	t.Fatalf("missing %s edge %s -> %s with %s=%s", relation, sourceName, targetName, attribute, value)
}

func assertNamedUnresolved(t *testing.T, records []map[string]any, nodes map[string]map[string]any, sourceName, relation, expression, reason string, candidateCount int) {
	t.Helper()
	source := nodes[sourceName]
	if source == nil {
		t.Fatalf("missing unresolved source %s", sourceName)
	}
	for _, record := range records {
		if record["record"] != "unresolved" || record["source"] != source["id"] || record["relation"] != relation || record["expression"] != expression || record["reason"] != reason {
			continue
		}
		attributes, _ := record["attributes"].(map[string]any)
		if candidateCount == 0 || attributes["candidate_count"] == float64(candidateCount) {
			return
		}
	}
	t.Fatalf("missing unresolved %s %s %q (%s)", sourceName, relation, expression, reason)
}
