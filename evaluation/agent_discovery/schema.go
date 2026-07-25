// Package agentdiscovery scores recorded repository-discovery traces.
package agentdiscovery

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const SchemaVersion = 1

type Corpus struct {
	Version    int    `json:"version"`
	Repository string `json:"repository"`
	Revision   string `json:"revision"`
	Cases      []Case `json:"cases"`
}

type Case struct {
	ID                string      `json:"id"`
	Task              string      `json:"task"`
	Ownership         string      `json:"ownership_boundary"`
	OwnershipEvidence []Evidence  `json:"ownership_evidence"`
	Required          []Evidence  `json:"required_evidence"`
	Structural        []Evidence  `json:"required_structural_evidence"`
	Forbidden         []Forbidden `json:"forbidden_unsupported_conclusions"`
	Completion        []string    `json:"completion_criteria"`
	RelevantBranches  []string    `json:"relevant_branches"`
}

type Evidence struct {
	Path    string   `json:"path"`
	Symbols []string `json:"symbols,omitempty"`
	Reason  string   `json:"reason"`
}

type Forbidden struct {
	ID      string `json:"id"`
	Pattern string `json:"pattern"`
}

type Event struct {
	TimeMS       int    `json:"time_ms"`
	Kind         string `json:"kind"`
	InputTokens  int    `json:"input_tokens,omitempty"`
	OutputTokens int    `json:"output_tokens,omitempty"`
	InputID      string `json:"input_id,omitempty"`
	InputText    string `json:"input_text,omitempty"`
	Path         string `json:"path,omitempty"`
	StartLine    int    `json:"start_line,omitempty"`
	EndLine      int    `json:"end_line,omitempty"`
	Symbol       string `json:"symbol,omitempty"`
	Claim        string `json:"claim,omitempty"`
	Branch       string `json:"branch,omitempty"`
	Relevant     *bool  `json:"relevant,omitempty"`
}

type Transcript struct {
	Adapter string  `json:"adapter"`
	RunID   string  `json:"run_id"`
	CaseID  string  `json:"case_id"`
	Events  []Event `json:"events"`
}

func LoadCorpus(path string) (Corpus, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Corpus{}, err
	}
	var corpus Corpus
	if err := json.Unmarshal(data, &corpus); err != nil {
		return Corpus{}, fmt.Errorf("decode corpus: %w", err)
	}
	if corpus.Version != SchemaVersion || strings.TrimSpace(corpus.Repository) == "" {
		return Corpus{}, fmt.Errorf("corpus must use version %d and name a repository", SchemaVersion)
	}
	seen := map[string]bool{}
	for i := range corpus.Cases {
		entry := &corpus.Cases[i]
		if entry.ID == "" || entry.Task == "" || entry.Ownership == "" || len(entry.Required) == 0 || seen[entry.ID] {
			return Corpus{}, fmt.Errorf("case %q needs unique id, task, ownership boundary, and required evidence", entry.ID)
		}
		seen[entry.ID] = true
		for _, group := range [][]Evidence{entry.OwnershipEvidence, entry.Required, entry.Structural} {
			for _, evidence := range group {
				if strings.TrimSpace(evidence.Path) == "" {
					return Corpus{}, fmt.Errorf("case %q has evidence without a path", entry.ID)
				}
			}
		}
	}
	sort.Slice(corpus.Cases, func(i, j int) bool { return corpus.Cases[i].ID < corpus.Cases[j].ID })
	return corpus, nil
}

func findCase(corpus Corpus, id string) (Case, bool) {
	for _, entry := range corpus.Cases {
		if entry.ID == id {
			return entry, true
		}
	}
	return Case{}, false
}

func normalizePath(path string) string {
	return strings.TrimPrefix(filepath.ToSlash(filepath.Clean(strings.TrimSpace(path))), "./")
}
