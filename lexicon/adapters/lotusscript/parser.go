package main

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	callablePattern = regexp.MustCompile(`(?i)^((?:(?:Public|Private|Protected|Static)\s+)*)(Sub|Function|Property\s+(?:Get|Set))\s+([A-Za-z_][A-Za-z0-9_]*)\s*(\([^)]*\))?`)
	classPattern    = regexp.MustCompile(`(?i)^(?:(Public|Private|Protected)\s+)?Class\s+([A-Za-z_][A-Za-z0-9_]*)(?:\s+As\s+([A-Za-z_][A-Za-z0-9_.]*))?`)
	constPattern    = regexp.MustCompile(`(?i)^((?:(?:Public|Private|Protected)\s+)*)Const\s+(.+)$`)
	externalPattern = regexp.MustCompile(`(?i)^Declare\s+((?:(?:Public|Private)\s+)*)(Sub|Function)\s+([A-Za-z_][A-Za-z0-9_]*)\s*(\([^)]*\))?(.*)$`)
	typePattern     = regexp.MustCompile(`(?i)^(?:(Public|Private)\s+)?Type\s+([A-Za-z_][A-Za-z0-9_]*)`)
	usePattern      = regexp.MustCompile(`(?i)^(UseLSX|Use)\s+(.+)$`)
)

func parseLotusScript(state *analysisState, file *parsedFile) {
	var currentClass *declaration
	var currentCallable *declaration
	for _, line := range logicalLines(file.path, string(file.content)) {
		lower := strings.ToLower(strings.TrimSpace(line.text))
		switch lower {
		case "end sub", "end function", "end property":
			currentCallable = nil
			continue
		case "end class", "end type":
			currentCallable = nil
			currentClass = nil
			continue
		}

		if match := usePattern.FindStringSubmatch(line.text); match != nil {
			state.addUse(file, line, match[1], match[2])
			continue
		}
		if match := classPattern.FindStringSubmatch(line.text); match != nil {
			decl := state.addType(file, line, "type", match[2], match[1], match[3])
			currentClass = &decl
			currentCallable = nil
			continue
		}
		if match := typePattern.FindStringSubmatch(line.text); match != nil {
			decl := state.addType(file, line, "type", match[2], match[1], "")
			currentClass = &decl
			currentCallable = nil
			continue
		}
		if match := externalPattern.FindStringSubmatch(line.text); match != nil && containsWord(match[5], "lib") {
			state.addCallable(file, line, currentClass, match[2], match[3], match[1], match[4], true, match[5])
			continue
		}
		if match := callablePattern.FindStringSubmatch(line.text); match != nil {
			decl := state.addCallable(file, line, currentClass, match[2], match[3], match[1], match[4], false, "")
			currentCallable = &decl
			continue
		}

		if state.addVariables(file, line, currentClass, currentCallable) {
			if currentCallable != nil {
				state.collectCalls(line, *currentCallable, currentClass)
			}
			continue
		}
		if currentCallable != nil {
			state.collectCalls(line, *currentCallable, currentClass)
		}
	}
}

func (state *analysisState) addType(file *parsedFile, line logicalLine, kind, name, visibility, base string) declaration {
	qualified := file.path + "::" + name
	attributes := visibilityAttributes(visibility)
	if base != "" {
		attributes["base"] = base
	}
	lineSpan := line.span
	id := state.facts.addNode(kind, name, file.path, qualified, qualified, file.path, &lineSpan, attributes, "")
	state.facts.addEdge(file.moduleID, id, "defines", file.path, &lineSpan, nil)
	decl := declaration{id: id, kind: kind, name: name, ownerID: file.moduleID, ownerPath: file.path, qualifiedName: qualified, span: &lineSpan}
	classKey := strings.ToLower(name)
	state.classesByName[classKey] = append(state.classesByName[classKey], decl)
	if base != "" {
		state.basesByClass[classKey] = append(state.basesByClass[classKey], strings.ToLower(base))
		state.extends = append(state.extends, extendsEvidence{base: base, classID: id, ownerPath: file.path, span: &lineSpan})
	}
	return decl
}

func (state *analysisState) addCallable(file *parsedFile, line logicalLine, class *declaration, form, name, modifiers, parameters string, external bool, tail string) declaration {
	ownerID := file.moduleID
	ownerQualified := file.path
	className := ""
	kind := "function"
	if class != nil {
		ownerID = class.id
		ownerQualified = class.qualifiedName
		className = class.name
		kind = "method"
	}
	lowerForm := strings.ToLower(form)
	if class != nil && strings.EqualFold(name, "New") {
		kind = "constructor"
	}
	attributes := visibilityAttributes(modifiers)
	attributes["form"] = lowerForm
	if strings.HasPrefix(lowerForm, "property") {
		attributes["property_accessor"] = strings.TrimSpace(strings.TrimPrefix(lowerForm, "property"))
	}
	if external {
		attributes["external"] = true
		if library := libraryName(tail); library != "" {
			attributes["library"] = library
		}
	}
	identity := fmt.Sprintf("%s::%s::%s::%s", ownerQualified, kind, strings.ToLower(name), lowerForm)
	qualified := ownerQualified + "::" + name
	lineSpan := line.span
	id := state.facts.addNode(kind, name, file.path, qualified, identity, file.path, &lineSpan, attributes, "")
	state.facts.addEdge(ownerID, id, "defines", file.path, &lineSpan, nil)
	decl := declaration{
		className: className, external: external, id: id, kind: kind, name: name,
		ownerID: ownerID, ownerPath: file.path, qualifiedName: qualified, span: &lineSpan,
	}
	key := strings.ToLower(name)
	state.callablesByName[key] = append(state.callablesByName[key], decl)
	if className == "" {
		state.globalsByName[key] = append(state.globalsByName[key], decl)
	} else {
		classKey := strings.ToLower(className)
		if state.methodsByClass[classKey] == nil {
			state.methodsByClass[classKey] = make(map[string][]declaration)
		}
		state.methodsByClass[classKey][key] = append(state.methodsByClass[classKey][key], decl)
	}
	state.addParameters(file, decl, parameters)
	return decl
}

func (state *analysisState) addParameters(file *parsedFile, callable declaration, parameters string) {
	parameters = strings.TrimSpace(parameters)
	if len(parameters) < 2 {
		return
	}
	parameters = parameters[1 : len(parameters)-1]
	for index, part := range splitTopLevel(parameters, ',') {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		name := parameterName(part)
		if name == "" {
			continue
		}
		attributes := map[string]any{"position": index}
		dataType := declaredType(part)
		if dataType != "" {
			attributes["type"] = dataType
		}
		state.recordCallableVariable(callable.id, name, dataType)
		for _, modifier := range []string{"byval", "byref", "optional", "paramarray"} {
			if containsWord(part, modifier) {
				attributes[modifier] = true
			}
		}
		identity := fmt.Sprintf("%s::parameter::%d::%s", callable.qualifiedName, index, strings.ToLower(name))
		id := state.facts.addNode("parameter", name, file.path, callable.qualifiedName+"::"+name, identity, file.path, callable.span, attributes, "")
		state.facts.addEdge(callable.id, id, "contains", file.path, callable.span, nil)
	}
}

func (state *analysisState) addVariables(file *parsedFile, line logicalLine, class, callable *declaration) bool {
	text := strings.TrimSpace(line.text)
	kind := ""
	body := ""
	visibility := ""
	if match := constPattern.FindStringSubmatch(text); match != nil {
		kind, body, visibility = "constant", match[2], match[1]
	} else {
		fields := strings.Fields(text)
		index := 0
		for index < len(fields) && isVariableModifier(fields[index]) {
			if !strings.EqualFold(fields[index], "Dim") {
				visibility += fields[index] + " "
			}
			index++
		}
		if index == 0 || index >= len(fields) {
			return false
		}
		body = strings.TrimSpace(strings.Join(fields[index:], " "))
		if callable != nil {
			kind = "variable"
		} else {
			kind = "field"
		}
	}

	ownerID := file.moduleID
	ownerQualified := file.path
	if class != nil {
		ownerID, ownerQualified = class.id, class.qualifiedName
	}
	if callable != nil {
		ownerID, ownerQualified = callable.id, callable.qualifiedName
	}
	for _, part := range splitTopLevel(body, ',') {
		name := identifierPrefix(part)
		if name == "" {
			continue
		}
		attributes := visibilityAttributes(visibility)
		dataType := declaredType(part)
		if dataType != "" {
			attributes["type"] = dataType
		}
		switch {
		case callable != nil:
			state.recordCallableVariable(callable.id, name, dataType)
		case class != nil:
			state.recordClassField(class.name, name, dataType)
		default:
			state.recordModuleVariable(file.path, name, dataType)
		}
		identity := fmt.Sprintf("%s::%s::%s::%d", ownerQualified, kind, strings.ToLower(name), line.span.StartLine)
		qualified := ownerQualified + "::" + name
		lineSpan := line.span
		id := state.facts.addNode(kind, name, file.path, qualified, identity, file.path, &lineSpan, attributes, "")
		relation := "defines"
		if callable != nil {
			relation = "contains"
		}
		state.facts.addEdge(ownerID, id, relation, file.path, &lineSpan, nil)
	}
	return true
}

func (state *analysisState) addUse(file *parsedFile, line logicalLine, keyword, expression string) {
	target, literal := literalValue(expression)
	identityTarget := strings.TrimSpace(expression)
	if literal {
		identityTarget = target
	}
	identity := fmt.Sprintf("%s::import::%d::%s::%s", file.path, line.span.StartLine, strings.ToLower(keyword), strings.ToLower(identityTarget))
	lineSpan := line.span
	id := state.facts.addNode("import", identityTarget, file.path, identity, identity, file.path, &lineSpan, map[string]any{"keyword": keyword, "target": identityTarget}, "")
	state.facts.addEdge(file.moduleID, id, "defines", file.path, &lineSpan, nil)
	state.uses = append(state.uses, useEvidence{
		dynamic: !literal, expression: strings.TrimSpace(expression), importID: id,
		keyword: keyword, ownerPath: file.path, span: &lineSpan, target: target,
	})
}
