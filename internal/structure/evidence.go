package structure

import "github.com/Lokee86/grimoire/internal/evidence"

// Span identifies a repository source range.
type Span struct {
	Path        string `json:"path"`
	StartLine   int    `json:"start_line"`
	StartColumn int    `json:"start_column,omitempty"`
	EndLine     int    `json:"end_line"`
	EndColumn   int    `json:"end_column,omitempty"`
}

// Node is a provider-neutral structural symbol reference. Identity is durable
// when supplied by Lexicon; NodeID is snapshot-local when supplied by Arcana.
type Node struct {
	Identity      string  `json:"identity,omitempty"`
	NodeID        *uint32 `json:"node_id,omitempty"`
	Kind          string  `json:"kind,omitempty"`
	Name          string  `json:"name,omitempty"`
	QualifiedName string  `json:"qualified_name,omitempty"`
	Path          string  `json:"path,omitempty"`
	Span          *Span   `json:"span,omitempty"`
}

// Relationship records one directed structural relation around an evidence
// node. Certainty is "definite" or "possible" when the provider distinguishes it.
type Relationship struct {
	Direction string `json:"direction"`
	Relation  string `json:"relation"`
	Certainty string `json:"certainty,omitempty"`
	Node      Node   `json:"node"`
}

// DepthNode records a graph node and its traversal distance from the subject.
type DepthNode struct {
	Depth int  `json:"depth"`
	Node  Node `json:"node"`
}

// Path records an ordered graph path and the relation between each node pair.
type Path struct {
	Depth     int      `json:"depth"`
	Nodes     []Node   `json:"nodes"`
	Relations []string `json:"relations"`
}

// Unresolved records one unresolved structural reference owned by a node.
type Unresolved struct {
	Relation           string `json:"relation"`
	Expression         string `json:"expression"`
	CandidateNamespace string `json:"candidate_namespace,omitempty"`
	CandidateName      string `json:"candidate_name,omitempty"`
	Reason             string `json:"reason"`
	Span               *Span  `json:"span,omitempty"`
}

// ProviderState identifies the immutable provider snapshot used to produce
// retained structural evidence.
type ProviderState struct {
	Provider string `json:"provider"`
	Snapshot string `json:"snapshot"`
}

// Evidence is a bounded, inspectable structural fact retained in a context
// package. Kind identifies the concrete shape: symbol, operational_role,
// impact, call_chain, or unresolved.
type Evidence struct {
	Provider      string               `json:"provider"`
	Kind          string               `json:"kind"`
	Rank          int                  `json:"rank"`
	Score         float64              `json:"score,omitempty"`
	Reasons       []string             `json:"reasons,omitempty"`
	Node          *Node                `json:"node,omitempty"`
	Summary       string               `json:"summary,omitempty"`
	Relationships []Relationship       `json:"relationships,omitempty"`
	Dependents    []DepthNode          `json:"dependents,omitempty"`
	Chain         *Path                `json:"chain,omitempty"`
	Unresolved    []Unresolved         `json:"unresolved,omitempty"`
	Truncated     bool                 `json:"truncated,omitempty"`
	Context       *evidence.Descriptor `json:"context,omitempty"`
}
