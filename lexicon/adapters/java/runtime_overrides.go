package main

func (state *analysisState) emitOverrides() {
	for _, method := range state.callables {
		if method.constructor || hasModifier(method.modifiers, "static") || hasModifier(method.modifiers, "private") {
			continue
		}
		for _, parentID := range state.directParents(method.ownerID) {
			for _, target := range state.overrideCandidates(parentID, method.name, method.signature) {
				if !state.overrideSignaturesMatch(method, target, parentID) {
					continue
				}
				state.facts.addEdge(method.id, target.id, "overrides", method.path, nil, map[string]any{
					"declared_signature": method.name + "(" + method.signature + ")",
					"direct_parent":      parentID,
				})
			}
		}
	}
}

func (state *analysisState) overrideSignaturesMatch(method, target callableDeclaration, parentID string) bool {
	if target.constructor || target.ownerID != parentID || target.name != method.name || target.signature != method.signature {
		return false
	}
	if hasModifier(target.modifiers, "static") || hasModifier(target.modifiers, "private") || hasModifier(target.modifiers, "final") {
		return false
	}
	if hasModifier(target.modifiers, "public") || hasModifier(target.modifiers, "protected") || state.interfaceType(target.ownerID) {
		return true
	}
	return method.context.packageName == target.context.packageName
}

func (state *analysisState) interfaceType(id string) bool {
	declarationKind := state.typeKindsByID[id]
	return declarationKind == "interface" || declarationKind == "annotation"
}
