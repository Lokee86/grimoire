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
		resolveMacroCall(facts, index, indirectCalls, observation, macros)
		return
	}

	resolution := resolveCallCandidateResolution(index, observation)
	candidates := resolution.Candidates
	if len(candidates) == 1 {
		addResolvedCallEdges(facts, index, observation, candidates[0], callEdgeAttributes(resolution))
		return
	}
	if len(candidates) > 1 {
		resolution = resolution.prune(len(observation.Arguments))
		candidates = resolution.Candidates
		if len(candidates) == 1 {
			addResolvedCallEdges(facts, index, observation, candidates[0], callEdgeAttributes(resolution))
			return
		}
		for _, candidate := range candidates {
			addCallEdge(facts, observation, candidate, "possible-calls", callEdgeAttributes(resolution))
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
			attributes := map[string]any{
				"candidate_count": len(targets), "evidence": []string{"function-pointer"},
				"indirect": "function-pointer", "via": pointerIDs,
			}
			for _, target := range targets {
				addCallEdge(facts, observation, target, "possible-calls", attributes)
			}
		}
		facts.addUnresolved(observation.Path, unresolvedRecord(observation.SourceID, "calls", observation.Expression, "dynamic-target", observation.Path, observation.Span))
		return
	}

	reason := "missing-target"
	if qualifier := explicitCallQualifier(observation.Candidate); qualifier != "" && len(directQualifiedTypes(index, qualifier, observation.Path)) == 0 {
		reason = "external-target"
	} else if !hasCallableDeclaration(index, observation.Candidate) {
		reason = "external-target"
	}
	facts.addUnresolved(observation.Path, unresolvedRecord(observation.SourceID, "calls", observation.Expression, reason, observation.Path, observation.Span))
}

func addCallEdge(facts *factSet, observation callObservation, target *declaration, relation string, attributes map[string]any) {
	edge := map[string]any{
		"owner": observation.Path, "record": "edge", "relation": relation, "source": observation.SourceID,
		"span": observation.Span.record(), "target": target.ID,
	}
	edgeAttributes := map[string]any{
		"candidate_count": 1, "evidence": []string{}, "resolution": "possible",
	}
	if relation == "calls" {
		edgeAttributes["resolution"] = "definite"
	}
	for key, value := range attributes {
		edgeAttributes[key] = value
	}
	edge["attributes"] = edgeAttributes
	facts.addEdge(observation.Path, edge)
}

func callEdgeAttributes(resolution callCandidateResolution) map[string]any {
	return map[string]any{
		"candidate_count": len(resolution.Candidates), "evidence": resolution.Evidence,
	}
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
