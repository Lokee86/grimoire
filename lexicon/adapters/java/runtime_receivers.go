package main

import "strings"

type localReceiverDeclaration struct {
	declaredAt int
	name       string
	scopeEnd   int
	scopeStart int
	typeName   string
}

type receiverTypes struct {
	locals     []localReceiverDeclaration
	parameters map[string]string
}

type receiverBlock struct {
	end   int
	start int
}

func receiverDeclaredType(tokens []token, start, end int) string {
	typeName, next, valid := receiverTypePrefix(tokens, start, end)
	if !valid || next != end {
		return ""
	}
	return typeName
}

func receiverTypePrefix(tokens []token, start, end int) (string, int, bool) {
	if start >= end || !identifierToken(tokens[start].text) {
		return "", start, false
	}
	first := tokens[start].text
	if javaKeywords[first] && !primitiveTypes[first] {
		return "", start, false
	}
	typeName := first
	index := start + 1
	for {
		if index < end && tokens[index].text == "<" {
			close := matchingToken(tokens, index, end, "<", ">")
			if close < 0 {
				return "", start, false
			}
			index = close + 1
		}
		if index+1 >= end || tokens[index].text != "." || !identifierToken(tokens[index+1].text) {
			break
		}
		typeName += "." + tokens[index+1].text
		index += 2
	}
	return typeName, index, true
}

func collectReceiverTypes(callable *callableDeclaration, ignored map[int]bool) receiverTypes {
	result := receiverTypes{parameters: callable.parameterReceiverTypes}
	blocks := []receiverBlock{{start: callable.bodyStart, end: callable.bodyEnd}}
	paren, bracket := 0, 0
	statementStart := true
	for index := callable.bodyStart; index < callable.bodyEnd; index++ {
		text := callable.tokens[index].text
		if !ignored[index] && statementStart && paren == 0 && bracket == 0 {
			if declarations, valid := simpleLocalReceiverDeclarations(callable.tokens, index, callable.bodyEnd); valid {
				block := blocks[len(blocks)-1]
				for _, declaration := range declarations {
					declaration.scopeStart = block.start
					declaration.scopeEnd = block.end
					result.locals = append(result.locals, declaration)
				}
			}
			statementStart = false
		}
		switch text {
		case "(":
			paren++
		case ")":
			if paren > 0 {
				paren--
			}
		case "[":
			bracket++
		case "]":
			if bracket > 0 {
				bracket--
			}
		case "{":
			if close := matchingToken(callable.tokens, index, callable.bodyEnd, "{", "}"); close >= 0 {
				blocks = append(blocks, receiverBlock{start: index + 1, end: close})
			}
			statementStart = true
		case "}":
			if len(blocks) > 1 && blocks[len(blocks)-1].end == index {
				blocks = blocks[:len(blocks)-1]
			}
			statementStart = true
		case ";":
			if paren == 0 && bracket == 0 {
				statementStart = true
			}
		}
	}
	return result
}

func simpleLocalReceiverDeclarations(tokens []token, start, end int) ([]localReceiverDeclaration, bool) {
	statementEnd := localStatementEnd(tokens, start, end)
	if statementEnd < 0 {
		return nil, false
	}
	core := start
	if tokens[core].text == "final" {
		core++
	}
	typeName, nameIndex, valid := receiverTypePrefix(tokens, core, statementEnd)
	if !valid || nameIndex >= statementEnd || !identifierToken(tokens[nameIndex].text) {
		return nil, false
	}
	if next := nameIndex + 1; next < statementEnd && tokens[next].text != "=" && tokens[next].text != "," {
		return nil, false
	}
	result := []localReceiverDeclaration{{declaredAt: nameIndex, name: tokens[nameIndex].text, typeName: typeName}}
	paren, bracket, braces := 0, 0, 0
	for index := nameIndex + 1; index < statementEnd; index++ {
		switch tokens[index].text {
		case "(":
			paren++
		case ")":
			paren--
		case "[":
			bracket++
		case "]":
			bracket--
		case "{":
			braces++
		case "}":
			braces--
		case ",":
			if paren == 0 && bracket == 0 && braces == 0 {
				candidate := index + 1
				if candidate < statementEnd && identifierToken(tokens[candidate].text) {
					result = append(result, localReceiverDeclaration{declaredAt: candidate, name: tokens[candidate].text, typeName: typeName})
				}
			}
		}
	}
	return result, true
}

func localStatementEnd(tokens []token, start, end int) int {
	paren, bracket, braces := 0, 0, 0
	for index := start; index < end; index++ {
		switch tokens[index].text {
		case "(":
			paren++
		case ")":
			paren--
		case "[":
			bracket++
		case "]":
			bracket--
		case "{":
			braces++
		case "}":
			if braces == 0 {
				return -1
			}
			braces--
		case ";":
			if paren == 0 && bracket == 0 && braces == 0 {
				return index
			}
		}
	}
	return -1
}

func (receivers receiverTypes) declaredType(name string, use int) (string, bool) {
	best := -1
	for index, declaration := range receivers.locals {
		if declaration.name != name || declaration.declaredAt >= use || use >= declaration.scopeEnd || use < declaration.scopeStart {
			continue
		}
		if best < 0 || declaration.scopeStart > receivers.locals[best].scopeStart ||
			(declaration.scopeStart == receivers.locals[best].scopeStart && declaration.declaredAt > receivers.locals[best].declaredAt) {
			best = index
		}
	}
	if best >= 0 {
		return receivers.locals[best].typeName, true
	}
	typeName, found := receivers.parameters[name]
	return typeName, found
}

func (state *analysisState) qualifiedCallCandidates(callable *callableDeclaration, invocation invocationEvidence, receivers receiverTypes) ([]callableDeclaration, string) {
	if invocation.qualifier == "this" {
		return state.callableCandidates(callable.ownerID, invocation.name, invocation.arity, false, false), "dynamic-target"
	}
	if !strings.Contains(invocation.qualifier, ".") {
		if typeName, found := receivers.declaredType(invocation.qualifier, invocation.start); found {
			if typeName == "" || primitiveTypes[typeName] {
				return nil, "dynamic-target"
			}
			return state.typedReceiverCandidates(callable, invocation, typeName)
		}
	} else {
		root := invocation.qualifier[:strings.Index(invocation.qualifier, ".")]
		if _, found := receivers.declaredType(root, invocation.start); found {
			return nil, "dynamic-target"
		}
	}
	types := state.resolveTypeDeclarations(invocation.qualifier, callable.ownerQualifiedName, callable.context)
	if len(types) == 0 {
		if typeLikeQualifier(invocation.qualifier) {
			return nil, "external-target"
		}
		return nil, "dynamic-target"
	}
	if len(types) > 1 {
		return nil, "ambiguous-target"
	}
	return state.callableCandidates(types[0].id, invocation.name, invocation.arity, false, true), "unsupported-form"
}

func (state *analysisState) typedReceiverCandidates(callable *callableDeclaration, invocation invocationEvidence, typeName string) ([]callableDeclaration, string) {
	types := state.resolveTypeDeclarations(typeName, callable.ownerQualifiedName, callable.context)
	if len(types) == 0 {
		return nil, "external-target"
	}
	if len(types) > 1 {
		return nil, "ambiguous-target"
	}
	return state.instanceCallableCandidates(types[0].id, invocation.name, invocation.arity), "unsupported-form"
}

func (state *analysisState) typedIdentifierReceiver(callable *callableDeclaration, invocation invocationEvidence, receivers receiverTypes) bool {
	if invocation.qualifier == "this" || strings.Contains(invocation.qualifier, ".") {
		return false
	}
	typeName, found := receivers.declaredType(invocation.qualifier, invocation.start)
	if !found || typeName == "" || primitiveTypes[typeName] {
		return false
	}
	return len(state.resolveTypeDeclarations(typeName, callable.ownerQualifiedName, callable.context)) == 1
}
