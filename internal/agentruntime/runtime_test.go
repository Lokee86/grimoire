package agentruntime

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/Lokee86/grimoire/internal/agentquery"
	"github.com/Lokee86/grimoire/internal/investigation"
	"github.com/Lokee86/grimoire/internal/knowledge"
	"github.com/Lokee86/grimoire/internal/repostate"
)

func TestExecuteReturnsFullEvidenceWithoutSession(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "docs/design.md", "# Graph ownership\nArcana owns repository graph traversal.\n")
	response, err := Execute(context.Background(), Request{
		Request: agentquery.Request{Mode: "search", Root: root, Query: "graph ownership", Limit: 4},
	}, testOptions(root))
	if err != nil {
		t.Fatal(err)
	}
	if len(response.SourceMatches) != 1 {
		t.Fatalf("source response = %#v", response.SourceMatches)
	}
	if len(response.DocumentMatches) == 0 || response.DocumentMatches[0].Path != "docs/design.md" {
		t.Fatalf("document response = %#v", response.DocumentMatches)
	}
	if response.Delta != nil {
		t.Fatalf("unexpected session delta: %#v", response.Delta)
	}
}

func TestDocumentMatchesRemainSeparateAndInspectable(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "internal/graph.go", "package internal\n\nfunc TraverseGraph() {}\n")
	writeTestFile(t, root, "docs/design.md", "# Graph ownership\nArcana owns repository graph traversal.\n")
	response, err := Execute(context.Background(), Request{
		Request: agentquery.Request{Mode: "search", Root: root, Query: "graph ownership", Limit: 4},
	}, testOptions(root))
	if err != nil {
		t.Fatal(err)
	}
	if len(response.SourceMatches) != 1 || len(response.DocumentMatches) != 1 {
		t.Fatalf("discovery lanes were not separate: source=%#v documents=%#v", response.SourceMatches, response.DocumentMatches)
	}
	if response.DocumentMatches[0].Handle == "" || response.DocumentMatches[0].StartLine <= 0 {
		t.Fatalf("document match has no inspectable handle and range: %#v", response.DocumentMatches[0])
	}

	inspection, err := Execute(context.Background(), Request{
		Request: agentquery.Request{
			Mode: "inspect", Root: root, Anchor: response.DocumentMatches[0].Handle, Limit: 4,
		},
	}, testOptions(root))
	if err != nil {
		t.Fatal(err)
	}
	if len(inspection.DocumentMatches) != 1 || inspection.DocumentMatches[0].Handle != response.DocumentMatches[0].Handle {
		t.Fatalf("document handle did not inspect exact evidence: %#v", inspection.DocumentMatches)
	}
	if len(inspection.SourceMatches) != 0 || len(inspection.SymbolMatches) != 0 {
		t.Fatalf("document inspection leaked into code lanes: %#v", inspection)
	}
}

func TestRationaleSearchReturnsArchitectureAndPlanningDocumentsWithoutDisplacingCode(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "docs/architecture/system.md", "# Architecture rationale\nRepository architecture decisions and ownership rationale.\n")
	writeTestFile(t, root, "docs/planning/roadmap.md", "# Planning rationale\nRepository planning decisions and implementation rationale.\n")

	result := func(provider, value string) agentquery.Result {
		return agentquery.Result{
			Rank: 1, Provider: provider,
			Node: agentquery.Node{Handle: agentquery.Handle{Value: value, Provider: provider}},
		}
	}
	options := testOptions(root)
	options.ExecuteQuery = func(_ context.Context, request agentquery.Request) (agentquery.Response, error) {
		return agentquery.Response{
			Schema:   agentquery.SchemaVersion,
			Mode:     request.Mode,
			Snapshot: agentquery.Snapshot{Source: "source-1"},
			ExactMatches: []agentquery.Result{
				result("exact", "exact-1"), result("exact", "exact-2"),
			},
			SourceMatches: []agentquery.Result{
				result("lexical", "source-1"), result("lexical", "source-2"),
			},
			SymbolMatches: []agentquery.Result{
				result("lexicon", "symbol-1"), result("lexicon", "symbol-2"),
			},
		}, nil
	}

	response, err := Execute(context.Background(), Request{
		Request: agentquery.Request{
			Mode: "search", Root: root,
			Query: "repository architecture planning rationale", Limit: 2,
		},
	}, options)
	if err != nil {
		t.Fatal(err)
	}
	if len(response.ExactMatches) != 2 || len(response.SourceMatches) != 2 || len(response.SymbolMatches) != 2 {
		t.Fatalf("code lanes were reduced by documentation: %+v", response)
	}
	paths := make(map[string]bool)
	for _, document := range response.DocumentMatches {
		paths[document.Path] = true
	}
	if !paths["docs/architecture/system.md"] || !paths["docs/planning/roadmap.md"] {
		t.Fatalf("rationale search did not retain both document families: %+v", response.DocumentMatches)
	}
}

func TestDocumentCoverageReportsDeferredRankedResults(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "docs/one.md", "# Repository rationale one\nArchitecture planning rationale ownership.\n")
	writeTestFile(t, root, "docs/two.md", "# Repository rationale two\nArchitecture planning rationale boundaries.\n")
	writeTestFile(t, root, "docs/three.md", "# Repository rationale three\nArchitecture planning rationale dependencies.\n")

	options := testOptions(root)
	options.ExecuteQuery = func(_ context.Context, request agentquery.Request) (agentquery.Response, error) {
		return agentquery.Response{
			Schema: agentquery.SchemaVersion, Mode: request.Mode,
			Snapshot: agentquery.Snapshot{Source: "source-1"},
		}, nil
	}
	response, err := Execute(context.Background(), Request{
		Request: agentquery.Request{
			Mode: "search", Root: root,
			Query: "repository architecture planning rationale", Limit: 2,
		},
	}, options)
	if err != nil {
		t.Fatal(err)
	}
	if len(response.DocumentMatches) != 2 {
		t.Fatalf("document result count = %d, want 2", len(response.DocumentMatches))
	}
	var coverage agentquery.LaneCoverage
	for _, lane := range response.Coverage {
		if lane.Lane == "document_matches" {
			coverage = lane
		}
	}
	if coverage.Available < 3 || coverage.Returned != 2 || coverage.Deferred < 1 {
		t.Fatalf("document coverage did not report deferred results: %+v", coverage)
	}
	if !response.Truncated || !slices.Contains(response.TruncatedLanes, "document_matches") {
		t.Fatalf("deferred document lane was not reported: %+v", response)
	}
}

func TestExecuteResolvesSessionNodeHandlesBeforeQuery(t *testing.T) {
	root := t.TempDir()
	originalHandle := "grimoire:v1:stable-node-handle"
	var requests []agentquery.Request
	options := testOptions(root)
	options.ExecuteQuery = func(_ context.Context, request agentquery.Request) (agentquery.Response, error) {
		requests = append(requests, request)
		response := agentquery.Response{
			Schema:   agentquery.SchemaVersion,
			Mode:     request.Mode,
			Snapshot: agentquery.Snapshot{Source: "source-1"},
		}
		if len(requests) == 1 {
			span := agentquery.Range{
				Path: "internal/target.go", StartLine: 10, EndLine: 12,
				Handle: agentquery.NewSourceHandle("source-1", "internal/target.go", 10, 12),
			}
			response.SourceMatches = []agentquery.Result{{
				Rank: 1, Provider: "lexicon", Kind: "function",
				Node: agentquery.Node{
					Handle: agentquery.Handle{Value: originalHandle, Provider: "lexicon", Snapshot: "lexicon-1", NodeIdentity: "node-one"},
					Kind:   "function", Name: "Target", Path: "internal/target.go", Span: &span,
				},
			}}
		}
		return response, nil
	}

	first, err := Execute(context.Background(), Request{
		Request: agentquery.Request{Mode: "search", Root: root, Query: "target", Limit: 4},
		Session: "stable-handles",
	}, options)
	if err != nil {
		t.Fatal(err)
	}
	if first.Delta == nil || len(first.Delta.NewNodes) != 1 || len(first.Delta.NewSourceRanges) != 1 {
		t.Fatalf("search delta = %#v", first.Delta)
	}
	opaque := first.Delta.NewNodes[0].Handle.String()
	if !strings.HasPrefix(opaque, "g2_") {
		t.Fatalf("session handle = %q", opaque)
	}

	_, err = Execute(context.Background(), Request{
		Request: agentquery.Request{Mode: "inspect", Root: root, Handles: []string{opaque}, Limit: 4},
		Session: "stable-handles",
	}, options)
	if err != nil {
		t.Fatal(err)
	}
	if len(requests) < 2 || len(requests[1].Handles) != 1 || requests[1].Handles[0] != originalHandle {
		t.Fatalf("inspect request did not restore original handle: %#v", requests)
	}

	_, err = Execute(context.Background(), Request{
		Request: agentquery.Request{Mode: "trace", Root: root, Anchor: opaque, Limit: 4},
		Session: "stable-handles",
	}, options)
	if err != nil {
		t.Fatal(err)
	}
	if len(requests) < 3 || requests[2].Anchor != originalHandle {
		t.Fatalf("trace request did not restore original handle: %#v", requests)
	}

	sourceOpaque := first.Delta.NewSourceRanges[0].Handle.String()
	_, err = Execute(context.Background(), Request{
		Request: agentquery.Request{Mode: "inspect", Root: root, Handles: []string{sourceOpaque}, Limit: 4},
		Session: "stable-handles",
	}, options)
	if err != nil {
		t.Fatal(err)
	}
	expectedSource := agentquery.NewSourceHandle("source-1", "internal/target.go", 10, 12).Value
	if len(requests) < 4 || len(requests[3].Handles) != 1 || requests[3].Handles[0] != expectedSource {
		t.Fatalf("source-range handle did not restore exact source handle: %#v", requests)
	}
}

func TestDiscoveryDefaultsUseIndependentLaneLimit(t *testing.T) {
	search := normalizeRequest(Request{Request: agentquery.Request{Mode: "search"}}, Options{})
	if search.Limit != 12 {
		t.Fatalf("search default limit = %d, want 12", search.Limit)
	}
	trace := normalizeRequest(Request{Request: agentquery.Request{Mode: "trace"}}, Options{})
	if trace.Limit != 8 {
		t.Fatalf("trace default limit = %d, want 8", trace.Limit)
	}
	enabled := true
	if includeDocuments(Request{Request: agentquery.Request{CodeOnly: true}, IncludeDocuments: &enabled}) {
		t.Fatal("code_only must suppress documents even when include_documents is true")
	}
}

func TestTraceAndImpactDoNotSearchDocumentsFromOpaqueHandlesByDefault(t *testing.T) {
	for _, mode := range []string{"trace", "impact", "inspect"} {
		request := Request{Request: agentquery.Request{Mode: mode, Anchor: "g1_opaque-session-handle"}}
		if includeDocuments(request) {
			t.Fatalf("%s unexpectedly enabled document retrieval from an opaque handle", mode)
		}
	}
	enabled := true
	if !includeDocuments(Request{Request: agentquery.Request{Mode: "trace", Anchor: "g1_handle"}, IncludeDocuments: &enabled}) {
		t.Fatal("explicit trace document opt-in was ignored")
	}
}

func TestDocumentVectorsAreExplicitOptIn(t *testing.T) {
	if useDocumentVectors(Request{}) {
		t.Fatal("document vectors should be disabled by default")
	}
	enabled := true
	if !useDocumentVectors(Request{UseDocumentVectors: &enabled}) {
		t.Fatal("explicit document vector opt-in was ignored")
	}
	disabled := false
	if useDocumentVectors(Request{UseDocumentVectors: &disabled}) {
		t.Fatal("explicit document vector disable was ignored")
	}
}

func TestCompactKnowledgeResultsBoundsAgentPayload(t *testing.T) {
	links := make([]knowledge.CodeLink, 12)
	for index := range links {
		links[index] = knowledge.CodeLink{Kind: "symbol", Value: "Symbol", SourcePath: "pkg/file.go"}
	}
	result := knowledge.Result{
		Text:      strings.Repeat("évidence ", 300),
		Reasons:   []string{"one", "two", "three", "four", "five"},
		CodeLinks: links,
	}

	compacted := compactKnowledgeResults([]knowledge.Result{result})
	if len(compacted) != 1 {
		t.Fatalf("result count = %d", len(compacted))
	}
	if len(compacted[0].Text) > 1203 || !utf8.ValidString(compacted[0].Text) || !strings.HasSuffix(compacted[0].Text, "…") {
		t.Fatalf("text was not compacted safely: bytes=%d valid=%v suffix=%q", len(compacted[0].Text), utf8.ValidString(compacted[0].Text), compacted[0].Text[len(compacted[0].Text)-3:])
	}
	if len(compacted[0].CodeLinks) != 8 {
		t.Fatalf("code links = %d, want 8", len(compacted[0].CodeLinks))
	}
	if len(compacted[0].Reasons) != 4 {
		t.Fatalf("reasons = %d, want 4", len(compacted[0].Reasons))
	}
	if len(result.CodeLinks) != 12 || len(result.Reasons) != 5 {
		t.Fatal("compaction mutated caller-owned result")
	}
}

func TestInvestigationResponseRecordsDirectRelationships(t *testing.T) {
	subject := agentquery.Node{Handle: agentquery.Handle{Value: "subject", Provider: "lexicon"}, Kind: "function", Name: "Caller"}
	object := agentquery.Node{Handle: agentquery.Handle{Value: "object", Provider: "lexicon"}, Kind: "function", Name: "Callee"}
	seed := agentquery.Node{Handle: agentquery.Handle{Value: "seed", Provider: "lexicon"}, Kind: "function", Name: "Seed"}
	converted := investigationResponse(agentquery.Response{
		Snapshot: agentquery.Snapshot{Source: "source-1"},
		RelationshipMatches: []agentquery.RelationshipMatch{{
			Rank: 1, Provider: "lexicon", Subject: subject, Direction: "outgoing", Relation: "calls", Object: object,
			Seed: &seed, SeedLane: "exact_matches", SeedRank: 1, SeedReasons: []string{"literal source match"},
		}},
	}, nil)
	if len(converted.GraphPaths) != 1 || len(converted.GraphPaths[0].Edges) != 1 || converted.GraphPaths[0].Edges[0] != "calls" {
		t.Fatalf("direct relationship was not recorded as graph evidence: %#v", converted.GraphPaths)
	}
	if len(converted.RetrievalHits) != 1 {
		t.Fatalf("relationship retrieval hit count = %d, want 1", len(converted.RetrievalHits))
	}
	hit := converted.RetrievalHits[0]
	if hit.Lane != "relationship_matches" || hit.Rank != 1 || hit.Relation != "calls" || hit.Seed == nil {
		t.Fatalf("relationship occurrence was not preserved: %#v", hit)
	}
	if hit.Seed.Lane != "exact_matches" || hit.Seed.Rank != 1 || len(hit.Seed.Reasons) != 1 || hit.Seed.Reasons[0] != "literal source match" {
		t.Fatalf("relationship seed tuple was not preserved: %#v", hit.Seed)
	}
	if len(converted.GraphPaths[0].Metadata) != 0 {
		t.Fatalf("query occurrence metadata leaked into graph evidence: %#v", converted.GraphPaths[0].Metadata)
	}
}

func TestInvestigationResponsePreservesRankedLaneMetadataAndOrder(t *testing.T) {
	firstSpan := agentquery.Range{
		Path: "internal/z.go", StartLine: 3, EndLine: 5,
		Handle: agentquery.Handle{Value: "range-z", Provider: "source"},
	}
	secondSpan := agentquery.Range{
		Path: "internal/a.go", StartLine: 7, EndLine: 8,
		Handle: agentquery.Handle{Value: "range-a", Provider: "source"},
	}
	converted := investigationResponse(agentquery.Response{
		Snapshot: agentquery.Snapshot{Source: "source-1"},
		ExactMatches: []agentquery.Result{
			{
				Rank: 1, Provider: "exact", Score: 9, Reasons: []string{"literal source match"}, Excerpt: "first excerpt",
				Node: agentquery.Node{Handle: agentquery.Handle{Value: "z-first", Provider: "source"}, Name: "First", Path: firstSpan.Path, Span: &firstSpan},
			},
			{
				Rank: 2, Provider: "exact", Score: 8, Reasons: []string{"literal source match"}, Excerpt: "second excerpt",
				Node: agentquery.Node{Handle: agentquery.Handle{Value: "a-second", Provider: "source"}, Name: "Second", Path: secondSpan.Path, Span: &secondSpan},
			},
		},
		SymbolMatches: []agentquery.Result{{
			Rank: 7, Provider: "lexicon", Score: 3, Reasons: []string{"qualified name match"}, Excerpt: "first excerpt",
			Node: agentquery.Node{Handle: agentquery.Handle{Value: "z-first", Provider: "lexicon"}, Name: "First", Path: firstSpan.Path, Span: &firstSpan},
		}},
	}, nil)
	if len(converted.Nodes) != 2 || converted.Nodes[0].ID != "z-first" || converted.Nodes[1].ID != "a-second" {
		t.Fatalf("conversion reordered evidence by opaque handle: %#v", converted.Nodes)
	}
	if len(converted.RetrievalHits) != 3 {
		t.Fatalf("retrieval hit count = %d, want 3", len(converted.RetrievalHits))
	}
	firstHit, repeatedHit := converted.RetrievalHits[0], converted.RetrievalHits[2]
	if firstHit.Lane != "exact_matches" || firstHit.Rank != 1 || firstHit.Score != 9 || firstHit.Reasons[0] != "literal source match" {
		t.Fatalf("first retrieval tuple changed: %#v", firstHit)
	}
	if repeatedHit.Lane != "symbol_matches" || repeatedHit.Rank != 7 || repeatedHit.Score != 3 || repeatedHit.Reasons[0] != "qualified name match" {
		t.Fatalf("repeated retrieval tuple changed: %#v", repeatedHit)
	}
	if firstHit.Evidence != repeatedHit.Evidence {
		t.Fatalf("same canonical source did not share an evidence reference: %#v != %#v", firstHit.Evidence, repeatedHit.Evidence)
	}
	if len(converted.SourceRanges) != 2 || converted.SourceRanges[0].Path != "internal/z.go" || converted.SourceRanges[0].Text != "first excerpt" {
		t.Fatalf("ranked source evidence was reordered or stripped: %#v", converted.SourceRanges)
	}
	if len(converted.Nodes[0].Metadata) != 0 || len(converted.SourceRanges[0].Metadata) != 0 {
		t.Fatalf("retrieval metadata leaked into canonical evidence: node=%#v source=%#v", converted.Nodes[0].Metadata, converted.SourceRanges[0].Metadata)
	}
	payload, err := json.Marshal(converted)
	if err != nil {
		t.Fatal(err)
	}
	if count := strings.Count(string(payload), "first excerpt"); count != 1 {
		t.Fatalf("excerpt was serialized %d times, want 1", count)
	}
}

func TestKnowledgePreviewsKeepLowerRanksInspectable(t *testing.T) {
	documents := []knowledge.Result{
		{Handle: "doc-1", Text: "first", CodeLinks: []knowledge.CodeLink{{Value: "one"}}},
		{Handle: "doc-2", Text: "second", CodeLinks: []knowledge.CodeLink{{Value: "two"}}},
		{Handle: "doc-3", Text: "third", CodeLinks: []knowledge.CodeLink{{Value: "three"}}},
		{Handle: "doc-4", Text: "fourth", CodeLinks: []knowledge.CodeLink{{Value: "four"}}},
	}
	previewed := applyKnowledgePreviews(documents, 0, "")
	if len(previewed) != 4 || previewed[0].Text == "" {
		t.Fatalf("top document preview was lost: %+v", previewed)
	}
	for _, document := range previewed[1:] {
		if document.Handle == "" || document.Text != "" || len(document.CodeLinks) != 0 {
			t.Fatalf("lower-ranked document was dropped instead of becoming handle-only: %+v", document)
		}
	}
	full := applyKnowledgePreviews([]knowledge.Result{{Handle: "doc", Text: "full"}}, 0, "full")
	if full[0].Text != "full" {
		t.Fatalf("full detail suppressed document content: %+v", full)
	}
}

func TestRuntimeKeepsDocumentationIndependentFromCodeLanes(t *testing.T) {
	result := func(value string) agentquery.Result {
		return agentquery.Result{Node: agentquery.Node{Handle: agentquery.Handle{Value: value}}}
	}
	query := agentquery.Response{
		Mode:          "search",
		ExactMatches:  []agentquery.Result{result("exact-1"), result("exact-2")},
		SourceMatches: []agentquery.Result{result("source-1"), result("source-2")},
		SymbolMatches: []agentquery.Result{result("symbol-1"), result("symbol-2")},
	}
	documents := []knowledge.Result{{Handle: "doc-1"}, {Handle: "doc-2"}}
	recordRuntimeEvidenceCoverage(&query, documents, 2)
	if len(query.ExactMatches) != 2 || len(query.SourceMatches) != 2 || len(query.SymbolMatches) != 2 || len(documents) != 2 {
		t.Fatalf("documentation displaced code evidence: query=%+v documents=%+v", query, documents)
	}
	if query.Truncated || len(query.TruncatedLanes) != 0 {
		t.Fatalf("independent documentation lane was reported as truncated: %+v", query)
	}
	coverage := query.Coverage[len(query.Coverage)-1]
	if coverage.Lane != "document_matches" || coverage.Available != 2 || coverage.Returned != 2 || coverage.Deferred != 0 {
		t.Fatalf("document coverage = %+v", coverage)
	}
}

func TestInvestigationBudgetAppliesSemanticBoundsWithoutEmergencyTruncation(t *testing.T) {
	snapshot := investigation.Snapshot{Repository: "repo:normal"}
	ledger, err := investigation.Create(t.TempDir(), "normal", snapshot)
	if err != nil {
		t.Fatal(err)
	}
	text := strings.Repeat("normal evidence ", 120)
	reason := strings.Repeat("normal reason ", 30)
	response := investigation.Response{
		Snapshot: snapshot,
		SourceRanges: []investigation.SourceRange{{
			Path: "internal/normal.go", StartLine: 1, EndLine: 20, Text: text,
		}},
		RetrievalHits: []investigation.RetrievalHit{{
			Evidence: investigation.EvidenceRef{Kind: "source", Index: 0},
			Lane:     "source_matches", Rank: 1, Reasons: []string{reason},
		}},
	}
	bounded, truncated, err := boundInvestigationResponse(ledger, response, 8)
	if err != nil {
		t.Fatal(err)
	}
	if truncated {
		t.Fatal("semantic shaping was incorrectly reported as emergency truncation")
	}
	if len(bounded.SourceRanges[0].Text) > maxInvestigationEvidenceTextBytes+len("…") || !utf8.ValidString(bounded.SourceRanges[0].Text) {
		t.Fatalf("source evidence was not semantically bounded: %q", bounded.SourceRanges[0].Text)
	}
	if len(bounded.RetrievalHits[0].Reasons[0]) > maxInvestigationReasonTextBytes+len("…") {
		t.Fatalf("retrieval reason was not semantically bounded: %q", bounded.RetrievalHits[0].Reasons[0])
	}
}

func TestNormalLimitEightInvestigationAvoidsEmergencyCompaction(t *testing.T) {
	snapshot := investigation.Snapshot{Repository: "repo:normal-limit-eight"}
	ledger, err := investigation.Create(t.TempDir(), "normal-limit-eight", snapshot)
	if err != nil {
		t.Fatal(err)
	}
	response := investigation.Response{Snapshot: snapshot}
	for index := 0; index < 8; index++ {
		text := ""
		if index < 2 {
			text = strings.Repeat("source evidence ", 35)
		}
		response.SourceRanges = append(response.SourceRanges, investigation.SourceRange{
			Path: fmt.Sprintf("internal/source_%d.go", index), StartLine: index*10 + 1, EndLine: index*10 + 6,
			Text: text,
		})
		response.RetrievalHits = append(response.RetrievalHits, investigation.RetrievalHit{
			Evidence: investigation.EvidenceRef{Kind: "source", Index: index},
			Lane:     "source_matches", Provider: "source", Rank: index + 1,
			Reasons: []string{"prepared source BM25 match", "matched query terms"},
		})
	}
	for index := 0; index < 8; index++ {
		response.Nodes = append(response.Nodes, investigation.Node{
			ID: fmt.Sprintf("symbol-%d", index), Kind: "function",
			Label: fmt.Sprintf("pkg.Symbol%d", index), Path: fmt.Sprintf("internal/symbol_%d.go", index),
		})
		response.RetrievalHits = append(response.RetrievalHits, investigation.RetrievalHit{
			Evidence: investigation.EvidenceRef{Kind: "node", Index: index},
			Lane:     "symbol_matches", Provider: "lexicon", Rank: index + 1,
			Reasons: []string{"qualified symbol match", "implementation-level function"},
		})
	}
	for index := 0; index < 8; index++ {
		content := ""
		if index < defaultDocumentPreviewCount {
			content = strings.Repeat("architecture planning rationale ", 24)
		}
		response.Documents = append(response.Documents, investigation.Document{
			ID:  fmt.Sprintf("knowledge://docs/design_%d.md#section", index),
			URI: fmt.Sprintf("docs/design_%d.md", index), Title: fmt.Sprintf("Design rationale %d", index),
			Content: content,
		})
		response.RetrievalHits = append(response.RetrievalHits, investigation.RetrievalHit{
			Evidence: investigation.EvidenceRef{Kind: "document", Index: index},
			Lane:     "document_matches", Provider: "knowledge", Rank: index + 1,
			Reasons: []string{"documentation BM25 match", "rationale section"},
		})
	}

	bounded, truncated, err := boundInvestigationResponse(ledger, response, 8)
	if err != nil {
		t.Fatal(err)
	}
	if truncated {
		delta, deltaErr := ledger.DeltaFor(bounded)
		if deltaErr != nil {
			t.Fatal(deltaErr)
		}
		payload, _ := json.Marshal(delta)
		t.Fatalf("normal limit-eight working set triggered emergency compaction: bytes=%d hits=%d", len(payload), len(bounded.RetrievalHits))
	}
	if len(bounded.RetrievalHits) != 24 {
		t.Fatalf("normal working set lost retrieval hits: %d", len(bounded.RetrievalHits))
	}
}

func TestInvestigationDeltaBudgetBoundsLimitEightPayload(t *testing.T) {
	snapshot := investigation.Snapshot{Repository: "repo:budget"}
	ledger, err := investigation.Create(t.TempDir(), "budget", snapshot)
	if err != nil {
		t.Fatal(err)
	}
	response := investigation.Response{Snapshot: snapshot}
	for index := 0; index < 8; index++ {
		response.SourceRanges = append(response.SourceRanges, investigation.SourceRange{
			Path:      fmt.Sprintf("internal/file_%d.go", index),
			StartLine: index*10 + 1,
			EndLine:   index*10 + 8,
			Text:      strings.Repeat("évidence payload ", 800),
		})
		response.RetrievalHits = append(response.RetrievalHits, investigation.RetrievalHit{
			Evidence: investigation.EvidenceRef{Kind: "source", Index: index},
			Lane:     "source_matches",
			Provider: "source",
			Rank:     index + 1,
			Reasons:  []string{strings.Repeat("ranked reason ", 100)},
			Support:  []string{strings.Repeat("supporting detail ", 100)},
		})
	}

	bounded, truncated, err := boundInvestigationResponse(ledger, response, 8)
	if err != nil {
		t.Fatal(err)
	}
	if truncated {
		t.Fatal("semantic compaction should preserve the ranked response without emergency truncation")
	}
	if len(bounded.RetrievalHits) == 0 || bounded.RetrievalHits[0].Rank != 1 {
		t.Fatalf("highest-ranked retrieval hit was not preserved: %#v", bounded.RetrievalHits)
	}
	for index, hit := range bounded.RetrievalHits {
		if hit.Rank != index+1 {
			t.Fatalf("retrieval order changed at %d: %#v", index, bounded.RetrievalHits)
		}
	}
	for _, source := range bounded.SourceRanges {
		if len(source.Text) > maxInvestigationEvidenceTextBytes+len("…") || !utf8.ValidString(source.Text) {
			t.Fatalf("source excerpt was not compacted safely: bytes=%d valid=%v", len(source.Text), utf8.ValidString(source.Text))
		}
	}
	delta, err := ledger.DeltaFor(bounded)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(delta)
	if err != nil {
		t.Fatal(err)
	}
	if len(payload) > maxInvestigationDeltaBytes {
		t.Fatalf("serialized delta = %d bytes, budget = %d", len(payload), maxInvestigationDeltaBytes)
	}
	if count := investigationDeltaEvidenceCount(delta); count > investigationEvidenceLimit(8) {
		t.Fatalf("delta evidence count = %d, budget = %d", count, investigationEvidenceLimit(8))
	}
}

func TestExecuteSessionReturnsOnlyNewEvidenceThenPriorHandles(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "docs/design.md", "# Graph ownership\nArcana owns repository graph traversal.\n")
	request := Request{
		Request: agentquery.Request{Mode: "search", Root: root, Query: "graph ownership", Limit: 4},
		Session: "agent-session",
	}
	first, err := Execute(context.Background(), request, testOptions(root))
	if err != nil {
		t.Fatal(err)
	}
	if len(first.SourceMatches) != 0 || len(first.DocumentMatches) != 0 || first.Delta == nil {
		t.Fatalf("session response leaked full evidence: %#v", first)
	}
	if len(first.Delta.NewNodes) == 0 || len(first.Delta.NewDocuments) == 0 {
		t.Fatalf("first delta = %#v", first.Delta)
	}
	second, err := Execute(context.Background(), request, testOptions(root))
	if err != nil {
		t.Fatal(err)
	}
	if second.Delta == nil || len(second.Delta.NewNodes) != 0 || len(second.Delta.NewDocuments) != 0 {
		t.Fatalf("second delta replayed evidence: %#v", second.Delta)
	}
	if len(second.Delta.PriorNodeHandles) == 0 || len(second.Delta.PriorDocuments) == 0 {
		t.Fatalf("second delta lacks internal prior handles: %#v", second.Delta)
	}
	payload, err := json.Marshal(second.Delta)
	if err != nil {
		t.Fatal(err)
	}
	serialized := string(payload)
	if strings.Contains(serialized, "prior_node_handles") || strings.Contains(serialized, "prior_document_handles") {
		t.Fatalf("serialized delta repeated prior handles: %s", serialized)
	}
	if !strings.Contains(serialized, "prior_evidence") {
		t.Fatalf("serialized delta omitted compact prior summary: %s", serialized)
	}
}

func testOptions(root string) Options {
	return Options{
		DefaultMode: repostate.CurrentOnly,
		EnsureRepository: func(context.Context, repostate.Options) (repostate.Status, error) {
			return repostate.Status{
				Repository:              repostate.RepositoryStatus{Root: root},
				Grimoire:                repostate.ComponentStatus{Status: "current", Prepared: true},
				DeterministicQueryReady: true,
			}, nil
		},
		ExecuteQuery: func(context.Context, agentquery.Request) (agentquery.Response, error) {
			handle := agentquery.Handle{Value: "grimoire://source/source-1/node/caller", Provider: "source", Snapshot: "source-1", NodeIdentity: "caller"}
			return agentquery.Response{
				Schema:        agentquery.SchemaVersion,
				Mode:          "search",
				Snapshot:      agentquery.Snapshot{Source: "source-1"},
				SourceMatches: []agentquery.Result{{Rank: 1, Provider: "source", Kind: "function", Node: agentquery.Node{Handle: handle, Kind: "function", Name: "caller", Path: filepath.ToSlash(filepath.Join("internal", "caller.go"))}, Reasons: []string{"exact symbol"}}},
			}, nil
		},
	}
}

func writeTestFile(t *testing.T, root, relative, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
