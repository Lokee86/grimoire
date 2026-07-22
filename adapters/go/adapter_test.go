package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFNV1aMatchesRustHasher(t *testing.T) {
	const want uint64 = 0xe71fa2190541574b
	if got := hashBytes([]byte("abc")); got != want {
		t.Fatalf("hashBytes(abc) = %016x, want %016x", got, want)
	}
}

func TestScanIsDeterministicAndCoversGoFacts(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"go.mod": "module example.com/demo\n\ngo 1.20\n",
		"main.go": `package demo

import (
	"fmt"
	"example.com/demo/internal/sub"
)

type Widget int
func helper() {}
func caller() { helper(); fmt.Println(sub.Value) }
func (Widget) Method() {}
`,
		"main_test.go": `package demo
import "testing"
func TestCaller(t *testing.T) { helper() }
`,
		"internal/sub/sub.go": "package sub\nconst Value = 1\n",
		"vendor/ignored.go":   "package ignored\nfunc Nope() {}\n",
		".git/ignored.go":     "package ignored\nfunc Nope() {}\n",
	})

	first, summary, err := scanRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	second, _, err := scanRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := encodeFacts(first), encodeFacts(second); got != want {
		t.Fatal("scanning the same repository was not deterministic")
	}
	if summary.CallExpressions != 3 || summary.DirectCalls != 3 || summary.UnresolvedCalls != 0 {
		t.Fatalf("call counts = total %d direct %d unresolved %d, want 3/3/0", summary.CallExpressions, summary.DirectCalls, summary.UnresolvedCalls)
	}
	if summary.Files != 4 {
		t.Fatalf("file count = %d, want 4", summary.Files)
	}
	if !hasNode(first, KindType, "Widget") || !hasNode(first, KindMethod, "Method") || !hasNode(first, KindTest, "TestCaller") {
		t.Fatal("missing type, method, or test node")
	}
	if !hasNode(first, KindImport, "fmt") || !hasNode(first, KindImport, "example.com/demo/internal/sub") {
		t.Fatal("missing external or internal import node")
	}
	if hasNode(first, KindFunction, "Nope") {
		t.Fatal("ignored vendor or .git source was scanned")
	}
	if countRelation(first, RelCalls) != 3 {
		t.Fatalf("call edge count = %d, want 3", countRelation(first, RelCalls))
	}
	if len(first.Unresolved) != 0 {
		t.Fatalf("unresolved fact count = %d, want 0", len(first.Unresolved))
	}
	if !hasNode(first, KindFunction, "Println") {
		t.Fatal("missing resolved standard-library function node")
	}
	packageKey := hashIdentity("package:example.com/demo:demo")
	fileKey := hashIdentity("file:main.go")
	if !hasEdge(first, packageKey, fileKey, RelContains) || hasEdge(first, fileKey, packageKey, RelContains) {
		t.Fatal("package/file containment direction is incorrect")
	}
	if !strings.Contains(encodeFacts(first), `"record":"lexicon"`) ||
		!strings.Contains(encodeFacts(first), `"schema_version":1`) {
		t.Fatal("fact output has no canonical Lexicon header")
	}
}

func writeFixture(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, contents := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func hasNode(facts RepositoryFacts, kind NodeKind, name string) bool {
	for _, node := range facts.Nodes {
		if node.Kind == kind && node.Name == name {
			return true
		}
	}
	return false
}

func countRelation(facts RepositoryFacts, relation RelationKind) int {
	count := 0
	for _, edge := range facts.Edges {
		if edge.Relation == relation {
			count++
		}
	}
	return count
}

func hasEdge(facts RepositoryFacts, source, target NodeKey, relation RelationKind) bool {
	for _, edge := range facts.Edges {
		if edge.Source == source && edge.Target == target && edge.Relation == relation {
			return true
		}
	}
	return false
}
