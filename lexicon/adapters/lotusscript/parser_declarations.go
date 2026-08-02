package main

import (
	"fmt"
	"strings"
)

func (state *analysisState) addType(file *parsedFile, line logicalLine, kind, name, visibility, base string, typeMembersPublic bool) declaration {
	qualified := file.path + "::" + name
	attributes := visibilityAttributes(visibility)
	if base != "" {
		attributes["base"] = base
	}
	lineSpan := line.span
	id := state.facts.addNode(kind, name, file.path, qualified, qualified, file.path, &lineSpan, attributes, "")
	state.facts.addEdge(file.moduleID, id, "defines", file.path, &lineSpan, nil)
	decl := declaration{
		id: id, kind: kind, name: name, ownerID: file.moduleID, ownerPath: file.path,
		public: declarationPublic(visibility, state.modulePublic[file.path]), qualifiedName: qualified,
		span: &lineSpan, typeMembersPublic: typeMembersPublic,
	}
	classKey := strings.ToLower(name)
	state.classesByName[classKey] = append(state.classesByName[classKey], decl)
	if base != "" {
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
	classID := ""
	if class != nil {
		classID = class.id
	}
	defaultPublic := state.modulePublic[file.path]
	if class != nil {
		defaultPublic = true
	}
	decl := declaration{
		classID: classID, className: className, external: external, id: id, kind: kind, name: name,
		ownerID: ownerID, ownerPath: file.path, public: declarationPublic(modifiers, defaultPublic),
		qualifiedName: qualified, span: &lineSpan,
	}
	key := strings.ToLower(name)
	state.callablesByName[key] = append(state.callablesByName[key], decl)
	if className == "" {
		state.globalsByName[key] = append(state.globalsByName[key], decl)
	} else {
		if state.methodsByClass[classID] == nil {
			state.methodsByClass[classID] = make(map[string][]declaration)
		}
		state.methodsByClass[classID][key] = append(state.methodsByClass[classID][key], decl)
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
		state.recordCallableSymbol(callable.id, name, dataType, id, file.path)
		state.facts.addEdge(callable.id, id, "contains", file.path, callable.span, nil)
	}
}
