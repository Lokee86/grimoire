package main

import (
	"fmt"
	"strings"
)

func (state *analysisState) addVariables(file *parsedFile, line logicalLine, class, callable *declaration) bool {
	text := strings.TrimSpace(line.text)
	kind := ""
	body := ""
	visibility := ""
	if redimBody, ok := redimVariableBody(text); ok {
		kind, body = "variable", redimBody
	} else if match := constPattern.FindStringSubmatch(text); match != nil {
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
		if index == 0 {
			if class == nil || callable != nil || !looksLikeMemberDeclaration(text) {
				return false
			}
			body = text
		} else {
			if index >= len(fields) {
				return false
			}
			body = strings.TrimSpace(strings.Join(fields[index:], " "))
		}
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
		identity := fmt.Sprintf("%s::%s::%s::%d", ownerQualified, kind, strings.ToLower(name), line.span.StartLine)
		qualified := ownerQualified + "::" + name
		lineSpan := line.span
		id := state.facts.addNode(kind, name, file.path, qualified, identity, file.path, &lineSpan, attributes, "")
		defaultPublic := state.modulePublic[file.path]
		if class != nil {
			defaultPublic = class.typeMembersPublic
		}
		symbolPublic := declarationPublic(visibility, defaultPublic)
		switch {
		case callable != nil:
			state.recordCallableVariable(callable.id, name, dataType)
			state.recordCallableSymbol(callable.id, name, dataType, id, file.path)
		case class != nil:
			state.recordClassField(class.id, name, dataType)
			state.recordClassFieldSymbol(class.id, name, dataType, id, file.path, symbolPublic)
		default:
			state.recordModuleVariable(file.path, name, dataType)
			state.recordModuleSymbol(file.path, name, dataType, id, symbolPublic)
		}
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
