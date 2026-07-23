package lexiconfacts

import (
	"github.com/Lokee86/grimoire/internal/retrieve"
	"github.com/Lokee86/grimoire/internal/structure"
)

type Span struct {
	Path      string `json:"path"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

type Node struct {
	ID            string `json:"id"`
	Kind          string `json:"kind"`
	Name          string `json:"name"`
	Path          string `json:"path"`
	QualifiedName string `json:"qualified_name"`
	Owner         string `json:"owner,omitempty"`
	Span          *Span  `json:"span,omitempty"`
}

type Edge struct {
	Source   string `json:"source"`
	Target   string `json:"target"`
	Relation string `json:"relation"`
}

type library struct {
	nodes map[string]Node
	edges []Edge
}

type recordHeader struct {
	Record string `json:"record"`
}

type scoredNode struct {
	node    Node
	score   float64
	reasons []string
}

type Corpus struct {
	facts library
}

// Result preserves both source candidates and the structural facts that caused
// Lexicon to select them. Seeds are the directly matched symbols used for
// bounded Arcana graph queries.
type Result struct {
	Candidates []retrieve.Candidate
	Evidence   []structure.Evidence
	Seeds      []structure.Node
}
