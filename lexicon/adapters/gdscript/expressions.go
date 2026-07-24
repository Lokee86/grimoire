package main

func (m *semanticModel) inferExpressionOwners(context analysisContext, expression []token) ownerSet {
	expression = trimExpression(expression)
	if len(expression) == 0 {
		return nil
	}
	if cast := topLevelToken(expression, "as"); cast > 0 && cast+1 < len(expression) {
		return m.typeOwners(context.file.path, joinTokens(expression[cast+1:]))
	}
	if branch := topLevelToken(expression, "if"); branch > 0 {
		if otherwise := topLevelTokenAfter(expression, "else", branch); otherwise > branch {
			return unionOwners(
				m.inferExpressionOwners(context, expression[:branch]),
				m.inferExpressionOwners(context, expression[otherwise+1:]),
			)
		}
	}
	if owners, ok := m.staticLoadOwners(context, expression); ok {
		return owners
	}
	if call := terminalCall(expression); call != nil {
		resolution := m.resolveCall(context, *call)
		owners := make(ownerSet)
		for _, owner := range resolution.constructorOwners {
			owners[owner] = struct{}{}
		}
		for _, target := range resolution.functionTargets {
			for owner := range m.returns[target] {
				owners[owner] = struct{}{}
			}
		}
		if len(owners) > 0 {
			return owners
		}
		if call.name == "duplicate" && len(call.receiver) > 0 {
			return m.inferExpressionOwners(context, call.receiver)
		}
		return nil
	}
	if base, property, ok := terminalPropertyAccess(expression); ok {
		owners := m.inferExpressionOwners(context, base)
		var result ownerSet
		for owner := range owners {
			result = unionOwners(result, m.memberOwners(owner, property, make(map[string]bool)))
			for _, nested := range m.facts.typeByOwnerID[owner][property] {
				result = unionOwners(result, ownerSet{nested: {}})
			}
		}
		return result
	}
	if name := simpleIdentifier(expression); name != "" {
		if values, known := m.preloadAliasOwners(context.file.path, name); known {
			return values
		}
		if values := m.typeAliasOwners(context.file.path, name); len(values) > 0 {
			return values
		}
		if name == "self" {
			return ownerSet{context.ownerID: {}}
		}
		if values := m.bindingOwners(context, name); len(values) > 0 {
			return values
		}
		if values, ambiguous := m.classOwners(context.file.path, name); len(values) > 0 && !ambiguous {
			return values
		}
		if m.bindingDeclared(context, name) {
			return nil
		}
		if owner := m.autoloadOwner(context.file.path, name); owner != "" {
			return ownerSet{owner: {}}
		}
		return nil
	}
	if parts, ok := propertyChain(expression); ok {
		var owners ownerSet
		if parts[0] == "self" {
			owners = ownerSet{context.ownerID: {}}
		} else {
			owners = m.bindingOwners(context, parts[0])
		}
		for _, property := range parts[1:] {
			var next ownerSet
			for owner := range owners {
				next = unionOwners(next, m.memberOwners(owner, property, make(map[string]bool)))
				for _, nested := range m.facts.typeByOwnerID[owner][property] {
					next = unionOwners(next, ownerSet{nested: {}})
				}
			}
			owners = next
			if len(owners) == 0 {
				break
			}
		}
		return owners
	}
	return nil
}

func (m *semanticModel) inferTypeReferenceOwners(context analysisContext, tokens []token) ownerSet {
	tokens = trimExpression(tokens)
	if owners, ok := m.staticLoadOwners(context, tokens); ok {
		return owners
	}
	if name := simpleIdentifier(tokens); name != "" {
		if owners, known := m.preloadAliasOwners(context.file.path, name); known {
			return owners
		}
		if owners := m.typeAliasOwners(context.file.path, name); len(owners) > 0 {
			return owners
		}
		if owners, ambiguous := m.classOwners(context.file.path, name); !ambiguous && len(owners) > 0 {
			return owners
		}
		return nil
	}
	parts, ok := propertyChain(tokens)
	if !ok || len(parts) < 2 {
		return nil
	}
	var owners ownerSet
	if base, ambiguous := m.classOwners(context.file.path, parts[0]); !ambiguous {
		owners = base
	}
	if len(owners) == 0 {
		owners, _ = m.preloadAliasOwners(context.file.path, parts[0])
	}
	if len(owners) == 0 {
		owners = m.typeAliasOwners(context.file.path, parts[0])
	}
	for _, nestedName := range parts[1:] {
		var nestedOwners ownerSet
		for owner := range owners {
			for _, nested := range m.facts.typeByOwnerID[owner][nestedName] {
				nestedOwners = unionOwners(nestedOwners, ownerSet{nested: {}})
			}
		}
		owners = nestedOwners
		if len(owners) == 0 {
			return nil
		}
	}
	return owners
}

func (m *semanticModel) staticLoadOwners(context analysisContext, tokens []token) (ownerSet, bool) {
	tokens = trimExpression(tokens)
	if len(tokens) != 4 || (tokens[0].text != "preload" && tokens[0].text != "load") || tokens[1].text != "(" || tokens[2].kind != tokenString || tokens[3].text != ")" {
		return nil, false
	}
	path, ok := normalizeImportPath(tokens[2].text)
	if !ok {
		return nil, true
	}
	path = projectResourcePath(context.file.projectRoot, path)
	owner := m.facts.scriptOwnerByPath[path]
	if owner == "" {
		return nil, true
	}
	return ownerSet{owner: {}}, true
}

func terminalPropertyAccess(tokens []token) ([]token, string, bool) {
	tokens = trimExpression(tokens)
	if len(tokens) < 3 || tokens[len(tokens)-2].text != "." || tokens[len(tokens)-1].kind != tokenIdentifier {
		return nil, "", false
	}
	base := trimExpression(tokens[:len(tokens)-2])
	if len(base) == 0 {
		return nil, "", false
	}
	return base, tokens[len(tokens)-1].text, true
}

func propertyChain(tokens []token) ([]string, bool) {
	tokens = trimExpression(tokens)
	if len(tokens) < 3 || len(tokens)%2 == 0 {
		return nil, false
	}
	parts := make([]string, 0, (len(tokens)+1)/2)
	for index, tok := range tokens {
		if index%2 == 0 {
			if tok.kind != tokenIdentifier {
				return nil, false
			}
			parts = append(parts, tok.text)
		} else if tok.text != "." {
			return nil, false
		}
	}
	return parts, true
}

func topLevelToken(tokens []token, text string) int { return topLevelTokenAfter(tokens, text, -1) }

func topLevelTokenAfter(tokens []token, text string, start int) int {
	depth := 0
	for index := start + 1; index < len(tokens); index++ {
		switch tokens[index].text {
		case "(", "[", "{":
			depth++
		case ")", "]", "}":
			if depth > 0 {
				depth--
			}
		default:
			if depth == 0 && tokens[index].text == text {
				return index
			}
		}
	}
	return -1
}
