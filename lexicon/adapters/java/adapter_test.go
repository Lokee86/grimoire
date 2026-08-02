package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const foundationFixture = "testdata/foundation"

func TestFoundationDeclarationsImportsAndContainment(t *testing.T) {
	data := analyzeFixture(t, foundationFixture)
	records := decodeRecords(t, data)
	if header := records[0]; header["record"] != "lexicon" || header["language"] != "java" || header["mode"] != "full" || header["schema_version"] != float64(1) {
		t.Fatalf("unexpected header: %#v", header)
	}

	nodes := nodeIndex(records)
	assertDeclaration(t, nodes, "com.acme.Service", "type", "class")
	assertDeclaration(t, nodes, "com.acme.Worker", "interface", "interface")
	assertDeclaration(t, nodes, "com.acme.Mode", "type", "enum")
	assertDeclaration(t, nodes, "com.acme.Result", "type", "record")
	assertDeclaration(t, nodes, "com.acme.Marker", "interface", "annotation")
	assertDeclaration(t, nodes, "com.acme.Service.<init>(Helper)", "constructor", "constructor")
	assertDeclaration(t, nodes, "com.acme.Result.<init>(String,int)", "constructor", "compact-constructor")
	assertDeclaration(t, nodes, "com.acme.Service.execute(String,int...)", "method", "method")
	assertDeclaration(t, nodes, "com.acme.Service.helper", "field", "field")
	assertDeclaration(t, nodes, "com.acme.Service.count", "field", "field")
	assertDeclaration(t, nodes, "com.acme.Service.limit", "field", "field")
	assertDeclaration(t, nodes, "com.acme.Result.value", "field", "record-component")
	assertDeclaration(t, nodes, "com.acme.Result.count", "field", "record-component")
	assertDeclaration(t, nodes, "com.acme.Service.Nested", "type", "class")

	parameterNames := map[string]bool{}
	for _, node := range nodes {
		if node["kind"] == "parameter" {
			parameterNames[node["name"].(string)] = true
		}
	}
	for _, name := range []string{"helper", "input", "retries", "value"} {
		if !parameterNames[name] {
			t.Errorf("missing parameter %q", name)
		}
	}

	assertAllEndpoints(t, records, nodes)
	assertContains(t, records, nodes, "com.acme.Service", "com.acme.Service.execute(String,int...)")
	assertContains(t, records, nodes, "com.acme.Service.execute(String,int...)", "com.acme.Service.execute(String,int...)#parameter:0:input")
	assertContains(t, records, nodes, "com.acme.Service", "com.acme.Service.Nested")
	assertImport(t, records, nodes, "com.acme.support.Helper")
	assertImport(t, records, nodes, "com.acme.support.Helper.DEFAULT_NAME")
	assertImportExpression(t, records, nodes, "import com.acme.support.*;", "com.acme.support")
	assertImportExpression(t, records, nodes, "import static com.acme.support.Helper.*;", "com.acme.support.Helper")
	assertUnresolved(t, records, "imports", "import java.util.List;", "external-target")
	assertUnresolvedRelation(t, records, "defines", "unsupported-form", "src/main/java/com/acme/broken/Broken.java")

	for _, node := range nodes {
		path := node["path"].(string)
		if strings.Contains(path, "\\") {
			t.Errorf("path is not normalized: %q", path)
		}
		if strings.Contains(path, "vendor") || node["name"] == "IgnoredVendorSource" {
			t.Errorf("permanently excluded source was emitted: %#v", node)
		}
	}
}

func TestFoundationIsCanonicalDeterministicAndCheckoutIndependent(t *testing.T) {
	first := analyzeFixture(t, foundationFixture)
	second := analyzeFixture(t, foundationFixture)
	if !bytes.Equal(first, second) {
		t.Fatal("repeat analysis was not byte-identical")
	}
	assertCanonicalJSONL(t, first)

	temporary := t.TempDir()
	left := filepath.Join(temporary, "left", "foundation")
	right := filepath.Join(temporary, "right", "foundation")
	copyTree(t, foundationFixture, left)
	copyTree(t, foundationFixture, right)
	leftData := analyzeFixture(t, left)
	rightData := analyzeFixture(t, right)
	if !bytes.Equal(leftData, rightData) {
		t.Fatal("absolute checkout path affected adapter output")
	}

	records := decodeRecords(t, first)
	service := nodeIndex(records)["com.acme.Service"]
	payload := []byte("lexicon:v1\x00java\x00type\x00com.acme.Service")
	digest := sha256.Sum256(payload)
	expected := "sha256:" + hex.EncodeToString(digest[:])
	if service["id"] != expected {
		t.Fatalf("Service id = %v, want %s", service["id"], expected)
	}
}

func TestCLIWritesStdoutAndFile(t *testing.T) {
	var stdout bytes.Buffer
	if err := run([]string{"--repo", foundationFixture, "--output", "-"}, &stdout); err != nil {
		t.Fatal(err)
	}
	if records := decodeRecords(t, stdout.Bytes()); len(records) < 2 {
		t.Fatalf("stdout emitted %d records", len(records))
	}

	output := filepath.Join(t.TempDir(), "facts.jsonl")
	if err := run([]string{"--repo", foundationFixture, "--output", output}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	written, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(stdout.Bytes(), written) {
		t.Fatal("stdout and file output differ")
	}
}
