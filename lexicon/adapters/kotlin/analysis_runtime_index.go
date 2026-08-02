package main

import (
	"fmt"
	"strings"
)

type runtimeIndex struct {
	callables               []*runtimeCallable
	callablesByKey          map[string][]*runtimeCallable
	constructors            map[string][]*runtimeCallable
	directCompanionsByOwner map[string][]*runtimeType
	extensionsByQN          map[string][]*runtimeCallable
	ordinaryMembersByKey    map[string]*runtimeAcceptedArities
	propertiesByKey         map[string][]runtimeProperty
	typesByID               map[string]*runtimeType
	typesByQN               map[string][]*runtimeType
}

type runtimeAcceptedArities struct {
	exact        map[int]struct{}
	variadicFrom int
}

type runtimeCallable struct {
	declaration *declaration
	file        *parsedKotlinFile
	id          string
	kind        string
	ownerKind   string
	ownerQN     string
	parameters  map[string][]string
	signature   string
}

type runtimeProperty struct {
	declaration *declaration
	file        *parsedKotlinFile
	id          string
	ownerQN     string
}

type runtimeType struct {
	declaration *declaration
	file        *parsedKotlinFile
	form        string
	id          string
	kind        string
	ownerQN     string
	qualified   string
}

func newRuntimeIndex() *runtimeIndex {
	return &runtimeIndex{
		callablesByKey:          make(map[string][]*runtimeCallable),
		constructors:            make(map[string][]*runtimeCallable),
		directCompanionsByOwner: make(map[string][]*runtimeType),
		extensionsByQN:          make(map[string][]*runtimeCallable),
		ordinaryMembersByKey:    make(map[string]*runtimeAcceptedArities),
		propertiesByKey:         make(map[string][]runtimeProperty),
		typesByID:               make(map[string]*runtimeType),
		typesByQN:               make(map[string][]*runtimeType),
	}
}

func (state *analysis) indexRuntimeDeclaration(
	file *parsedKotlinFile,
	declaration *declaration,
	id, kind, ownerQN, ownerKind, qualifiedBase string,
) *runtimeCallable {
	switch kind {
	case "type", "interface":
		target := &runtimeType{
			declaration: declaration, file: file, form: declaration.form, id: id,
			kind: kind, ownerQN: ownerQN, qualified: qualifiedBase,
		}
		state.runtime.typesByID[id] = target
		state.runtime.typesByQN[qualifiedBase] = append(state.runtime.typesByQN[qualifiedBase], target)
		if target.form == "companion_object" {
			state.runtime.directCompanionsByOwner[ownerQN] = append(
				state.runtime.directCompanionsByOwner[ownerQN], target,
			)
		}
	case "field":
		property := runtimeProperty{declaration: declaration, file: file, id: id, ownerQN: ownerQN}
		key := runtimeMemberKey(ownerQN, declaration.name)
		state.runtime.propertiesByKey[key] = append(state.runtime.propertiesByKey[key], property)
	case "function", "method", "constructor":
		callable := &runtimeCallable{
			declaration: declaration, file: file, id: id, kind: kind,
			ownerKind: ownerKind, ownerQN: ownerQN, parameters: make(map[string][]string),
			signature: normalizedParameterSignature(declaration.parameters),
		}
		state.runtime.callables = append(state.runtime.callables, callable)
		if declaration.receiver == "" {
			state.runtime.indexOrdinaryMember(ownerQN, declaration)
		}
		if kind == "constructor" {
			state.runtime.constructors[runtimeArityKey(ownerQN, len(declaration.parameters))] = append(
				state.runtime.constructors[runtimeArityKey(ownerQN, len(declaration.parameters))], callable,
			)
		} else {
			key := runtimeCallableKey(ownerQN, declaration.name, len(declaration.parameters))
			state.runtime.callablesByKey[key] = append(state.runtime.callablesByKey[key], callable)
			if declaration.receiver != "" {
				state.runtime.extensionsByQN[qualifiedBase] = append(
					state.runtime.extensionsByQN[qualifiedBase], callable,
				)
			}
		}
		return callable
	}
	return nil
}

func (index *runtimeIndex) indexOrdinaryMember(owner string, declaration *declaration) {
	key := runtimeMemberKey(owner, declaration.name)
	arities := index.ordinaryMembersByKey[key]
	if arities == nil {
		arities = &runtimeAcceptedArities{exact: make(map[int]struct{}), variadicFrom: -1}
		index.ordinaryMembersByKey[key] = arities
	}
	for arity := 0; arity <= len(declaration.parameters); arity++ {
		if runtimeCallableAcceptsArity(declaration, arity) {
			arities.exact[arity] = struct{}{}
		}
	}
	variadicFrom := len(declaration.parameters) + 1
	if runtimeCallableAcceptsArity(declaration, variadicFrom) &&
		(arities.variadicFrom < 0 || variadicFrom < arities.variadicFrom) {
		arities.variadicFrom = variadicFrom
	}
}

func (index *runtimeIndex) hasOrdinaryMember(owner, name string, arity int) bool {
	arities := index.ordinaryMembersByKey[runtimeMemberKey(owner, name)]
	if arities == nil {
		return false
	}
	if _, exists := arities.exact[arity]; exists {
		return true
	}
	return arities.variadicFrom >= 0 && arity >= arities.variadicFrom
}

func runtimeCallableKey(owner, name string, arity int) string {
	return fmt.Sprintf("%s\x00%s\x00%d", owner, name, arity)
}

func runtimeArityKey(owner string, arity int) string {
	return fmt.Sprintf("%s\x00%d", owner, arity)
}

func runtimeMemberKey(owner, name string) string {
	return owner + "\x00" + name
}

func normalizedParameterSignature(parameters []parameterDecl) string {
	parts := make([]string, len(parameters))
	for index, parameter := range parameters {
		parts[index] = normalizeRuntimeType(parameter.typeName)
	}
	return strings.Join(parts, ",")
}

func normalizeRuntimeType(value string) string {
	return strings.Join(strings.Fields(value), "")
}

func normalizedReceiver(declaration *declaration) string {
	return normalizeRuntimeType(declaration.receiver)
}
