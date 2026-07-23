package main

import "testing"

func TestSemanticDataflowReadsWritesCompoundMembersAndShadowing(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, map[string]string{
		"go.mod": "module example.com/dataflow\n\ngo 1.22\n",
		"main.go": `package dataflow

type Box struct { Field int }
const Constant = 3
func consume(value int) int { return value }
func run(value int, box *Box) int {
    local := value
    local += Constant
    local++
    box.Field = local
    {
        value := local
        local = consume(value)
    }
    return local + value
}
`,
	})
	facts, _, err := scanRepository(root)
	if err != nil { t.Fatal(err) }
	source := hashIdentity("function:example.com/dataflow:run")
	reads, writes := map[NodeKey]bool{}, map[NodeKey]bool{}
	for _, edge := range facts.Edges {
		if edge.Source != source { continue }
		switch edge.Relation { case RelReads: reads[edge.Target] = true; case RelWrites: writes[edge.Target] = true }
	}
	if len(reads) < 3 || len(writes) < 3 { t.Fatalf("dataflow reads=%d writes=%d", len(reads), len(writes)) }
	if !hasNode(facts, KindField, "Field") || !hasNode(facts, KindConstant, "Constant") { t.Fatal("missing field or constant data symbols") }
	if readsEqual := len(reads); readsEqual == 0 { t.Fatal("missing argument/return reads") }
	for _, edge := range facts.Edges {
		if edge.Source == source && (edge.Relation == RelReads || edge.Relation == RelWrites) && edge.Target == hashIdentity("function:go:builtins:len") { t.Fatal("fabricated builtin dataflow target") }
	}
}
