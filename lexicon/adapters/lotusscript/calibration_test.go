package main

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
)

func TestLogicalLinesSkipPercentRemBlocks(t *testing.T) {
	lines := logicalLines("sample.lss", `Public Sub Run()
%REM
    Call Ghost()
    Public Sub Fake()
%END REM
    Call Real()
End Sub
`)
	var text []string
	for _, line := range lines {
		text = append(text, line.text)
	}
	joined := strings.Join(text, "\n")
	if strings.Contains(joined, "Ghost") || strings.Contains(joined, "Fake") {
		t.Fatalf("block comment leaked into logical lines: %s", joined)
	}
	if !strings.Contains(joined, "Call Real()") {
		t.Fatalf("source following block comment was lost: %s", joined)
	}
}

func TestAnalyzeRepositoryResolvesTypedReceiversAndSkipsIndexing(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "json.ls", `Public Class JSONArray
    Public Sub New()
    End Sub

    Public Sub AddItem(value As String)
    End Sub
End Class

Public Function Build() As JSONArray
    Dim result As New JSONArray
    Dim values List As String
    Call result.AddItem("value")
    values("key") = "value"
    Set Build = result
End Function
`)

	data, err := analyzeRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	records := decodeRecords(t, data)
	assertRelation(t, records, "calls", "function", "Build", "constructor", "New")
	assertRelation(t, records, "calls", "function", "Build", "method", "AddItem")
	assertNoUnresolvedCandidate(t, records, "JSONArray")
	assertNoUnresolvedCandidate(t, records, "result.AddItem")
	assertNoUnresolvedCandidate(t, records, "values")
}

func TestAnalyzeRepositoryExtractsODPAgentSource(t *testing.T) {
	root := t.TempDir()
	source := `'++LotusScript Development Environment:2:5:(Options):0:74
%REM
    Call Ghost()
%END REM
Option Public
Option Declare
Use "Core"

Sub Initialize
    Call RunAgent()
End Sub

Private Sub RunAgent
End Sub
`
	payload := append([]byte{0x81, 0x02, 0x85, 0xff, 0x20, 0x00, 0x00, 0x00}, []byte(source)...)
	payload = append(payload, 0)
	dxl := fmt.Sprintf("<?xml version='1.0'?><note><item name='$AssistAction'><rawitemdata type='10'>%s</rawitemdata></item></note>", base64.StdEncoding.EncodeToString(payload))
	writeFixture(t, root, "Code/Agents/TestAgent.lsa", dxl)
	writeFixture(t, root, "Code/ScriptLibraries/Core.lss", "Public Sub CoreHelper\nEnd Sub\n")

	data, err := analyzeRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	records := decodeRecords(t, data)
	assertNode(t, records, "function", "Initialize")
	assertNode(t, records, "function", "RunAgent")
	assertRelation(t, records, "calls", "function", "Initialize", "function", "RunAgent")
	assertRelation(t, records, "imports", "import", "Core", "module", "Core")
	assertNoUnresolvedCandidate(t, records, "Ghost")
}

func TestAnalyzeRepositoryExtractsStructuredDXLAgentSource(t *testing.T) {
	root := t.TempDir()
	dxl := `<?xml version='1.0'?>
<agent xmlns='http://www.lotus.com/dxl'>
<code event='options'><lotusscript>%REM
    Call Ghost()
%END REM
Option Public
Option Declare
</lotusscript></code>
<code event='initialize'><lotusscript>Sub Initialize
    Call ExecuteAgent()
End Sub

Private Sub ExecuteAgent
End Sub
</lotusscript></code>
</agent>`
	writeFixture(t, root, "Code/Agents/Structured.lsa", dxl)

	data, err := analyzeRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	records := decodeRecords(t, data)
	assertRelation(t, records, "calls", "function", "Initialize", "function", "ExecuteAgent")
	assertNoUnresolvedCandidate(t, records, "Ghost")
}

func assertNoUnresolvedCandidate(t *testing.T, records []map[string]any, candidate string) {
	t.Helper()
	for _, record := range records {
		if record["record"] != "unresolved" {
			continue
		}
		attributes, _ := record["attributes"].(map[string]any)
		if strings.EqualFold(fmt.Sprint(attributes["candidate_name"]), candidate) {
			t.Fatalf("unexpected unresolved candidate %q: %#v", candidate, record)
		}
	}
}
