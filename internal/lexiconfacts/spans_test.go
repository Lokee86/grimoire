package lexiconfacts

import "testing"

func TestSourceSpansReturnsDeclarationBoundariesAndOmitsFunctionOwnedDetails(t *testing.T) {
	corpus := &Corpus{facts: library{nodes: map[string]Node{
		"function": {
			ID: "function", Kind: "function", Name: "Resolve", Path: "resolve.go",
			Span: &Span{Path: "resolve.go", StartLine: 10, EndLine: 30},
		},
		"closure": {
			ID: "closure", Kind: "function", Name: "closure@15:2", Path: "resolve.go",
			Span: &Span{Path: "resolve.go", StartLine: 15, EndLine: 18},
		},
		"type": {
			ID: "type", Kind: "type", Name: "Resolver", Path: "resolver.go",
			Span: &Span{Path: "resolver.go", StartLine: 3, EndLine: 40},
		},
		"method": {
			ID: "method", Kind: "method", Name: "Run", Path: "resolver.go", Owner: "type",
			Span: &Span{Path: "resolver.go", StartLine: 12, EndLine: 20},
		},
		"external": {
			ID: "external", Kind: "function", Name: "Error", Path: "@stdlib/errors",
			Span: &Span{Path: "@stdlib/errors", StartLine: 1, EndLine: 1},
		},
	}, edges: []Edge{{Source: "function", Target: "closure", Relation: "defines"}}}}

	spans := corpus.SourceSpans()
	if len(spans) != 3 {
		t.Fatalf("source spans = %+v, want function, type, and method", spans)
	}
	if spans[0].Path != "resolve.go" || spans[0].Name != "Resolve" {
		t.Fatalf("unexpected first span: %+v", spans[0])
	}
	if spans[1].Path != "resolver.go" || spans[1].Name != "Resolver" {
		t.Fatalf("unexpected second span: %+v", spans[1])
	}
	if spans[2].Path != "resolver.go" || spans[2].Name != "Run" {
		t.Fatalf("unexpected third span: %+v", spans[2])
	}
}
