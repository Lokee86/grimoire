package main

import "sort"

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

func resolveCall(facts *factSet, index declarationIndex, indirectCalls indirectCallIndex, observation callObservation) {
	if macros := resolveMacroDeclarations(index, observation.Candidate, observation.Path); len(macros) > 0 {
		resolveMacroCall(facts, index, observation, macros)
		return
	}

	candidates := resolveDeclarations(index, observation.Candidate, observation.SourceScope, observation.Path, func(declaration *declaration) bool {
		return declaration.Callable
	})
	if len(candidates) == 1 {
		addCallEdge(facts, observation, candidates[0], "calls", nil)
		return
	}
	if len(candidates) > 1 {
		for _, candidate := range candidates {
			addCallEdge(facts, observation, candidate, "possible-calls", nil)
		}
		facts.addUnresolved(observation.Path, unresolvedRecord(observation.SourceID, "calls", observation.Expression, "ambiguous-target", observation.Path, observation.Span))
		return
	}

	if functionPointers := resolvePointerDeclarations(index, observation); len(functionPointers) > 0 {
		targets := indirectCalls.targets(functionPointers)
		if len(targets) > 0 {
			pointerIDs := make([]string, 0, len(functionPointers))
			for _, pointer := range functionPointers {
				pointerIDs = append(pointerIDs, pointer.ID)
			}
			sort.Strings(pointerIDs)
			attributes := map[string]any{"indirect": "function-pointer", "via": pointerIDs}
			for _, target := range targets {
				addCallEdge(facts, observation, target, "possible-calls", attributes)
			}
		}
		facts.addUnresolved(observation.Path, unresolvedRecord(observation.SourceID, "calls", observation.Expression, "dynamic-target", observation.Path, observation.Span))
		return
	}

	reason := "missing-target"
	if !hasCallableDeclaration(index, observation.Candidate) {
		reason = "external-target"
	}
	facts.addUnresolved(observation.Path, unresolvedRecord(observation.SourceID, "calls", observation.Expression, reason, observation.Path, observation.Span))
}

func resolveMacroCall(facts *factSet, index declarationIndex, observation callObservation, macros []*declaration) {
	macroIDs := make([]string, 0, len(macros))
	for _, macro := range macros {
		macroIDs = append(macroIDs, macro.ID)
		facts.addEdge(observation.Path, map[string]any{
			"attributes": map[string]any{"role": "macro-expansion"},
			"owner":      observation.Path, "record": "edge", "relation": "references", "source": observation.SourceID,
			"span": observation.Span.record(), "target": macro.ID,
		})
	}
	sort.Strings(macroIDs)

	targets := macroCallableTargets(index, macros, observation, map[string]struct{}{})
	attributes := map[string]any{"indirect": "macro", "via": macroIDs}
	if len(targets) == 0 {
		return
	}
	conditional := false
	for _, macro := range macros {
		if value, _ := macro.Attributes["conditional"].(bool); value {
			conditional = true
			break
		}
	}
	if len(targets) == 1 && !conditional {
		addCallEdge(facts, observation, targets[0], "calls", attributes)
		return
	}
	attributes["conditional"] = conditional
	for _, target := range targets {
		addCallEdge(facts, observation, target, "possible-calls", attributes)
	}
}

func macroCallableTargets(index declarationIndex, macros []*declaration, observation callObservation, visited map[string]struct{}) []*declaration {
	unique := map[string]*declaration{}
	for _, macro := range macros {
		if _, seen := visited[macro.ID]; seen || macro.MacroTarget == "" {
			continue
		}
		visited[macro.ID] = struct{}{}
		callables := resolveDeclarations(index, macro.MacroTarget, observation.SourceScope, observation.Path, func(declaration *declaration) bool {
			return declaration.Callable
		})
		for _, callable := range callables {
			unique[callable.ID] = callable
		}
		if len(callables) > 0 {
			continue
		}
		nested := resolveMacroDeclarations(index, macro.MacroTarget, observation.Path)
		for _, callable := range macroCallableTargets(index, nested, observation, visited) {
			unique[callable.ID] = callable
		}
	}
	result := make([]*declaration, 0, len(unique))
	for _, declaration := range unique {
		result = append(result, declaration)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func addCallEdge(facts *factSet, observation callObservation, target *declaration, relation string, attributes map[string]any) {
	edge := map[string]any{
		"owner": observation.Path, "record": "edge", "relation": relation, "source": observation.SourceID,
		"span": observation.Span.record(), "target": target.ID,
	}
	if len(attributes) > 0 {
		edge["attributes"] = attributes
	}
	facts.addEdge(observation.Path, edge)
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
