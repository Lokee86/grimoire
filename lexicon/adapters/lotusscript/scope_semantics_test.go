package main

import "testing"

func TestAnalyzeRepositoryResolvesInheritedPrivateMethodFromMe(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "Base.lss", `Option Public
Class BaseWorker
    Private Sub Hidden()
    End Sub
End Class
`)
	writeFixture(t, root, "Derived.lss", `Use "Base"
Public Class Worker As BaseWorker
    Public Sub Run()
        Call Me.Hidden()
    End Sub
End Class
`)

	records := analyzeFixtureRecords(t, root)
	assertRelationAtPaths(t, records, "calls", "Derived.lss", "Run", "Base.lss", "Hidden")
	assertNoUnresolvedCandidate(t, records, "Me.Hidden")
}

func TestAnalyzeRepositoryResolvesPrivateMethodThroughSameClassReceiver(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "Library.lss", `Public Class Worker
    Private Sub Hidden()
    End Sub

    Public Sub Run(other As Worker)
        Call other.Hidden()
    End Sub
End Class
`)

	records := analyzeFixtureRecords(t, root)
	assertRelationAtPaths(t, records, "calls", "Library.lss", "Run", "Library.lss", "Hidden")
	assertNoUnresolvedCandidate(t, records, "other.Hidden")
}

func TestAnalyzeRepositoryScopesDuplicateClassNames(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "A.lss", `Public Class Worker
    Public Sub New()
    End Sub
    Public Sub Run()
    End Sub
End Class
`)
	writeFixture(t, root, "B.lss", `Public Class Worker
    Public Sub New()
    End Sub
    Public Sub Run()
    End Sub
End Class
`)
	writeFixture(t, root, "Caller.lss", `Use "A"
Public Sub Execute()
    Dim worker As New Worker
    Call worker.Run()
End Sub
`)

	records := analyzeFixtureRecords(t, root)
	assertRelationAtPaths(t, records, "calls", "Caller.lss", "Execute", "A.lss", "New")
	assertRelationAtPaths(t, records, "calls", "Caller.lss", "Execute", "A.lss", "Run")
	assertNoRelationAtPaths(t, records, "calls", "Caller.lss", "Execute", "B.lss", "New")
	assertNoRelationAtPaths(t, records, "calls", "Caller.lss", "Execute", "B.lss", "Run")
	assertNoUnresolvedCandidate(t, records, "Worker.New")
	assertNoUnresolvedCandidate(t, records, "worker.Run")
}

func TestAnalyzeRepositoryResolvesWithStatementMembers(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "Library.lss", `Public Class Worker
    Public Sub Run()
    End Sub
End Class

Public Sub Execute()
    Dim worker As New Worker
    With worker
        Call .Run()
    End With
End Sub
`)

	records := analyzeFixtureRecords(t, root)
	assertRelationAtPaths(t, records, "calls", "Library.lss", "Execute", "Library.lss", "Run")
	assertNoUnresolvedCandidate(t, records, "worker.Run")
}

func TestAnalyzeRepositoryCapturesCallsFromColonSeparatedStatements(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "Library.lss", `Public Sub First()
End Sub
Public Sub Second()
End Sub
Public Sub Run()
    Call First() : Call Second()
End Sub
`)

	records := analyzeFixtureRecords(t, root)
	assertRelationAtPaths(t, records, "calls", "Library.lss", "Run", "Library.lss", "First")
	assertRelationAtPaths(t, records, "calls", "Library.lss", "Run", "Library.lss", "Second")
}

func TestAnalyzeRepositoryTreatsStaticAndRedimAsDeclarations(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "Library.lss", `Public Sub Run()
    Static current As String
    ReDim values(0 To 2)
    values(0) = current
End Sub
`)

	records := analyzeFixtureRecords(t, root)
	assertNode(t, records, "variable", "current")
	assertNode(t, records, "variable", "values")
	assertNoUnresolvedCandidate(t, records, "Static")
	assertNoUnresolvedCandidate(t, records, "values")
}
