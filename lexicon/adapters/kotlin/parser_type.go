package main

func (state *parser) parseType(prefix declarationPrefix) *declaration {
	keywordIndex := state.index
	keyword := state.current().text
	state.index++
	name := ""
	if keyword == "object" && containsString(prefix.modifiers, "companion") {
		name = "Companion"
		if state.current().kind == tokenIdentifier && !state.at(":") {
			name = identifierText(state.current())
			state.index++
		}
	} else if state.current().kind == tokenIdentifier {
		name = identifierText(state.current())
		state.index++
	} else {
		state.addDiagnostic(keywordIndex, "type declaration is missing a name")
		state.skipStatement()
		return nil
	}

	if state.at("<") {
		state.skipBalanced("<", ">")
	}
	for state.at("@") || isVisibility(state.current().text) {
		if state.at("@") {
			state.parseAnnotation()
		} else {
			state.index++
		}
		state.skipNewlines()
	}
	explicitConstructor := false
	if state.at("constructor") {
		explicitConstructor = true
		state.index++
		state.skipNewlines()
	}
	var parameters []parameterDecl
	if state.at("(") {
		var valid bool
		parameters, valid = state.parseParameterList()
		if !valid {
			state.skipStatement()
			return nil
		}
		explicitConstructor = true
	}

	form := keyword
	kind := "type"
	if keyword == "interface" {
		kind = "interface"
	}
	switch {
	case keyword == "object" && containsString(prefix.modifiers, "companion"):
		form = "companion_object"
	case keyword == "object" && containsString(prefix.modifiers, "data"):
		form = "data_object"
	case keyword == "object":
		form = "object"
	case containsString(prefix.modifiers, "enum"):
		form = "enum_class"
	case containsString(prefix.modifiers, "data"):
		form = "data_class"
	case containsString(prefix.modifiers, "value"):
		form = "value_class"
	case containsString(prefix.modifiers, "annotation"):
		form = "annotation_class"
	case containsString(prefix.modifiers, "sealed") && keyword == "interface":
		form = "sealed_interface"
	case containsString(prefix.modifiers, "sealed"):
		form = "sealed_class"
	}

	typeDeclaration := &declaration{
		annotations: prefix.annotations,
		form:        form,
		kind:        kind,
		modifiers:   prefix.modifiers,
		name:        name,
		parameters:  parameters,
		span:        state.span(prefix.start, maxInt(prefix.start, state.index-1)),
	}

	headerEnd := state.findTypeHeaderEnd()
	if state.at(":") {
		typeDeclaration.supertypes = state.parseSupertypes(state.index+1, headerEnd)
	}
	state.index = headerEnd

	if state.at("{") {
		state.index++
		typeDeclaration.children = state.parseScope(true)
		typeDeclaration.span = state.span(prefix.start, maxInt(prefix.start, state.index-1))
	}
	if kind == "type" && keyword == "class" {
		constructor := &declaration{
			form: "primary_constructor", kind: "constructor", name: name,
			parameters: parameters, primary: true, span: typeDeclaration.span,
		}
		if !explicitConstructor {
			constructor.modifiers = []string{"implicit"}
		}
		typeDeclaration.children = append([]*declaration{constructor}, typeDeclaration.children...)
	}
	return typeDeclaration
}
