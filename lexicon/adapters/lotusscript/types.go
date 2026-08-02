package main

import (
	"sort"
	"strings"
)

func (state *analysisState) recordCallableVariable(callableID, name, dataType string) {
	recordVariable(state.typesByCallable, callableID, name, dataType)
}

func (state *analysisState) recordCallableSymbol(callableID, name, dataType, id, ownerPath string) {
	recordSymbol(state.variableSymbols, callableID, name, variableSymbol{dataType: normalizeTypeName(dataType), id: id, ownerPath: ownerPath})
}

func (state *analysisState) recordClassField(classID, name, dataType string) {
	recordVariable(state.fieldsByClass, classID, name, dataType)
}

func (state *analysisState) recordClassFieldSymbol(classID, name, dataType, id, ownerPath string, public bool) {
	recordSymbol(state.fieldSymbols, classID, name, variableSymbol{dataType: normalizeTypeName(dataType), id: id, ownerPath: ownerPath, public: public})
}

func (state *analysisState) recordModuleVariable(path, name, dataType string) {
	recordVariable(state.moduleVariables, path, name, dataType)
}

func (state *analysisState) recordModuleSymbol(path, name, dataType, id string, public bool) {
	recordSymbol(state.moduleSymbols, path, name, variableSymbol{dataType: normalizeTypeName(dataType), id: id, ownerPath: path, public: public})
}

func recordVariable(target map[string]map[string]string, owner, name, dataType string) {
	if target[owner] == nil {
		target[owner] = make(map[string]string)
	}
	target[owner][strings.ToLower(name)] = normalizeTypeName(dataType)
}

func recordSymbol(target map[string]map[string]variableSymbol, owner, name string, symbol variableSymbol) {
	if target[owner] == nil {
		target[owner] = make(map[string]variableSymbol)
	}
	target[owner][strings.ToLower(name)] = symbol
}

func (state *analysisState) declaredVariable(evidence callEvidence, name string) bool {
	_, declared := state.receiverType(evidence, strings.ToLower(name))
	return declared
}

func (state *analysisState) receiverType(evidence callEvidence, name string) (string, bool) {
	symbol, ok := state.resolveVariableSymbol(evidence.ownerID, evidence.classID, evidence.ownerPath, name)
	if !ok {
		return "", false
	}
	return symbol.dataType, true
}

func (state *analysisState) resolveVariableSymbol(ownerID, classID, ownerPath, name string) (variableSymbol, bool) {
	name = strings.ToLower(name)
	if symbol, ok := state.variableSymbols[ownerID][name]; ok {
		return symbol, true
	}
	if classID != "" {
		if symbols := state.fieldCandidates(classID, name, make(map[string]struct{})); len(symbols) == 1 {
			return symbols[0], true
		}
	}
	if symbol, ok := state.moduleSymbols[ownerPath][name]; ok {
		return symbol, true
	}

	visible := state.visiblePaths(ownerPath)
	paths := make([]string, 0, len(visible))
	for path := range visible {
		if path != ownerPath {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	var candidates []variableSymbol
	for _, path := range paths {
		if symbol, ok := state.moduleSymbols[path][name]; ok && symbol.public {
			candidates = append(candidates, symbol)
		}
	}
	if len(candidates) == 1 {
		return candidates[0], true
	}
	return variableSymbol{}, false
}

func (state *analysisState) fieldCandidates(classID, name string, visited map[string]struct{}) []variableSymbol {
	if classID == "" {
		return nil
	}
	if _, seen := visited[classID]; seen {
		return nil
	}
	visited[classID] = struct{}{}
	if symbol, ok := state.fieldSymbols[classID][strings.ToLower(name)]; ok {
		return []variableSymbol{symbol}
	}
	var result []variableSymbol
	for _, baseID := range state.basesByClass[classID] {
		result = append(result, state.fieldCandidates(baseID, name, visited)...)
	}
	return result
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
