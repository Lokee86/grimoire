package main

import (
	"sort"
	"strings"
)

type callableDeclaration struct {
	arity                  int
	bodyEnd                int
	bodyStart              int
	constructor            bool
	context                resolutionContext
	id                     string
	modifiers              []string
	name                   string
	ownerID                string
	ownerQualifiedName     string
	parameterIDs           map[string]string
	parameterReceiverTypes map[string]string
	parameterTypes         []string
	path                   string
	signature              string
	source                 string
	tokens                 []token
}

type callableIndexKey struct {
	name      string
	ownerID   string
	signature string
}

type callableLookupKey struct {
	constructor bool
	name        string
	ownerID     string
}

type fieldDeclaration struct {
	id        string
	modifiers []string
	name      string
	ownerID   string
}

func (context resolutionContext) clone() resolutionContext {
	return resolutionContext{
		explicitImports: append([]string(nil), context.explicitImports...),
		packageName:     context.packageName,
		wildcardImports: append([]string(nil), context.wildcardImports...),
	}
}

func (state *analysisState) registerCallable(declaration callableDeclaration) {
	state.callables = append(state.callables, declaration)
}

func (state *analysisState) registerField(declaration fieldDeclaration) {
	if state.fields[declaration.ownerID] == nil {
		state.fields[declaration.ownerID] = make(map[string][]fieldDeclaration)
	}
	state.fields[declaration.ownerID][declaration.name] = append(
		state.fields[declaration.ownerID][declaration.name], declaration,
	)
}

func (state *analysisState) resolveRuntimeSemantics() {
	sort.Slice(state.callables, func(left, right int) bool {
		return state.callables[left].id < state.callables[right].id
	})
	state.indexDirectParents()
	state.indexCallables()
	state.emitOverrides()
	for index := range state.callables {
		callable := &state.callables[index]
		if callable.bodyStart < 0 || callable.bodyEnd < callable.bodyStart {
			continue
		}
		state.emitCalls(callable)
		state.emitDataflow(callable)
	}
}

func (state *analysisState) indexCallables() {
	state.callablesByOwnerKindName = make(map[callableLookupKey][]callableDeclaration)
	state.callablesByOwnerNameSignature = make(map[callableIndexKey][]callableDeclaration)
	for _, declaration := range state.callables {
		overrideKey := callableIndexKey{ownerID: declaration.ownerID, name: declaration.name, signature: declaration.signature}
		state.callablesByOwnerNameSignature[overrideKey] = append(state.callablesByOwnerNameSignature[overrideKey], declaration)
		lookupName := declaration.name
		if declaration.constructor {
			lookupName = ""
		}
		lookupKey := callableLookupKey{ownerID: declaration.ownerID, constructor: declaration.constructor, name: lookupName}
		state.callablesByOwnerKindName[lookupKey] = append(state.callablesByOwnerKindName[lookupKey], declaration)
	}
}

func (state *analysisState) overrideCandidates(ownerID, name, signature string) []callableDeclaration {
	return state.callablesByOwnerNameSignature[callableIndexKey{ownerID: ownerID, name: name, signature: signature}]
}

func (state *analysisState) indexDirectParents() {
	state.parentIDs = make(map[string][]string)
	state.superclassIDs = make(map[string][]string)
	for _, edge := range state.facts.edges {
		relation, _ := edge["relation"].(string)
		if relation != "extends" && relation != "implements" {
			continue
		}
		source, _ := edge["source"].(string)
		target, _ := edge["target"].(string)
		state.parentIDs[source] = appendUnique(state.parentIDs[source], target)
		if relation == "extends" {
			state.superclassIDs[source] = appendUnique(state.superclassIDs[source], target)
		}
	}
	for source := range state.parentIDs {
		sort.Strings(state.parentIDs[source])
	}
	for source := range state.superclassIDs {
		sort.Strings(state.superclassIDs[source])
	}
}

func (state *analysisState) callableCandidates(ownerID, name string, arity int, constructor, staticOnly bool) []callableDeclaration {
	var result []callableDeclaration
	seen := make(map[string]bool)
	if constructor {
		name = ""
	}
	key := callableLookupKey{ownerID: ownerID, constructor: constructor, name: name}
	for _, declaration := range state.callablesByOwnerKindName[key] {
		if seen[declaration.id] {
			continue
		}
		if staticOnly && !hasModifier(declaration.modifiers, "static") {
			continue
		}
		if !callableAcceptsArity(declaration, arity) {
			continue
		}
		seen[declaration.id] = true
		result = append(result, declaration)
	}
	sort.Slice(result, func(left, right int) bool { return result[left].id < result[right].id })
	return result
}

func (state *analysisState) instanceCallableCandidates(ownerID, name string, arity int) []callableDeclaration {
	var result []callableDeclaration
	for _, declaration := range state.callableCandidates(ownerID, name, arity, false, false) {
		if !hasModifier(declaration.modifiers, "static") {
			result = append(result, declaration)
		}
	}
	return result
}

func callableAcceptsArity(declaration callableDeclaration, arity int) bool {
	if declaration.arity == arity {
		return true
	}
	if declaration.arity == 0 || len(declaration.parameterTypes) == 0 {
		return false
	}
	return strings.Contains(declaration.parameterTypes[len(declaration.parameterTypes)-1], "...") && arity >= declaration.arity-1
}

func (state *analysisState) fieldCandidates(ownerID, name string, staticOnly bool) []fieldDeclaration {
	var result []fieldDeclaration
	seen := make(map[string]bool)
	for _, declaration := range state.fields[ownerID][name] {
		if seen[declaration.id] || (staticOnly && !hasModifier(declaration.modifiers, "static")) {
			continue
		}
		seen[declaration.id] = true
		result = append(result, declaration)
	}
	sort.Slice(result, func(left, right int) bool { return result[left].id < result[right].id })
	return result
}

func (state *analysisState) resolveTypeDeclarations(name, lexicalOwner string, context resolutionContext) []typeDeclaration {
	qualifiedNames := candidateQualifiedNames(name, lexicalOwner, context)
	if state.exactQualifiedReference(name) {
		qualifiedNames = appendUnique(qualifiedNames, name)
	}
	var result []typeDeclaration
	seen := make(map[string]bool)
	for _, qualifiedName := range qualifiedNames {
		for _, declaration := range state.types[qualifiedName] {
			if !seen[declaration.id] {
				seen[declaration.id] = true
				result = append(result, declaration)
			}
		}
	}
	sort.Slice(result, func(left, right int) bool { return result[left].id < result[right].id })
	return result
}

func (state *analysisState) directParents(ownerID string) []string {
	return state.parentIDs[ownerID]
}

func hasModifier(modifiers []string, wanted string) bool {
	for _, modifier := range modifiers {
		if modifier == wanted {
			return true
		}
	}
	return false
}

func runtimeSpan(path string, tokens []token, start, end int) *span {
	if start < 0 || start >= len(tokens) || end <= start {
		return nil
	}
	if end > len(tokens) {
		end = len(tokens)
	}
	first, last := tokens[start], tokens[end-1]
	return &span{
		EndColumn: last.endColumn, EndLine: last.endLine, Path: path,
		StartColumn: first.column, StartLine: first.line,
	}
}
