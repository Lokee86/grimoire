package agentdiscovery

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
)

type SourceRange struct {
	Path      string `json:"path"`
	StartLine int    `json:"start_line,omitempty"`
	EndLine   int    `json:"end_line,omitempty"`
}

type Score struct {
	Adapter                string        `json:"adapter"`
	RunID                  string        `json:"run_id"`
	CaseID                 string        `json:"case_id"`
	TotalInputTokens       int           `json:"total_input_tokens"`
	TotalOutputTokens      int           `json:"total_output_tokens"`
	RepeatedInputTokens    int           `json:"repeated_input_tokens"`
	DiscoveryCalls         int           `json:"discovery_calls"`
	ToolCalls              int           `json:"tool_calls"`
	Opened                 []SourceRange `json:"opened"`
	FirstOwnershipMS       int           `json:"time_to_first_correct_ownership_ms,omitempty"`
	EvidenceCompleteMS     int           `json:"time_to_complete_evidence_ms,omitempty"`
	IrrelevantBranches     []string      `json:"irrelevant_branches,omitempty"`
	UnsupportedConclusions []string      `json:"unsupported_conclusions,omitempty"`
	RequiredRecovered      int           `json:"required_recovered"`
	RequiredTotal          int           `json:"required_total"`
	StructuralRecovered    int           `json:"structural_recovered"`
	StructuralTotal        int           `json:"structural_total"`
	Correct                bool          `json:"correct"`
	ownershipFound         bool
	evidenceComplete       bool
}

func ScoreTranscript(entry Case, transcript Transcript) Score {
	result := Score{Adapter: transcript.Adapter, RunID: transcript.RunID, CaseID: entry.ID, RequiredTotal: len(entry.Required), StructuralTotal: len(entry.Structural)}
	events := append([]Event(nil), transcript.Events...)
	sort.SliceStable(events, func(i, j int) bool { return events[i].TimeMS < events[j].TimeMS })
	inputs, opened, irrelevant := map[string]bool{}, map[string]SourceRange{}, map[string]bool{}
	matchedRequired, matchedStructural := make([]bool, len(entry.Required)), make([]bool, len(entry.Structural))
	for _, event := range events {
		result.TotalInputTokens += event.InputTokens
		result.TotalOutputTokens += event.OutputTokens
		if key := inputKey(event); key != "" && event.InputTokens > 0 {
			if inputs[key] {
				result.RepeatedInputTokens += event.InputTokens
			}
			inputs[key] = true
		}
		if isDiscovery(event.Kind) {
			result.DiscoveryCalls++
		}
		if isTool(event.Kind) {
			result.ToolCalls++
		}
		if event.Path != "" {
			path := normalizePath(event.Path)
			rangeValue := SourceRange{Path: path, StartLine: event.StartLine, EndLine: event.EndLine}
			opened[fmt.Sprintf("%s:%d:%d", path, event.StartLine, event.EndLine)] = rangeValue
			for i, evidence := range entry.Required {
				matchedRequired[i] = matchedRequired[i] || matches(evidence, event)
			}
			for i, evidence := range entry.Structural {
				matchedStructural[i] = matchedStructural[i] || matches(evidence, event)
			}
			if !result.ownershipFound && anyMatch(entry.OwnershipEvidence, event) {
				result.FirstOwnershipMS = event.TimeMS
				result.ownershipFound = true
			}
			if !relevant(entry, event) {
				irrelevant[branchName(event)] = true
			}
		}
		if event.Branch != "" && event.Relevant != nil && !*event.Relevant {
			irrelevant[event.Branch] = true
		}
		for _, forbidden := range entry.Forbidden {
			if event.Kind == "claim" && forbidden.Pattern != "" && strings.Contains(strings.ToLower(event.Claim), strings.ToLower(forbidden.Pattern)) {
				result.UnsupportedConclusions = appendUnique(result.UnsupportedConclusions, forbidden.ID)
			}
		}
		if !result.evidenceComplete && all(matchedRequired) && all(matchedStructural) {
			result.EvidenceCompleteMS = event.TimeMS
			result.evidenceComplete = true
		}
	}
	for _, found := range matchedRequired {
		if found {
			result.RequiredRecovered++
		}
	}
	for _, found := range matchedStructural {
		if found {
			result.StructuralRecovered++
		}
	}
	for _, value := range opened {
		result.Opened = append(result.Opened, value)
	}
	for value := range irrelevant {
		if value != "" {
			result.IrrelevantBranches = append(result.IrrelevantBranches, value)
		}
	}
	sort.Slice(result.Opened, func(i, j int) bool { return sourceKey(result.Opened[i]) < sourceKey(result.Opened[j]) })
	sort.Strings(result.IrrelevantBranches)
	sort.Strings(result.UnsupportedConclusions)
	result.Correct = result.RequiredRecovered == result.RequiredTotal && result.StructuralRecovered == result.StructuralTotal && len(result.UnsupportedConclusions) == 0 && (len(entry.OwnershipEvidence) == 0 || result.ownershipFound)
	return result
}

func inputKey(event Event) string {
	if event.InputID != "" {
		return "id:" + event.InputID
	}
	if event.InputText != "" {
		return "text:" + event.InputText
	}
	if event.Path != "" {
		return "source:" + sourceKey(SourceRange{normalizePath(event.Path), event.StartLine, event.EndLine})
	}
	return ""
}

func matches(evidence Evidence, event Event) bool {
	if normalizePath(evidence.Path) != normalizePath(event.Path) {
		return false
	}
	if len(evidence.Symbols) == 0 {
		return true
	}
	for _, symbol := range evidence.Symbols {
		if symbol == event.Symbol || strings.Contains(event.Claim, symbol) {
			return true
		}
	}
	return false
}

func anyMatch(evidence []Evidence, event Event) bool {
	for _, item := range evidence {
		if matches(item, event) {
			return true
		}
	}
	return false
}
func all(values []bool) bool {
	for _, value := range values {
		if !value {
			return false
		}
	}
	return true
}
func isDiscovery(kind string) bool {
	switch kind {
	case "query", "search", "browse", "source_open":
		return true
	}
	return false
}
func isTool(kind string) bool { return kind == "tool_call" || kind == "source_open" }
func sourceKey(value SourceRange) string {
	return fmt.Sprintf("%s:%d:%d", value.Path, value.StartLine, value.EndLine)
}
func branchName(event Event) string {
	if event.Branch != "" {
		return event.Branch
	}
	return strings.Split(normalizePath(event.Path), "/")[0]
}
func appendUnique(values []string, value string) []string {
	for _, current := range values {
		if current == value {
			return values
		}
	}
	return append(values, value)
}

func relevant(entry Case, event Event) bool {
	if event.Relevant != nil {
		return *event.Relevant
	}
	if anyMatch(entry.Required, event) || anyMatch(entry.Structural, event) || anyMatch(entry.OwnershipEvidence, event) {
		return true
	}
	for _, prefix := range entry.RelevantBranches {
		if strings.HasPrefix(normalizePath(event.Path), strings.TrimSuffix(normalizePath(prefix), "/")+"/") {
			return true
		}
	}
	return false
}

func scoreFingerprint(score Score) string {
	copy := score
	copy.RunID = ""
	data := fmt.Sprintf("%+v", copy)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(data)))
}
