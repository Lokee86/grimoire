package main

import "strings"

func (parser *javaParser) parseMembers(ownerID, ownerQualifiedName, ownerName string, start, limit int) {
	for index := start; index < limit; {
		if parser.tokens[index].text == ";" {
			index++
			continue
		}
		core := skipPrefix(parser.tokens, index, limit)
		if _, _, ok := typeKeywordAt(parser.tokens, core); ok {
			next := parser.parseType(index, core, limit, ownerID, ownerQualifiedName, "contains")
			if next > index {
				index = next
				continue
			}
		}
		delimiter, delimiterText := findMemberDelimiter(parser.tokens, index, limit)
		if delimiter < 0 {
			parser.facts.addUnresolved(ownerID, "defines", sourceExcerpt(parser.source, parser.tokens, index, limit), "unsupported-form", parser.path, parser.tokenSpan(index, limit), nil)
			return
		}
		if delimiterText == "{" {
			close := matchingToken(parser.tokens, delimiter, limit, "{", "}")
			if close < 0 {
				parser.facts.addUnresolved(ownerID, "defines", sourceExcerpt(parser.source, parser.tokens, index, delimiter+1), "unsupported-form", parser.path, parser.tokenSpan(index, delimiter+1), nil)
				return
			}
			if parser.parseCallable(ownerID, ownerQualifiedName, ownerName, index, delimiter, close+1) || initializerHeader(parser.tokens, core, delimiter) {
				index = close + 1
				continue
			}
			parser.facts.addUnresolved(ownerID, "defines", sourceExcerpt(parser.source, parser.tokens, index, close+1), "unsupported-form", parser.path, parser.tokenSpan(index, close+1), nil)
			index = close + 1
			continue
		}
		if parser.parseCallable(ownerID, ownerQualifiedName, ownerName, index, delimiter, delimiter+1) {
			index = delimiter + 1
			continue
		}
		if parser.parseFields(ownerID, ownerQualifiedName, index, delimiter) {
			index = delimiter + 1
			continue
		}
		parser.facts.addUnresolved(ownerID, "defines", sourceExcerpt(parser.source, parser.tokens, index, delimiter+1), "unsupported-form", parser.path, parser.tokenSpan(index, delimiter+1), nil)
		index = delimiter + 1
	}
}

func (parser *javaParser) parseCallable(ownerID, ownerQualifiedName, ownerName string, start, headerEnd, declarationEnd int) bool {
	core := skipPrefix(parser.tokens, start, headerEnd)
	if core >= headerEnd {
		return false
	}
	open, close := callableParameters(parser.tokens, core, headerEnd)
	compact := false
	nameIndex := -1
	if open >= 0 {
		nameIndex = open - 1
		if nameIndex < core || !identifierToken(parser.tokens[nameIndex].text) {
			return false
		}
	} else if headerEnd == core+1 && parser.tokens[core].text == ownerName {
		compact = true
		nameIndex = core
	} else {
		return false
	}
	name := parser.tokens[nameIndex].text
	constructor := name == ownerName
	if !constructor && nameIndex == core {
		return false
	}
	parameterTypes := []string{}
	parameters := []parameterDeclaration{}
	if compact {
		parameterTypes = append(parameterTypes, parser.recordComponentTypes[ownerQualifiedName]...)
	} else {
		parameters = parseParameters(parser.tokens, open+1, close)
		if parameters == nil && close > open+1 {
			return false
		}
		for _, parameter := range parameters {
			parameterTypes = append(parameterTypes, parameter.typeName)
		}
	}
	signature := strings.Join(parameterTypes, ",")
	kind := "method"
	qualifiedName := ownerQualifiedName + "." + name + "(" + signature + ")"
	declarationKind := "method"
	if constructor {
		kind = "constructor"
		qualifiedName = ownerQualifiedName + ".<init>(" + signature + ")"
		declarationKind = "constructor"
		if compact {
			declarationKind = "compact-constructor"
		}
	}
	attributes := map[string]any{"declaration_kind": declarationKind, "parameter_types": parameterTypes}
	if !constructor {
		returnType := normalizedTokens(parser.tokens[core:nameIndex])
		if returnType == "" {
			return false
		}
		attributes["return_type"] = returnType
	}
	modifiers := modifierList(parser.tokens, start, core)
	if len(modifiers) != 0 {
		attributes["modifiers"] = modifiers
	}
	callableSpan := parser.tokenSpan(core, declarationEnd)
	id := parser.facts.addNode(kind, name, parser.path, qualifiedName, qualifiedName, parser.path, callableSpan, attributes, "")
	parser.state.registerDeclaration(qualifiedName, id)
	parser.facts.addEdge(ownerID, id, "contains", parser.path, callableSpan, nil)
	parser.queueAnnotations(id, ownerQualifiedName, start, core)
	parameterIDs := make(map[string]string, len(parameters))
	parameterReceiverTypes := make(map[string]string, len(parameters))
	for index, parameter := range parameters {
		parameterQualifiedName := qualifiedName + "#parameter:" + decimal(index) + ":" + parameter.name
		parameterSpan := parser.tokenSpan(parameter.start, parameter.end)
		parameterID := parser.facts.addNode("parameter", parameter.name, parser.path, parameterQualifiedName, parameterQualifiedName, parser.path, parameterSpan, map[string]any{
			"index": index, "type": parameter.typeName, "varargs": parameter.varargs,
		}, "")
		parser.state.registerDeclaration(parameterQualifiedName, parameterID)
		parameterIDs[parameter.name] = parameterID
		parameterReceiverTypes[parameter.name] = parameter.receiverType
		parser.facts.addEdge(id, parameterID, "contains", parser.path, parameterSpan, nil)
		parameterCore := skipPrefix(parser.tokens, parameter.start, parameter.end)
		parser.queueAnnotations(parameterID, ownerQualifiedName, parameter.start, parameterCore)
	}
	bodyStart, bodyEnd := -1, -1
	if headerEnd < declarationEnd && parser.tokens[headerEnd].text == "{" {
		bodyStart, bodyEnd = headerEnd+1, declarationEnd-1
	}
	parser.state.registerCallable(callableDeclaration{
		arity: len(parameterTypes), bodyEnd: bodyEnd, bodyStart: bodyStart, constructor: constructor,
		context: parser.resolution.clone(), id: id, modifiers: modifiers, name: name,
		ownerID: ownerID, ownerQualifiedName: ownerQualifiedName, parameterIDs: parameterIDs,
		parameterReceiverTypes: parameterReceiverTypes,
		parameterTypes:         append([]string(nil), parameterTypes...), path: parser.path,
		signature: signature, source: parser.source, tokens: parser.tokens,
	})
	return true
}

func (parser *javaParser) parseFields(ownerID, ownerQualifiedName string, start, end int) bool {
	core := skipPrefix(parser.tokens, start, end)
	segments := splitTopLevel(parser.tokens, core, end, ",")
	if len(segments) == 0 {
		return false
	}
	firstName := variableName(parser.tokens, segments[0][0], segments[0][1])
	if firstName < 0 || firstName == core {
		return false
	}
	typeName := normalizedTokens(parser.tokens[core:firstName])
	if typeName == "" {
		return false
	}
	modifiers := modifierList(parser.tokens, start, core)
	for index, segment := range segments {
		nameIndex := variableName(parser.tokens, segment[0], segment[1])
		if nameIndex < 0 {
			return false
		}
		name := parser.tokens[nameIndex].text
		qualifiedName := ownerQualifiedName + "." + name
		attributes := map[string]any{"declaration_kind": "field", "type": typeName}
		if len(modifiers) != 0 {
			attributes["modifiers"] = modifiers
		}
		fieldSpan := parser.tokenSpan(segment[0], segment[1])
		id := parser.facts.addNode("field", name, parser.path, qualifiedName, qualifiedName, parser.path, fieldSpan, attributes, "")
		parser.state.registerDeclaration(qualifiedName, id)
		parser.state.registerField(fieldDeclaration{id: id, modifiers: modifiers, name: name, ownerID: ownerID})
		parser.facts.addEdge(ownerID, id, "contains", parser.path, fieldSpan, map[string]any{"declarator_index": index})
		parser.queueAnnotations(id, ownerQualifiedName, start, core)
	}
	return true
}

func (parser *javaParser) parseRecordComponents(ownerID, ownerQualifiedName string, start, end int) []string {
	open := findToken(parser.tokens, start, end, "(")
	if open < 0 {
		return nil
	}
	close := matchingToken(parser.tokens, open, end, "(", ")")
	if close < 0 {
		parser.facts.addUnresolved(ownerID, "defines", "record component list", "unsupported-form", parser.path, parser.tokenSpan(open, min(end, open+1)), nil)
		return nil
	}
	components := parseParameters(parser.tokens, open+1, close)
	if components == nil && close > open+1 {
		parser.facts.addUnresolved(ownerID, "defines", sourceExcerpt(parser.source, parser.tokens, open, close+1), "unsupported-form", parser.path, parser.tokenSpan(open, close+1), nil)
		return nil
	}
	types := make([]string, 0, len(components))
	for index, component := range components {
		types = append(types, component.typeName)
		qualifiedName := ownerQualifiedName + "." + component.name
		componentSpan := parser.tokenSpan(component.start, component.end)
		id := parser.facts.addNode("field", component.name, parser.path, qualifiedName, qualifiedName, parser.path, componentSpan, map[string]any{
			"declaration_kind": "record-component", "index": index, "type": component.typeName,
		}, "")
		parser.state.registerDeclaration(qualifiedName, id)
		parser.state.registerField(fieldDeclaration{id: id, name: component.name, ownerID: ownerID})
		parser.facts.addEdge(ownerID, id, "contains", parser.path, componentSpan, nil)
		componentCore := skipPrefix(parser.tokens, component.start, component.end)
		parser.queueAnnotations(id, ownerQualifiedName, component.start, componentCore)
	}
	return types
}

func (parser *javaParser) tokenSpan(start, end int) *span {
	if start < 0 || start >= len(parser.tokens) || end <= start {
		return nil
	}
	if end > len(parser.tokens) {
		end = len(parser.tokens)
	}
	first, last := parser.tokens[start], parser.tokens[end-1]
	return &span{EndColumn: last.endColumn, EndLine: last.endLine, Path: parser.path, StartColumn: first.column, StartLine: first.line}
}
