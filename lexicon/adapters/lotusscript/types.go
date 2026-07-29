package main

import "strings"

func (state *analysisState) recordCallableVariable(callableID, name, dataType string) {
	recordVariable(state.typesByCallable, callableID, name, dataType)
}

func (state *analysisState) recordClassField(className, name, dataType string) {
	recordVariable(state.fieldsByClass, strings.ToLower(className), name, dataType)
}

func (state *analysisState) recordModuleVariable(path, name, dataType string) {
	recordVariable(state.moduleVariables, path, name, dataType)
}

func recordVariable(target map[string]map[string]string, owner, name, dataType string) {
	if target[owner] == nil {
		target[owner] = make(map[string]string)
	}
	target[owner][strings.ToLower(name)] = normalizeTypeName(dataType)
}

func (state *analysisState) declaredVariable(evidence callEvidence, name string) bool {
	_, declared := state.receiverType(evidence, strings.ToLower(name))
	return declared
}

func (state *analysisState) receiverType(evidence callEvidence, name string) (string, bool) {
	name = strings.ToLower(name)
	if dataType, ok := state.typesByCallable[evidence.ownerID][name]; ok {
		return dataType, true
	}
	if evidence.className != "" {
		if dataType, ok := state.fieldsByClass[strings.ToLower(evidence.className)][name]; ok {
			return dataType, true
		}
	}
	dataType, ok := state.moduleVariables[evidence.ownerPath][name]
	return dataType, ok
}

func normalizeTypeName(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 4 && strings.EqualFold(value[:4], "new ") {
		value = strings.TrimSpace(value[4:])
	}
	value = strings.TrimSpace(strings.TrimSuffix(value, "()"))
	if index := strings.LastIndex(value, "."); index >= 0 {
		value = value[index+1:]
	}
	return strings.ToLower(value)
}
