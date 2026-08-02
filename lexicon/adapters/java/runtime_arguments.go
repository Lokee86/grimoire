package main

import (
	"regexp"
	"strings"
)

type argumentEvidence struct {
	kind     string
	typeName string
}

const (
	argumentIncompatible = -1
	argumentUnknown      = 100
)

var (
	integralLiteral = regexp.MustCompile(`^[+-]?(?:0[xX][0-9a-fA-F](?:_?[0-9a-fA-F])*|0[bB][01](?:_?[01])*|[0-9](?:_?[0-9])*)[lL]?$`)
	floatingLiteral = regexp.MustCompile(`^[+-]?(?:[0-9](?:_?[0-9])*(?:\.[0-9](?:_?[0-9])*)?|\.[0-9](?:_?[0-9])*)(?:[eE][+-]?[0-9](?:_?[0-9])*)?[fFdD]?$`)
	primitiveOrder  = map[string]int{"byte": 0, "short": 1, "char": 1, "int": 2, "long": 3, "float": 4, "double": 5}
)

func (state *analysisState) narrowArgumentCandidates(callable *callableDeclaration, invocation invocationEvidence, receivers receiverTypes, candidates []callableDeclaration) []callableDeclaration {
	if len(candidates) < 2 || len(invocation.arguments) != invocation.arity {
		return candidates
	}
	for _, candidate := range candidates {
		if candidate.arity != invocation.arity || len(candidate.parameterTypes) != invocation.arity {
			return candidates
		}
	}
	evidence := make([]argumentEvidence, len(invocation.arguments))
	known := false
	for index, argument := range invocation.arguments {
		evidence[index] = invocationArgumentEvidence(callable, receivers, argument)
		known = known || evidence[index].kind != ""
	}
	if !known {
		return candidates
	}
	ranks := make([][]int, len(candidates))
	viable := make([]bool, len(candidates))
	for index, candidate := range candidates {
		ranks[index], viable[index] = state.argumentRanks(callable, candidate, evidence)
	}
	result := make([]callableDeclaration, 0, len(candidates))
	for index, candidate := range candidates {
		if !viable[index] || dominatedArgumentRanks(index, ranks, viable) {
			continue
		}
		result = append(result, candidate)
	}
	if len(result) == 0 {
		return candidates
	}
	return result
}

func invocationArgumentEvidence(callable *callableDeclaration, receivers receiverTypes, bounds [2]int) argumentEvidence {
	tokens := callable.tokens[bounds[0]:bounds[1]]
	text := normalizedTokens(tokens)
	if len(tokens) == 1 {
		switch text {
		case "null":
			return argumentEvidence{kind: "null"}
		case "true", "false":
			return argumentEvidence{kind: "typed", typeName: "boolean"}
		}
		if strings.HasPrefix(text, `"`) && strings.HasSuffix(text, `"`) {
			return argumentEvidence{kind: "typed", typeName: "String"}
		}
		if strings.HasPrefix(text, `'`) && strings.HasSuffix(text, `'`) {
			return argumentEvidence{kind: "typed", typeName: "char"}
		}
		if identifierToken(text) {
			if typeName, found := receivers.declaredType(text, bounds[0]); found && typeName != "" && typeName != "var" {
				return argumentEvidence{kind: "typed", typeName: typeName}
			}
		}
	}
	if integralLiteral.MatchString(text) {
		typeName := "int"
		if strings.HasSuffix(text, "l") || strings.HasSuffix(text, "L") {
			typeName = "long"
		}
		return argumentEvidence{kind: "typed", typeName: typeName}
	}
	if floatingLiteral.MatchString(text) {
		typeName := "double"
		if strings.HasSuffix(text, "f") || strings.HasSuffix(text, "F") {
			typeName = "float"
		}
		return argumentEvidence{kind: "typed", typeName: typeName}
	}
	return argumentEvidence{}
}

func (state *analysisState) argumentRanks(callable *callableDeclaration, candidate callableDeclaration, evidence []argumentEvidence) ([]int, bool) {
	ranks := make([]int, len(evidence))
	for index, item := range evidence {
		ranks[index] = state.argumentConversionRank(callable, candidate, item, candidate.parameterTypes[index])
		if ranks[index] == argumentIncompatible {
			return ranks, false
		}
	}
	return ranks, true
}

func (state *analysisState) argumentConversionRank(callable *callableDeclaration, candidate callableDeclaration, evidence argumentEvidence, parameterType string) int {
	if evidence.kind == "" || strings.Contains(parameterType, "...") {
		return argumentUnknown
	}
	parameterPrimitive := primitiveTypes[parameterType] && parameterType != "var"
	if evidence.kind == "null" {
		if parameterPrimitive {
			return argumentIncompatible
		}
		return 0
	}
	if primitiveTypes[evidence.typeName] {
		if !parameterPrimitive && !unsafeArgumentType(parameterType) &&
			len(state.resolveTypeDeclarations(parameterType, candidate.ownerQualifiedName, candidate.context)) == 1 {
			return argumentIncompatible
		}
		return primitiveArgumentRank(evidence.typeName, parameterType)
	}
	if parameterPrimitive {
		return state.referenceToPrimitiveRank(callable, evidence.typeName, parameterType)
	}
	if state.sameDeclaredArgumentType(callable, candidate, evidence.typeName, parameterType) {
		return 0
	}
	return state.referenceArgumentRank(callable, candidate, evidence.typeName, parameterType)
}

func primitiveArgumentRank(source, target string) int {
	if source == target {
		return 0
	}
	targetName := javaLangType(target)
	if targetName == boxedType(source) {
		return 20
	}
	if numericPrimitive(source) && targetName == "Number" {
		return 21
	}
	if targetName == "Object" {
		return 22
	}
	if targetName != "" {
		return argumentIncompatible
	}
	if !primitiveTypes[target] {
		return argumentUnknown
	}
	if source == "boolean" || target == "boolean" {
		return argumentIncompatible
	}
	if target == "char" || primitiveOrder[target] <= primitiveOrder[source] || (source == "char" && target == "short") {
		return argumentIncompatible
	}
	return primitiveOrder[target] - primitiveOrder[source]
}

func (state *analysisState) referenceToPrimitiveRank(callable *callableDeclaration, source, target string) int {
	sourceTypes := state.resolveTypeDeclarations(source, callable.ownerQualifiedName, callable.context)
	if len(sourceTypes) == 1 {
		return argumentIncompatible
	}
	if len(sourceTypes) > 1 {
		return argumentUnknown
	}
	sourceName := javaLangType(source)
	if sourceName == "" {
		return argumentUnknown
	}
	unboxed := unboxedType(sourceName)
	if unboxed == "" {
		return argumentIncompatible
	}
	rank := primitiveArgumentRank(unboxed, target)
	if rank >= 0 && rank < argumentUnknown {
		return rank + 20
	}
	return rank
}

func (state *analysisState) sameDeclaredArgumentType(callable *callableDeclaration, candidate callableDeclaration, source, target string) bool {
	if unsafeArgumentType(source) || unsafeArgumentType(target) {
		return false
	}
	sourceTypes := state.resolveTypeDeclarations(source, callable.ownerQualifiedName, callable.context)
	targetTypes := state.resolveTypeDeclarations(target, candidate.ownerQualifiedName, candidate.context)
	if len(sourceTypes) == 1 && len(targetTypes) == 1 {
		return sourceTypes[0].id == targetTypes[0].id
	}
	if len(sourceTypes) != 0 || len(targetTypes) != 0 {
		return false
	}
	return javaLangType(source) != "" && javaLangType(source) == javaLangType(target)
}

func (state *analysisState) referenceArgumentRank(callable *callableDeclaration, candidate callableDeclaration, source, target string) int {
	if unsafeArgumentType(source) || unsafeArgumentType(target) {
		return argumentUnknown
	}
	sourceTypes := state.resolveTypeDeclarations(source, callable.ownerQualifiedName, callable.context)
	targetTypes := state.resolveTypeDeclarations(target, candidate.ownerQualifiedName, candidate.context)
	if len(sourceTypes) == 1 && len(targetTypes) == 1 {
		return 10
	}
	sourceKnown := len(sourceTypes) == 1 || (len(sourceTypes) == 0 && javaLangType(source) != "")
	targetKnown := len(targetTypes) == 1 || (len(targetTypes) == 0 && javaLangType(target) != "")
	if sourceKnown && targetKnown {
		return 10
	}
	if len(sourceTypes) != 0 || len(targetTypes) != 0 {
		return argumentUnknown
	}
	sourceName, targetName := javaLangType(source), javaLangType(target)
	if sourceName == "" || targetName == "" {
		return argumentUnknown
	}
	if targetName == "Object" {
		return 1
	}
	if numericWrapper(sourceName) && targetName == "Number" {
		return 1
	}
	return argumentIncompatible
}

func dominatedArgumentRanks(index int, ranks [][]int, viable []bool) bool {
	for other := range ranks {
		if other == index || !viable[other] || argumentRanksUnknown(ranks[index]) || argumentRanksUnknown(ranks[other]) {
			continue
		}
		better, strict := true, false
		for argument := range ranks[index] {
			if ranks[other][argument] > ranks[index][argument] {
				better = false
			}
			strict = strict || ranks[other][argument] < ranks[index][argument]
		}
		if better && strict {
			return true
		}
	}
	return false
}

func argumentRanksUnknown(ranks []int) bool {
	for _, rank := range ranks {
		if rank == argumentUnknown {
			return true
		}
	}
	return false
}

func unsafeArgumentType(typeName string) bool {
	return typeName == "" || strings.ContainsAny(typeName, "<>?[]&")
}

func javaLangType(typeName string) string {
	name := strings.TrimPrefix(typeName, "java.lang.")
	switch name {
	case "Boolean", "Byte", "Character", "Double", "Float", "Integer", "Long", "Number", "Object", "Short", "String":
		return name
	}
	return ""
}

func boxedType(primitive string) string {
	return map[string]string{
		"boolean": "Boolean", "byte": "Byte", "char": "Character", "double": "Double",
		"float": "Float", "int": "Integer", "long": "Long", "short": "Short",
	}[primitive]
}

func unboxedType(reference string) string {
	return map[string]string{
		"Boolean": "boolean", "Byte": "byte", "Character": "char", "Double": "double",
		"Float": "float", "Integer": "int", "Long": "long", "Short": "short",
	}[reference]
}

func numericPrimitive(typeName string) bool {
	return typeName != "" && typeName != "boolean" && typeName != "char" && typeName != "var" && primitiveTypes[typeName]
}

func numericWrapper(typeName string) bool {
	switch typeName {
	case "Byte", "Double", "Float", "Integer", "Long", "Short":
		return true
	}
	return false
}
