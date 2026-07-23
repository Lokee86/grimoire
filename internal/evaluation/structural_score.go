package evaluation

import (
	"strings"

	"github.com/Lokee86/grimoire/internal/structure"
)

func scoreStructuralGroup(group []StructuralExpectation, stages Stages) []StructuralEvidenceStatus {
	result := make([]StructuralEvidenceStatus, 0, len(group))
	for _, expected := range group {
		status := StructuralEvidenceStatus{
			Evidence:  expected,
			Produced:  structuralEvidencePresent(expected, stages.StructuralProduced),
			Composed:  structuralEvidencePresent(expected, stages.StructuralComposed),
			Assembled: structuralEvidencePresent(expected, stages.StructuralAssembled),
			Included:  structuralEvidencePresent(expected, stages.StructuralIncluded),
		}
		switch {
		case status.Included:
		case status.Assembled:
			status.FailureStage = FailureStructuralBudgetFittingLoss
		case status.Composed:
			status.FailureStage = FailureStructuralAssemblyLoss
		case status.Produced:
			status.FailureStage = FailureStructuralCompositionLoss
		default:
			status.FailureStage = FailureStructuralProviderMiss
		}
		result = append(result, status)
	}
	return result
}

func structuralEvidencePresent(expected StructuralExpectation, evidence []structure.Evidence) bool {
	for _, item := range evidence {
		if structuralExpectationMatches(expected, item) {
			return true
		}
	}
	return false
}

func structuralMatchesAny(group []StructuralExpectation, item structure.Evidence) bool {
	for _, expected := range group {
		if structuralExpectationMatches(expected, item) {
			return true
		}
	}
	return false
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
		if !evidenceHasRelatedTarget(item, expected) {
			return false
		}
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
	return filepathKey(actual) == filepathKey(path)
}

func nodeNameMatches(node structure.Node, expected string) bool {
	return sameFold(node.Name, expected) || sameFold(node.QualifiedName, expected)
}

func sameFold(left, right string) bool {
	return strings.EqualFold(strings.TrimSpace(left), strings.TrimSpace(right))
}

func structuralRecall(statuses []StructuralEvidenceStatus) float64 {
	if len(statuses) == 0 {
		return 0
	}
	included := 0
	for _, status := range statuses {
		if status.Included {
			included++
		}
	}
	return float64(included) / float64(len(statuses))
}

func requiredSatisfied(statuses []EvidenceStatus) bool {
	return len(statuses) == 0 || recall(statuses) == 1
}

func structuralRequiredSatisfied(statuses []StructuralEvidenceStatus) bool {
	return len(statuses) == 0 || structuralRecall(statuses) == 1
}
