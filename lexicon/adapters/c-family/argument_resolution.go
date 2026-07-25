package main

import "strings"

func addResolvedCallEdges(facts *factSet, index declarationIndex, observation callObservation, target *declaration, attributes map[string]any) {
	addCallEdge(facts, observation, target, "calls", attributes)
	parameters := index.byCallableParameters[target.ID]
	for argumentIndex, expression := range observation.ArgumentExpressions {
		if !isSimpleIdentifier(expression) {
			continue
		}
		parameter := parameterAtIndex(parameters, argumentIndex)
		if parameter == nil {
			continue
		}
		argument := resolveDirectArgument(index, observation, expression)
		if argument == nil {
			continue
		}
		facts.addEdge(observation.Path, map[string]any{
			"attributes": map[string]any{
				"argument_index": argumentIndex,
				"expression":     expression,
				"via_call":       target.ID,
			},
			"owner": observation.Path, "record": "edge", "relation": "passes-to",
			"source": argument.ID, "span": observation.Span.record(), "target": parameter.ID,
		})
	}
}

func parameterAtIndex(parameters []*declaration, index int) *declaration {
	for _, parameter := range parameters {
		value, _ := parameter.Attributes["index"].(int)
		if value == index {
			return parameter
		}
	}
	return nil
}

func resolveDirectArgument(index declarationIndex, observation callObservation, name string) *declaration {
	local := filterDeclarations(index.byContainerName[observation.SourceID+"\x00"+name], func(value *declaration) bool {
		return value.Kind == "parameter" || value.Kind == "variable" || value.Kind == "constant"
	})
	if len(local) == 1 {
		return local[0]
	}
	if len(local) > 1 {
		return nil
	}

	source := index.byID[observation.SourceID]
	if source != nil && source.ParentTypeID != "" {
		fields := visibleDeclarations(index, index.byContainerName[source.ParentTypeID+"\x00"+name], observation.Path, func(value *declaration) bool {
			return value.Kind == "field"
		})
		if len(fields) == 1 {
			return fields[0]
		}
		if len(fields) > 1 {
			return nil
		}
	}

	constants := resolveDeclarations(index, name, observation.SourceScope, observation.Path, func(value *declaration) bool {
		return value.Kind == "constant"
	})
	if len(constants) == 1 {
		return constants[0]
	}
	return nil
}

func isSimpleIdentifier(expression string) bool {
	expression = strings.TrimSpace(expression)
	if expression == "" {
		return false
	}
	for index, character := range expression {
		if index == 0 {
			if character != '_' && (character < 'a' || character > 'z') && (character < 'A' || character > 'Z') {
				return false
			}
			continue
		}
		if character != '_' && (character < 'a' || character > 'z') && (character < 'A' || character > 'Z') && (character < '0' || character > '9') {
			return false
		}
	}
	return true
}
