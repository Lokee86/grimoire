package main

import "testing"

func TestAnalyzeRepositoryScopesCallsToImportedLibraries(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "A.lss", `Public Sub Helper()
End Sub
`)
	writeFixture(t, root, "B.lss", `Public Sub Helper()
End Sub
`)
	writeFixture(t, root, "Caller.lss", `Use "A"
Public Sub Run()
    Call Helper()
End Sub
`)

	records := analyzeFixtureRecords(t, root)
	assertRelationAtPaths(t, records, "calls", "Caller.lss", "Run", "A.lss", "Helper")
	assertNoRelationAtPaths(t, records, "calls", "Caller.lss", "Run", "B.lss", "Helper")
	assertNoUnresolvedCandidate(t, records, "Helper")
}

func TestAnalyzeRepositoryDoesNotResolveUnimportedGlobal(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "Library.lss", `Public Sub RemoteOnly()
End Sub
`)
	writeFixture(t, root, "Caller.lss", `Public Sub Run()
    Call RemoteOnly()
End Sub
`)

	records := analyzeFixtureRecords(t, root)
	assertNoRelationAtPaths(t, records, "calls", "Caller.lss", "Run", "Library.lss", "RemoteOnly")
	assertUnresolved(t, records, "calls", "external-target", "RemoteOnly")
}

func TestAnalyzeRepositoryKeepsPrivateImportedDeclarationsOutOfScope(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "Library.lss", `Private Sub Hidden()
End Sub

Public Sub Visible()
End Sub
`)
	writeFixture(t, root, "Caller.lss", `Use "Library"
Public Sub Run()
    Call Hidden()
    Call Visible()
End Sub
`)

	records := analyzeFixtureRecords(t, root)
	assertNoRelationAtPaths(t, records, "calls", "Caller.lss", "Run", "Library.lss", "Hidden")
	assertRelationAtPaths(t, records, "calls", "Caller.lss", "Run", "Library.lss", "Visible")
	assertUnresolved(t, records, "calls", "external-target", "Hidden")
	assertNoUnresolvedCandidate(t, records, "Visible")
}

func TestAnalyzeRepositoryUsesOptionPublicForImportedDeclarations(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "Library.lss", `Option Public
Sub VisibleByDefault()
End Sub
`)
	writeFixture(t, root, "Caller.lss", `Use "Library"
Public Sub Run()
    Call VisibleByDefault()
End Sub
`)

	records := analyzeFixtureRecords(t, root)
	assertRelationAtPaths(t, records, "calls", "Caller.lss", "Run", "Library.lss", "VisibleByDefault")
	assertNoUnresolvedCandidate(t, records, "VisibleByDefault")
}

func TestAnalyzeRepositoryUsesTransitiveLibraryScope(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "Core.lss", `Public Sub CoreHelper()
End Sub
`)
	writeFixture(t, root, "Facade.lss", `Use "Core"
Public Sub FacadeHelper()
End Sub
`)
	writeFixture(t, root, "Caller.lss", `Use "Facade"
Public Sub Run()
    Call CoreHelper()
End Sub
`)

	records := analyzeFixtureRecords(t, root)
	assertRelationAtPaths(t, records, "calls", "Caller.lss", "Run", "Core.lss", "CoreHelper")
	assertNoUnresolvedCandidate(t, records, "CoreHelper")
}
