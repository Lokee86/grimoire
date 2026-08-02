package main

import "strings"

func (state *analysis) emitRuntimeSemantics() {
	state.emitOverrides()
	for _, callable := range state.runtime.callables {
		state.emitCallableCalls(callable)
		state.emitCallableDataflow(callable)
	}
}

func (state *analysis) emitCallableCalls(callable *runtimeCallable) {
	shadowed := runtimeShadowedNames(callable)
	invocations := runtimeInvocations(callable.file, callable.declaration.delegation, callable.declaration.body)
	for _, invocation := range invocations {
		targets, reason := state.resolveInvocation(callable, invocation, shadowed)
		switch len(targets) {
		case 0:
			state.facts.addUnresolved(
				callable.id, "calls", invocation.expression, reason, callable.file.path, &invocation.span, nil,
			)
		case 1:
			state.facts.addEdge(callable.id, targets[0].id, "calls", callable.file.path, &invocation.span, nil)
		default:
			for _, target := range targets {
				state.facts.addEdge(callable.id, target.id, "possible-calls", callable.file.path, &invocation.span, nil)
			}
		}
	}
}

func (state *analysis) resolveInvocation(
	callable *runtimeCallable,
	invocation runtimeInvocation,
	shadowed map[string]struct{},
) ([]*runtimeCallable, string) {
	if invocation.unsupported {
		return nil, "unsupported-form"
	}
	if invocation.qualifier == "" {
		return state.resolveUnqualifiedInvocation(callable, invocation, shadowed)
	}
	if invocation.qualifier == "this" {
		targets := state.sameOwnerCallables(callable, invocation.name, invocation.arity)
		return classifyRuntimeCallables(targets, "external-target")
	}
	if invocation.qualifier == "super" {
		return nil, "unsupported-form"
	}
	return state.resolveQualifiedInvocation(callable, invocation, shadowed)
}

func (state *analysis) resolveUnqualifiedInvocation(
	callable *runtimeCallable,
	invocation runtimeInvocation,
	shadowed map[string]struct{},
) ([]*runtimeCallable, string) {
	if invocation.name == "this" && callable.kind == "constructor" {
		return classifyRuntimeCallables(state.constructorsFor(callable.ownerQN, invocation.arity), "external-target")
	}
	if invocation.name == "super" && callable.kind == "constructor" {
		var targets []*runtimeCallable
		for _, target := range state.directRuntimeSupertypes(callable.ownerQN) {
			targets = append(targets, state.constructorsFor(target.qualified, invocation.arity)...)
		}
		return classifyRuntimeCallables(uniqueRuntimeCallables(targets), "external-target")
	}
	if _, exists := shadowed[invocation.name]; exists || len(callable.parameters[invocation.name]) != 0 {
		return nil, "dynamic-target"
	}
	if targets := state.sameOwnerCallables(callable, invocation.name, invocation.arity); len(targets) != 0 {
		return classifyRuntimeCallables(targets, "external-target")
	}
	if targets := state.packageCallables(callable, invocation.name, invocation.arity); len(targets) != 0 {
		return classifyRuntimeCallables(targets, "external-target")
	}
	types, reason := state.resolveRuntimeTypes(callable.file, callable.ownerQN, invocation.name)
	if len(types) != 1 {
		return nil, reason
	}
	return classifyRuntimeCallables(state.constructorsFor(types[0].qualified, invocation.arity), "external-target")
}

func (state *analysis) resolveQualifiedInvocation(
	callable *runtimeCallable,
	invocation runtimeInvocation,
	shadowed map[string]struct{},
) ([]*runtimeCallable, string) {
	if targets, reason, receiver := state.resolveExtensionInvocation(callable, invocation); receiver {
		return targets, reason
	}
	first := invocation.qualifier
	if dot := strings.IndexByte(first, '.'); dot >= 0 {
		first = first[:dot]
	}
	if _, exists := shadowed[first]; exists || len(callable.parameters[first]) != 0 || state.hasRuntimeValue(callable, first) {
		return nil, "dynamic-target"
	}
	owners, reason := state.resolveRuntimeTypes(callable.file, callable.ownerQN, invocation.qualifier)
	if len(owners) != 1 {
		return nil, reason
	}
	targets := state.qualifiedCallables(owners[0], invocation.name, invocation.arity)
	fullName := invocation.qualifier + "." + invocation.name
	if nested, nestedReason := state.resolveRuntimeTypes(callable.file, callable.ownerQN, fullName); len(nested) == 1 {
		targets = append(targets, state.constructorsFor(nested[0].qualified, invocation.arity)...)
	} else if len(targets) == 0 && nestedReason == "ambiguous-target" {
		return nil, nestedReason
	}
	return classifyRuntimeCallables(uniqueRuntimeCallables(targets), "external-target")
}

func classifyRuntimeCallables(targets []*runtimeCallable, emptyReason string) ([]*runtimeCallable, string) {
	targets = uniqueRuntimeCallables(targets)
	if len(targets) == 0 {
		return nil, emptyReason
	}
	if len(targets) > 1 {
		return targets, "ambiguous-target"
	}
	return targets, ""
}
