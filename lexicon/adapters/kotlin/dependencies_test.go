package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

const dependencyFixture = "testdata/dependencies"

func TestGradleKotlinDependencyManifestEvidence(t *testing.T) {
	records := dependencyFixtureRecords(t)
	nodes := factRecords(records, "node")
	edges := dependencyEdges(records)
	rootID := stableID("repository", "dependencies")

	for _, expected := range []struct {
		coordinate, configuration string
	}{
		{"org.jetbrains.kotlin:kotlin-stdlib:2.0.21", "implementation"},
		{"com.example:public-api:1.0", "api"},
		{"org.jetbrains:annotations:24.1.0", "compileOnly"},
		{"com.example:runtime:3.0", "runtimeOnly"},
		{"org.junit.jupiter:junit-jupiter:5.11.0", "testImplementation"},
		{"com.google.dagger:dagger-compiler:2.52", "kapt"},
		{"com.google.devtools.ksp:symbol-processing-api:2.0.21-1.0.28", "ksp"},
	} {
		edge := requireDependencyEdge(t, edges, rootID, expected.coordinate, expected.configuration)
		assertManifestEvidence(t, edge, "build.gradle.kts", "gradle-kotlin")
		target := requireNodeByID(t, nodes, edge["target"].(string))
		if target["kind"] != "module" || target["qualified_name"] != expected.coordinate {
			t.Fatalf("coordinate target is not an external module: %#v", target)
		}
		assertAttribute(t, target, "external", true)
		assertAttribute(t, target, "ecosystem", "maven")
	}
	if countDependencyEdges(edges, rootID, "com.example:public-api:1.0") != 2 {
		t.Fatal("duplicate Kotlin DSL declarations were collapsed")
	}
	for _, expression := range []string{
		"libs.kotlin.coroutines", `"com.example:dynamic:$dynamicVersion"`, `project(":shared")`,
		`platform("com.example:bom:1.0")`, `files("libs/local.jar")`,
		`group = "com.example", name = "named", version = "1.0"`,
	} {
		requireDependencyUnresolved(t, records, rootID, "build.gradle.kts", expression)
	}
	if findNode(nodes, "module", "build.gradle", "build.gradle.kts") != nil {
		t.Fatal("Gradle Kotlin manifest was also analyzed as Kotlin source")
	}
}

func TestGradleGroovyDependencyManifestEvidence(t *testing.T) {
	records := dependencyFixtureRecords(t)
	edges := dependencyEdges(records)
	sourceID := stableID("module", "build-module:groovy-module")
	for _, expected := range []struct {
		coordinate, configuration string
	}{
		{"org.codehaus.groovy:groovy:3.0.22", "implementation"},
		{"com.acme:groovy-api:1.2", "api"},
		{"com.acme:runtime:2.0", "runtimeOnly"},
	} {
		edge := requireDependencyEdge(t, edges, sourceID, expected.coordinate, expected.configuration)
		assertManifestEvidence(t, edge, "groovy-module/build.gradle", "gradle-groovy")
	}
	requireDependencyUnresolved(t, records, sourceID, "groovy-module/build.gradle", "project(':test-support')")
}

func TestMavenDependencyManifestEvidence(t *testing.T) {
	records := dependencyFixtureRecords(t)
	nodes := factRecords(records, "node")
	edges := dependencyEdges(records)
	sourceID := stableID("module", "build-module:maven-service")

	compile := requireDependencyEdge(t, edges, sourceID, "org.slf4j:slf4j-api:2.0.16", "compile")
	assertManifestEvidence(t, compile, "maven-service/pom.xml", "maven")
	assertAttribute(t, compile, "optional", false)
	testEdge := requireDependencyEdge(t, edges, sourceID, "org.junit.jupiter:junit-jupiter:5.11.0", "test")
	assertAttribute(t, testEdge, "optional", true)
	requireDependencyEdge(t, edges, sourceID, "com.example:versionless", "")
	if countDependencyEdges(edges, sourceID, "org.slf4j:slf4j-api:2.0.16") != 2 {
		t.Fatal("duplicate Maven declarations were collapsed")
	}
	requireDependencyUnresolved(t, records, sourceID, "maven-service/pom.xml", "property-version")
	requireDependencyUnresolved(t, records, sourceID, "maven-service/pom.xml", "fixture-bom")
	for _, ignored := range []string{"fixture-bom", "managed-bom", "profile-only", "plugin", "ignored-parent"} {
		if findNode(nodes, "module", ignored, "") != nil {
			t.Fatalf("unsupported Maven section emitted dependency %q", ignored)
		}
	}
}

func TestDependencyFixtureExclusionsEndpointsAndByteDeterminism(t *testing.T) {
	first, err := analyzeRepository(filepath.FromSlash(dependencyFixture))
	if err != nil {
		t.Fatal(err)
	}
	second, err := analyzeRepository(filepath.FromSlash(dependencyFixture))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("dependency analysis changed byte output")
	}
	records := decodeFacts(t, first)
	nodes := factRecords(records, "node")
	edges := factRecords(records, "edge")
	assertEdgeEndpoints(t, nodes, edges)
	repository := requireNode(t, nodes, "repository", "dependencies", ".")
	assertAttribute(t, repository, "source_file_count", float64(1))
	assertAttribute(t, repository, "dependency_manifest_count", float64(3))
	requireNode(t, nodes, "type", "App", "src/main/kotlin/fixture/App.kt")
	if len(dependencyEdges(records)) != 15 {
		t.Fatalf("depends-on edges = %d, want 15", len(dependencyEdges(records)))
	}
	for _, ignored := range []string{"ignored:gradle-cache:1.0", "ignored:build-output:1", "ignored:vendor:1.0"} {
		if findNode(nodes, "module", "", ignored) != nil || hasQualifiedNode(nodes, ignored) {
			t.Fatalf("excluded manifest emitted %q", ignored)
		}
	}
	for _, node := range nodes {
		path, _ := node["path"].(string)
		if strings.Contains(path, "\\") || filepath.IsAbs(path) {
			t.Fatalf("invalid dependency fixture node path: %q", path)
		}
	}
}

func dependencyFixtureRecords(t *testing.T) []map[string]any {
	t.Helper()
	data, err := analyzeRepository(filepath.FromSlash(dependencyFixture))
	if err != nil {
		t.Fatal(err)
	}
	return decodeFacts(t, data)
}

func dependencyEdges(records []map[string]any) []map[string]any {
	var edges []map[string]any
	for _, edge := range factRecords(records, "edge") {
		if edge["relation"] == "depends-on" {
			edges = append(edges, edge)
		}
	}
	return edges
}

func requireDependencyEdge(t *testing.T, edges []map[string]any, source, coordinate, configuration string) map[string]any {
	t.Helper()
	for _, edge := range edges {
		attrs := attributes(edge)
		if edge["source"] == source && attrs["coordinate"] == coordinate && attrs["configuration"] == configuration {
			return edge
		}
	}
	t.Fatalf("missing %s dependency %q from %s", configuration, coordinate, source)
	return nil
}

func countDependencyEdges(edges []map[string]any, source, coordinate string) int {
	count := 0
	for _, edge := range edges {
		if edge["source"] == source && attributes(edge)["coordinate"] == coordinate {
			count++
		}
	}
	return count
}

func requireDependencyUnresolved(t *testing.T, records []map[string]any, source, manifest, expressionPart string) {
	t.Helper()
	for _, record := range factRecords(records, "unresolved") {
		if record["source"] == source && record["relation"] == "depends-on" &&
			attributes(record)["manifest"] == manifest && strings.Contains(record["expression"].(string), expressionPart) {
			if record["reason"] != "unsupported-form" {
				t.Fatalf("unexpected dependency unresolved reason: %#v", record)
			}
			assertManifestSpan(t, record, manifest)
			return
		}
	}
	t.Fatalf("missing unresolved dependency %q in %s", expressionPart, manifest)
}

func assertManifestEvidence(t *testing.T, record map[string]any, manifest, manifestType string) {
	t.Helper()
	assertAttribute(t, record, "manifest", manifest)
	assertAttribute(t, record, "manifest_type", manifestType)
	assertAttribute(t, record, "evidence", "dependency-manifest")
	assertManifestSpan(t, record, manifest)
}

func assertManifestSpan(t *testing.T, record map[string]any, manifest string) {
	t.Helper()
	span, _ := record["span"].(map[string]any)
	if span["path"] != manifest || span["start_line"].(float64) < 1 || span["end_line"].(float64) < span["start_line"].(float64) {
		t.Fatalf("invalid manifest span: %#v", record)
	}
}

func requireNodeByID(t *testing.T, nodes []map[string]any, id string) map[string]any {
	t.Helper()
	for _, node := range nodes {
		if node["id"] == id {
			return node
		}
	}
	t.Fatalf("missing node ID %s", id)
	return nil
}

func hasQualifiedNode(nodes []map[string]any, qualifiedName string) bool {
	for _, node := range nodes {
		if node["qualified_name"] == qualifiedName {
			return true
		}
	}
	return false
}
