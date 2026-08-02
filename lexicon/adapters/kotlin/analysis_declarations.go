package main

import (
	"fmt"
	"strings"
)

func (state *analysis) emitDeclaration(file *parsedKotlinFile, declaration *declaration, ownerID, ownerQN, ownerCanonical, ownerKind string, occurrences map[string]int) string {
	kind := declaration.kind
	if kind == "function" && (ownerKind == "type" || ownerKind == "interface") {
		kind = "method"
	}
	qualifiedBase := qualify(ownerQN, declaration.name)
	canonical := ""
	qualifiedName := qualifiedBase
	switch kind {
	case "function", "method":
		signature := parameterSignature(declaration.parameters)
		receiver := declaration.receiver
		canonical = fmt.Sprintf("%s::callable:%s::receiver:%s(%s)", ownerCanonical, declaration.name, receiver, signature)
		qualifiedName = fmt.Sprintf("%s(%s)", qualifiedBase, signature)
	case "constructor":
		signature := parameterSignature(declaration.parameters)
		canonical = fmt.Sprintf("%s::constructor(%s)", ownerCanonical, signature)
		qualifiedName = fmt.Sprintf("%s.<init>(%s)", ownerQN, signature)
		declaration.name = simpleQualifiedName(ownerQN)
	case "field":
		canonical = fmt.Sprintf("%s::property:%s::receiver:%s", ownerCanonical, declaration.name, declaration.receiver)
	default:
		canonical = fmt.Sprintf("%s::%s:%s", ownerCanonical, kind, declaration.name)
	}
	canonical = disambiguateCanonical(canonical, occurrences)
	attributes := declarationAttributes(declaration)
	id := state.facts.addNode(kind, canonical, declaration.name, file.path, qualifiedName, file.path, &declaration.span, attributes)
	state.facts.addEdge(ownerID, id, "contains", file.path, &declaration.span, nil)
	state.facts.addEdge(ownerID, id, "defines", file.path, &declaration.span, nil)
	state.indexRelationshipTarget(id, qualifiedBase, kind, declaration.form)
	runtimeCallable := state.indexRuntimeDeclaration(file, declaration, id, kind, ownerQN, ownerKind, qualifiedBase)
	state.queueAnnotations(file, id, ownerQN, declaration.annotations, declaration.span)
	if declaration.kind == "type" || declaration.kind == "interface" {
		state.queueSupertypes(file, id, ownerQN, declaration.supertypes)
	}

	if kind == "function" || kind == "method" || kind == "constructor" {
		for parameterIndex, parameter := range declaration.parameters {
			parameterCanonical := fmt.Sprintf("%s::parameter::%04d:%s", canonical, parameterIndex, parameter.name)
			parameterQN := fmt.Sprintf("%s::parameter:%s", qualifiedName, parameter.name)
			parameterAttributes := parameterAttributes(parameter, parameterIndex)
			parameterID := state.facts.addNode("parameter", parameterCanonical, parameter.name, file.path, parameterQN, file.path, &parameter.span, parameterAttributes)
			state.facts.addEdge(id, parameterID, "contains", file.path, &parameter.span, nil)
			state.facts.addEdge(id, parameterID, "defines", file.path, &parameter.span, nil)
			if runtimeCallable != nil {
				runtimeCallable.parameters[parameter.name] = append(runtimeCallable.parameters[parameter.name], parameterID)
			}
			state.queueAnnotations(file, parameterID, ownerQN, parameter.annotations, parameter.span)
		}
	}

	childOccurrences := make(map[string]int)
	for _, child := range declaration.children {
		state.emitDeclaration(file, child, id, qualifiedBase, canonical, declaration.kind, childOccurrences)
	}
	if declaration.kind == "type" {
		for _, parameter := range declaration.parameters {
			if !parameter.property {
				continue
			}
			property := constructorParameterProperty(parameter)
			propertyID := state.emitDeclaration(file, property, id, qualifiedBase, canonical, declaration.kind, childOccurrences)
			if record := state.facts.nodes[propertyID]; record != nil {
				attrs, _ := record["attributes"].(map[string]any)
				if attrs == nil {
					attrs = make(map[string]any)
					record["attributes"] = attrs
				}
				attrs["constructor_parameter"] = true
			}
		}
	}
	return id
}

func constructorParameterProperty(parameter parameterDecl) *declaration {
	return &declaration{
		form: "constructor_parameter_property", kind: "field", mutable: parameter.mutable,
		name: parameter.name, span: parameter.span, typeName: parameter.typeName,
	}
}

func declarationAttributes(declaration *declaration) map[string]any {
	attributes := make(map[string]any)
	if declaration.form != "" {
		attributes["declaration_kind"] = declaration.form
	}
	if len(declaration.annotations) != 0 {
		attributes["annotations"] = append([]string(nil), declaration.annotations...)
	}
	if len(declaration.modifiers) != 0 {
		attributes["modifiers"] = append([]string(nil), declaration.modifiers...)
	}
	if declaration.kind == "function" {
		attributes["suspend"] = containsString(declaration.modifiers, "suspend")
		if declaration.receiver != "" {
			attributes["extension_receiver"] = declaration.receiver
			attributes["extension_receiver_nullable"] = nullableType(declaration.receiver)
		}
		if declaration.returnType != "" {
			attributes["return_type"] = declaration.returnType
			attributes["return_nullable"] = nullableType(declaration.returnType)
		}
	}
	if declaration.kind == "constructor" {
		attributes["primary"] = declaration.primary
	}
	if declaration.kind == "field" {
		attributes["mutable"] = declaration.mutable
		if declaration.receiver != "" {
			attributes["extension_receiver"] = declaration.receiver
			attributes["extension_receiver_nullable"] = nullableType(declaration.receiver)
		}
		if declaration.typeName != "" {
			attributes["type"] = declaration.typeName
			attributes["nullable"] = nullableType(declaration.typeName)
		}
	}
	return attributes
}

func parameterAttributes(parameter parameterDecl, index int) map[string]any {
	attributes := map[string]any{
		"has_default": parameter.hasDefault,
		"index":       index,
		"nullable":    nullableType(parameter.typeName),
		"property":    parameter.property,
		"type":        parameter.typeName,
	}
	if parameter.property {
		attributes["mutable"] = parameter.mutable
	}
	if len(parameter.annotations) != 0 {
		attributes["annotations"] = append([]string(nil), parameter.annotations...)
	}
	if len(parameter.modifiers) != 0 {
		attributes["modifiers"] = append([]string(nil), parameter.modifiers...)
	}
	return attributes
}

func disambiguateCanonical(canonical string, occurrences map[string]int) string {
	occurrence := occurrences[canonical]
	occurrences[canonical] = occurrence + 1
	if occurrence == 0 {
		return canonical
	}
	return fmt.Sprintf("%s#%d", canonical, occurrence+1)
}

func parameterSignature(parameters []parameterDecl) string {
	parts := make([]string, len(parameters))
	for index, parameter := range parameters {
		parts[index] = parameter.typeName
	}
	return strings.Join(parts, ",")
}
