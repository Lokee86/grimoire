package main

import (
	"regexp"
	"sort"
	"strings"
)

var (
	callStatementPattern = regexp.MustCompile(`(?i)\bCall\s+([A-Za-z_][A-Za-z0-9_]*(?:\.[A-Za-z_][A-Za-z0-9_]*)*)`)
	functionCallPattern  = regexp.MustCompile(`(?i)([A-Za-z_][A-Za-z0-9_]*(?:\.[A-Za-z_][A-Za-z0-9_]*)*)\s*\(`)
	newPattern           = regexp.MustCompile(`(?i)\bNew\s+([A-Za-z_][A-Za-z0-9_]*)`)
)

var reservedCallWords = map[string]struct{}{
	"and": {}, "call": {}, "case": {}, "class": {}, "const": {}, "declare": {}, "dim": {},
	"do": {}, "else": {}, "elseif": {}, "end": {}, "error": {}, "exit": {},
	"for": {}, "forall": {}, "function": {}, "goto": {}, "if": {}, "is": {},
	"let": {}, "like": {}, "loop": {}, "mod": {}, "next": {}, "not": {}, "on": {},
	"open": {}, "option": {}, "or": {}, "print": {}, "property": {}, "redim": {},
	"resume": {}, "select": {}, "set": {}, "stop": {}, "sub": {}, "then": {},
	"type": {}, "use": {}, "uselsx": {}, "wend": {}, "while": {}, "with": {}, "xor": {},
}

var builtinFunctions = map[string]struct{}{
	"array": {}, "cbool": {}, "cbyte": {}, "ccur": {}, "cdate": {}, "cdat": {},
	"cdbl": {}, "cint": {}, "clng": {}, "createobject": {}, "csng": {}, "cstr": {},
	"chr": {}, "date": {}, "environ": {}, "evaluate": {}, "execute": {}, "format": {},
	"getobject": {}, "getthreadinfo": {}, "implode": {}, "instr": {}, "instrrev": {},
	"isarray": {}, "iselement": {}, "isempty": {}, "isnull": {}, "isobject": {},
	"join": {}, "lbound": {}, "lcase": {}, "left": {}, "len": {}, "lsi_info": {},
	"mid": {}, "msgbox": {}, "now": {}, "replace": {}, "right": {}, "run": {},
	"shell": {}, "split": {}, "strleft": {}, "strright": {}, "trim": {},
	"typename": {}, "ubound": {}, "ucase": {},
}

func (state *analysisState) collectCalls(line logicalLine, callable declaration, class *declaration) {
	seen := make(map[string]struct{})
	add := func(candidate, expression string) {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			return
		}
		key := strings.ToLower(candidate)
		if _, reserved := reservedCallWords[key]; reserved {
			return
		}
		if _, exists := seen[key]; exists {
			return
		}
		seen[key] = struct{}{}
		className := ""
		if class != nil {
			className = class.name
		}
		lineSpan := line.span
		state.calls = append(state.calls, callEvidence{
			candidate: candidate, className: className, expression: expression,
			ownerID: callable.id, ownerPath: callable.ownerPath, span: &lineSpan,
		})
	}
	for _, match := range callStatementPattern.FindAllStringSubmatch(line.text, -1) {
		add(match[1], match[0])
	}
	for _, match := range functionCallPattern.FindAllStringSubmatchIndex(line.text, -1) {
		candidate := line.text[match[2]:match[3]]
		prefix := strings.TrimSpace(line.text[:match[0]])
		if strings.HasSuffix(strings.ToLower(prefix), "new") {
			continue
		}
		add(candidate, line.text[match[0]:match[1]])
	}
	for _, match := range newPattern.FindAllStringSubmatch(line.text, -1) {
		add(match[1]+".New", match[0])
	}
	if candidate := bareCallCandidate(line.text); candidate != "" {
		add(candidate, line.text)
	}
}

func bareCallCandidate(line string) string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.Contains(strings.SplitN(trimmed, " ", 2)[0], ":") {
		return ""
	}
	candidate := identifierPrefix(trimmed)
	if candidate == "" {
		return ""
	}
	lower := strings.ToLower(candidate)
	if _, reserved := reservedCallWords[lower]; reserved {
		return ""
	}
	remainder := strings.TrimSpace(trimmed[len(candidate):])
	if remainder == "" || strings.HasPrefix(remainder, "=") || strings.HasPrefix(remainder, ".") {
		return ""
	}
	return candidate
}

func (state *analysisState) resolveCalls() {
	sort.Slice(state.calls, func(left, right int) bool {
		if state.calls[left].ownerPath != state.calls[right].ownerPath {
			return state.calls[left].ownerPath < state.calls[right].ownerPath
		}
		if state.calls[left].span.StartLine != state.calls[right].span.StartLine {
			return state.calls[left].span.StartLine < state.calls[right].span.StartLine
		}
		return strings.ToLower(state.calls[left].candidate) < strings.ToLower(state.calls[right].candidate)
	})
	for _, evidence := range state.calls {
		candidates, dynamic, indexed := state.callCandidates(evidence)
		if indexed {
			continue
		}
		sortDeclarations(candidates)
		switch {
		case dynamic:
			state.facts.addUnresolved(evidence.ownerID, "calls", evidence.expression, "dynamic-target", evidence.ownerPath, evidence.span, map[string]any{"candidate_name": evidence.candidate})
		case len(candidates) == 1:
			state.facts.addEdge(evidence.ownerID, candidates[0].id, "calls", evidence.ownerPath, evidence.span, nil)
		case len(candidates) > 1:
			state.facts.addUnresolved(evidence.ownerID, "calls", evidence.expression, "ambiguous-target", evidence.ownerPath, evidence.span, map[string]any{"candidate_count": len(candidates), "candidate_name": evidence.candidate})
		default:
			reason := "external-target"
			if _, builtin := builtinFunctions[strings.ToLower(evidence.candidate)]; builtin {
				reason = "builtin-target"
			}
			state.facts.addUnresolved(evidence.ownerID, "calls", evidence.expression, reason, evidence.ownerPath, evidence.span, map[string]any{"candidate_name": evidence.candidate})
		}
	}
}

func (state *analysisState) callCandidates(evidence callEvidence) ([]declaration, bool, bool) {
	parts := strings.Split(evidence.candidate, ".")
	if len(parts) > 1 {
		methodName := strings.ToLower(parts[len(parts)-1])
		qualifier := strings.ToLower(parts[len(parts)-2])
		if qualifier == "me" && evidence.className != "" {
			if state.declaredVariable(evidence, methodName) {
				return nil, false, true
			}
			return state.methodCandidates(strings.ToLower(evidence.className), methodName, make(map[string]struct{})), false, false
		}
		if methods := state.methodsByClass[qualifier]; methods != nil {
			return state.methodCandidates(qualifier, methodName, make(map[string]struct{})), false, false
		}
		if receiverType, declared := state.receiverType(evidence, qualifier); declared {
			if receiverType == "" {
				return nil, true, false
			}
			methods := state.methodCandidates(receiverType, methodName, make(map[string]struct{}))
			if len(methods) > 0 {
				return methods, false, false
			}
			return nil, true, false
		}
		return nil, true, false
	}
	name := strings.ToLower(evidence.candidate)
	if state.declaredVariable(evidence, name) {
		return nil, false, true
	}
	if evidence.className != "" {
		if methods := state.methodCandidates(strings.ToLower(evidence.className), name, make(map[string]struct{})); len(methods) > 0 {
			return methods, false, false
		}
	}
	return append([]declaration(nil), state.globalsByName[name]...), false, false
}

func (state *analysisState) methodCandidates(className, methodName string, visited map[string]struct{}) []declaration {
	if _, seen := visited[className]; seen {
		return nil
	}
	visited[className] = struct{}{}
	if methods := state.methodsByClass[className][methodName]; len(methods) > 0 {
		return append([]declaration(nil), methods...)
	}
	var result []declaration
	for _, base := range state.basesByClass[className] {
		result = append(result, state.methodCandidates(base, methodName, visited)...)
	}
	return result
}
