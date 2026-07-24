package main

func resolvePointerDeclarations(index declarationIndex, observation callObservation) []*declaration {
	accept := func(declaration *declaration) bool {
		functionPointer, _ := declaration.Attributes["function_pointer"].(bool)
		return functionPointer
	}
	if !observation.Member {
		if local := filterDeclarations(index.byContainerName[observation.SourceID+"\x00"+observation.Candidate], accept); len(local) > 0 {
			return local
		}
		return filterDeclarations(index.byPathName[observation.Path+"\x00"+observation.Candidate], func(declaration *declaration) bool {
			return accept(declaration) && declaration.Kind != "field"
		})
	}
	return filterDeclarations(index.byName[observation.Candidate], func(declaration *declaration) bool {
		return accept(declaration) && declaration.Kind == "field"
	})
}

func hasCallableDeclaration(index declarationIndex, candidate string) bool {
	for _, declaration := range index.byName[lastQualifiedPart(candidate)] {
		if declaration.Callable {
			return true
		}
	}
	return false
}
