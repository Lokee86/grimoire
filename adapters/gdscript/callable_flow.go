package main

func (m *semanticModel) inferExpressionCallables(context analysisContext, expression []token) ownerSet {
	expression = trimExpression(expression)
	if len(expression) == 0 {
		return nil
	}
	if lambda := lambdaExpressionTarget(context.file, expression); lambda != "" {
		return ownerSet{lambda: {}}
	}
	if branch := topLevelToken(expression, "if"); branch > 0 {
		if otherwise := topLevelTokenAfter(expression, "else", branch); otherwise > branch {
			return unionOwners(
				m.inferExpressionCallables(context, expression[:branch]),
				m.inferExpressionCallables(context, expression[otherwise+1:]),
			)
		}
	}
	if call := terminalCall(expression); call != nil {
		if call.name == "get" && len(call.args) > 0 {
			targets := m.callableMapLookup(context, call.receiver, stringLiteral(call.args[0]))
			if len(call.args) > 1 {
				targets = unionOwners(targets, m.inferExpressionCallables(context, call.args[1]))
			}
			if len(targets) > 0 {
				return targets
			}
		}
		if call.name == "Callable" && len(call.args) >= 2 {
			return m.callableConstructorTargets(context, call.args[0], call.args[1])
		}
		if call.name == "bind" || call.name == "bindv" || call.name == "unbind" {
			return m.inferExpressionCallables(context, call.receiver)
		}
		resolution := m.resolveCall(context, *call)
		var targets ownerSet
		for _, function := range resolution.functionTargets {
			targets = unionOwners(targets, m.returnCallables[function])
		}
		return targets
	}
	if name := simpleIdentifier(expression); name != "" {
		if values := m.callableBindingTargets(context, name); len(values) > 0 {
			return values
		}
		return sliceOwners(m.methodTargets(context.ownerID, name, false, false))
	}
	if parts, ok := propertyChain(expression); ok && len(parts) == 2 {
		if parts[0] == "self" {
			return sliceOwners(m.methodTargets(context.ownerID, parts[1], false, false))
		}
		owners := m.bindingOwners(context, parts[0])
		var targets ownerSet
		for owner := range owners {
			targets = unionOwners(targets, sliceOwners(m.methodTargets(owner, parts[1], false, false)))
		}
		return targets
	}
	return nil
}

func lambdaExpressionTarget(file *parsedFile, expression []token) string {
	for _, tok := range expression {
		if tok.text != "func" {
			continue
		}
		for i := range file.declarations {
			decl := &file.declarations[i]
			if decl.keyword == "lambda" && spanInt(decl.span, "start_line") == tok.line && spanInt(decl.span, "start_column") == tok.column {
				return decl.nodeID
			}
		}
	}
	return ""
}

func (m *semanticModel) callableConstructorTargets(context analysisContext, receiver, methodExpression []token) ownerSet {
	method := stringLiteral(methodExpression)
	if method == "" {
		return nil
	}
	owners := m.inferExpressionOwners(context, receiver)
	if len(owners) == 0 && simpleIdentifier(receiver) == "self" {
		owners = ownerSet{context.ownerID: {}}
	}
	var targets ownerSet
	for owner := range owners {
		targets = unionOwners(targets, sliceOwners(m.methodTargets(owner, method, false, false)))
	}
	return targets
}

func (m *semanticModel) callableBindingTargets(context analysisContext, name string) ownerSet {
	if context.functionID != "" {
		if values := m.localCallables[context.functionID][name]; len(values) > 0 {
			return cloneOwners(values)
		}
	}
	return m.memberCallableTargets(context.ownerID, name, make(map[string]bool))
}

func (m *semanticModel) memberCallableTargets(owner, name string, seen map[string]bool) ownerSet {
	if owner == "" || seen[owner] {
		return nil
	}
	seen[owner] = true
	if values := m.memberCallables[owner][name]; len(values) > 0 {
		return cloneOwners(values)
	}
	var result ownerSet
	for _, parent := range m.facts.parentByOwnerID[owner] {
		result = unionOwners(result, m.memberCallableTargets(parent, name, seen))
	}
	return result
}

func (m *semanticModel) addMemberCallable(owner, name string, values ownerSet) bool {
	if owner == "" || name == "" || len(values) == 0 {
		return false
	}
	if m.memberCallables[owner] == nil {
		m.memberCallables[owner] = make(map[string]ownerSet)
	}
	return mergeOwners(m.memberCallables[owner], name, values)
}

func (m *semanticModel) addLocalCallable(function, name string, values ownerSet) bool {
	if function == "" || name == "" || len(values) == 0 {
		return false
	}
	if m.localCallables[function] == nil {
		m.localCallables[function] = make(map[string]ownerSet)
	}
	return mergeOwners(m.localCallables[function], name, values)
}

func (m *semanticModel) addReturnCallable(function string, values ownerSet) bool {
	if function == "" || len(values) == 0 {
		return false
	}
	if m.returnCallables[function] == nil {
		m.returnCallables[function] = make(ownerSet)
	}
	changed := false
	for value := range values {
		if _, exists := m.returnCallables[function][value]; !exists {
			m.returnCallables[function][value] = struct{}{}
			changed = true
		}
	}
	return changed
}

func sliceOwners(values []string) ownerSet {
	if len(values) == 0 {
		return nil
	}
	result := make(ownerSet, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}
