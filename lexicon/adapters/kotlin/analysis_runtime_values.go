package main

import "strings"

func runtimeShadowedNames(callable *runtimeCallable) map[string]struct{} {
	result := make(map[string]struct{})
	tokens := callable.file.tokens
	bounds := callable.declaration.body
	for index := bounds.start; index < bounds.end && index < len(tokens); index++ {
		if tokens[index].text == "val" || tokens[index].text == "var" {
			next := nextRuntimeToken(tokens, index+1, bounds.end)
			if next >= bounds.end {
				continue
			}
			if tokens[next].kind == tokenIdentifier {
				result[identifierText(tokens[next])] = struct{}{}
				continue
			}
			if tokens[next].text == "(" {
				if close := matchingRuntimeDelimiter(tokens, next, bounds.end, "(", ")"); close >= 0 {
					for item := next + 1; item < close; item++ {
						if tokens[item].kind == tokenIdentifier {
							result[identifierText(tokens[item])] = struct{}{}
						}
					}
				}
			}
		}
		if tokens[index].text == "fun" || tokens[index].text == "class" || tokens[index].text == "object" {
			next := nextRuntimeToken(tokens, index+1, bounds.end)
			if next < bounds.end && tokens[next].kind == tokenIdentifier {
				result[identifierText(tokens[next])] = struct{}{}
			}
		}
		if tokens[index].kind == tokenIdentifier {
			minus := nextRuntimeToken(tokens, index+1, bounds.end)
			arrow := nextRuntimeToken(tokens, minus+1, bounds.end)
			if minus < bounds.end && arrow < bounds.end && tokens[minus].text == "-" && tokens[arrow].text == ">" {
				result[identifierText(tokens[index])] = struct{}{}
			}
		}
	}
	return result
}

func (state *analysis) hasRuntimeValue(callable *runtimeCallable, name string) bool {
	if len(callable.parameters[name]) != 0 {
		return true
	}
	owners := []string{callable.ownerQN}
	if callable.ownerKind == "type" || callable.ownerKind == "interface" {
		owners = append(owners, runtimePackage(callable.file))
	}
	for _, owner := range owners {
		if len(state.runtime.propertiesByKey[runtimeMemberKey(owner, name)]) != 0 {
			return true
		}
	}
	return false
}

func (state *analysis) resolveRuntimeValue(
	callable *runtimeCallable,
	qualifier, name string,
	shadowed map[string]struct{},
) string {
	if qualifier == "" {
		if _, exists := shadowed[name]; exists {
			return ""
		}
		if targets := callable.parameters[name]; len(targets) == 1 {
			return targets[0]
		} else if len(targets) > 1 {
			return ""
		}
		if callable.ownerKind == "type" || callable.ownerKind == "interface" {
			if target := state.uniqueRuntimeProperty(callable.ownerQN, name); target != "" {
				return target
			}
		}
		return state.uniqueRuntimeProperty(runtimePackage(callable.file), name)
	}
	if qualifier == "this" {
		return state.uniqueRuntimeProperty(callable.ownerQN, name)
	}
	first := qualifier
	if dot := strings.IndexByte(first, '.'); dot >= 0 {
		first = first[:dot]
	}
	if _, exists := shadowed[first]; exists || len(callable.parameters[first]) != 0 || state.hasRuntimeValue(callable, first) {
		return ""
	}
	owners, _ := state.resolveRuntimeTypes(callable.file, callable.ownerQN, qualifier)
	if len(owners) != 1 {
		return ""
	}
	owner := owners[0]
	if owner.form == "object" || owner.form == "data_object" || owner.form == "companion_object" {
		return state.uniqueRuntimeProperty(owner.qualified, name)
	}
	var targets []string
	for _, companion := range state.runtime.directCompanionsByOwner[owner.qualified] {
		if target := state.uniqueRuntimeProperty(companion.qualified, name); target != "" {
			targets = append(targets, target)
		}
	}
	if len(uniqueStrings(targets)) == 1 {
		return targets[0]
	}
	return ""
}

func (state *analysis) uniqueRuntimeProperty(owner, name string) string {
	var targets []string
	for _, property := range state.runtime.propertiesByKey[runtimeMemberKey(owner, name)] {
		if property.declaration.receiver == "" && !property.declaration.delegated {
			targets = append(targets, property.id)
		}
	}
	targets = uniqueStrings(targets)
	if len(targets) == 1 {
		return targets[0]
	}
	return ""
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	var result []string
	for _, value := range values {
		if _, exists := seen[value]; !exists {
			seen[value] = struct{}{}
			result = append(result, value)
		}
	}
	return result
}
