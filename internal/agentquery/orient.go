package agentquery

import (
	"path/filepath"
	"strings"

	"github.com/Lokee86/grimoire/internal/structure"
)

func (engine *Engine) orient(request Request, response *Response) error {
	limit := request.Limit
	seen := make(map[string]bool)
	add := func(result Result) {
		key := handleKey(result.Node.Handle)
		if seen[key] || len(response.Results) >= limit {
			return
		}
		seen[key] = true
		result.Rank = len(response.Results) + 1
		response.Results = append(response.Results, result)
	}

	if engine.lexicon != nil {
		var nodes []structure.Node
		if strings.TrimSpace(request.Query) != "" {
			for _, match := range engine.lexicon.Find(request.Query, limit) {
				nodes = append(nodes, match.Node)
			}
		} else {
			nodes = engine.lexicon.Anchors(max(1, limit/2))
		}
		for _, value := range nodes {
			add(Result{
				Provider: "lexicon", Kind: value.Kind,
				Node:    engine.node("lexicon", engine.lexiconSnapshot, value),
				Reasons: []string{"compact Lexicon symbol or contract anchor"},
			})
		}
	}

	for _, file := range engine.source.Files {
		if len(file.Chunks) == 0 {
			continue
		}
		kind, reason := classifyPath(file.Path)
		if kind == "file" && len(response.Results) >= max(1, limit/2) {
			continue
		}
		first, last := file.Chunks[0], file.Chunks[len(file.Chunks)-1]
		path := normalizePath(file.Path)
		span := Range{
			Path: path, StartLine: first.StartLine, EndLine: last.EndLine,
			Handle: sourceHandle(engine.source.Identity(), path, first.StartLine, last.EndLine),
		}
		add(Result{
			Provider: "source", Kind: kind,
			Node:    Node{Handle: span.Handle, Kind: kind, Name: filepath.Base(path), Path: path, Span: &span},
			Reasons: []string{reason},
		})
	}

	for _, result := range response.Results {
		response.Suggestions = append(response.Suggestions,
			Suggestion{Mode: "inspect", Anchor: result.Node.Handle.Value, Reason: "inspect exact source for this anchor"},
		)
		if result.Provider == "lexicon" || result.Provider == "arcana" {
			response.Suggestions = append(response.Suggestions,
				Suggestion{Mode: "trace", Anchor: result.Node.Handle.Value, Reason: "expand bounded structural paths"},
			)
		}
		if len(response.Suggestions) >= min(6, limit) {
			break
		}
	}
	if request.Query == "" {
		response.Suggestions = append(response.Suggestions,
			Suggestion{Mode: "search", Query: "<specific symbol, contract, or behavior>", Reason: "narrow the repository with deterministic evidence"},
		)
	}
	response.Truncated = len(response.Results) == limit
	return nil
}
