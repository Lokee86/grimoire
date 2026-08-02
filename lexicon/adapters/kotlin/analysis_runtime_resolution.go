package main

import "strings"

func (state *analysis) resolveRuntimeTypes(file *parsedKotlinFile, lexicalOwner, name string) ([]*runtimeType, string) {
	resolveNames := func(names []string) []*runtimeType {
		var candidates []*runtimeType
		for _, qualifiedName := range names {
			candidates = append(candidates, state.runtime.typesByQN[qualifiedName]...)
		}
		return uniqueRuntimeTypes(candidates)
	}
	classify := func(targets []*runtimeType) ([]*runtimeType, string) {
		if len(targets) == 1 {
			return targets, ""
		}
		if len(targets) > 1 {
			return targets, "ambiguous-target"
		}
		return nil, "external-target"
	}

	if name == "" {
		return nil, "unsupported-form"
	}
	if strings.Contains(name, ".") {
		if targets := resolveNames([]string{name}); len(targets) != 0 {
			return classify(targets)
		}
	}
	if expanded, bound := importedAliasNames(file, name); bound {
		return classify(resolveNames(expanded))
	}
	packageName := runtimePackage(file)
	if targets := state.resolveRuntimeLexical(lexicalOwner, packageName, name); len(targets) != 0 {
		return classify(targets)
	}
	if imported, bound := explicitImportNames(file, name); bound {
		return classify(resolveNames(imported))
	}
	if targets := resolveNames([]string{qualify(packageName, name)}); len(targets) != 0 {
		return classify(targets)
	}
	var wildcardNames []string
	for _, imported := range file.imports {
		if imported.wildcard {
			wildcardNames = append(wildcardNames, strings.TrimSuffix(imported.path, ".*")+"."+name)
		}
	}
	return classify(resolveNames(wildcardNames))
}

func (state *analysis) resolveRuntimeLexical(owner, packageName, name string) []*runtimeType {
	for owner != "" && owner != "<default>" && owner != packageName {
		if candidates := uniqueRuntimeTypes(state.runtime.typesByQN[owner+"."+name]); len(candidates) != 0 {
			return candidates
		}
		dot := strings.LastIndexByte(owner, '.')
		if dot < 0 {
			break
		}
		owner = owner[:dot]
	}
	return nil
}

func uniqueRuntimeTypes(targets []*runtimeType) []*runtimeType {
	seen := make(map[string]struct{}, len(targets))
	result := make([]*runtimeType, 0, len(targets))
	for _, target := range targets {
		if _, exists := seen[target.id]; exists {
			continue
		}
		seen[target.id] = struct{}{}
		result = append(result, target)
	}
	return result
}

func uniqueRuntimeCallables(targets []*runtimeCallable) []*runtimeCallable {
	seen := make(map[string]struct{}, len(targets))
	result := make([]*runtimeCallable, 0, len(targets))
	for _, target := range targets {
		if _, exists := seen[target.id]; exists {
			continue
		}
		seen[target.id] = struct{}{}
		result = append(result, target)
	}
	return result
}

func runtimePackage(file *parsedKotlinFile) string {
	if file.packageName == "" {
		return "<default>"
	}
	return file.packageName
}

func ordinaryCallables(targets []*runtimeCallable) []*runtimeCallable {
	var result []*runtimeCallable
	for _, target := range targets {
		if target.declaration.receiver == "" {
			result = append(result, target)
		}
	}
	return uniqueRuntimeCallables(result)
}

func (state *analysis) sameOwnerCallables(callable *runtimeCallable, name string, arity int) []*runtimeCallable {
	if callable.ownerKind != "type" && callable.ownerKind != "interface" {
		return nil
	}
	return ordinaryCallables(state.runtime.callablesByKey[runtimeCallableKey(callable.ownerQN, name, arity)])
}

func (state *analysis) packageCallables(callable *runtimeCallable, name string, arity int) []*runtimeCallable {
	return ordinaryCallables(state.runtime.callablesByKey[runtimeCallableKey(runtimePackage(callable.file), name, arity)])
}

func (state *analysis) constructorsFor(owner string, arity int) []*runtimeCallable {
	return uniqueRuntimeCallables(state.runtime.constructors[runtimeArityKey(owner, arity)])
}

func (state *analysis) qualifiedCallables(owner *runtimeType, name string, arity int) []*runtimeCallable {
	if owner.form == "object" || owner.form == "data_object" || owner.form == "companion_object" {
		return ordinaryCallables(state.runtime.callablesByKey[runtimeCallableKey(owner.qualified, name, arity)])
	}
	var targets []*runtimeCallable
	for _, companion := range state.runtime.directCompanionsByOwner[owner.qualified] {
		targets = append(targets, state.runtime.callablesByKey[runtimeCallableKey(companion.qualified, name, arity)]...)
	}
	return ordinaryCallables(targets)
}
