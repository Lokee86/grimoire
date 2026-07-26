package agentruntime

import (
	"context"

	"github.com/Lokee86/grimoire/internal/agentquery"
	"github.com/Lokee86/grimoire/internal/investigation"
	"github.com/Lokee86/grimoire/internal/knowledge"
	"github.com/Lokee86/grimoire/internal/repostate"
)

const SchemaVersion = "grimoire.agent.v1"

type Request struct {
	agentquery.Request
	Session          string         `json:"session,omitempty"`
	StateMode        repostate.Mode `json:"state_mode,omitempty"`
	IncludeKnowledge *bool          `json:"include_knowledge,omitempty"`
}

type Response struct {
	Schema           string                  `json:"schema"`
	Mode             string                  `json:"mode"`
	Snapshot         agentquery.Snapshot     `json:"snapshot"`
	Preparation      *repostate.Status       `json:"preparation,omitempty"`
	Query            *agentquery.Response    `json:"query,omitempty"`
	Knowledge        []knowledge.Result      `json:"knowledge,omitempty"`
	Delta            *investigation.Delta    `json:"delta,omitempty"`
	Handles          []agentquery.Handle     `json:"handles,omitempty"`
	KnowledgeHandles []string                `json:"knowledge_handles,omitempty"`
	Suggestions      []agentquery.Suggestion `json:"suggestions,omitempty"`
	Warnings         []string                `json:"warnings,omitempty"`
	Truncated        bool                    `json:"truncated,omitempty"`
}

type Options struct {
	DefaultRoot      string
	DefaultState     string
	DefaultMode      repostate.Mode
	GrimoireCommand  string
	EnsureRepository func(context.Context, repostate.Options) (repostate.Status, error)
	ExecuteQuery     func(context.Context, agentquery.Request) (agentquery.Response, error)
}
