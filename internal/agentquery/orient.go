package agentquery

import (
	"path/filepath"
	"strings"

	"github.com/Lokee86/grimoire/internal/structure"
)

func (engine *Engine) orient(request Request, response *Response) error {
	if engine.lexicon != nil {
		var nodes []structure.Node
		if strings.TrimSpace(request.Query) != "" {
			for _, match := range engine.lexicon.Find(request.Query, request.Limit) {
				nodes = append(nodes, match.Node)
			}
		} else {
			nodes = engine.lexicon.Anchors(request.Limit)
		}
		for _, value := range nodes {
			if isDocumentationPath(value.Path) {
				continue
			}
			response.SymbolMatches = append(response.SymbolMatches, Result{
				Rank:     len(response.SymbolMatches) + 1,
				Provider: "lexicon",
				Kind:     value.Kind,
				Node:     engine.node("lexicon", engine.lexiconSnapshot, value),
				Reasons:  []string{"compact Lexicon symbol or contract anchor"},
			})
			if len(response.SymbolMatches) == request.Limit {
				break
			}
		}
		markLaneTruncated(response, "symbol_matches", len(nodes) > len(response.SymbolMatches))
	}

	for _, file := range engine.source.Files {
		if len(file.Chunks) == 0 || isDocumentationPath(file.Path) {
			continue
		}
		kind, reason := classifyPath(file.Path)
		first, last := file.Chunks[0], file.Chunks[len(file.Chunks)-1]
		path := normalizePath(file.Path)
		span := Range{
			Path:      path,
			StartLine: first.StartLine,
			EndLine:   last.EndLine,
			Handle:    sourceHandle(engine.source.Identity(), path, first.StartLine, last.EndLine),
		}
		response.SourceMatches = append(response.SourceMatches, Result{
			Rank:     len(response.SourceMatches) + 1,
			Provider: "source",
			Kind:     kind,
			Node: Node{
				Handle: span.Handle,
				Kind:   kind,
				Name:   filepath.Base(path),
				Path:   path,
				Span:   &span,
			},
			Reasons: []string{reason},
		})
		if len(response.SourceMatches) == request.Limit {
			markLaneTruncated(response, "source_matches", true)
			break
		}
	}

	appendSuggestions := func(results []Result, structural bool) {
		for _, result := range results {
			response.Suggestions = append(response.Suggestions, Suggestion{
				Mode:   "inspect",
				Anchor: result.Node.Handle.Value,
				Reason: "inspect exact evidence for this anchor",
			})
			if structural {
				response.Suggestions = append(response.Suggestions, Suggestion{
					Mode:   "trace",
					Anchor: result.Node.Handle.Value,
					Reason: "expand bounded structural paths",
				})
			}
			if len(response.Suggestions) >= min(6, request.Limit) {
				return
			}
		}
	}
	appendSuggestions(response.SymbolMatches, true)
	if len(response.Suggestions) < min(6, request.Limit) {
		appendSuggestions(response.SourceMatches, false)
	}
	if request.Query == "" {
		response.Suggestions = append(response.Suggestions, Suggestion{
			Mode:   "search",
			Query:  "<specific symbol, contract, behavior, or literal>",
			Reason: "discover evidence in independent source, document, symbol, exact, and relationship lanes",
		})
	}
	return nil
}
