package main

type keyedCallables map[string]ownerSet

func (m *semanticModel) inferExpressionCallableMap(context analysisContext, expression []token) keyedCallables {
	expression = trimExpression(expression)
	if len(expression) == 0 {
		return nil
	}
	if branch := topLevelToken(expression, "if"); branch > 0 {
		if otherwise := topLevelTokenAfter(expression, "else", branch); otherwise > branch {
			return mergeKeyedCallables(
				m.inferExpressionCallableMap(context, expression[:branch]),
				m.inferExpressionCallableMap(context, expression[otherwise+1:]),
			)
		}
	}
	if expression[0].text == "{" && expression[len(expression)-1].text == "}" {
		return m.dictionaryLiteralCallables(context, expression[1:len(expression)-1])
	}
	if name := simpleIdentifier(expression); name != "" {
		return m.callableMapBinding(context, name)
	}
	if call := terminalCall(expression); call != nil {
		resolution := m.resolveCall(context, *call)
		var result keyedCallables
		for _, function := range resolution.functionTargets {
			result = mergeKeyedCallables(result, m.returnCallableMaps[function])
		}
		return result
	}
	return nil
}

func (m *semanticModel) dictionaryLiteralCallables(context analysisContext, entries []token) keyedCallables {
	result := make(keyedCallables)
	for _, entry := range splitArguments(entries) {
		colon := topLevelToken(entry, ":")
		if colon <= 0 || colon+1 >= len(entry) {
			continue
		}
		key := stringLiteral(entry[:colon])
		if key == "" {
			continue
		}
		if targets := m.inferExpressionCallables(context, entry[colon+1:]); len(targets) > 0 {
			result[key] = unionOwners(result[key], targets)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func (m *semanticModel) callableMapBinding(context analysisContext, name string) keyedCallables {
	if context.functionID != "" {
		if values := m.localCallableMaps[context.functionID][name]; len(values) > 0 {
			return cloneKeyedCallables(values)
		}
	}
	return m.memberCallableMap(context.ownerID, name, make(map[string]bool))
}

func (m *semanticModel) memberCallableMap(owner, name string, seen map[string]bool) keyedCallables {
	if owner == "" || seen[owner] {
		return nil
	}
	seen[owner] = true
	if values := m.memberCallableMaps[owner][name]; len(values) > 0 {
		return cloneKeyedCallables(values)
	}
	var result keyedCallables
	for _, parent := range m.facts.parentByOwnerID[owner] {
		result = mergeKeyedCallables(result, m.memberCallableMap(parent, name, seen))
	}
	return result
}

func (m *semanticModel) callableMapLookup(context analysisContext, receiver []token, key string) ownerSet {
	if key == "" {
		return nil
	}
	if name := simpleIdentifier(receiver); name != "" {
		return cloneOwners(m.callableMapBinding(context, name)[key])
	}
	if call := terminalCall(receiver); call != nil {
		var result ownerSet
		resolution := m.resolveCall(context, *call)
		for _, function := range resolution.functionTargets {
			result = unionOwners(result, m.returnCallableMaps[function][key])
		}
		return result
	}
	return nil
}

func (m *semanticModel) addMemberCallableMap(owner, name string, values keyedCallables) bool {
	if owner == "" || name == "" || len(values) == 0 {
		return false
	}
	if m.memberCallableMaps[owner] == nil {
		m.memberCallableMaps[owner] = make(map[string]keyedCallables)
	}
	merged, changed := mergeKeyedCallablesChanged(m.memberCallableMaps[owner][name], values)
	m.memberCallableMaps[owner][name] = merged
	return changed
}

func (m *semanticModel) addLocalCallableMap(function, name string, values keyedCallables) bool {
	if function == "" || name == "" || len(values) == 0 {
		return false
	}
	if m.localCallableMaps[function] == nil {
		m.localCallableMaps[function] = make(map[string]keyedCallables)
	}
	merged, changed := mergeKeyedCallablesChanged(m.localCallableMaps[function][name], values)
	m.localCallableMaps[function][name] = merged
	return changed
}

func (m *semanticModel) addReturnCallableMap(function string, values keyedCallables) bool {
	if function == "" || len(values) == 0 {
		return false
	}
	merged, changed := mergeKeyedCallablesChanged(m.returnCallableMaps[function], values)
	m.returnCallableMaps[function] = merged
	return changed
}

func mergeKeyedCallables(values ...keyedCallables) keyedCallables {
	result, _ := mergeKeyedCallablesChanged(nil, values...)
	return result
}

func mergeKeyedCallablesChanged(existing keyedCallables, values ...keyedCallables) (keyedCallables, bool) {
	result := cloneKeyedCallables(existing)
	if result == nil {
		result = make(keyedCallables)
	}
	changed := false
	for _, value := range values {
		for key, targets := range value {
			before := len(result[key])
			result[key] = unionOwners(result[key], targets)
			if len(result[key]) != before {
				changed = true
			}
		}
	}
	if len(result) == 0 {
		return nil, changed
	}
	return result, changed
}

func cloneKeyedCallables(values keyedCallables) keyedCallables {
	if len(values) == 0 {
		return nil
	}
	result := make(keyedCallables, len(values))
	for key, targets := range values {
		result[key] = cloneOwners(targets)
	}
	return result
}
