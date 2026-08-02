package main

import (
	"bytes"
	"path/filepath"
	"reflect"
	"testing"
)

func TestRelationshipFixtureResolvesRepositoryLocalEvidence(t *testing.T) {
	first, err := analyzeRepository(filepath.FromSlash("testdata/relationships"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := analyzeRepository(filepath.FromSlash("testdata/relationships"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("relationship analysis changed byte output")
	}

	records := decodeFacts(t, first)
	nodes := factRecords(records, "node")
	edges := factRecords(records, "edge")
	assertEdgeEndpoints(t, nodes, edges)

	base := requireQualifiedNode(t, nodes, "relationships.contracts.Base")
	contract := requireQualifiedNode(t, nodes, "relationships.contracts.Contract")
	marker := requireQualifiedNode(t, nodes, "relationships.contracts.Marker")
	aliasChild := requireQualifiedNode(t, nodes, "relationships.app.AliasChild")
	assertRelationshipEdge(t, edges, aliasChild, base, "extends")
	assertRelationshipEdge(t, edges, aliasChild, contract, "implements")
	assertRelationshipEdge(t, edges, aliasChild, marker, "annotates")
	assertRelationshipEdge(t, edges,
		requireQualifiedNode(t, nodes, "relationships.app.DirectChild"),
		requireQualifiedNode(t, nodes, "relationships.contracts.DirectContract"), "implements")
	assertRelationshipEdge(t, edges,
		requireQualifiedNode(t, nodes, "relationships.contracts.ChildContract"), contract, "implements")

	assertRelationshipEdge(t, edges,
		requireQualifiedNode(t, nodes, "relationships.app.ExactChild"), base, "extends")
	assertRelationshipEdge(t, edges,
		requireQualifiedNode(t, nodes, "relationships.app.LocalChild"),
		requireQualifiedNode(t, nodes, "relationships.app.LocalBase"), "extends")
	assertRelationshipEdge(t, edges,
		requireQualifiedNode(t, nodes, "relationships.app.LocalChild"),
		requireQualifiedNode(t, nodes, "relationships.app.LocalMarker"), "annotates")

	wildChild := requireQualifiedNode(t, nodes, "relationships.app.WildChild")
	assertRelationshipEdge(t, edges, wildChild,
		requireQualifiedNode(t, nodes, "relationships.wild.WildContract"), "implements")
	assertRelationshipEdge(t, edges, wildChild,
		requireQualifiedNode(t, nodes, "relationships.wild.WildMarker"), "annotates")
	assertRelationshipEdge(t, edges,
		requireQualifiedNode(t, nodes, "relationships.app.NestedAliasChild"),
		requireQualifiedNode(t, nodes, "relationships.contracts.Outer.NestedContract"), "implements")

	innerChild := requireQualifiedNode(t, nodes, "relationships.app.Lexical.InnerChild")
	assertRelationshipEdge(t, edges, innerChild,
		requireQualifiedNode(t, nodes, "relationships.app.Lexical.InnerContract"), "implements")
	assertRelationshipEdge(t, edges, innerChild,
		requireQualifiedNode(t, nodes, "relationships.app.Lexical.InnerMarker"), "annotates")

	delegating := requireQualifiedNode(t, nodes, "relationships.app.Delegating")
	delegation := requireRelationshipEdge(t, edges, delegating, contract, "implements")
	assertRelationshipAttributes(t, delegation, map[string]any{
		"delegate_expression": "delegate", "delegated": true,
	})
}

func TestRelationshipFixturePreservesAmbiguousAndExternalEvidence(t *testing.T) {
	data, err := analyzeRepository(filepath.FromSlash("testdata/relationships"))
	if err != nil {
		t.Fatal(err)
	}
	records := decodeFacts(t, data)
	nodes := factRecords(records, "node")
	unresolved := factRecords(records, "unresolved")

	ambiguousChild := requireQualifiedNode(t, nodes, "relationships.app.AmbiguousChild")
	requireRelationshipUnresolved(t, unresolved, ambiguousChild, "extends", "Shared", "ambiguous-target")
	ambiguousAnnotated := requireQualifiedNode(t, nodes, "relationships.app.AmbiguousAnnotated")
	requireRelationshipUnresolved(t, unresolved, ambiguousAnnotated, "annotates", "SharedMarker", "ambiguous-target")

	externalChild := requireQualifiedNode(t, nodes, "relationships.app.ExternalChild")
	requireRelationshipUnresolved(t, unresolved, externalChild, "extends", "external.Base()", "external-target")
	delegation := requireRelationshipUnresolved(t, unresolved, externalChild, "extends", "ExternalContract by externalDelegate", "external-target")
	assertRelationshipAttributes(t, delegation, map[string]any{
		"delegate_expression": "externalDelegate", "delegated": true,
	})
	externalAnnotated := requireQualifiedNode(t, nodes, "relationships.app.ExternalAnnotated")
	requireRelationshipUnresolved(t, unresolved, externalAnnotated, "annotates", "ExternalMarker", "external-target")
}

func requireQualifiedNode(t *testing.T, nodes []map[string]any, qualifiedName string) map[string]any {
	t.Helper()
	for _, node := range nodes {
		if node["qualified_name"] == qualifiedName && (node["kind"] == "type" || node["kind"] == "interface") {
			return node
		}
	}
	t.Fatalf("missing node %s", qualifiedName)
	return nil
}

func assertRelationshipEdge(t *testing.T, edges []map[string]any, source, target map[string]any, relation string) {
	t.Helper()
	requireRelationshipEdge(t, edges, source, target, relation)
}

func requireRelationshipEdge(t *testing.T, edges []map[string]any, source, target map[string]any, relation string) map[string]any {
	t.Helper()
	for _, edge := range edges {
		if edge["source"] == source["id"] && edge["target"] == target["id"] && edge["relation"] == relation {
			return edge
		}
	}
	t.Fatalf("missing %s edge from %s to %s", relation, source["qualified_name"], target["qualified_name"])
	return nil
}

func requireRelationshipUnresolved(t *testing.T, records []map[string]any, source map[string]any, relation, expression, reason string) map[string]any {
	t.Helper()
	for _, record := range records {
		if record["source"] == source["id"] && record["relation"] == relation && record["expression"] == expression && record["reason"] == reason {
			return record
		}
	}
	t.Fatalf("missing unresolved %s %q (%s) for %s", relation, expression, reason, source["qualified_name"])
	return nil
}

func assertRelationshipAttributes(t *testing.T, record map[string]any, expected map[string]any) {
	t.Helper()
	if actual := attributes(record); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("attributes = %#v, want %#v on %#v", actual, expected, record)
	}
}
