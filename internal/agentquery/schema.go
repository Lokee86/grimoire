package agentquery

import "github.com/Lokee86/grimoire/internal/structure"

const SchemaVersion = "grimoire.discovery.v1"

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
	Detail       string   `json:"detail,omitempty"`
	LexiconFacts string   `json:"lexicon_facts,omitempty"`
	LexiconState string   `json:"lexicon_state,omitempty"`
	LexiconCmd   string   `json:"lexicon_command,omitempty"`
	ArcanaState  string   `json:"arcana_state,omitempty"`
	ArcanaCmd    string   `json:"arcana_command,omitempty"`

	PreparedSnapshot Snapshot `json:"-"`
}

type Response struct {
	Schema              string              `json:"schema"`
	Mode                string              `json:"mode"`
	Snapshot            Snapshot            `json:"snapshot"`
	ExactMatches        []Result            `json:"exact_matches,omitempty"`
	SourceMatches       []Result            `json:"source_matches,omitempty"`
	SymbolMatches       []Result            `json:"symbol_matches,omitempty"`
	RelationshipMatches []RelationshipMatch `json:"relationship_matches,omitempty"`
	Paths               []Path              `json:"paths,omitempty"`
	Dependents          []Dependent         `json:"dependents,omitempty"`
	Inspections         []Inspection        `json:"inspections,omitempty"`
	Suggestions         []Suggestion        `json:"suggestions,omitempty"`
	Unresolved          []Unresolved        `json:"unresolved,omitempty"`
	Warnings            []string            `json:"warnings,omitempty"`
	Coverage            []LaneCoverage      `json:"coverage,omitempty"`
	DeferredExpansions  []DeferredExpansion `json:"deferred_expansions,omitempty"`
	TruncatedLanes      []string            `json:"truncated_lanes,omitempty"`
	Truncated           bool                `json:"truncated,omitempty"`
}

type LaneCoverage struct {
	Lane                 string `json:"lane"`
	Available            int    `json:"available"`
	Returned             int    `json:"returned"`
	Previewed            int    `json:"previewed,omitempty"`
	Deferred             int    `json:"deferred,omitempty"`
	SuppressedDuplicates int    `json:"suppressed_duplicates,omitempty"`
}

type DeferredExpansion struct {
	Kind           string   `json:"kind"`
	CandidateCount int      `json:"candidate_count"`
	FollowUpModes  []string `json:"follow_up_modes"`
	Reason         string   `json:"reason"`
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
	Rank        int      `json:"rank"`
	Provider    string   `json:"provider"`
	Kind        string   `json:"kind"`
	Node        Node     `json:"node"`
	Excerpt     string   `json:"excerpt,omitempty"`
	DuplicateOf string   `json:"duplicate_of,omitempty"`
	Score       float64  `json:"score,omitempty"`
	Reasons     []string `json:"reasons"`
}

type RelationshipMatch struct {
	Rank        int      `json:"rank"`
	Provider    string   `json:"provider"`
	Subject     Node     `json:"subject"`
	Direction   string   `json:"direction"`
	Relation    string   `json:"relation"`
	Certainty   string   `json:"certainty,omitempty"`
	Object      Node     `json:"object"`
	Reasons     []string `json:"reasons,omitempty"`
	Evidence    []string `json:"evidence,omitempty"`
	Spans       []Range  `json:"spans,omitempty"`
	Seed        *Node    `json:"seed,omitempty"`
	SeedLane    string   `json:"seed_lane,omitempty"`
	SeedRank    int      `json:"seed_rank,omitempty"`
	SeedScore   float64  `json:"seed_score,omitempty"`
	SeedReasons []string `json:"seed_reasons,omitempty"`
}

type Path struct {
	Rank                int             `json:"rank"`
	Score               float64         `json:"score,omitempty"`
	Summary             string          `json:"summary,omitempty"`
	ContinuationHandles []string        `json:"continuation_handles,omitempty"`
	Relations           []string        `json:"relations,omitempty"`
	Evidence            []TraceEvidence `json:"evidence,omitempty"`
	Nodes               []Node          `json:"nodes,omitempty"`
	Steps               []PathStep      `json:"steps,omitempty"`
}

type TraceEvidence struct {
	Relation  string `json:"relation"`
	Path      string `json:"path,omitempty"`
	StartLine int    `json:"start_line,omitempty"`
	EndLine   int    `json:"end_line,omitempty"`
	Handle    string `json:"handle,omitempty"`
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
