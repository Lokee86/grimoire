package repostate

import "context"

// Mode controls whether Ensure may invoke repository preparation commands.
type Mode string

const (
	CurrentOnly     Mode = "current-only"
	RefreshIfNeeded Mode = "refresh-if-needed"
	ForceRefresh    Mode = "force-refresh"
)

// CommandRunner is injectable so callers and tests can control the existing
// Lexicon, Arcana, and Grimoire command boundaries without owning analyzers.
type CommandRunner func(context.Context, string, ...string) error

type Options struct {
	Root string
	Mode Mode

	LexiconState  string
	ArcanaState   string
	GrimoireState string

	LexiconCommand  string
	ArcanaCommand   string
	GrimoireCommand string
	Run             CommandRunner
}

type Status struct {
	Version                 int              `json:"version"`
	Mode                    Mode             `json:"mode"`
	Repository              RepositoryStatus `json:"repository"`
	Lexicon                 ComponentStatus  `json:"lexicon"`
	Arcana                  ComponentStatus  `json:"arcana"`
	Grimoire                ComponentStatus  `json:"grimoire"`
	Vectors                 VectorStatus     `json:"vectors"`
	Actions                 []Action         `json:"actions,omitempty"`
	Warnings                []string         `json:"warnings,omitempty"`
	DeterministicQueryReady bool             `json:"deterministic_query_ready"`
	ElapsedMS               int64            `json:"elapsed_ms"`
	Error                   string           `json:"error,omitempty"`
}

type RepositoryStatus struct {
	Root              string `json:"root"`
	GitHead           string `json:"git_head,omitempty"`
	GitDirty          bool   `json:"git_dirty"`
	SourceFingerprint string `json:"source_fingerprint,omitempty"`
	GitAvailable      bool   `json:"git_available"`
}

type ComponentStatus struct {
	Status       string   `json:"status"`
	Snapshot     string   `json:"snapshot,omitempty"`
	Expected     string   `json:"expected_snapshot,omitempty"`
	StaleReasons []string `json:"stale_reasons,omitempty"`
	Prepared     bool     `json:"prepared"`
}

type VectorStatus struct {
	Status   string `json:"status"`
	Snapshot string `json:"snapshot,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

type Action struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	ElapsedMS int64  `json:"elapsed_ms"`
	Error     string `json:"error,omitempty"`
}
