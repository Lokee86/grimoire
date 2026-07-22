package main

func processCalls(facts *factSet, model *semanticModel, pf *parsedFile) {
	for _, stmt := range pf.statements {
		for _, call := range findCalls(stmt, pf.path) {
			context := contextForPosition(pf, spanInt(call.span, "start_line"), spanInt(call.span, "start_column"))
			owner := context.ownerID
			if context.functionID != "" {
				owner = context.functionID
			}
			emitResolvedCall(facts, model, context, owner, call)
			emitCallbackEdges(facts, model, context, owner, call)
		}
	}
}

func emitResolvedCall(facts *factSet, model *semanticModel, context analysisContext, owner string, call callReference) {
	resolution := model.resolveCall(context, call)
	targets := append([]string(nil), resolution.functionTargets...)
	targets = append(targets, resolution.constructorOwners...)
	targets = uniqueSorted(targets)
	if len(targets) == 1 {
		facts.addEdge(edge(owner, targets[0], "calls", call.span))
		return
	}
	if len(targets) > 1 {
		for _, target := range targets {
			facts.addEdge(edgeWithAttributes(owner, target, "possible-calls", call.span, map[string]any{"dispatch": true}))
		}
		return
	}
	reason := resolution.reason
	if reason == "" {
		reason = "dynamic-target"
	}
	record := unresolved(owner, "calls", call.expr, reason, call.span)
	if call.name != "" {
		record["candidate_name"] = call.callee
	}
	facts.addUnresolved(record)
}

func emitCallbackEdges(facts *factSet, model *semanticModel, context analysisContext, owner string, call callReference) {
	if call.name == "Callable" && len(call.args) >= 2 {
		method := stringLiteral(call.args[1])
		if method != "" {
			emitPossibleMethods(facts, model, owner, model.inferExpressionOwners(context, call.args[0]), method, call.span, "callable")
		}
		return
	}
	if call.name == "call" || call.name == "call_deferred" || call.name == "rpc" || call.name == "rpc_id" {
		index := 0
		if call.name == "rpc_id" {
			index = 1
		}
		if index < len(call.args) {
			method := stringLiteral(call.args[index])
			if method != "" {
				owners := ownerSet{context.ownerID: {}}
				if len(call.receiver) > 0 {
					owners = model.inferExpressionOwners(context, call.receiver)
				}
				emitPossibleMethods(facts, model, owner, owners, method, call.span, "dynamic-invocation")
			}
		}
	}
	if callbackArgumentIndex(call.name) < 0 {
		return
	}
	index := callbackArgumentIndex(call.name)
	if index >= len(call.args) {
		return
	}
	argument := call.args[index]
	if targets := ownerSlice(model.inferExpressionCallables(context, argument)); len(targets) > 0 {
		emitPossibleTargets(facts, owner, targets, call.span, "callback")
		return
	}
	if name := simpleIdentifier(argument); name != "" {
		emitPossibleTargets(facts, owner, model.methodTargets(context.ownerID, name, false, false), call.span, "callback")
		return
	}
	if parts, ok := propertyChain(argument); ok && len(parts) == 2 && parts[0] == "self" {
		emitPossibleTargets(facts, owner, model.methodTargets(context.ownerID, parts[1], false, false), call.span, "callback")
	}
}

func emitPossibleMethods(facts *factSet, model *semanticModel, source string, owners ownerSet, method string, span sourceSpan, reason string) {
	var targets []string
	for owner := range owners {
		targets = append(targets, model.methodTargets(owner, method, false, false)...)
	}
	emitPossibleTargets(facts, source, uniqueSorted(targets), span, reason)
}

func emitPossibleTargets(facts *factSet, source string, targets []string, span sourceSpan, reason string) {
	for _, target := range uniqueSorted(targets) {
		facts.addEdge(edgeWithAttributes(source, target, "possible-calls", span, map[string]any{"callback": true, "reason": reason}))
	}
}

func callbackArgumentIndex(name string) int {
	switch name {
	case "connect", "map", "filter", "any", "all", "sort_custom", "bsearch_custom":
		return 0
	case "reduce":
		return 0
	default:
		return -1
	}
}
