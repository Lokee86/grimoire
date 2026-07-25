package main

import "sort"

func macroCallableTargets(index declarationIndex, macros []*declaration, observation callObservation, active map[string]struct{}) []*declaration {
	unique := make(map[string]*declaration)
	for _, macro := range macros {
		collectMacroCallableTargets(index, macro, observation, active, 0, unique)
	}
	result := make([]*declaration, 0, len(unique))
	for _, target := range unique {
		result = append(result, target)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func collectMacroCallableTargets(
	index declarationIndex,
	macro *declaration,
	observation callObservation,
	active map[string]struct{},
	depth int,
	result map[string]*declaration,
) {
	if depth >= maxMacroExpansionDepth {
		return
	}
	if _, exists := active[macro.ID]; exists {
		return
	}
	active[macro.ID] = struct{}{}
	defer delete(active, macro.ID)

	bindings, valid := macroInvocationBindings(macro, observation.ArgumentExpressions)
	if !valid {
		return
	}
	calls := macro.MacroCalls
	alias := false
	if len(calls) == 0 && macro.MacroTarget != "" {
		calls = []macroCallExpression{{
			Callee:    macro.MacroTarget,
			Arguments: append([]string(nil), observation.ArgumentExpressions...),
		}}
		alias = true
	}
	for _, call := range calls {
		if call.Unsupported {
			continue
		}
		expanded := call
		if !alias {
			expanded = substituteMacroCall(call, bindings)
		}
		nestedObservation := observation
		nestedObservation.Candidate = expanded.Callee
		nestedObservation.Arguments = macroArgumentCandidates(expanded.Arguments)
		nestedObservation.ArgumentExpressions = append([]string(nil), expanded.Arguments...)
		if nested := resolveMacroDeclarations(index, expanded.Callee, observation.Path); len(nested) > 0 {
			for _, nestedMacro := range nested {
				collectMacroCallableTargets(index, nestedMacro, nestedObservation, active, depth+1, result)
			}
			continue
		}
		resolution := resolveCallCandidateResolution(index, nestedObservation)
		if len(resolution.Candidates) > 1 {
			resolution = resolution.prune(len(expanded.Arguments))
		}
		for _, target := range resolution.Candidates {
			result[target.ID] = target
		}
	}
}
