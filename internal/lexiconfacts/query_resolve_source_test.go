package lexiconfacts

import "testing"

func TestResolveSourcePrefersExactDeclarationSpan(t *testing.T) {
	path := "pkg/example.go"
	corpus := &Corpus{facts: library{nodes: map[string]Node{
		"function": {
			ID: "function", Kind: "function", Name: "target", Path: path,
			Span: &Span{Path: path, StartLine: 10, EndLine: 20},
		},
		"variable": {
			ID: "variable", Kind: "variable", Name: "target", Path: path,
			Span: &Span{Path: path, StartLine: 10, EndLine: 10},
		},
	}}}

	resolved := corpus.ResolveSource(path, 10, 20, 8)
	if len(resolved) != 1 || resolved[0].Identity != "function" {
		t.Fatalf("resolved source anchors = %+v", resolved)
	}
}

func TestResolveSourceKeepsExactLocalRange(t *testing.T) {
	path := "pkg/example.go"
	corpus := &Corpus{facts: library{nodes: map[string]Node{
		"function": {
			ID: "function", Kind: "function", Name: "target", Path: path,
			Span: &Span{Path: path, StartLine: 10, EndLine: 20},
		},
		"local": {
			ID: "local", Kind: "local", Name: "value", Path: path,
			Span: &Span{Path: path, StartLine: 14, EndLine: 14},
		},
	}}}

	resolved := corpus.ResolveSource(path, 14, 14, 8)
	if len(resolved) != 1 || resolved[0].Identity != "local" {
		t.Fatalf("resolved local source anchor = %+v", resolved)
	}
}
