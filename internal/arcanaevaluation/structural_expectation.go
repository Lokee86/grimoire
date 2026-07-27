package arcanaevaluation

import (
	"strings"

	"github.com/Lokee86/grimoire/internal/structure"
)

// StructuralExpectation describes one Arcana fact required by an evaluation
// case. Empty optional fields are wildcards. Chain is matched as an ordered
// subsequence of call-chain node names or qualified names.
type StructuralExpectation struct {
	Provider     string   `json:"provider"`
	Kind         string   `json:"kind"`
	Symbol       string   `json:"symbol,omitempty"`
	Path         string   `json:"path,omitempty"`
	Relation     string   `json:"relation,omitempty"`
	Direction    string   `json:"direction,omitempty"`
	Certainty    string   `json:"certainty,omitempty"`
	TargetSymbol string   `json:"target_symbol,omitempty"`
	TargetPath   string   `json:"target_path,omitempty"`
	Chain        []string `json:"chain,omitempty"`
	Expression   string   `json:"expression,omitempty"`
	Reason       string   `json:"reason,omitempty"`
}

func structuralExpectationMatches(expected StructuralExpectation, item structure.Evidence) bool {
	if !sameFold(expected.Provider, item.Provider) || !sameFold(expected.Kind, item.Kind) {
		return false
	}
	if expected.Symbol != "" && !evidenceHasSubject(item, expected.Symbol, expected.Path) {
		return false
	}
	if expected.Path != "" && expected.Symbol == "" && !evidenceHasPath(item, expected.Path) {
		return false
	}
	if len(expected.Chain) > 0 && !chainContains(item, expected.Chain) {
		return false
	}
	if expected.Expression != "" && !unresolvedContains(item, expected.Expression) {
		return false
	}
	if expected.Relation != "" || expected.Direction != "" || expected.Certainty != "" ||
		expected.TargetSymbol != "" || expected.TargetPath != "" {
		return evidenceHasRelatedTarget(item, expected)
	}
	return true
}

func evidenceHasSubject(item structure.Evidence, symbol, path string) bool {
	if item.Node != nil && nodeMatches(*item.Node, symbol, path) {
		return true
	}
	if item.Chain != nil {
		for _, node := range item.Chain.Nodes {
			if nodeMatches(node, symbol, path) {
				return true
			}
		}
	}
	return false
}

func evidenceHasPath(item structure.Evidence, path string) bool {
	if item.Node != nil && nodeMatches(*item.Node, "", path) {
		return true
	}
	if item.Chain != nil {
		for _, node := range item.Chain.Nodes {
			if nodeMatches(node, "", path) {
				return true
			}
		}
	}
	for _, related := range item.Relationships {
		if nodeMatches(related.Node, "", path) {
			return true
		}
	}
	for _, dependent := range item.Dependents {
		if nodeMatches(dependent.Node, "", path) {
			return true
		}
	}
	return false
}

func evidenceHasRelatedTarget(item structure.Evidence, expected StructuralExpectation) bool {
	for _, related := range item.Relationships {
		if expected.Relation != "" && !sameFold(expected.Relation, related.Relation) {
			continue
		}
		if expected.Direction != "" && !sameFold(expected.Direction, related.Direction) {
			continue
		}
		if expected.Certainty != "" && !sameFold(expected.Certainty, related.Certainty) {
			continue
		}
		if nodeMatches(related.Node, expected.TargetSymbol, expected.TargetPath) {
			return true
		}
	}
	for _, dependent := range item.Dependents {
		if expected.Relation != "" && !sameFold(expected.Relation, "impact") {
			continue
		}
		if nodeMatches(dependent.Node, expected.TargetSymbol, expected.TargetPath) {
			return true
		}
	}
	if item.Chain != nil {
		for index, node := range item.Chain.Nodes {
			if !nodeMatches(node, expected.TargetSymbol, expected.TargetPath) {
				continue
			}
			if expected.Relation == "" {
				return true
			}
			if index > 0 && index-1 < len(item.Chain.Relations) && sameFold(expected.Relation, item.Chain.Relations[index-1]) {
				return true
			}
		}
	}
	return false
}

func chainContains(item structure.Evidence, expected []string) bool {
	if item.Chain == nil || len(expected) == 0 {
		return false
	}
	position := 0
	for _, node := range item.Chain.Nodes {
		if nodeNameMatches(node, expected[position]) {
			position++
			if position == len(expected) {
				return true
			}
		}
	}
	return false
}

func unresolvedContains(item structure.Evidence, expected string) bool {
	for _, unresolved := range item.Unresolved {
		if strings.Contains(strings.ToLower(unresolved.Expression), strings.ToLower(expected)) ||
			sameFold(unresolved.CandidateName, expected) {
			return true
		}
	}
	return false
}

func nodeMatches(node structure.Node, symbol, path string) bool {
	if symbol != "" && !nodeNameMatches(node, symbol) {
		return false
	}
	if path == "" {
		return true
	}
	actual := node.Path
	if actual == "" && node.Span != nil {
		actual = node.Span.Path
	}
	return seedPathKey(actual) == seedPathKey(path)
}

func nodeNameMatches(node structure.Node, expected string) bool {
	return sameFold(node.Name, expected) || sameFold(node.QualifiedName, expected)
}

func sameFold(left, right string) bool {
	return strings.EqualFold(strings.TrimSpace(left), strings.TrimSpace(right))
}
