package main

func (state *analysis) emitOverrides() {
	for _, callable := range state.runtime.callables {
		if callable.kind != "method" {
			continue
		}
		for _, supertype := range state.directRuntimeSupertypes(callable.ownerQN) {
			candidates := state.runtime.callablesByKey[runtimeCallableKey(
				supertype.qualified, callable.declaration.name, len(callable.declaration.parameters),
			)]
			var matches []*runtimeCallable
			for _, candidate := range candidates {
				if candidate.kind == "method" &&
					candidate.signature == callable.signature &&
					normalizedReceiver(candidate.declaration) == normalizedReceiver(callable.declaration) {
					matches = append(matches, candidate)
				}
			}
			matches = uniqueRuntimeCallables(matches)
			if len(matches) == 1 {
				state.facts.addEdge(
					callable.id, matches[0].id, "overrides", callable.file.path,
					&callable.declaration.span, nil,
				)
			}
		}
	}
}

func (state *analysis) directRuntimeSupertypes(ownerQN string) []*runtimeType {
	owners := uniqueRuntimeTypes(state.runtime.typesByQN[ownerQN])
	if len(owners) != 1 {
		return nil
	}
	owner := owners[0]
	var targets []*runtimeType
	for _, supertype := range owner.declaration.supertypes {
		resolved, _ := state.resolveRuntimeTypes(owner.file, owner.ownerQN, supertype.targetName)
		if len(resolved) == 1 {
			targets = append(targets, resolved[0])
		}
	}
	return uniqueRuntimeTypes(targets)
}
