package main

import "strings"

func (state *analysis) resolveExtensionInvocation(
	callable *runtimeCallable,
	invocation runtimeInvocation,
) ([]*runtimeCallable, string, bool) {
	if invocation.fluent || strings.Contains(invocation.qualifier, ".") {
		return nil, "unsupported-form", false
	}
	evidence, found := state.runtimeReceiverEvidence(callable, invocation)
	if !found {
		return nil, "dynamic-target", false
	}
	if evidence.spelling == "" {
		return nil, "dynamic-target", true
	}
	receiver, nullable, reason := state.resolveRuntimeDeclaredType(evidence)
	if receiver == nil {
		if reason == "external-target" {
			reason = "dynamic-target"
		}
		return nil, reason, true
	}
	if state.runtimeTypeHasOrdinaryMember(receiver, invocation.name, invocation.arity) {
		return nil, "dynamic-target", true
	}
	visible, visibleReason := state.visibleExtensions(callable, invocation.name)
	var targets []*runtimeCallable
	for _, target := range visible {
		if !runtimeCallableAcceptsArity(target.declaration, invocation.arity) {
			continue
		}
		targetType, targetNullable, _ := state.resolveRuntimeDeclaredType(runtimeDeclaredType{
			file: target.file, lexicalOwner: target.ownerQN, spelling: target.declaration.receiver,
		})
		if targetType != nil && targetType.id == receiver.id && targetNullable == nullable {
			targets = append(targets, target)
		}
	}
	if len(targets) == 0 {
		if visibleReason != "" {
			return nil, visibleReason, true
		}
		return nil, "dynamic-target", true
	}
	targets, reason = classifyRuntimeCallables(targets, "dynamic-target")
	return targets, reason, true
}

func (state *analysis) runtimeTypeHasOrdinaryMember(receiver *runtimeType, name string, arity int) bool {
	queue := []*runtimeType{receiver}
	seen := make(map[string]struct{})
	for len(queue) != 0 {
		current := queue[0]
		queue = queue[1:]
		if _, exists := seen[current.id]; exists {
			continue
		}
		seen[current.id] = struct{}{}
		if state.runtime.hasOrdinaryMember(current.qualified, name, arity) {
			return true
		}
		queue = append(queue, state.directRuntimeSupertypes(current.qualified)...)
	}
	return false
}

func (state *analysis) visibleExtensions(callable *runtimeCallable, name string) ([]*runtimeCallable, string) {
	resolve := func(names []string) []*runtimeCallable {
		var targets []*runtimeCallable
		for _, qualified := range names {
			targets = append(targets, state.runtime.extensionsByQN[qualified]...)
		}
		return uniqueRuntimeCallables(targets)
	}
	if names, bound := importedAliasNames(callable.file, name); bound {
		if targets := resolve(names); len(targets) != 0 {
			return targets, ""
		}
		return nil, "external-target"
	}
	packageName := runtimePackage(callable.file)
	for owner := callable.ownerQN; owner != "" && owner != "<default>" && owner != packageName; {
		if targets := resolve([]string{owner + "." + name}); len(targets) != 0 {
			return targets, ""
		}
		dot := strings.LastIndexByte(owner, '.')
		if dot < 0 {
			break
		}
		owner = owner[:dot]
	}
	if names, bound := explicitImportNames(callable.file, name); bound {
		if targets := resolve(names); len(targets) != 0 {
			return targets, ""
		}
		return nil, "external-target"
	}
	if targets := resolve([]string{qualify(packageName, name)}); len(targets) != 0 {
		return targets, ""
	}
	var wildcardNames []string
	for _, imported := range callable.file.imports {
		if imported.wildcard {
			wildcardNames = append(wildcardNames, strings.TrimSuffix(imported.path, ".*")+"."+name)
		}
	}
	if targets := resolve(wildcardNames); len(targets) != 0 {
		return targets, ""
	}
	return nil, "dynamic-target"
}

func runtimeCallableAcceptsArity(declaration *declaration, arity int) bool {
	parameters := declaration.parameters
	vararg := -1
	for index, parameter := range parameters {
		if containsString(parameter.modifiers, "vararg") {
			if vararg >= 0 || index != len(parameters)-1 {
				return arity == len(parameters)
			}
			vararg = index
		}
	}
	if vararg >= 0 && arity > vararg {
		return true
	}
	if arity > len(parameters) {
		return false
	}
	end := len(parameters)
	if vararg >= 0 {
		end = vararg
	}
	for index := arity; index < end; index++ {
		if !parameters[index].hasDefault {
			return false
		}
	}
	return arity <= end || vararg >= 0
}
