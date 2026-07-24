package main

import "sort"

type indirectCallIndex struct {
	byPointer map[string]map[string]*declaration
}

func buildIndirectCallIndex(model *repositoryModel, index declarationIndex) indirectCallIndex {
	bindings := indirectCallIndex{byPointer: map[string]map[string]*declaration{}}
	for _, declaration := range model.Declarations {
		functionPointer, _ := declaration.Attributes["function_pointer"].(bool)
		if !functionPointer {
			continue
		}
		target, _ := declaration.Attributes["pointer_target"].(string)
		if target == "" {
			continue
		}
		for _, callable := range resolveCallableReference(index, target, declaration.ContainerQualified, declaration.Path) {
			bindings.add(declaration.ID, callable)
		}
	}

	for _, file := range model.Files {
		for _, binding := range file.PointerBindings {
			pointers := resolvePointerDeclarations(index, callObservation{
				SourceID: binding.SourceID, SourceScope: binding.SourceScope, Path: binding.Path,
				Candidate: binding.Candidate, Member: binding.Member,
			})
			if binding.Member && len(pointers) != 1 {
				continue
			}
			for _, callable := range resolveCallableReference(index, binding.Target, binding.SourceScope, binding.Path) {
				for _, pointer := range pointers {
					bindings.add(pointer.ID, callable)
				}
			}
		}
		for _, observation := range file.Calls {
			if len(observation.Arguments) == 0 {
				continue
			}
			callees := directCallableTargets(index, observation)
			if len(callees) != 1 {
				continue
			}
			parameters := index.byCallableParameters[callees[0].ID]
			for argumentIndex, argument := range observation.Arguments {
				if argument == "" || argumentIndex >= len(parameters) {
					continue
				}
				parameter := parameters[argumentIndex]
				functionPointer, _ := parameter.Attributes["function_pointer"].(bool)
				if !functionPointer {
					continue
				}
				for _, callable := range resolveCallableReference(index, argument, observation.SourceScope, observation.Path) {
					bindings.add(parameter.ID, callable)
				}
			}
		}
	}
	return bindings
}

func (index indirectCallIndex) add(pointerID string, target *declaration) {
	if index.byPointer[pointerID] == nil {
		index.byPointer[pointerID] = map[string]*declaration{}
	}
	index.byPointer[pointerID][target.ID] = target
}

func (index indirectCallIndex) targets(pointers []*declaration) []*declaration {
	unique := map[string]*declaration{}
	for _, pointer := range pointers {
		for id, target := range index.byPointer[pointer.ID] {
			unique[id] = target
		}
	}
	result := make([]*declaration, 0, len(unique))
	for _, target := range unique {
		result = append(result, target)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func directCallableTargets(index declarationIndex, observation callObservation) []*declaration {
	if macros := resolveMacroDeclarations(index, observation.Candidate, observation.Path); len(macros) > 0 {
		return macroCallableTargets(index, macros, observation, map[string]struct{}{})
	}
	return resolveDeclarations(index, observation.Candidate, observation.SourceScope, observation.Path, func(declaration *declaration) bool {
		return declaration.Callable
	})
}

func resolveCallableReference(index declarationIndex, candidate, scope, path string) []*declaration {
	if macros := resolveMacroDeclarations(index, candidate, path); len(macros) > 0 {
		observation := callObservation{Candidate: candidate, SourceScope: scope, Path: path}
		return macroCallableTargets(index, macros, observation, map[string]struct{}{})
	}
	return resolveDeclarations(index, candidate, scope, path, func(declaration *declaration) bool {
		return declaration.Callable
	})
}
