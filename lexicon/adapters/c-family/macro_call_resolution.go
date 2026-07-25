package main

import (
	"sort"
	"strings"
)

const maxMacroExpansionDepth = 8

func resolveMacroCall(facts *factSet, index declarationIndex, indirectCalls indirectCallIndex, observation callObservation, macros []*declaration) {
	possible := len(macros) != 1
	for _, macro := range macros {
		conditional, _ := macro.Attributes["conditional"].(bool)
		expandMacroInvocation(
			facts, index, indirectCalls, observation, macro,
			nil, map[string]struct{}{}, 0, possible || conditional,
		)
	}
}

func expandMacroInvocation(
	facts *factSet,
	index declarationIndex,
	indirectCalls indirectCallIndex,
	observation callObservation,
	macro *declaration,
	chain []*declaration,
	active map[string]struct{},
	depth int,
	possible bool,
) {
	chain = append(append([]*declaration(nil), chain...), macro)
	addMacroReference(facts, observation, chain, depth)
	if depth >= maxMacroExpansionDepth {
		addMacroUnresolved(facts, observation, "macro-expansion-depth", chain, nil, nil)
		return
	}
	if _, exists := active[macro.ID]; exists {
		addMacroUnresolved(facts, observation, "macro-expansion-cycle", chain, nil, nil)
		return
	}
	active[macro.ID] = struct{}{}
	defer delete(active, macro.ID)

	if macroHasUnsupportedSubstitution(macro) {
		addMacroUnresolved(facts, observation, "unsupported-macro-expansion", chain, nil, nil)
		return
	}
	bindings, valid := macroInvocationBindings(macro, observation.ArgumentExpressions)
	if !valid {
		addMacroUnresolved(facts, observation, "macro-argument-mismatch", chain, nil, nil)
		return
	}
	if len(macro.MacroCalls) == 0 {
		if macro.MacroTarget == "" {
			return
		}
		call := macroCallExpression{
			Callee:    macro.MacroTarget,
			Arguments: append([]string(nil), observation.ArgumentExpressions...),
		}
		resolveExpandedMacroCall(facts, index, indirectCalls, observation, call, bindings, chain, active, depth, possible, 0, true)
		return
	}
	for callIndex, call := range macro.MacroCalls {
		if call.Unsupported {
			addMacroUnresolved(facts, observation, "unsupported-macro-expansion", chain, &call, bindings)
			continue
		}
		resolveExpandedMacroCall(facts, index, indirectCalls, observation, call, bindings, chain, active, depth, possible, callIndex, false)
	}
}

func resolveExpandedMacroCall(
	facts *factSet,
	index declarationIndex,
	indirectCalls indirectCallIndex,
	invocation callObservation,
	bodyCall macroCallExpression,
	bindings map[string]string,
	chain []*declaration,
	active map[string]struct{},
	depth int,
	possible bool,
	callIndex int,
	alias bool,
) {
	expanded := bodyCall
	if !alias {
		expanded = substituteMacroCall(bodyCall, bindings)
	}
	if expanded.Callee == "" {
		addMacroUnresolved(facts, invocation, "unsupported-macro-expansion", chain, &bodyCall, bindings)
		return
	}
	observation := invocation
	observation.Candidate = expanded.Callee
	observation.Expression = renderMacroCall(expanded)
	observation.Arguments = macroArgumentCandidates(expanded.Arguments)
	observation.ArgumentExpressions = append([]string(nil), expanded.Arguments...)
	observation.Member = false
	observation.Receiver = ""
	observation.ReceiverTypeID = ""

	if nested := resolveMacroDeclarations(index, expanded.Callee, invocation.Path); len(nested) > 0 {
		nestedPossible := possible || len(nested) != 1
		for _, nestedMacro := range nested {
			conditional, _ := nestedMacro.Attributes["conditional"].(bool)
			expandMacroInvocation(facts, index, indirectCalls, observation, nestedMacro, chain, active, depth+1, nestedPossible || conditional)
		}
		return
	}

	resolution := resolveCallCandidateResolution(index, observation)
	if len(resolution.Candidates) > 1 {
		resolution = resolution.prune(len(observation.ArgumentExpressions))
	}
	attributes := macroCallAttributes(chain, bodyCall, expanded, bindings, resolution.Evidence, callIndex, alias)
	attributes["candidate_count"] = len(resolution.Candidates)
	if len(resolution.Candidates) == 1 && !possible {
		addResolvedCallEdges(facts, index, observation, resolution.Candidates[0], attributes)
		return
	}
	if len(resolution.Candidates) > 0 {
		for _, candidate := range resolution.Candidates {
			addCallEdge(facts, observation, candidate, "possible-calls", attributes)
		}
		return
	}

	if pointers := resolvePointerDeclarations(index, observation); len(pointers) > 0 {
		targets := indirectCalls.targets(pointers)
		attributes["candidate_count"] = len(targets)
		attributes["evidence"] = appendEvidence(attributes["evidence"], "function-pointer")
		for _, target := range targets {
			addCallEdge(facts, observation, target, "possible-calls", attributes)
		}
		if len(targets) > 0 {
			return
		}
	}

	reason := "missing-target"
	if !hasCallableDeclaration(index, observation.Candidate) {
		reason = "external-target"
	}
	addMacroUnresolved(facts, observation, reason, chain, &bodyCall, bindings)
}

func addMacroReference(facts *factSet, observation callObservation, chain []*declaration, depth int) {
	macro := chain[len(chain)-1]
	facts.addEdge(observation.Path, map[string]any{
		"attributes": map[string]any{
			"expansion_depth": depth,
			"role":            "macro-expansion",
			"via":             macroChainIDs(chain),
		},
		"owner": observation.Path, "record": "edge", "relation": "references", "source": observation.SourceID,
		"span": observation.Span.record(), "target": macro.ID,
	})
}

func macroCallAttributes(chain []*declaration, original, expanded macroCallExpression, bindings map[string]string, resolutionEvidence []string, callIndex int, alias bool) map[string]any {
	evidence := []string{"macro-mediation"}
	if alias {
		evidence = append(evidence, "macro-alias")
	} else {
		evidence = append(evidence, "macro-body")
	}
	if len(bindings) > 0 {
		evidence = append(evidence, "argument-substitution")
	}
	evidence = append(evidence, resolutionEvidence...)
	current := chain[len(chain)-1]
	return map[string]any{
		"evidence":              uniqueStrings(evidence),
		"expansion_depth":       len(chain) - 1,
		"indirect":              "macro",
		"macro_body_callee":     original.Callee,
		"macro_call_index":      callIndex,
		"macro_definition_span": current.Span.record(),
		"substituted_arguments": append([]string(nil), expanded.Arguments...),
		"substitutions":         macroSubstitutionAttributes(bindings),
		"via":                   macroChainIDs(chain),
	}
}

func addMacroUnresolved(facts *factSet, observation callObservation, reason string, chain []*declaration, call *macroCallExpression, bindings map[string]string) {
	record := unresolvedRecord(observation.SourceID, "calls", observation.Expression, reason, observation.Path, observation.Span)
	attributes := map[string]any{"expansion_depth": len(chain) - 1, "via": macroChainIDs(chain)}
	if call != nil {
		attributes["macro_body_callee"] = call.Callee
		attributes["substituted_arguments"] = append([]string(nil), call.Arguments...)
	}
	if substitutions := macroSubstitutionAttributes(bindings); len(substitutions) > 0 {
		attributes["substitutions"] = substitutions
	}
	record["attributes"] = attributes
	facts.addUnresolved(observation.Path, record)
}

func macroChainIDs(chain []*declaration) []string {
	result := make([]string, 0, len(chain))
	for _, macro := range chain {
		result = append(result, macro.ID)
	}
	return result
}

func renderMacroCall(call macroCallExpression) string {
	return call.Callee + "(" + strings.Join(call.Arguments, ", ") + ")"
}

func appendEvidence(value any, evidence string) []string {
	var values []string
	switch typed := value.(type) {
	case []string:
		values = append(values, typed...)
	case []any:
		for _, item := range typed {
			if text, ok := item.(string); ok {
				values = append(values, text)
			}
		}
	}
	return uniqueStrings(append(values, evidence))
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
