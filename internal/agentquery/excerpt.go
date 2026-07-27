package agentquery

import (
	"strings"
	"unicode/utf8"

	"github.com/Lokee86/grimoire/internal/index"
	"github.com/Lokee86/grimoire/internal/structure"
)

const maxDiscoveryExcerptBytes = 1200

func compactExcerpt(text string) string {
	text = strings.TrimSpace(strings.ReplaceAll(text, "\r\n", "\n"))
	if len(text) <= maxDiscoveryExcerptBytes {
		return text
	}
	cut := maxDiscoveryExcerptBytes
	for cut > 0 && !utf8.RuneStart(text[cut]) {
		cut--
	}
	text = strings.TrimSpace(text[:cut])
	if boundary := strings.LastIndexAny(text, "\n.!?; }"); boundary >= maxDiscoveryExcerptBytes/2 {
		text = strings.TrimSpace(text[:boundary+1])
	}
	return text + "…"
}

func (engine *Engine) nodeExcerpt(node structure.Node) string {
	if node.Span == nil || node.Path == "" {
		return ""
	}
	source, _, err := sourceRange(
		engine.source,
		node.Path,
		node.Span.StartLine,
		node.Span.EndLine,
		0,
	)
	if err != nil {
		return ""
	}
	return compactExcerpt(source)
}

func chunkExcerpt(chunk index.Chunk) string {
	return compactExcerpt(chunk.Text)
}
