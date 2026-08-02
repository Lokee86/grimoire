package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func analyzeFixture(t *testing.T, path string) []byte {
	t.Helper()
	data, err := analyzeRepository(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func decodeRecords(t *testing.T, data []byte) []map[string]any {
	t.Helper()
	var records []map[string]any
	decoder := json.NewDecoder(bytes.NewReader(data))
	for decoder.More() {
		var record map[string]any
		if err := decoder.Decode(&record); err != nil {
			t.Fatal(err)
		}
		records = append(records, record)
	}
	return records
}

func nodeIndex(records []map[string]any) map[string]map[string]any {
	result := make(map[string]map[string]any)
	for _, record := range records {
		if record["record"] == "node" {
			result[record["qualified_name"].(string)] = record
		}
	}
	return result
}

func assertDeclaration(t *testing.T, nodes map[string]map[string]any, qualifiedName, kind, declarationKind string) {
	t.Helper()
	node := nodes[qualifiedName]
	if node == nil {
		t.Fatalf("missing declaration %s", qualifiedName)
	}
	attributes, _ := node["attributes"].(map[string]any)
	if node["kind"] != kind || attributes["declaration_kind"] != declarationKind {
		t.Fatalf("declaration %s = %#v", qualifiedName, node)
	}
}

func assertAllEndpoints(t *testing.T, records []map[string]any, nodes map[string]map[string]any) {
	t.Helper()
	ids := make(map[string]bool, len(nodes))
	for _, node := range nodes {
		ids[node["id"].(string)] = true
	}
	for _, record := range records {
		if record["record"] == "edge" && (!ids[record["source"].(string)] || !ids[record["target"].(string)]) {
			t.Fatalf("edge has missing endpoint: %#v", record)
		}
		if record["record"] == "unresolved" && !ids[record["source"].(string)] {
			t.Fatalf("unresolved record has missing source: %#v", record)
		}
	}
}

func assertContains(t *testing.T, records []map[string]any, nodes map[string]map[string]any, sourceName, targetName string) {
	t.Helper()
	source, target := nodes[sourceName], nodes[targetName]
	if source == nil || target == nil {
		t.Fatalf("missing containment endpoint %s -> %s", sourceName, targetName)
	}
	for _, record := range records {
		if record["record"] == "edge" && record["relation"] == "contains" && record["source"] == source["id"] && record["target"] == target["id"] {
			return
		}
	}
	t.Fatalf("missing contains edge %s -> %s", sourceName, targetName)
}

func assertImport(t *testing.T, records []map[string]any, nodes map[string]map[string]any, targetName string) {
	t.Helper()
	target := nodes[targetName]
	if target == nil {
		t.Fatalf("missing import target %s", targetName)
	}
	for _, record := range records {
		if record["record"] == "edge" && record["relation"] == "imports" && record["target"] == target["id"] {
			return
		}
	}
	t.Fatalf("missing resolved import to %s", targetName)
}

func assertImportExpression(t *testing.T, records []map[string]any, nodes map[string]map[string]any, expression, targetName string) {
	t.Helper()
	target := nodes[targetName]
	if target == nil {
		t.Fatalf("missing import target %s", targetName)
	}
	importIDs := make(map[string]bool)
	for _, node := range nodes {
		attributes, _ := node["attributes"].(map[string]any)
		if node["kind"] == "import" && attributes["expression"] == expression {
			importIDs[node["id"].(string)] = true
		}
	}
	for _, record := range records {
		if record["record"] == "edge" && record["relation"] == "imports" && importIDs[record["source"].(string)] && record["target"] == target["id"] {
			return
		}
	}
	t.Fatalf("missing resolved import %q to %s", expression, targetName)
}

func assertUnresolved(t *testing.T, records []map[string]any, relation, expression, reason string) {
	t.Helper()
	for _, record := range records {
		if record["record"] == "unresolved" && record["relation"] == relation && record["expression"] == expression && record["reason"] == reason {
			return
		}
	}
	t.Fatalf("missing unresolved %s %q (%s)", relation, expression, reason)
}

func assertUnresolvedRelation(t *testing.T, records []map[string]any, relation, reason, owner string) {
	t.Helper()
	for _, record := range records {
		if record["record"] == "unresolved" && record["relation"] == relation && record["reason"] == reason && record["owner"] == owner {
			return
		}
	}
	t.Fatalf("missing unresolved %s (%s) owned by %s", relation, reason, owner)
}

func copyTree(t *testing.T, source, destination string) {
	t.Helper()
	if err := filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	}); err != nil {
		t.Fatal(err)
	}
}
