package main

type ownerSet map[string]struct{}

type semanticModel struct {
	facts              *factSet
	files              []*parsedFile
	members            map[string]map[string]ownerSet
	locals             map[string]map[string]ownerSet
	returns            map[string]ownerSet
	memberCallables    map[string]map[string]ownerSet
	localCallables     map[string]map[string]ownerSet
	returnCallables    map[string]ownerSet
	memberCallableMaps map[string]map[string]keyedCallables
	localCallableMaps  map[string]map[string]keyedCallables
	returnCallableMaps map[string]keyedCallables
}

type analysisContext struct {
	file       *parsedFile
	functionID string
	ownerID    string
}

func buildSemanticModel(facts *factSet, files []*parsedFile) *semanticModel {
	model := &semanticModel{
		facts: facts, files: files,
		members:            make(map[string]map[string]ownerSet),
		locals:             make(map[string]map[string]ownerSet),
		returns:            make(map[string]ownerSet),
		memberCallables:    make(map[string]map[string]ownerSet),
		localCallables:     make(map[string]map[string]ownerSet),
		returnCallables:    make(map[string]ownerSet),
		memberCallableMaps: make(map[string]map[string]keyedCallables),
		localCallableMaps:  make(map[string]map[string]keyedCallables),
		returnCallableMaps: make(map[string]keyedCallables),
	}
	model.seedDeclaredTypes()
	for iteration := 0; iteration < 16; iteration++ {
		changed := false
		for _, file := range files {
			changed = model.inferDeclarations(file) || changed
			changed = model.inferStatements(file) || changed
		}
		for _, file := range files {
			changed = model.propagateArguments(file) || changed
		}
		if !changed {
			break
		}
	}
	return model
}

func (m *semanticModel) seedDeclaredTypes() {
	for _, file := range m.files {
		for i := range file.declarations {
			decl := &file.declarations[i]
			if decl.kind == "function" {
				context := analysisContext{file: file, functionID: decl.nodeID, ownerID: decl.ownerClassID}
				for _, name := range decl.parameterNames {
					m.addLocal(decl.nodeID, name, m.typeOwners(file.path, decl.parameterTypes[name]))
					if defaultValue := decl.parameterDefaults[name]; len(defaultValue) > 0 {
						m.addLocal(decl.nodeID, name, m.inferExpressionOwners(context, defaultValue))
						m.addLocalCallable(decl.nodeID, name, m.inferExpressionCallables(context, defaultValue))
						m.addLocalCallableMap(decl.nodeID, name, m.inferExpressionCallableMap(context, defaultValue))
					}
				}
				m.addReturn(decl.nodeID, m.typeOwners(file.path, decl.returnType))
				continue
			}
			if decl.kind != "variable" && decl.kind != "constant" {
				continue
			}
			owners := m.typeOwners(file.path, decl.typeName)
			if decl.ownerFunction != "" {
				m.addLocal(decl.ownerFunction, decl.name, owners)
			} else {
				m.addMember(decl.ownerID, decl.name, owners)
			}
		}
	}
}

func (m *semanticModel) inferDeclarations(file *parsedFile) bool {
	changed := false
	for i := range file.declarations {
		decl := &file.declarations[i]
		if len(decl.initializer) == 0 || (decl.kind != "variable" && decl.kind != "constant") {
			continue
		}
		context := analysisContext{file: file, functionID: decl.ownerFunction, ownerID: decl.ownerClassID}
		if context.ownerID == "" {
			context.ownerID = file.scriptOwnerID
		}
		owners := m.inferExpressionOwners(context, decl.initializer)
		callables := m.inferExpressionCallables(context, decl.initializer)
		callableMap := m.inferExpressionCallableMap(context, decl.initializer)
		if decl.ownerFunction != "" {
			changed = m.addLocal(decl.ownerFunction, decl.name, owners) || changed
			changed = m.addLocalCallable(decl.ownerFunction, decl.name, callables) || changed
			changed = m.addLocalCallableMap(decl.ownerFunction, decl.name, callableMap) || changed
		} else {
			changed = m.addMember(decl.ownerID, decl.name, owners) || changed
			changed = m.addMemberCallable(decl.ownerID, decl.name, callables) || changed
			changed = m.addMemberCallableMap(decl.ownerID, decl.name, callableMap) || changed
		}
	}
	return changed
}

func (m *semanticModel) inferStatements(file *parsedFile) bool {
	changed := false
	for _, stmt := range file.statements {
		context := contextForStatement(file, stmt)
		if len(stmt.tokens) == 0 {
			continue
		}
		if stmt.tokens[0].text == "return" && context.functionID != "" {
			changed = m.addReturn(context.functionID, m.inferExpressionOwners(context, stmt.tokens[1:])) || changed
			changed = m.addReturnCallable(context.functionID, m.inferExpressionCallables(context, stmt.tokens[1:])) || changed
			changed = m.addReturnCallableMap(context.functionID, m.inferExpressionCallableMap(context, stmt.tokens[1:])) || changed
		}
		assignment := topLevelAssignment(stmt.tokens)
		if assignment < 0 || parseDeclaration(stmt) != nil {
			continue
		}
		left := stmt.tokens[:assignment]
		owners := m.inferExpressionOwners(context, stmt.tokens[assignment+1:])
		callables := m.inferExpressionCallables(context, stmt.tokens[assignment+1:])
		callableMap := m.inferExpressionCallableMap(context, stmt.tokens[assignment+1:])
		if memberName, targetOwners := m.assignmentMemberTargets(context, left); memberName != "" && len(targetOwners) > 0 {
			for targetOwner := range targetOwners {
				changed = m.addMember(targetOwner, memberName, owners) || changed
				changed = m.addMemberCallable(targetOwner, memberName, callables) || changed
				changed = m.addMemberCallableMap(targetOwner, memberName, callableMap) || changed
			}
			continue
		}
		name, forceMember := assignmentName(left)
		if name == "" {
			continue
		}
		if context.functionID == "" || forceMember || m.memberDeclared(context.ownerID, name) && !m.localDeclared(context.functionID, name) {
			changed = m.addMember(context.ownerID, name, owners) || changed
			changed = m.addMemberCallable(context.ownerID, name, callables) || changed
			changed = m.addMemberCallableMap(context.ownerID, name, callableMap) || changed
		} else {
			changed = m.addLocal(context.functionID, name, owners) || changed
			changed = m.addLocalCallable(context.functionID, name, callables) || changed
			changed = m.addLocalCallableMap(context.functionID, name, callableMap) || changed
		}
	}
	return changed
}

func (m *semanticModel) propagateArguments(file *parsedFile) bool {
	changed := false
	for _, stmt := range file.statements {
		context := contextForStatement(file, stmt)
		for _, call := range findCalls(stmt, file.path) {
			resolution := m.resolveCall(context, call)
			for _, target := range resolution.functionTargets {
				changed = m.addArgumentsToFunction(context, target, call.args) || changed
			}
			for _, owner := range resolution.constructorOwners {
				for _, target := range m.methodTargets(owner, "_init", false, false) {
					changed = m.addArgumentsToFunction(context, target, call.args) || changed
				}
			}
		}
	}
	return changed
}

func (m *semanticModel) addArgumentsToFunction(context analysisContext, target string, args [][]token) bool {
	decl := m.facts.declarationByID[target]
	if decl == nil {
		return false
	}
	changed := false
	for index, argument := range args {
		if index >= len(decl.parameterNames) {
			break
		}
		changed = m.addLocal(target, decl.parameterNames[index], m.inferExpressionOwners(context, argument)) || changed
		changed = m.addLocalCallable(target, decl.parameterNames[index], m.inferExpressionCallables(context, argument)) || changed
		changed = m.addLocalCallableMap(target, decl.parameterNames[index], m.inferExpressionCallableMap(context, argument)) || changed
	}
	return changed
}

func contextForStatement(file *parsedFile, stmt statement) analysisContext {
	return contextForPosition(file, stmt.start.line, stmt.indent+1)
}

func contextForPosition(file *parsedFile, line, column int) analysisContext {
	context := analysisContext{file: file, ownerID: file.scriptOwnerID}
	bestLine, bestColumn, bestIndent := -1, -1, -1
	positionIndent := column - 1
	for i := range file.declarations {
		decl := &file.declarations[i]
		declLine := spanInt(decl.span, "start_line")
		declColumn := spanInt(decl.span, "start_column")
		if declLine > line || declLine == line && declColumn >= column || decl.indent >= positionIndent {
			continue
		}
		if decl.kind == "function" && (declLine > bestLine || declLine == bestLine && declColumn > bestColumn || declLine == bestLine && declColumn == bestColumn && decl.indent > bestIndent) {
			context.functionID, context.ownerID = decl.nodeID, decl.ownerClassID
			bestLine, bestColumn, bestIndent = declLine, declColumn, decl.indent
		}
	}
	if context.ownerID == "" {
		context.ownerID = file.scriptOwnerID
	}
	return context
}

func topLevelAssignment(tokens []token) int {
	depth := 0
	for i, tok := range tokens {
		switch tok.text {
		case "(", "[", "{":
			depth++
		case ")", "]", "}":
			if depth > 0 {
				depth--
			}
		case "=", ":=":
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func (m *semanticModel) assignmentMemberTargets(context analysisContext, tokens []token) (string, ownerSet) {
	tokens = trimExpression(tokens)
	parts, ok := propertyChain(tokens)
	if !ok || len(parts) < 2 {
		return "", nil
	}
	member := parts[len(parts)-1]
	receiver := tokens[:len(tokens)-2]
	if len(parts) == 2 && parts[0] == "self" {
		return member, ownerSet{context.ownerID: {}}
	}
	return member, m.inferExpressionOwners(context, receiver)
}

func assignmentName(tokens []token) (string, bool) {
	tokens = trimExpression(tokens)
	if len(tokens) == 1 && tokens[0].kind == tokenIdentifier {
		return tokens[0].text, false
	}
	if len(tokens) == 3 && tokens[0].text == "self" && tokens[1].text == "." && tokens[2].kind == tokenIdentifier {
		return tokens[2].text, true
	}
	return "", false
}
