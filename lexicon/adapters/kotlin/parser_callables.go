package main

func (state *parser) parseFunction(prefix declarationPrefix) *declaration {
	start := prefix.start
	state.index++
	state.skipNewlines()
	if state.at("<") {
		state.skipBalanced("<", ">")
		state.skipNewlines()
	}
	nameStart := state.index
	open := state.findHeaderToken("(")
	if open < 0 {
		state.addDiagnostic(start, "function declaration has no parameter list")
		state.skipStatement()
		return nil
	}
	nameIndex := lastIdentifier(state.tokens, nameStart, open)
	if nameIndex < 0 {
		state.addDiagnostic(start, "function declaration is missing a name")
		state.index = open
		state.skipBalanced("(", ")")
		state.skipStatement()
		return nil
	}
	name := identifierText(state.tokens[nameIndex])
	receiver := ""
	if dot := lastToken(state.tokens, nameStart, nameIndex, "."); dot >= 0 {
		receiver = state.tokenText(nameStart, dot)
	}
	state.index = open
	parameters, valid := state.parseParameterList()
	if !valid {
		state.skipStatement()
		return nil
	}
	returnType := ""
	state.skipNewlines()
	if state.at(":") {
		state.index++
		typeStart := state.index
		end := state.findTypeEnd()
		returnType = state.tokenText(typeStart, end)
		state.index = end
	}
	for state.at("where") {
		state.skipStatementHeader()
	}
	body := tokenRange{}
	if state.at("{") {
		open := state.index
		if close := state.matchingDelimiter(open, "{", "}"); close >= 0 {
			body = tokenRange{start: open + 1, end: close}
		}
		state.skipBalanced("{", "}")
	} else if state.at("=") {
		body.start = state.index + 1
		state.skipExpression()
		body.end = state.index
	}
	end := maxInt(start, state.index-1)
	return &declaration{
		annotations: prefix.annotations, form: "function", kind: "function",
		modifiers: prefix.modifiers, name: name, parameters: parameters,
		body: body, receiver: receiver, returnType: returnType, span: state.span(start, end),
	}
}

func (state *parser) parseProperty(prefix declarationPrefix) *declaration {
	start := prefix.start
	mutable := state.at("var")
	state.index++
	state.skipNewlines()
	if state.at("(") {
		state.addDiagnostic(start, "destructuring property declarations are not modeled")
		state.skipStatement()
		return nil
	}
	headerEnd := state.findPropertyHeaderEnd()
	colon := firstTopLevelToken(state.tokens, state.index, headerEnd, ":")
	equals := firstTopLevelToken(state.tokens, state.index, headerEnd, "=")
	by := firstTopLevelToken(state.tokens, state.index, headerEnd, "by")
	nameEnd := headerEnd
	for _, candidate := range []int{colon, equals, by} {
		if candidate >= 0 && candidate < nameEnd {
			nameEnd = candidate
		}
	}
	nameIndex := lastIdentifier(state.tokens, state.index, nameEnd)
	if nameIndex < 0 {
		state.addDiagnostic(start, "property declaration is missing a name")
		state.skipStatement()
		return nil
	}
	name := identifierText(state.tokens[nameIndex])
	receiver := ""
	if dot := lastToken(state.tokens, state.index, nameIndex, "."); dot >= 0 {
		receiver = state.tokenText(state.index, dot)
	}
	typeName := ""
	if colon >= 0 {
		typeEnd := headerEnd
		for _, candidate := range []int{equals, by} {
			if candidate > colon && candidate < typeEnd {
				typeEnd = candidate
			}
		}
		typeName = state.tokenText(colon+1, typeEnd)
	}
	state.index = headerEnd
	if state.at("=") || state.at("by") {
		state.skipExpression()
	}
	return &declaration{
		annotations: prefix.annotations, form: "property", kind: "field", modifiers: prefix.modifiers,
		delegated: by >= 0, mutable: mutable, name: name, receiver: receiver,
		span: state.span(start, maxInt(start, state.index-1)), typeName: typeName,
	}
}

func (state *parser) parseConstructor(prefix declarationPrefix, primary bool) *declaration {
	start := prefix.start
	state.index++
	state.skipNewlines()
	if !state.at("(") {
		state.addDiagnostic(start, "constructor declaration has no parameter list")
		state.skipStatement()
		return nil
	}
	parameters, valid := state.parseParameterList()
	if !valid {
		state.skipStatement()
		return nil
	}
	delegation := tokenRange{}
	for state.current().kind != tokenEOF && state.current().kind != tokenNewline && !state.at("{") && !state.at(";") {
		if (state.at("this") || state.at("super")) && state.index+1 < len(state.tokens) && state.tokens[state.index+1].text == "(" {
			if close := state.matchingDelimiter(state.index+1, "(", ")"); close >= 0 {
				delegation = tokenRange{start: state.index, end: close + 1}
			}
		}
		if state.at("(") {
			state.skipBalanced("(", ")")
		} else {
			state.index++
		}
	}
	body := tokenRange{}
	if state.at("{") {
		open := state.index
		if close := state.matchingDelimiter(open, "{", "}"); close >= 0 {
			body = tokenRange{start: open + 1, end: close}
		}
		state.skipBalanced("{", "}")
	}
	return &declaration{
		annotations: prefix.annotations, form: "secondary_constructor", kind: "constructor",
		body: body, delegation: delegation, modifiers: prefix.modifiers,
		name: "constructor", parameters: parameters, primary: primary,
		span: state.span(start, maxInt(start, state.index-1)),
	}
}
