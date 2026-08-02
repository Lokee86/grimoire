package main

import "strings"

type fieldAccess struct {
	end      int
	name     string
	receiver string
	start    int
	target   int
}

type writeAccess struct {
	access   fieldAccess
	compound bool
}

var javaKeywords = map[string]bool{
	"abstract": true, "assert": true, "boolean": true, "break": true, "byte": true,
	"case": true, "catch": true, "char": true, "class": true, "const": true,
	"continue": true, "default": true, "do": true, "double": true, "else": true,
	"enum": true, "extends": true, "false": true, "final": true, "finally": true,
	"float": true, "for": true, "goto": true, "if": true, "implements": true,
	"import": true, "instanceof": true, "int": true, "interface": true, "long": true,
	"native": true, "new": true, "null": true, "package": true, "private": true,
	"protected": true, "public": true, "record": true, "return": true, "sealed": true,
	"short": true, "static": true, "strictfp": true, "super": true, "switch": true,
	"synchronized": true, "this": true, "throw": true, "throws": true, "transient": true,
	"true": true, "try": true, "var": true, "void": true, "volatile": true, "while": true,
}

var primitiveTypes = map[string]bool{
	"boolean": true, "byte": true, "char": true, "double": true, "float": true,
	"int": true, "long": true, "short": true, "var": true,
}

func (state *analysisState) emitDataflow(callable *callableDeclaration) {
	ignored := ignoredRuntimeTokens(callable)
	declarationTokens, localNames := localDeclarations(callable, ignored)
	writes := findWrites(callable, ignored)
	for _, write := range writes {
		state.emitAccess(callable, write.access, "writes", localNames)
		if write.compound {
			state.emitAccess(callable, write.access, "reads", localNames)
		}
	}
	for index := callable.bodyStart; index < callable.bodyEnd; index++ {
		if ignored[index] || !identifierToken(callable.tokens[index].text) || javaKeywords[callable.tokens[index].text] || declarationTokens[index] {
			continue
		}
		if write, exists := writes[index]; exists {
			_ = write
			continue
		}
		if index+1 < callable.bodyEnd && (callable.tokens[index+1].text == "." || callable.tokens[index+1].text == "(") {
			continue
		}
		access, valid := accessEndingAt(callable.tokens, callable.bodyStart, index)
		if !valid {
			continue
		}
		if access.receiver == "" {
			if callable.parameterIDs[access.name] == "" && len(state.fields[callable.ownerID][access.name]) == 0 && !localNames[access.name] {
				continue
			}
		}
		state.emitAccess(callable, access, "reads", localNames)
	}
}

func findWrites(callable *callableDeclaration, ignored map[int]bool) map[int]writeAccess {
	result := make(map[int]writeAccess)
	tokens := callable.tokens
	for index := callable.bodyStart; index < callable.bodyEnd; index++ {
		if ignored[index] {
			continue
		}
		if tokens[index].text == "=" {
			targetEnd, compound, valid := assignmentTargetEnd(tokens, callable.bodyStart, index, callable.bodyEnd)
			if valid {
				if access, ok := accessEndingAt(tokens, callable.bodyStart, targetEnd); ok {
					result[access.target] = writeAccess{access: access, compound: compound}
				}
			}
		}
		if index+1 < callable.bodyEnd && (tokens[index].text == "+" || tokens[index].text == "-") && tokens[index+1].text == tokens[index].text {
			target := index - 1
			if target < callable.bodyStart || !identifierToken(tokens[target].text) {
				target = index + 2
			}
			if target < callable.bodyEnd {
				if access, ok := accessEndingAt(tokens, callable.bodyStart, target); ok {
					result[access.target] = writeAccess{access: access, compound: true}
				}
			}
			index++
		}
	}
	return result
}

func assignmentTargetEnd(tokens []token, start, equals, end int) (int, bool, bool) {
	if equals+1 < end && (tokens[equals+1].text == "=" || tokens[equals+1].text == ">") {
		return 0, false, false
	}
	if equals == start {
		return 0, false, false
	}
	previous := tokens[equals-1].text
	if previous == "=" || previous == "!" {
		return 0, false, false
	}
	if previous == "<" || previous == ">" {
		if equals < start+2 || tokens[equals-2].text != previous {
			return 0, false, false
		}
		return equals - 3, true, equals >= start+3
	}
	if strings.ContainsRune("+-*/%&|^", []rune(previous)[0]) {
		return equals - 2, true, equals >= start+2
	}
	return equals - 1, false, true
}

func accessEndingAt(tokens []token, bodyStart, target int) (fieldAccess, bool) {
	if target < bodyStart || target >= len(tokens) || !identifierToken(tokens[target].text) || javaKeywords[tokens[target].text] {
		return fieldAccess{}, false
	}
	access := fieldAccess{end: target + 1, name: tokens[target].text, start: target, target: target}
	if target >= bodyStart+2 && tokens[target-1].text == "." {
		start := qualifiedChainStart(tokens, target-2)
		receiver, valid := qualifiedTokenName(tokens[start : target-1])
		if !valid {
			return fieldAccess{}, false
		}
		access.receiver, access.start = receiver, start
	}
	return access, true
}

func (state *analysisState) emitAccess(callable *callableDeclaration, access fieldAccess, relation string, localNames map[string]bool) {
	candidates, reason := state.accessCandidates(callable, access, localNames)
	attributes := map[string]any{"access_form": "identifier"}
	if access.receiver != "" {
		attributes["access_form"] = "member"
		attributes["receiver"] = access.receiver
	}
	accessSpan := runtimeSpan(callable.path, callable.tokens, access.start, access.end)
	if len(candidates) == 1 {
		state.facts.addEdge(callable.id, candidates[0], relation, callable.path, accessSpan, attributes)
		return
	}
	if len(candidates) > 1 {
		reason = "ambiguous-target"
		attributes["candidate_count"] = len(candidates)
	}
	state.facts.addUnresolved(
		callable.id, relation, sourceExcerpt(callable.source, callable.tokens, access.start, access.end),
		reason, callable.path, accessSpan, attributes,
	)
}

func (state *analysisState) accessCandidates(callable *callableDeclaration, access fieldAccess, localNames map[string]bool) ([]string, string) {
	if access.receiver == "" {
		if localNames[access.name] {
			return nil, "dynamic-target"
		}
		if parameterID := callable.parameterIDs[access.name]; parameterID != "" {
			return []string{parameterID}, ""
		}
		return fieldIDs(state.fieldCandidates(callable.ownerID, access.name, false)), "dynamic-target"
	}
	if access.receiver == "this" {
		return fieldIDs(state.fieldCandidates(callable.ownerID, access.name, false)), "unsupported-form"
	}
	if access.receiver == "super" {
		return state.parentFieldIDs(callable.ownerID, access.name), "unsupported-form"
	}
	if localNames[access.receiver] || callable.parameterIDs[access.receiver] != "" {
		return nil, "dynamic-target"
	}
	types := state.resolveTypeDeclarations(access.receiver, callable.ownerQualifiedName, callable.context)
	if len(types) == 0 {
		if typeLikeQualifier(access.receiver) {
			return nil, "external-target"
		}
		return nil, "dynamic-target"
	}
	if len(types) > 1 {
		return nil, "ambiguous-target"
	}
	return fieldIDs(state.fieldCandidates(types[0].id, access.name, true)), "unsupported-form"
}

func fieldIDs(fields []fieldDeclaration) []string {
	result := make([]string, 0, len(fields))
	for _, field := range fields {
		result = append(result, field.id)
	}
	return result
}

func (state *analysisState) parentFieldIDs(ownerID, name string) []string {
	var result []string
	for _, parentID := range state.directSuperclasses(ownerID) {
		result = append(result, fieldIDs(state.fieldCandidates(parentID, name, false))...)
	}
	return result
}

func typeLikeQualifier(qualifier string) bool {
	name := simpleName(qualifier)
	return name != "" && name[0] >= 'A' && name[0] <= 'Z'
}

func localDeclarations(callable *callableDeclaration, ignored map[int]bool) (map[int]bool, map[string]bool) {
	declarationTokens := make(map[int]bool)
	names := make(map[string]bool)
	for index := callable.bodyStart + 1; index < callable.bodyEnd; index++ {
		if ignored[index] || !identifierToken(callable.tokens[index].text) || javaKeywords[callable.tokens[index].text] {
			continue
		}
		previous := callable.tokens[index-1].text
		typeBefore := primitiveTypes[previous] || (identifierToken(previous) && !javaKeywords[previous]) || previous == "]" || previous == ">"
		if !typeBefore || index+1 >= callable.bodyEnd {
			continue
		}
		next := callable.tokens[index+1].text
		if next != "=" && next != ";" && next != "," && next != ":" && next != ")" && next != "[" {
			continue
		}
		declarationTokens[index] = true
		names[callable.tokens[index].text] = true
		markAdditionalLocalDeclarations(callable.tokens, index+1, callable.bodyEnd, declarationTokens, names)
	}
	return declarationTokens, names
}

func markAdditionalLocalDeclarations(tokens []token, start, end int, declarationTokens map[int]bool, names map[string]bool) {
	paren, bracket, braces := 0, 0, 0
	for index := start; index < end; index++ {
		switch tokens[index].text {
		case "(":
			paren++
		case ")":
			if paren == 0 {
				return
			}
			paren--
		case "[":
			bracket++
		case "]":
			bracket--
		case "{":
			braces++
		case "}":
			braces--
		case ";", ":":
			if paren == 0 && bracket == 0 && braces == 0 {
				return
			}
		case ",":
			if paren != 0 || bracket != 0 || braces != 0 {
				continue
			}
			for candidate := index + 1; candidate < end; candidate++ {
				if identifierToken(tokens[candidate].text) && !javaKeywords[tokens[candidate].text] {
					declarationTokens[candidate] = true
					names[tokens[candidate].text] = true
					break
				}
				if tokens[candidate].text == ";" || tokens[candidate].text == "=" {
					break
				}
			}
		}
	}
}
