package main

import (
	"fmt"
	"strings"
	"testing"
)

func analyzeFixtureRecords(t *testing.T, root string) []map[string]any {
	t.Helper()
	data, err := analyzeRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	return decodeRecords(t, data)
}

func assertRelationAtPaths(t *testing.T, records []map[string]any, relation, sourcePath, sourceName, targetPath, targetName string) {
	t.Helper()
	source := findNodeAtPath(records, sourcePath, sourceName)
	target := findNodeAtPath(records, targetPath, targetName)
	if source == nil || target == nil {
		t.Fatalf("missing relation endpoint for %s: %s::%s -> %s::%s", relation, sourcePath, sourceName, targetPath, targetName)
	}
	for _, record := range records {
		if record["record"] == "edge" && record["relation"] == relation && record["source"] == source["id"] && record["target"] == target["id"] {
			return
		}
	}
	t.Fatalf("missing %s edge: %s::%s -> %s::%s", relation, sourcePath, sourceName, targetPath, targetName)
}

func assertNoRelationAtPaths(t *testing.T, records []map[string]any, relation, sourcePath, sourceName, targetPath, targetName string) {
	t.Helper()
	source := findNodeAtPath(records, sourcePath, sourceName)
	target := findNodeAtPath(records, targetPath, targetName)
	if source == nil || target == nil {
		return
	}
	for _, record := range records {
		if record["record"] == "edge" && record["relation"] == relation && record["source"] == source["id"] && record["target"] == target["id"] {
			t.Fatalf("unexpected %s edge: %s::%s -> %s::%s", relation, sourcePath, sourceName, targetPath, targetName)
		}
	}
}

func findNodeAtPath(records []map[string]any, path, name string) map[string]any {
	for _, record := range records {
		if record["record"] == "node" && record["path"] == path && strings.EqualFold(fmt.Sprint(record["name"]), name) {
			return record
		}
	}
	return nil
}
