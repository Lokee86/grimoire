package main

import "testing"

func TestAnalyzeRepositoryEmitsConservativeVariableAndFieldDataflow(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "Library.lss", `Public Class Counter
    total As Integer

    Public Sub Add(delta As Integer)
        Me.total = Me.total + delta
    End Sub
End Class

Public shared As Integer

Public Sub Run()
    Dim local As Integer
    local = shared
    shared = local
End Sub
`)

	records := analyzeFixtureRecords(t, root)
	assertNode(t, records, "field", "total")
	assertRelationAtPaths(t, records, "writes", "Library.lss", "Add", "Library.lss", "total")
	assertRelationAtPaths(t, records, "reads", "Library.lss", "Add", "Library.lss", "total")
	assertRelationAtPaths(t, records, "reads", "Library.lss", "Add", "Library.lss", "delta")
	assertRelationAtPaths(t, records, "writes", "Library.lss", "Run", "Library.lss", "local")
	assertRelationAtPaths(t, records, "reads", "Library.lss", "Run", "Library.lss", "shared")
	assertRelationAtPaths(t, records, "writes", "Library.lss", "Run", "Library.lss", "shared")
	assertRelationAtPaths(t, records, "reads", "Library.lss", "Run", "Library.lss", "local")
}

func TestAnalyzeRepositoryScopesImportedModuleVariableDataflow(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "Library.lss", `Public visible As Integer
Private hidden As Integer
`)
	writeFixture(t, root, "Caller.lss", `Use "Library"
Public Sub Run()
    Dim local As Integer
    local = visible
    hidden = local
End Sub
`)

	records := analyzeFixtureRecords(t, root)
	assertRelationAtPaths(t, records, "reads", "Caller.lss", "Run", "Library.lss", "visible")
	assertNoRelationAtPaths(t, records, "writes", "Caller.lss", "Run", "Library.lss", "hidden")
}

func TestAnalyzeRepositoryDoesNotReadIdentifiersInsideStrings(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "Library.lss", `Public Sub Run()
    Dim secret As String
    Print "secret"
End Sub
`)

	records := analyzeFixtureRecords(t, root)
	assertNoRelationAtPaths(t, records, "reads", "Library.lss", "Run", "Library.lss", "secret")
}
