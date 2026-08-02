package main

func (state *parser) parseParameterList() ([]parameterDecl, bool) {
	open := state.index
	close := state.matchingDelimiter(open, "(", ")")
	if close < 0 {
		state.addDiagnostic(open, "unclosed parameter list")
		state.index = state.findRecoveryBoundary(open + 1)
		return nil, false
	}
	segments := splitTopLevel(state.tokens, open+1, close, ",")
	parameters := make([]parameterDecl, 0, len(segments))
	valid := true
	for _, segment := range segments {
		if parameter, ok := state.parseParameterSegment(segment[0], segment[1]); ok {
			parameters = append(parameters, parameter)
		} else if skipTokenKind(state.tokens, segment[0], segment[1], tokenNewline) < segment[1] {
			valid = false
		}
	}
	state.index = close + 1
	return parameters, valid
}

func (state *parser) parseParameterSegment(start, end int) (parameterDecl, bool) {
	start = skipTokenKind(state.tokens, start, end, tokenNewline)
	end = trimTokenKind(state.tokens, start, end, tokenNewline)
	if start >= end {
		return parameterDecl{}, false
	}
	originalStart := start
	var annotations, modifiers []string
	property := false
	mutable := false
	for start < end {
		if state.tokens[start].text == "@" {
			annotationEnd := annotationEnd(state.tokens, start, end)
			annotations = append(annotations, compactTokenRange(state.tokens, start+1, annotationEnd))
			start = annotationEnd
			continue
		}
		if _, ok := parameterModifiers[state.tokens[start].text]; ok {
			modifiers = append(modifiers, state.tokens[start].text)
			start++
			continue
		}
		if state.tokens[start].text == "val" || state.tokens[start].text == "var" {
			property = true
			mutable = state.tokens[start].text == "var"
			start++
			continue
		}
		break
	}
	colon := firstTopLevelToken(state.tokens, start, end, ":")
	if colon < 0 {
		state.addDiagnostic(originalStart, "parameter declaration has no type")
		return parameterDecl{}, false
	}
	nameIndex := lastIdentifier(state.tokens, start, colon)
	if nameIndex < 0 {
		state.addDiagnostic(originalStart, "parameter declaration is missing a name")
		return parameterDecl{}, false
	}
	equals := firstTopLevelToken(state.tokens, colon+1, end, "=")
	typeEnd := end
	if equals >= 0 {
		typeEnd = equals
	}
	return parameterDecl{
		annotations: uniqueSorted(annotations), hasDefault: equals >= 0, modifiers: uniqueSorted(modifiers),
		mutable: mutable, name: identifierText(state.tokens[nameIndex]), property: property,
		span: state.span(originalStart, end-1), typeName: state.tokenText(colon+1, typeEnd),
	}, true
}
