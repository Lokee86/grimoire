package main

func (state *analysis) emitCallableDataflow(callable *runtimeCallable) {
	shadowed := runtimeShadowedNames(callable)
	for _, bounds := range []tokenRange{callable.declaration.delegation, callable.declaration.body} {
		state.emitCallableRangeDataflow(callable, bounds, shadowed)
	}
}

func (state *analysis) emitCallableRangeDataflow(
	callable *runtimeCallable,
	bounds tokenRange,
	shadowed map[string]struct{},
) {
	if bounds.end <= bounds.start {
		return
	}
	tokens := callable.file.tokens
	ignored := runtimeNestedRanges(tokens, bounds)
	for index := bounds.start; index < bounds.end && index < len(tokens); index++ {
		if runtimeTokenIgnored(index, ignored) {
			continue
		}
		if tokens[index].kind != tokenIdentifier || runtimeReferenceDeclaration(tokens, index, bounds.start) {
			continue
		}
		start, unsupported := runtimeCalleeStart(tokens, index, bounds.start)
		if unsupported {
			continue
		}
		qualifier := ""
		if dot := lastToken(tokens, start, index, "."); dot >= 0 {
			qualifier = compactTokenRange(tokens, start, dot)
		}
		name := identifierText(tokens[index])
		if name == "this" || name == "super" || runtimeNamedArgument(tokens, index, bounds) {
			continue
		}
		target := state.resolveRuntimeValue(callable, qualifier, name, shadowed)
		if target == "" {
			continue
		}
		span := runtimeTokenSpan(callable.file.path, tokens[start], tokens[index])
		write, compound := runtimeWriteKind(tokens, start, index, bounds)
		if !write || compound {
			state.facts.addEdge(callable.id, target, "reads", callable.file.path, &span, nil)
		}
		if write {
			state.facts.addEdge(callable.id, target, "writes", callable.file.path, &span, nil)
		}
	}
}

func runtimeReferenceDeclaration(tokens []token, index, lower int) bool {
	previous := previousRuntimeToken(tokens, index-1, lower)
	if previous < lower {
		return false
	}
	switch tokens[previous].text {
	case "class", "fun", "object", "val", "var":
		return true
	}
	return false
}

func runtimeNamedArgument(tokens []token, index int, bounds tokenRange) bool {
	next := nextRuntimeToken(tokens, index+1, bounds.end)
	if next >= bounds.end || tokens[next].text != "=" {
		return false
	}
	after := nextRuntimeToken(tokens, next+1, bounds.end)
	if after < bounds.end && tokens[after].text == "=" {
		return false
	}
	previous := previousRuntimeToken(tokens, index-1, bounds.start)
	return previous >= bounds.start && (tokens[previous].text == "(" || tokens[previous].text == ",")
}

func runtimeWriteKind(tokens []token, start, name int, bounds tokenRange) (bool, bool) {
	next := nextRuntimeToken(tokens, name+1, bounds.end)
	if next < bounds.end {
		after := nextRuntimeToken(tokens, next+1, bounds.end)
		if tokens[next].text == "=" {
			if after >= bounds.end || tokens[after].text != "=" {
				return true, false
			}
		}
		if runtimeUpdateOperator(tokens[next].text) && after < bounds.end && tokens[after].text == "=" {
			return true, true
		}
		if (tokens[next].text == "+" || tokens[next].text == "-") && after < bounds.end && tokens[after].text == tokens[next].text {
			return true, true
		}
	}
	previous := previousRuntimeToken(tokens, start-1, bounds.start)
	if previous < bounds.start || (tokens[previous].text != "+" && tokens[previous].text != "-") {
		return false, false
	}
	before := previousRuntimeToken(tokens, previous-1, bounds.start)
	if before >= bounds.start && tokens[before].text == tokens[previous].text {
		return true, true
	}
	return false, false
}

func runtimeUpdateOperator(value string) bool {
	switch value {
	case "+", "-", "*", "/", "%":
		return true
	}
	return false
}
