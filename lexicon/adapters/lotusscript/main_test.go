package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzeRepositoryExtractsLotusScriptFoundation(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "Core.lss", `Option Public
Public Class BaseWorker
    Public Sub Run()
        MsgBox("base")
    End Sub
End Class

Public Function Helper(value As String) As String
    Helper = UCase(value)
End Function
`)
	writeFixture(t, root, "agents/Worker.lss", `Option Declare
Use "Core"
Public Class Worker As BaseWorker
    Private mName As String

    Public Sub New(name As String)
        mName = name
    End Sub

    Public Sub Execute()
        Call Helper(mName)
        Call Me.Run()
        session.GetDatabase("", "")
    End Sub
End Class
`)

	data, err := analyzeRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	records := decodeRecords(t, data)
	assertHeader(t, records[0])
	assertNode(t, records, "type", "Worker")
	assertNode(t, records, "constructor", "New")
	assertNode(t, records, "method", "Execute")
	assertNode(t, records, "field", "mName")
	assertNode(t, records, "parameter", "name")
	assertRelation(t, records, "imports", "import", "Core", "module", "Core")
	assertRelation(t, records, "extends", "type", "Worker", "type", "BaseWorker")
	assertRelation(t, records, "calls", "method", "Execute", "function", "Helper")
	assertRelation(t, records, "calls", "method", "Execute", "method", "Run")
	assertUnresolved(t, records, "calls", "builtin-target", "MsgBox")
	assertUnresolved(t, records, "calls", "dynamic-target", "session.GetDatabase")
}

func TestAnalyzeRepositoryIsDeterministic(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "Library.ls", `Public Sub First()
    Call Second()
End Sub

Private Sub Second()
End Sub
`)
	first, err := analyzeRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	second, err := analyzeRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("identical input produced different facts")
	}
}

func TestLogicalLinesSplitColonStatementsWithoutSplittingLabelsOrDates(t *testing.T) {
	lines := logicalLines("sample.lss", "ErrorHandler:\nCall First() : Call Second(#12:30:00#)\n")
	if len(lines) != 2 {
		t.Fatalf("logical line count = %d, want 2: %#v", len(lines), lines)
	}
	if lines[0].text != "Call First()" {
		t.Fatalf("first statement = %q", lines[0].text)
	}
	if lines[1].text != "Call Second(#12:30:00#)" {
		t.Fatalf("second statement = %q", lines[1].text)
	}
}

func TestLogicalLinesPreservePipeStringsAndContinuations(t *testing.T) {
	lines := logicalLines("sample.lss", "Call Helper( _\n    |apostrophe ' remains|) ' comment\n")
	if len(lines) != 1 {
		t.Fatalf("logical line count = %d, want 1", len(lines))
	}
	if got, want := lines[0].text, "Call Helper( |apostrophe ' remains|)"; got != want {
		t.Fatalf("logical line = %q, want %q", got, want)
	}
	if lines[0].span.StartLine != 1 || lines[0].span.EndLine != 2 {
		t.Fatalf("logical span = %#v", lines[0].span)
	}
}

func writeFixture(t *testing.T, root, relative, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func decodeRecords(t *testing.T, data []byte) []map[string]any {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(data))
	var records []map[string]any
	for decoder.More() {
		var record map[string]any
		if err := decoder.Decode(&record); err != nil {
			t.Fatal(err)
		}
		records = append(records, record)
	}
	return records
}

func assertHeader(t *testing.T, header map[string]any) {
	t.Helper()
	if header["record"] != "lexicon" || header["language"] != language || header["mode"] != "full" {
		t.Fatalf("unexpected header: %#v", header)
	}
}

func assertNode(t *testing.T, records []map[string]any, kind, name string) {
	t.Helper()
	if findNode(records, kind, name) == nil {
		t.Fatalf("missing %s node %q", kind, name)
	}
}

func assertRelation(t *testing.T, records []map[string]any, relation, sourceKind, sourceName, targetKind, targetName string) {
	t.Helper()
	source := findNode(records, sourceKind, sourceName)
	target := findNode(records, targetKind, targetName)
	if source == nil || target == nil {
		t.Fatalf("missing relation endpoint for %s: %s %q -> %s %q", relation, sourceKind, sourceName, targetKind, targetName)
	}
	for _, record := range records {
		if record["record"] == "edge" && record["relation"] == relation && record["source"] == source["id"] && record["target"] == target["id"] {
			return
		}
	}
	t.Fatalf("missing %s edge: %s %q -> %s %q", relation, sourceKind, sourceName, targetKind, targetName)
}

func assertUnresolved(t *testing.T, records []map[string]any, relation, reason, candidate string) {
	t.Helper()
	for _, record := range records {
		if record["record"] != "unresolved" || record["relation"] != relation || record["reason"] != reason {
			continue
		}
		attributes, _ := record["attributes"].(map[string]any)
		if attributes["candidate_name"] == candidate {
			return
		}
	}
	t.Fatalf("missing unresolved %s %s for %q", relation, reason, candidate)
}

func findNode(records []map[string]any, kind, name string) map[string]any {
	for _, record := range records {
		if record["record"] == "node" && record["kind"] == kind && record["name"] == name {
			return record
		}
	}
	return nil
}
