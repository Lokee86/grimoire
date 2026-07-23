package queryshape

import (
	"github.com/Lokee86/grimoire/internal/retrieve"
	"github.com/Lokee86/grimoire/internal/structure"
)

// Level is a deterministic qualitative classification emitted by query analysis.
type Level string

const (
	LevelLow    Level = "low"
	LevelMedium Level = "medium"
	LevelHigh   Level = "high"
)

func ValidLevel(level Level) bool {
	return level == LevelLow || level == LevelMedium || level == LevelHigh
}

// Scope is the three-tier retrieval and assembly policy classification.
type Scope string

const (
	ScopeFocused     Scope = "focused"
	ScopeBounded     Scope = "bounded"
	ScopeExploratory Scope = "exploratory"
)

func ValidScope(scope Scope) bool {
	return scope == ScopeFocused || scope == ScopeBounded || scope == ScopeExploratory
}

// Input contains query and retrieval observations available before curation.
type Input struct {
	Query           string
	RequestedBudget int
	Exact           []retrieve.Candidate
	Ranked          []retrieve.Candidate
	Candidates      []retrieve.Candidate
	Structural      []structure.Evidence
}

// Profile describes the observed scope and certainty of a query without
// changing retrieval, curation, or package compilation.
type Profile struct {
	ExactSymbolMatches  int      `json:"exact_symbol_matches"`
	ExactPathMatches    int      `json:"exact_path_matches"`
	ExactErrorMatches   int      `json:"exact_error_matches"`
	RecognizedTaskTerms []string `json:"recognized_task_terms,omitempty"`
	MatchedSubsystems   []string `json:"matched_subsystems,omitempty"`
	MatchedGraphRegions []string `json:"matched_graph_regions,omitempty"`
	TopScoreGap         float64  `json:"top_score_gap"`
	CandidateDispersion float64  `json:"candidate_dispersion"`
	Specificity         Level    `json:"specificity"`
	Breadth             Level    `json:"breadth"`
	Ambiguity           Level    `json:"ambiguity"`
}

// RetrievalPolicy is the non-authoritative policy recommendation derived from
// a Profile. Shadow remains true until assembly consumes this contract.
type RetrievalPolicy struct {
	Shadow               bool     `json:"shadow"`
	Scope                Scope    `json:"scope"`
	BudgetMode           string   `json:"budget_mode"`
	TargetTokens         int      `json:"target_tokens"`
	MaximumTokens        int      `json:"maximum_tokens"`
	ExpansionRadius      int      `json:"expansion_radius"`
	RequiredEvidence     []string `json:"required_evidence_types"`
	DiversityRequirement int      `json:"diversity_requirement"`
	StopConditions       []string `json:"stop_conditions"`
}
