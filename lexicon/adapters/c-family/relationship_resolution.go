package main

import "strings"

func resolveInheritance(facts *factSet, index declarationIndex, observation inheritanceObservation) {
	candidates := resolveDeclarations(index, observation.Candidate, observation.SourceScope, observation.Path, func(declaration *declaration) bool {
		return declaration.Kind == "type"
	})
	if len(candidates) == 1 {
		facts.addEdge(observation.Path, map[string]any{
			"owner": observation.Path, "record": "edge", "relation": "extends", "source": observation.SourceID,
			"span": observation.Span.record(), "target": candidates[0].ID,
		})
		return
	}
	facts.addUnresolved(observation.Path, unresolvedRecord(observation.SourceID, "extends", observation.Expression, resolutionReason(candidates), observation.Path, observation.Span))
}

func resolveCall(facts *factSet, index declarationIndex, observation callObservation) {
	candidates := resolveDeclarations(index, observation.Candidate, observation.SourceScope, observation.Path, func(declaration *declaration) bool {
		return declaration.Callable
	})
	if len(candidates) == 1 {
		facts.addEdge(observation.Path, map[string]any{
			"owner": observation.Path, "record": "edge", "relation": "calls", "source": observation.SourceID,
			"span": observation.Span.record(), "target": candidates[0].ID,
		})
		return
	}
	if len(candidates) > 1 {
		for _, candidate := range candidates {
			facts.addEdge(observation.Path, map[string]any{
				"owner": observation.Path, "record": "edge", "relation": "possible-calls", "source": observation.SourceID,
				"span": observation.Span.record(), "target": candidate.ID,
			})
		}
	}
	if len(candidates) == 0 {
		macros := resolveDeclarations(index, observation.Candidate, observation.SourceScope, observation.Path, func(declaration *declaration) bool {
			return declaration.MacroFunction
		})
		if len(macros) == 1 {
			facts.addEdge(observation.Path, map[string]any{
				"attributes": map[string]any{"role": "macro-expansion"},
				"owner":      observation.Path, "record": "edge", "relation": "references", "source": observation.SourceID,
				"span": observation.Span.record(), "target": macros[0].ID,
			})
			return
		}
		functionPointers := resolveDeclarations(index, observation.Candidate, observation.SourceScope, observation.Path, func(declaration *declaration) bool {
			functionPointer, _ := declaration.Attributes["function_pointer"].(bool)
			return functionPointer
		})
		if len(functionPointers) > 0 {
			facts.addUnresolved(observation.Path, unresolvedRecord(observation.SourceID, "calls", observation.Expression, "dynamic-target", observation.Path, observation.Span))
			return
		}
	}
	reason := resolutionReason(candidates)
	if len(candidates) == 0 && (observation.Member || strings.Contains(observation.Candidate, "::")) {
		reason = "external-target"
	}
	facts.addUnresolved(observation.Path, unresolvedRecord(observation.SourceID, "calls", observation.Expression, reason, observation.Path, observation.Span))
}

func resolveAccess(facts *factSet, index declarationIndex, observation accessObservation) {
	filter := func(declaration *declaration) bool {
		switch declaration.Kind {
		case "parameter", "variable", "constant", "field":
			return true
		default:
			return false
		}
	}
	candidates := filterDeclarations(index.byContainerName[observation.SourceID+"\x00"+observation.Candidate], filter)
	if len(candidates) == 0 && observation.ParentType != "" {
		candidates = filterDeclarations(index.byContainerName[observation.ParentType+"\x00"+observation.Candidate], filter)
	}
	if len(candidates) == 0 {
		candidates = resolveDeclarations(index, observation.Candidate, observation.SourceScope, observation.Path, filter)
	}
	if len(candidates) != 1 {
		return
	}
	facts.addEdge(observation.Path, map[string]any{
		"owner": observation.Path, "record": "edge", "relation": observation.Relation, "source": observation.SourceID,
		"span": observation.Span.record(), "target": candidates[0].ID,
	})
}
