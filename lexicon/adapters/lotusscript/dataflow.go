package main

import (
	"regexp"
	"sort"
	"strings"
)

var identifierChainPattern = regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_]*(?:\.[A-Za-z_][A-Za-z0-9_]*)*`)

func (state *analysisState) collectAccess(line logicalLine, callable declaration) {
	lineSpan := line.span
	state.accesses = append(state.accesses, accessEvidence{
		classID: callable.classID, ownerID: callable.id, ownerPath: callable.ownerPath,
		span: &lineSpan, text: line.text,
	})
}

func (state *analysisState) resolveAccesses() {
	sort.Slice(state.accesses, func(left, right int) bool {
		if state.accesses[left].ownerPath != state.accesses[right].ownerPath {
			return state.accesses[left].ownerPath < state.accesses[right].ownerPath
		}
		if state.accesses[left].span.StartLine != state.accesses[right].span.StartLine {
			return state.accesses[left].span.StartLine < state.accesses[right].span.StartLine
		}
		return state.accesses[left].text < state.accesses[right].text
	})
	for _, evidence := range state.accesses {
		state.resolveAccess(evidence)
	}
}

func (state *analysisState) resolveAccess(evidence accessEvidence) {
	masked := maskLotusLiterals(evidence.text)
	assignment, assigned := parseAssignmentAccess(masked)
	if !assigned {
		state.addReadAccesses(evidence, masked)
		return
	}

	state.addWriteAccess(evidence, assignment.target)
	state.addReadAccesses(evidence, assignment.left)
	state.addReadAccesses(evidence, assignment.right)
}

func (state *analysisState) addWriteAccess(evidence accessEvidence, target string) {
	parts := splitIdentifierChain(target)
	if len(parts) == 0 {
		return
	}
	if len(parts) == 1 {
		if symbol, ok := state.resolveVariableSymbol(evidence.ownerID, evidence.classID, evidence.ownerPath, parts[0]); ok {
			state.facts.addEdge(evidence.ownerID, symbol.id, "writes", evidence.ownerPath, evidence.span, nil)
		}
		return
	}

	if !strings.EqualFold(parts[0], "Me") {
		if symbol, ok := state.resolveVariableSymbol(evidence.ownerID, evidence.classID, evidence.ownerPath, parts[0]); ok {
			state.facts.addEdge(evidence.ownerID, symbol.id, "reads", evidence.ownerPath, evidence.span, nil)
		}
	}
	if symbol, ok := state.resolveMemberSymbol(evidence, parts); ok {
		state.facts.addEdge(evidence.ownerID, symbol.id, "writes", evidence.ownerPath, evidence.span, nil)
	}
}

func (state *analysisState) addReadAccesses(evidence accessEvidence, text string) {
	for _, match := range identifierChainPattern.FindAllString(text, -1) {
		parts := splitIdentifierChain(match)
		if len(parts) == 0 {
			continue
		}
		if len(parts) == 1 {
			if symbol, ok := state.resolveVariableSymbol(evidence.ownerID, evidence.classID, evidence.ownerPath, parts[0]); ok {
				state.facts.addEdge(evidence.ownerID, symbol.id, "reads", evidence.ownerPath, evidence.span, nil)
			}
			continue
		}
		if !strings.EqualFold(parts[0], "Me") {
			if symbol, ok := state.resolveVariableSymbol(evidence.ownerID, evidence.classID, evidence.ownerPath, parts[0]); ok {
				state.facts.addEdge(evidence.ownerID, symbol.id, "reads", evidence.ownerPath, evidence.span, nil)
			}
		}
		if symbol, ok := state.resolveMemberSymbol(evidence, parts); ok {
			state.facts.addEdge(evidence.ownerID, symbol.id, "reads", evidence.ownerPath, evidence.span, nil)
		}
	}
}

func (state *analysisState) resolveMemberSymbol(evidence accessEvidence, parts []string) (variableSymbol, bool) {
	if len(parts) < 2 {
		return variableSymbol{}, false
	}

	var classes []declaration
	if strings.EqualFold(parts[0], "Me") {
		if class := state.classByID(evidence.classID); class != nil {
			classes = []declaration{*class}
		}
	} else {
		receiver, ok := state.resolveVariableSymbol(evidence.ownerID, evidence.classID, evidence.ownerPath, parts[0])
		if !ok || receiver.dataType == "" {
			return variableSymbol{}, false
		}
		classes = state.visibleDeclarations(evidence.ownerPath, state.classesByName[receiver.dataType])
	}
	if len(classes) != 1 {
		return variableSymbol{}, false
	}

	for index := 1; index < len(parts); index++ {
		class := classes[0]
		candidates := state.fieldCandidates(class.id, parts[index], make(map[string]struct{}))
		allowPrivate := strings.EqualFold(parts[0], "Me") || state.classIsCurrentOrBase(evidence.classID, class.id, make(map[string]struct{}))
		if !allowPrivate {
			filtered := candidates[:0]
			for _, candidate := range candidates {
				if candidate.public {
					filtered = append(filtered, candidate)
				}
			}
			candidates = filtered
		}
		if len(candidates) != 1 {
			return variableSymbol{}, false
		}
		if index == len(parts)-1 {
			return candidates[0], true
		}
		if candidates[0].dataType == "" {
			return variableSymbol{}, false
		}
		classes = state.visibleDeclarations(evidence.ownerPath, state.classesByName[candidates[0].dataType])
		if len(classes) != 1 {
			return variableSymbol{}, false
		}
	}
	return variableSymbol{}, false
}

func (state *analysisState) classByID(id string) *declaration {
	if id == "" {
		return nil
	}
	for _, declarations := range state.classesByName {
		for index := range declarations {
			if declarations[index].id == id {
				result := declarations[index]
				return &result
			}
		}
	}
	return nil
}

func (state *analysisState) classIsCurrentOrBase(currentID, targetID string, visited map[string]struct{}) bool {
	if currentID == "" || targetID == "" {
		return false
	}
	if currentID == targetID {
		return true
	}
	if _, seen := visited[currentID]; seen {
		return false
	}
	visited[currentID] = struct{}{}
	for _, baseID := range state.basesByClass[currentID] {
		if state.classIsCurrentOrBase(baseID, targetID, visited) {
			return true
		}
	}
	return false
}
