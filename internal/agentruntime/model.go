package agentruntime

import (
	"context"

	"github.com/Lokee86/grimoire/internal/agentquery"
	"github.com/Lokee86/grimoire/internal/investigation"
	"github.com/Lokee86/grimoire/internal/knowledge"
	"github.com/Lokee86/grimoire/internal/repostate"
)

const SchemaVersion = agentquery.SchemaVersion

type Request struct {
	agentquery.Request
	Session            string         `json:"session,omitempty"`
	StateMode          repostate.Mode `json:"state_mode,omitempty"`
	IncludeDocuments   *bool          `json:"include_documents,omitempty"`
	UseDocumentVectors *bool          `json:"use_document_vectors,omitempty"`
}

type Response struct {
	Schema              string                         `json:"schema"`
	Mode                string                         `json:"mode"`
	Snapshot            agentquery.Snapshot            `json:"snapshot"`
	Preparation         *repostate.Status              `json:"preparation,omitempty"`
	ExactMatches        []agentquery.Result            `json:"exact_matches,omitempty"`
	SourceMatches       []agentquery.Result            `json:"source_matches,omitempty"`
	DocumentMatches     []knowledge.Result             `json:"document_matches,omitempty"`
	SymbolMatches       []agentquery.Result            `json:"symbol_matches,omitempty"`
	RelationshipMatches []agentquery.RelationshipMatch `json:"relationship_matches,omitempty"`
	Paths               []agentquery.Path              `json:"paths,omitempty"`
	Dependents          []agentquery.Dependent         `json:"dependents,omitempty"`
	Inspections         []agentquery.Inspection        `json:"inspections,omitempty"`
	Delta               *investigation.Delta           `json:"delta,omitempty"`
	Suggestions         []agentquery.Suggestion        `json:"suggestions,omitempty"`
	Unresolved          []agentquery.Unresolved        `json:"unresolved,omitempty"`
	Warnings            []string                       `json:"warnings,omitempty"`
	Coverage            []agentquery.LaneCoverage      `json:"coverage,omitempty"`
	DeferredExpansions  []agentquery.DeferredExpansion `json:"deferred_expansions,omitempty"`
	TruncatedLanes      []string                       `json:"truncated_lanes,omitempty"`
	Truncated           bool                           `json:"truncated,omitempty"`
}

type Options struct {
	DefaultRoot      string
	DefaultState     string
	DefaultMode      repostate.Mode
	GrimoireCommand  string
	EnsureRepository func(context.Context, repostate.Options) (repostate.Status, error)
	ExecuteQuery     func(context.Context, agentquery.Request) (agentquery.Response, error)
}
