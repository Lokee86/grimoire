package agentquery

import (
	"strings"
	"testing"
)

func TestSearchUsesCompactDefaults(t *testing.T) {
	search := normalizeRequest(Request{Mode: "search"})
	narrow := normalizeRequest(Request{Mode: "search", Breadth: "narrow"})
	orient := normalizeRequest(Request{Mode: "orient"})
	trace := normalizeRequest(Request{Mode: "trace"})
	if search.Limit != 6 || search.Breadth != "balanced" || narrow.Limit != 4 ||
		narrow.Breadth != "narrow" || narrow.Detail != "handles" || orient.Limit != 6 || trace.Limit != 8 {
		t.Fatalf(
			"unexpected defaults: search=%d/%s narrow=%d/%s/%s orient=%d trace=%d",
			search.Limit,
			search.Breadth,
			narrow.Limit,
			narrow.Breadth,
			narrow.Detail,
			orient.Limit,
			trace.Limit,
		)
	}
}

func TestHandleOnlySearchShapeDefersSourceExpansion(t *testing.T) {
	handle := Handle{Provider: "source", Value: "source-owner", Path: "db/owner.cc", StartLine: 10, EndLine: 20}
	results := []Result{{
		Provider: "exact",
		Excerpt:  "full implementation evidence",
		Node: Node{
			Handle: handle,
			Path:   handle.Path,
			Span:   &Range{Path: handle.Path, StartLine: 10, EndLine: 20, Handle: handle},
		},
	}}
	if previewed := applyResultPreviews(results, "handles"); previewed != 0 {
		t.Fatalf("handle-only preview count = %d, want 0", previewed)
	}
	if results[0].Excerpt != "" || results[0].Node.Span != nil {
		t.Fatalf("handle-only result retained expanded evidence: %+v", results[0])
	}
	if results[0].Node.Handle.Path != handle.Path || results[0].Node.Handle.StartLine != 10 || results[0].Node.Handle.EndLine != 20 {
		t.Fatalf("handle-only result lost inspectable range identity: %+v", results[0].Node.Handle)
	}
}

func TestSearchBreadthValidation(t *testing.T) {
	if err := validateRequest(normalizeRequest(Request{Mode: "search", Query: "owner", Breadth: "wide"})); err == nil {
		t.Fatal("invalid search breadth was accepted")
	}
	if err := validateRequest(normalizeRequest(Request{Mode: "inspect", Anchor: "source:handle", Breadth: "narrow"})); err == nil {
		t.Fatal("breadth outside search was accepted")
	}
}

func TestNarrowSearchBudgetInterleavesAndDeduplicatesLanes(t *testing.T) {
	result := func(provider, value, path, name string, start, end int) Result {
		handle := Handle{Provider: provider, Value: value, Path: path, StartLine: start, EndLine: end}
		span := Range{Path: path, StartLine: start, EndLine: end, Handle: handle}
		return Result{Provider: provider, Node: Node{Handle: handle, Path: path, Name: name, Span: &span}}
	}
	response := Response{
		ExactMatches: []Result{
			result("source", "exact-owner", "db/owner.cc", "Owner", 10, 20),
			result("source", "exact-api", "include/api.h", "Pause", 5, 8),
		},
		SymbolMatches: []Result{
			result("lexicon", "symbol-owner", "db/owner.cc", "Owner", 10, 20),
			result("lexicon", "symbol-test", "db/owner_test.cc", "PauseTest", 30, 45),
		},
		SourceMatches: []Result{
			result("source", "source-helper", "db/helper.cc", "Schedule", 40, 55),
		},
	}

	suppressed := applyNarrowSearchBudget(&response, 3)
	if total := len(response.ExactMatches) + len(response.SymbolMatches) + len(response.SourceMatches); total != 3 {
		t.Fatalf("narrow result total = %d, want 3: %+v", total, response)
	}
	if len(response.ExactMatches) != 1 || len(response.SymbolMatches) != 1 || len(response.SourceMatches) != 1 {
		t.Fatalf("narrow lane selection was not interleaved: %+v", response)
	}
	if suppressed["symbol_matches"] != 1 {
		t.Fatalf("overlapping symbol suppression = %d, want 1", suppressed["symbol_matches"])
	}
}

func TestDiscoveryExcerptIsBounded(t *testing.T) {
	text := strings.Repeat("abcdefghij ", 100)
	excerpt := compactExcerpt(text)
	if len(excerpt) > maxDiscoveryExcerptBytes+len("…") || !strings.HasSuffix(excerpt, "…") {
		t.Fatalf("excerpt was not compacted: bytes=%d value=%q", len(excerpt), excerpt)
	}
}

func TestRelationshipBucketsAreInterleavedBySeed(t *testing.T) {
	node := func(value string) Node {
		return Node{Handle: Handle{Value: value, Provider: "lexicon", NodeIdentity: value}}
	}
	buckets := [][]RelationshipMatch{
		{
			{Subject: node("seed-a"), Direction: "outgoing", Relation: "calls", Object: node("a-1"), SeedLane: "exact_matches"},
			{Subject: node("seed-a"), Direction: "outgoing", Relation: "calls", Object: node("a-2"), SeedLane: "exact_matches"},
		},
		{
			{Subject: node("seed-b"), Direction: "outgoing", Relation: "uses", Object: node("b-1"), SeedLane: "symbol_matches"},
			{Subject: node("seed-b"), Direction: "outgoing", Relation: "uses", Object: node("b-2"), SeedLane: "symbol_matches"},
		},
	}

	matches, truncated := interleaveRelationshipBuckets(buckets, 2)
	if !truncated || len(matches) != 2 {
		t.Fatalf("interleaved matches = %+v truncated=%v", matches, truncated)
	}
	if matches[0].Object.Handle.Value != "a-1" || matches[1].Object.Handle.Value != "b-1" {
		t.Fatalf("first seed monopolized relationship budget: %+v", matches)
	}
	if matches[0].Rank != 1 || matches[1].Rank != 2 {
		t.Fatalf("relationship ranks were not assigned after interleaving: %+v", matches)
	}
}
