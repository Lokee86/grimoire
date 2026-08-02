package main

type invocationEvidence struct {
	arguments  [][2]int
	arity      int
	end        int
	expression string
	kind       string
	name       string
	qualifier  string
	start      int
}

var nonCallKeywords = map[string]bool{
	"assert": true, "catch": true, "do": true, "for": true, "if": true, "return": true,
	"switch": true, "synchronized": true, "throw": true, "try": true, "while": true,
}

func (state *analysisState) emitCalls(callable *callableDeclaration) {
	ignored := ignoredRuntimeTokens(callable)
	receivers := collectReceiverTypes(callable, ignored)
	for index := callable.bodyStart; index < callable.bodyEnd; index++ {
		if ignored[index] {
			continue
		}
		invocation, ok := callableInvocation(callable, index)
		if !ok {
			continue
		}
		state.emitInvocation(callable, invocation, receivers)
	}
}

func callableInvocation(callable *callableDeclaration, nameIndex int) (invocationEvidence, bool) {
	tokens := callable.tokens
	if nameIndex+1 >= callable.bodyEnd || !identifierToken(tokens[nameIndex].text) || tokens[nameIndex+1].text != "(" {
		return invocationEvidence{}, false
	}
	name := tokens[nameIndex].text
	if nonCallKeywords[name] || (nameIndex > callable.bodyStart && tokens[nameIndex-1].text == "@") {
		return invocationEvidence{}, false
	}
	close := matchingToken(tokens, nameIndex+1, callable.bodyEnd, "(", ")")
	if close < 0 {
		return invocationEvidence{
			arity: -1, end: nameIndex + 2,
			expression: sourceExcerpt(callable.source, tokens, nameIndex, nameIndex+2),
			kind:       "unsupported", name: name, start: nameIndex,
		}, true
	}
	arguments := splitTopLevel(tokens, nameIndex+2, close, ",")
	arity := len(arguments)
	if nameIndex+2 == close {
		arity = 0
	}
	start, kind, qualifier := nameIndex, "unqualified", ""
	chainStart := qualifiedChainStart(tokens, nameIndex)
	if chainStart > callable.bodyStart && tokens[chainStart-1].text == "new" {
		start, kind = chainStart-1, "constructor"
		qualifier, _ = qualifiedTokenName(tokens[chainStart:nameIndex])
		if qualifier == "" {
			qualifier = name
		} else {
			qualifier += "." + name
		}
	} else if name == "this" || name == "super" {
		kind = name + "-constructor"
	} else if nameIndex > callable.bodyStart && tokens[nameIndex-1].text == "." {
		qualifierEnd := nameIndex - 1
		start = qualifiedChainStart(tokens, nameIndex-2)
		if start > callable.bodyStart && tokens[start-1].text == "@" {
			return invocationEvidence{}, false
		}
		if value, valid := qualifiedTokenName(tokens[start:qualifierEnd]); valid {
			qualifier, kind = value, "qualified"
		} else {
			kind = "dynamic"
			start = nameIndex
		}
	}
	return invocationEvidence{
		arguments: arguments, arity: arity, end: close + 1, expression: sourceExcerpt(callable.source, tokens, start, close+1),
		kind: kind, name: name, qualifier: qualifier, start: start,
	}, true
}

func qualifiedChainStart(tokens []token, end int) int {
	start := end
	for start >= 2 && tokens[start-1].text == "." && identifierToken(tokens[start-2].text) {
		start -= 2
	}
	return start
}

func (state *analysisState) emitInvocation(callable *callableDeclaration, invocation invocationEvidence, receivers receiverTypes) {
	attributes := map[string]any{"arity": invocation.arity, "invocation_kind": invocation.kind}
	invocationSpan := runtimeSpan(callable.path, callable.tokens, invocation.start, invocation.end)
	if invocation.kind == "unsupported" {
		state.facts.addUnresolved(callable.id, "calls", invocation.expression, "unsupported-form", callable.path, invocationSpan, attributes)
		return
	}
	var candidates []callableDeclaration
	reason := "dynamic-target"
	switch invocation.kind {
	case "unqualified":
		candidates = state.callableCandidates(callable.ownerID, invocation.name, invocation.arity, false, false)
	case "qualified":
		candidates, reason = state.qualifiedCallCandidates(callable, invocation, receivers)
		if state.typedIdentifierReceiver(callable, invocation, receivers) {
			candidates = state.narrowArgumentCandidates(callable, invocation, receivers, candidates)
		}
	case "constructor":
		types := state.resolveTypeDeclarations(invocation.qualifier, callable.ownerQualifiedName, callable.context)
		if len(types) == 0 {
			reason = "external-target"
		} else if len(types) > 1 {
			reason = "ambiguous-target"
		} else {
			candidates = state.callableCandidates(types[0].id, "", invocation.arity, true, false)
			reason = "unsupported-form"
		}
	case "this-constructor":
		candidates = state.callableCandidates(callable.ownerID, "", invocation.arity, true, false)
		candidates = withoutCallable(candidates, callable.id)
		reason = "unsupported-form"
	case "super-constructor":
		parents := state.directSuperclasses(callable.ownerID)
		if len(parents) == 1 {
			candidates = state.callableCandidates(parents[0], "", invocation.arity, true, false)
			reason = "unsupported-form"
		} else if len(parents) == 0 {
			reason = "external-target"
		} else {
			reason = "ambiguous-target"
		}
	}
	state.emitCallCandidates(callable, invocation, candidates, reason, attributes, invocationSpan)
}

func (state *analysisState) emitCallCandidates(callable *callableDeclaration, invocation invocationEvidence, candidates []callableDeclaration, reason string, attributes map[string]any, invocationSpan *span) {
	if len(candidates) == 0 {
		state.facts.addUnresolved(callable.id, "calls", invocation.expression, reason, callable.path, invocationSpan, attributes)
		return
	}
	relation := "calls"
	if len(candidates) > 1 {
		relation = "possible-calls"
		attributes = cloneAttributes(attributes)
		attributes["candidate_count"] = len(candidates)
	}
	for _, candidate := range candidates {
		state.facts.addEdge(callable.id, candidate.id, relation, callable.path, invocationSpan, attributes)
	}
}

func withoutCallable(candidates []callableDeclaration, id string) []callableDeclaration {
	result := candidates[:0]
	for _, candidate := range candidates {
		if candidate.id != id {
			result = append(result, candidate)
		}
	}
	return result
}

func (state *analysisState) directSuperclasses(ownerID string) []string {
	return state.superclassIDs[ownerID]
}
