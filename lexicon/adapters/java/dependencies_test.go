package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

const dependenciesFixture = "testdata/dependencies"

func TestMavenDependenciesAreLiteralScopedEvidence(t *testing.T) {
	records := decodeRecords(t, analyzeFixture(t, dependenciesFixture))
	nodes := nodeIndex(records)
	assertAllEndpoints(t, records, nodes)
	assertDependencyNode(t, nodes, "org.slf4j:slf4j-api")
	assertDependencyEdge(t, records, nodes, "org.slf4j:slf4j-api", "pom.xml", "compile", "2.0.13", "runtime", 2)
	assertDependencyEdge(t, records, nodes, "junit:junit", "pom.xml", "test", "4.13.2", "test", 1)
	optional := assertDependencyEdge(t, records, nodes, "org.example:optional-tool", "pom.xml", "compile", "", "optional", 1)
	if attributes(optional)["optional"] != true {
		t.Fatalf("optional Maven dependency attributes = %#v", attributes(optional))
	}
	assertDependencySpan(t, optional, "pom.xml")

	assertUnsupportedDependency(t, records, "pom.xml", "${revision}", "compile")
	for _, forbidden := range []string{"ignored:managed-bom", "ignored:profile-only", "ignored:plugin-helper"} {
		if nodes["dependency:maven:"+forbidden] != nil || recordsContain(records, forbidden) {
			t.Errorf("Maven non-project dependency was emitted: %s", forbidden)
		}
	}
}

func TestGradleGroovyAndKotlinDependenciesPreserveUnsupportedForms(t *testing.T) {
	records := decodeRecords(t, analyzeFixture(t, dependenciesFixture))
	nodes := nodeIndex(records)
	checks := []struct {
		coordinate, manifest, configuration, constraint, category string
		count                                                     int
	}{
		{"com.google.guava:guava", "groovy/build.gradle", "implementation", "33.2.1-jre", "runtime", 2},
		{"org.apache.commons:commons-lang3", "groovy/build.gradle", "implementation", "3.14.0", "runtime", 1},
		{"com.squareup.okio:okio", "groovy/build.gradle", "api", "3.9.0", "runtime", 1},
		{"org.jetbrains:annotations", "groovy/build.gradle", "compileOnly", "24.1.0", "build", 1},
		{"ch.qos.logback:logback-classic", "groovy/build.gradle", "runtimeOnly", "1.5.6", "runtime", 1},
		{"org.junit.jupiter:junit-jupiter", "groovy/build.gradle", "testImplementation", "5.10.2", "test", 1},
		{"io.ktor:ktor-server-core", "kotlin/build.gradle.kts", "implementation", "2.3.11", "runtime", 1},
		{"io.ktor:ktor-server-test-host", "kotlin/build.gradle.kts", "api", "2.3.11", "runtime", 1},
		{"javax.servlet:javax.servlet-api", "kotlin/build.gradle.kts", "compileOnly", "4.0.1", "build", 1},
		{"org.postgresql:postgresql", "kotlin/build.gradle.kts", "runtimeOnly", "42.7.3", "runtime", 1},
		{"org.assertj:assertj-core", "kotlin/build.gradle.kts", "testImplementation", "3.26.0", "test", 1},
		{"com.google.dagger:dagger-compiler", "groovy/build.gradle", "kapt", "2.51.1", "build", 1},
		{"com.google.dagger:dagger-compiler", "kotlin/build.gradle.kts", "kapt", "2.51.1", "build", 1},
	}
	for _, check := range checks {
		assertDependencyNode(t, nodes, check.coordinate)
		assertDependencyEdge(t, records, nodes, check.coordinate, check.manifest, check.configuration, check.constraint, check.category, check.count)
	}

	for _, unsupported := range []struct{ manifest, expression, configuration string }{
		{"groovy/build.gradle", "libs.guava", "implementation"},
		{"groovy/build.gradle", "$version", "implementation"},
		{"groovy/build.gradle", "project(\":local\")", "implementation"},
		{"groovy/build.gradle", "platform(", "implementation"},
		{"kotlin/build.gradle.kts", "libs.ktor.server.core", "implementation"},
		{"kotlin/build.gradle.kts", "${property", "implementation"},
		{"kotlin/build.gradle.kts", "project(\":shared\")", "implementation"},
	} {
		assertUnsupportedDependency(t, records, unsupported.manifest, unsupported.expression, unsupported.configuration)
	}
	for _, forbidden := range []string{"ignored:commented", "ignored:string", "ignored:commented-kts", "ignored:string-kts"} {
		if recordsContain(records, forbidden) {
			t.Errorf("comment or string content was parsed as a dependency: %s", forbidden)
		}
	}
}

func TestDependencyDiscoveryExclusionsEndpointsAndDeterminism(t *testing.T) {
	first := analyzeFixture(t, dependenciesFixture)
	second := analyzeFixture(t, dependenciesFixture)
	if !bytes.Equal(first, second) {
		t.Fatal("repeat dependency analysis was not byte-identical")
	}
	assertCanonicalJSONL(t, first)
	records := decodeRecords(t, first)
	nodes := nodeIndex(records)
	assertAllEndpoints(t, records, nodes)
	if nodes["demo.App"] == nil {
		t.Fatal("existing Java source facts were not preserved")
	}
	for _, record := range records {
		path, _ := record["path"].(string)
		owner, _ := record["owner"].(string)
		if strings.Contains(path, "vendor/") || strings.Contains(path, "build/build.gradle") || strings.Contains(owner, "vendor/") || strings.Contains(owner, "build/build.gradle") {
			t.Fatalf("permanently excluded manifest was emitted: %#v", record)
		}
	}

	temporary := t.TempDir()
	left := filepath.Join(temporary, "left", "dependencies")
	right := filepath.Join(temporary, "right", "dependencies")
	copyTree(t, dependenciesFixture, left)
	copyTree(t, dependenciesFixture, right)
	if leftData, rightData := analyzeFixture(t, left), analyzeFixture(t, right); !bytes.Equal(leftData, rightData) {
		t.Fatal("absolute checkout path affected dependency output")
	}
}

func assertDependencyNode(t *testing.T, nodes map[string]map[string]any, coordinate string) {
	t.Helper()
	node := nodes["dependency:maven:"+coordinate]
	if node == nil {
		t.Fatalf("missing external dependency module %s", coordinate)
	}
	attributes := attributes(node)
	if node["kind"] != "module" || attributes["dependency"] != true || attributes["ecosystem"] != "maven" {
		t.Fatalf("external dependency module %s = %#v", coordinate, node)
	}
	expectedPath := "@dependencies/maven/" + strings.ReplaceAll(coordinate, ":", "/")
	if node["path"] != expectedPath {
		t.Fatalf("external dependency path = %v, want %s", node["path"], expectedPath)
	}
}

func assertDependencyEdge(t *testing.T, records []map[string]any, nodes map[string]map[string]any, coordinate, manifest, configuration, constraint, category string, expected int) map[string]any {
	t.Helper()
	target := nodes["dependency:maven:"+coordinate]
	source := nodes["build-module:"+manifest]
	if target == nil || source == nil {
		t.Fatalf("missing dependency endpoints for %s in %s", coordinate, manifest)
	}
	var matches []map[string]any
	for _, record := range records {
		attrs := attributes(record)
		if record["record"] == "edge" && record["relation"] == "depends-on" && record["source"] == source["id"] && record["target"] == target["id"] && attrs["manifest"] == manifest && attrs["configuration"] == configuration && attrs["constraint"] == constraint && attrs["category"] == category {
			matches = append(matches, record)
		}
	}
	if len(matches) != expected {
		t.Fatalf("dependency edge %s in %s count = %d, want %d", coordinate, manifest, len(matches), expected)
	}
	assertDependencySpan(t, matches[0], manifest)
	return matches[0]
}

func assertUnsupportedDependency(t *testing.T, records []map[string]any, manifest, expression, configuration string) {
	t.Helper()
	for _, record := range records {
		attrs := attributes(record)
		text, _ := record["expression"].(string)
		if record["record"] == "unresolved" && record["relation"] == "depends-on" && record["reason"] == "unsupported-form" && record["owner"] == manifest && strings.Contains(text, expression) && attrs["manifest"] == manifest && attrs["configuration"] == configuration {
			assertDependencySpan(t, record, manifest)
			return
		}
	}
	t.Fatalf("missing unsupported dependency %q in %s", expression, manifest)
}

func assertDependencySpan(t *testing.T, record map[string]any, manifest string) {
	t.Helper()
	value, _ := record["span"].(map[string]any)
	if value == nil || value["path"] != manifest || value["start_line"].(float64) < 1 || value["start_column"].(float64) < 1 || value["end_line"].(float64) < value["start_line"].(float64) {
		t.Fatalf("invalid dependency source span: %#v", record)
	}
}

func attributes(record map[string]any) map[string]any {
	value, _ := record["attributes"].(map[string]any)
	return value
}

func recordsContain(records []map[string]any, text string) bool {
	for _, record := range records {
		if strings.Contains(jsonText(record), text) {
			return true
		}
	}
	return false
}
