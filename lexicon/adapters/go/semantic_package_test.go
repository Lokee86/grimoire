package main

import "testing"

func TestPackageNodeForNamespacePrefersCanonicalPackage(t *testing.T) {
	scanner := &scanner{
		packages: map[string]packageInfo{
			"external-test": {key: NodeKey("external-test"), importKey: "example.com/project/tooling", name: "tooling_test"},
			"canonical":     {key: NodeKey("canonical"), importKey: "example.com/project/tooling", name: "tooling"},
		},
	}
	key, ok := scanner.packageNodeForNamespace("example.com/project/tooling")
	if !ok || key != NodeKey("canonical") {
		t.Fatalf("package key = %q, ok = %v", key, ok)
	}
}
