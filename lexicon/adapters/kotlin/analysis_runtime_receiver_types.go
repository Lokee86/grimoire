package main

import "strings"

type runtimeDeclaredType struct {
	file         *parsedKotlinFile
	lexicalOwner string
	spelling     string
}

func (state *analysis) runtimeReceiverEvidence(
	callable *runtimeCallable,
	invocation runtimeInvocation,
) (runtimeDeclaredType, bool) {
	name := invocation.qualifier
	if name == "" || strings.Contains(name, ".") {
		return runtimeDeclaredType{}, false
	}
	if spelling, found := runtimeLocalDeclaredType(callable, name, invocation.calleeStart); found {
		return runtimeDeclaredType{file: callable.file, lexicalOwner: callable.ownerQN, spelling: spelling}, true
	}
	var parameterTypes []string
	for _, parameter := range callable.declaration.parameters {
		if parameter.name == name {
			parameterTypes = append(parameterTypes, parameter.typeName)
		}
	}
	if len(parameterTypes) == 1 {
		return runtimeDeclaredType{file: callable.file, lexicalOwner: callable.ownerQN, spelling: parameterTypes[0]}, true
	}
	if len(parameterTypes) > 1 {
		return runtimeDeclaredType{}, true
	}
	owners := []string{callable.ownerQN}
	if callable.ownerKind == "type" || callable.ownerKind == "interface" {
		owners = append(owners, runtimePackage(callable.file))
	}
	for _, owner := range owners {
		if evidence, found := state.runtimePropertyDeclaredType(owner, name); found {
			return evidence, true
		}
	}
	return runtimeDeclaredType{}, false
}

func (state *analysis) runtimePropertyDeclaredType(owner, name string) (runtimeDeclaredType, bool) {
	var evidence []runtimeDeclaredType
	for _, property := range state.runtime.propertiesByKey[runtimeMemberKey(owner, name)] {
		declaration := property.declaration
		if declaration.receiver == "" && !declaration.delegated {
			evidence = append(evidence, runtimeDeclaredType{
				file: property.file, lexicalOwner: property.ownerQN, spelling: declaration.typeName,
			})
		}
	}
	if len(evidence) == 1 {
		return evidence[0], true
	}
	return runtimeDeclaredType{}, len(evidence) > 1
}

func (state *analysis) resolveRuntimeDeclaredType(evidence runtimeDeclaredType) (*runtimeType, bool, string) {
	name, nullable, ok := simpleRuntimeTypeName(evidence.spelling)
	if !ok {
		return nil, false, "unsupported-form"
	}
	targets, reason := state.resolveRuntimeTypes(evidence.file, evidence.lexicalOwner, name)
	if len(targets) != 1 {
		return nil, false, reason
	}
	return targets[0], nullable, ""
}

func simpleRuntimeTypeName(spelling string) (string, bool, bool) {
	name := normalizeRuntimeType(spelling)
	nullable := strings.HasSuffix(name, "?")
	name = strings.TrimSuffix(name, "?")
	if name == "" || strings.HasPrefix(name, ".") || strings.HasSuffix(name, ".") || strings.Contains(name, "..") {
		return "", false, false
	}
	for _, current := range name {
		if current != '.' && current != '_' && current != '`' && !isRuntimeTypeRune(current) {
			return "", false, false
		}
	}
	parts := strings.Split(name, ".")
	for index, part := range parts {
		parts[index] = strings.Trim(part, "`")
		if parts[index] == "" {
			return "", false, false
		}
	}
	return strings.Join(parts, "."), nullable, true
}

func isRuntimeTypeRune(value rune) bool {
	return value >= '0' && value <= '9' || value >= 'A' && value <= 'Z' || value >= 'a' && value <= 'z' || value > 127
}
