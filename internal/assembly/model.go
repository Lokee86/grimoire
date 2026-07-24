package assembly

import (
	"github.com/Lokee86/grimoire/internal/queryshape"
	"github.com/Lokee86/grimoire/internal/retrieve"
	"github.com/Lokee86/grimoire/internal/structure"
)

// Decision records why adaptive assembly stopped before package compilation.
type Decision struct {
	Scope                queryshape.Scope `json:"scope"`
	CandidatesConsidered int              `json:"candidates_considered"`
	CandidatesSelected   int              `json:"candidates_selected"`
	CandidateTokens      int              `json:"candidate_tokens"`
	StructuralConsidered int              `json:"structural_considered"`
	StructuralSelected   int              `json:"structural_selected"`
	RegionsRepresented   []string         `json:"regions_represented,omitempty"`
	RolesRepresented     []string         `json:"roles_represented,omitempty"`
	GroupsRepresented    int              `json:"groups_represented,omitempty"`
	StopReason           string           `json:"stop_reason"`
}

// Result is the bounded evidence set passed to the exact-budget compiler.
type Result struct {
	Candidates []retrieve.Candidate
	Structural []structure.Evidence
	Decision   Decision
}
