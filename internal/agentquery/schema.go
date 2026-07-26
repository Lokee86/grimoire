package agentquery

import "github.com/Lokee86/grimoire/internal/structure"

const SchemaVersion = "grimoire.query.v1"

type Request struct {
	Schema       string   `json:"schema"`
	Mode         string   `json:"mode"`
	Root         string   `json:"root,omitempty"`
	State        string   `json:"state,omitempty"`
	Query        string   `json:"query,omitempty"`
	Anchor       string   `json:"anchor,omitempty"`
	Target       string   `json:"target,omitempty"`
	Handles      []string `json:"handles,omitempty"`
	Limit        int      `json:"limit,omitempty"`
	Depth        int      `json:"depth,omitempty"`
	Direction    string   `json:"direction,omitempty"`
	Relations    []string `json:"relations,omitempty"`
	Adjacent     int      `json:"adjacent_context,omitempty"`
	CodeOnly     bool     `json:"code_only,omitempty"`
	LexiconFacts string   `json:"lexicon_facts,omitempty"`
	LexiconState string   `json:"lexicon_state,omitempty"`
	LexiconCmd   string   `json:"lexicon_command,omitempty"`
	ArcanaState  string   `json:"arcana_state,omitempty"`
	ArcanaCmd    string   `json:"arcana_command,omitempty"`
}

type Response struct {
	Schema      string       `json:"schema"`
	Mode        string       `json:"mode"`
	Snapshot    Snapshot     `json:"snapshot"`
	Results     []Result     `json:"results,omitempty"`
	Paths       []Path       `json:"paths,omitempty"`
	Dependents  []Dependent  `json:"dependents,omitempty"`
	Inspections []Inspection `json:"inspections,omitempty"`
	Suggestions []Suggestion `json:"suggestions,omitempty"`
	Unresolved  []Unresolved `json:"unresolved,omitempty"`
	Warnings    []string     `json:"warnings,omitempty"`
	Truncated   bool         `json:"truncated,omitempty"`
}

type Snapshot struct {
	Source    string            `json:"source"`
	Providers map[string]string `json:"providers,omitempty"`
}

type Handle struct {
	Value        string  `json:"value"`
	Provider     string  `json:"provider"`
	Snapshot     string  `json:"snapshot,omitempty"`
	NodeIdentity string  `json:"node_identity,omitempty"`
	NodeID       *uint32 `json:"node_id,omitempty"`
	Path         string  `json:"path,omitempty"`
	StartLine    int     `json:"start_line,omitempty"`
	EndLine      int     `json:"end_line,omitempty"`
}

type Range struct {
	Path        string `json:"path"`
	StartLine   int    `json:"start_line"`
	StartColumn int    `json:"start_column,omitempty"`
	EndLine     int    `json:"end_line"`
	EndColumn   int    `json:"end_column,omitempty"`
	Handle      Handle `json:"handle"`
}

type Node struct {
	Handle        Handle `json:"handle"`
	Kind          string `json:"kind,omitempty"`
	Name          string `json:"name,omitempty"`
	QualifiedName string `json:"qualified_name,omitempty"`
	Path          string `json:"path,omitempty"`
	Span          *Range `json:"span,omitempty"`
}

type Result struct {
	Rank     int      `json:"rank"`
	Provider string   `json:"provider"`
	Kind     string   `json:"kind"`
	Node     Node     `json:"node"`
	Score    float64  `json:"score,omitempty"`
	Reasons  []string `json:"reasons"`
}

type Path struct {
	Rank  int        `json:"rank"`
	Nodes []Node     `json:"nodes"`
	Steps []PathStep `json:"steps"`
}

type PathStep struct {
	From      Handle   `json:"from"`
	To        Handle   `json:"to"`
	Direction string   `json:"direction"`
	Relation  string   `json:"relation"`
	Certainty string   `json:"certainty"`
	Evidence  []string `json:"evidence,omitempty"`
	Spans     []Range  `json:"spans,omitempty"`
}

type Dependent struct {
	Depth     int      `json:"depth"`
	Direction string   `json:"direction"`
	Relation  string   `json:"relation"`
	Certainty string   `json:"certainty"`
	Node      Node     `json:"node"`
	Evidence  []string `json:"evidence,omitempty"`
	Spans     []Range  `json:"spans,omitempty"`
}

type Inspection struct {
	Handle         Handle `json:"handle"`
	Node           *Node  `json:"node,omitempty"`
	Declaration    *Range `json:"declaration,omitempty"`
	ContainingSpan Range  `json:"containing_span"`
	Source         string `json:"source"`
}

type Suggestion struct {
	Mode   string `json:"mode"`
	Anchor string `json:"anchor,omitempty"`
	Query  string `json:"query,omitempty"`
	Reason string `json:"reason"`
}

type Unresolved struct {
	Relation           string `json:"relation"`
	Expression         string `json:"expression"`
	CandidateNamespace string `json:"candidate_namespace,omitempty"`
	CandidateName      string `json:"candidate_name,omitempty"`
	Reason             string `json:"reason"`
	Span               *Range `json:"span,omitempty"`
}

func certainty(relation string) string {
	if len(relation) >= 9 && relation[:9] == "possible-" {
		return "possible"
	}
	return "definite"
}

func unresolvedFromStructure(value structure.Unresolved, sourceSnapshot string) Unresolved {
	result := Unresolved{
		Relation: value.Relation, Expression: value.Expression,
		CandidateNamespace: value.CandidateNamespace,
		CandidateName:      value.CandidateName, Reason: value.Reason,
	}
	if value.Span != nil {
		span := rangeFromStructure(*value.Span, sourceSnapshot)
		result.Span = &span
	}
	return result
}
