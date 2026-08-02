package main

type runtimeLocalType struct {
	name       string
	scopeEnd   int
	scopeStart int
	spelling   string
	visibleAt  int
}

func runtimeLocalDeclaredType(callable *runtimeCallable, name string, callAt int) (string, bool) {
	bounds := callable.declaration.body
	if callAt < bounds.start || callAt >= bounds.end {
		return "", false
	}
	tokens := callable.file.tokens
	ignored := runtimeNestedRanges(tokens, bounds)
	var best *runtimeLocalType
	for index := bounds.start; index < callAt; index++ {
		if nested, ok := runtimeRangeStartingAt(index, ignored); ok {
			index = nested.end - 1
			continue
		}
		if tokens[index].text != "val" && tokens[index].text != "var" {
			continue
		}
		local, next := parseRuntimeLocal(tokens, bounds, index)
		if next > index {
			index = next - 1
		}
		if local.name != name || local.visibleAt > callAt || callAt >= local.scopeEnd {
			continue
		}
		if best == nil || local.scopeStart > best.scopeStart ||
			(local.scopeStart == best.scopeStart && local.visibleAt > best.visibleAt) {
			candidate := local
			best = &candidate
		}
	}
	if best == nil {
		return "", false
	}
	return best.spelling, true
}

func parseRuntimeLocal(tokens []token, bounds tokenRange, start int) (runtimeLocalType, int) {
	local := runtimeLocalType{}
	name := nextRuntimeToken(tokens, start+1, bounds.end)
	statementEnd := runtimeLocalStatementEnd(tokens, name, bounds.end)
	if name >= bounds.end || tokens[name].kind != tokenIdentifier {
		return local, statementEnd
	}
	local.name = identifierText(tokens[name])
	local.visibleAt = statementEnd
	local.scopeStart, local.scopeEnd = runtimeLocalScope(tokens, bounds, start)
	colon := nextRuntimeToken(tokens, name+1, statementEnd)
	if colon >= statementEnd || tokens[colon].text != ":" {
		return local, statementEnd
	}
	typeStart := nextRuntimeToken(tokens, colon+1, statementEnd)
	typeEnd := typeStart
	depth := delimiterDepth{}
	for typeEnd < statementEnd {
		current := tokens[typeEnd]
		if depth.zero() && (current.text == "=" || current.text == "by" || current.kind == tokenNewline) {
			break
		}
		depth.update(current.text)
		typeEnd++
	}
	local.spelling = compactTokenRange(tokens, typeStart, typeEnd)
	return local, statementEnd
}

func runtimeLocalStatementEnd(tokens []token, start, upper int) int {
	depth := delimiterDepth{}
	for index := start; index < upper; index++ {
		current := tokens[index]
		if depth.zero() && (current.kind == tokenNewline || current.text == ";") {
			return index + 1
		}
		depth.update(current.text)
	}
	return upper
}

func runtimeLocalScope(tokens []token, bounds tokenRange, declaration int) (int, int) {
	start, end := bounds.start, bounds.end
	for index := bounds.start; index < declaration; index++ {
		if tokens[index].text != "{" {
			continue
		}
		close := matchingRuntimeDelimiter(tokens, index, bounds.end, "{", "}")
		if close > declaration && index >= start && close <= end {
			start, end = index+1, close
		}
	}
	return start, end
}

func runtimeRangeStartingAt(index int, ranges []tokenRange) (tokenRange, bool) {
	for _, bounds := range ranges {
		if bounds.start == index {
			return bounds, true
		}
	}
	return tokenRange{}, false
}
