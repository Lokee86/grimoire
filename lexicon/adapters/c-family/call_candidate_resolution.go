package main

import "strings"

func resolveCallCandidates(index declarationIndex, observation callObservation) []*declaration {
	accept := func(declaration *declaration) bool {
		return declaration.Callable
	}
	if qualifier := explicitCallQualifier(observation.Candidate); qualifier != "" {
		candidates := resolveDeclarations(index, observation.Candidate, observation.SourceScope, observation.Path, accept)
		if types := directQualifiedTypes(index, qualifier, observation.Path); len(types) > 0 {
			owned := callableDeclarationsOwnedByTypes(index, types, lastQualifiedPart(observation.Candidate), observation.Path, accept)
			if len(owned) > 0 {
				return owned
			}
			return candidates
		}
		free := filterDeclarations(candidates, func(declaration *declaration) bool {
			return !classOwnedCallable(index, declaration)
		})
		if len(free) > 0 {
			return free
		}
		return candidates
	}
	if receiverTypeID := resolveDirectReceiverTypeID(index, observation); receiverTypeID != "" {
		owned := selectDeclarations(index, index.byContainerName[receiverTypeID+"\x00"+lastQualifiedPart(observation.Candidate)], observation.Path, accept)
		if len(owned) > 0 {
			return owned
		}
	}
	if observation.ReceiverTypeID != "" {
		candidates := resolveDeclarations(index, observation.Candidate, observation.SourceScope, observation.Path, accept)
		owned := selectDeclarations(index, index.byContainerName[observation.ReceiverTypeID+"\x00"+lastQualifiedPart(observation.Candidate)], observation.Path, accept)
		if len(owned) > 0 {
			return owned
		}
		return candidates
	}
	return resolveDeclarations(index, observation.Candidate, observation.SourceScope, observation.Path, accept)
}

func explicitCallQualifier(candidate string) string {
	candidate = normalizeQualified(candidate)
	separator := strings.LastIndex(candidate, "::")
	if separator <= 0 || separator+2 >= len(candidate) {
		return ""
	}
	return candidate[:separator]
}

func directQualifiedTypes(index declarationIndex, qualifier, path string) []*declaration {
	return selectDeclarations(index, index.byQualified[normalizeQualified(qualifier)], path, func(declaration *declaration) bool {
		return declaration.Kind == "type"
	})
}

func callableDeclarationsOwnedByTypes(index declarationIndex, types []*declaration, name, path string, accept func(*declaration) bool) []*declaration {
	ownedTypes := map[string]struct{}{}
	var candidates []*declaration
	for _, typ := range types {
		ownedTypes[typ.ID] = struct{}{}
		candidates = append(candidates, selectDeclarations(index, index.byContainerName[typ.ID+"\x00"+name], path, accept)...)
	}
	return filterDeclarations(candidates, func(declaration *declaration) bool {
		_, owned := ownedTypes[declaration.ContainerID]
		return owned
	})
}

func classOwnedCallable(index declarationIndex, declaration *declaration) bool {
	if declaration.Kind == "method" || declaration.Kind == "constructor" || declaration.ParentTypeID != "" {
		return true
	}
	qualifier := explicitCallQualifier(declaration.QualifiedName)
	return qualifier != "" && len(directQualifiedTypes(index, qualifier, declaration.Path)) > 0
}

func resolveDirectReceiverTypeID(index declarationIndex, observation callObservation) string {
	if !observation.Member || observation.Receiver == "" || observation.Receiver == "this" || observation.Receiver == "self" {
		return ""
	}
	source := index.byID[observation.SourceID]
	if source == nil || source.FileLanguage != "cpp" {
		return ""
	}

	acceptReceiver := func(declaration *declaration) bool {
		return declaration.Kind == "parameter" || declaration.Kind == "variable" || declaration.Kind == "field"
	}
	candidates := filterDeclarations(index.byContainerName[observation.SourceID+"\x00"+observation.Receiver], acceptReceiver)
	if source.ParentTypeID != "" {
		candidates = append(candidates, visibleDeclarations(index, index.byContainerName[source.ParentTypeID+"\x00"+observation.Receiver], observation.Path, func(declaration *declaration) bool {
			return declaration.Kind == "field"
		})...)
		candidates = filterDeclarations(candidates, acceptReceiver)
	}
	if len(candidates) != 1 {
		return ""
	}
	typeText, _ := candidates[0].Attributes["type"].(string)
	typeName := directReceiverTypeName(typeText)
	if typeName == "" {
		return ""
	}
	types := resolveDeclarations(index, typeName, source.QualifiedName, observation.Path, func(declaration *declaration) bool {
		if declaration.Kind != "type" {
			return false
		}
		alias, _ := declaration.Attributes["alias"].(bool)
		return !alias
	})
	if len(types) != 1 {
		return ""
	}
	return types[0].ID
}

func directReceiverTypeName(value string) string {
	value = normalizeSpace(value)
	if value == "" || strings.ContainsAny(value, "<>()[],") {
		return ""
	}
	value = strings.ReplaceAll(value, "*", "")
	value = strings.ReplaceAll(value, "&", "")
	parts := strings.Fields(value)
	filtered := parts[:0]
	for _, part := range parts {
		if part == "const" || part == "volatile" || part == "struct" || part == "class" || part == "enum" {
			continue
		}
		filtered = append(filtered, part)
	}
	if len(filtered) != 1 {
		return ""
	}
	return normalizeQualified(filtered[0])
}

func pruneCallableCandidates(candidates []*declaration, argumentCount int) []*declaration {
	if len(candidates) < 2 {
		return candidates
	}
	compatible := make([]*declaration, 0, len(candidates))
	for _, candidate := range candidates {
		shape := candidate.CallableShape
		if shape == nil {
			compatible = append(compatible, candidate)
			continue
		}
		if argumentCount < shape.Minimum || !shape.Variadic && argumentCount > shape.Maximum {
			continue
		}
		compatible = append(compatible, candidate)
	}
	if len(compatible) == 0 {
		return candidates
	}
	return compatible
}
