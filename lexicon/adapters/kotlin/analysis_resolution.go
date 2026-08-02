package main

import "strings"

func (state *analysis) resolveRelationship(pending pendingRelationship) ([]relationshipTarget, string) {
	name := pending.targetName
	if name == "" {
		return nil, "unsupported-form"
	}
	filter := func(targets []relationshipTarget) []relationshipTarget {
		return filterRelationshipTargets(targets, pending.kind == "annotation")
	}
	resolveNames := func(names []string) ([]relationshipTarget, bool) {
		var candidates []relationshipTarget
		for _, qualifiedName := range names {
			candidates = append(candidates, filter(state.relationshipByQN[qualifiedName])...)
		}
		return uniqueRelationshipTargets(candidates), len(candidates) != 0
	}

	if strings.Contains(name, ".") {
		if targets, found := resolveNames([]string{name}); found {
			return classifyRelationshipTargets(targets)
		}
	}
	if expanded, bound := importedAliasNames(pending.file, name); bound {
		targets, _ := resolveNames(expanded)
		return classifyRelationshipTargets(targets)
	}
	packageName := pending.file.packageName
	if packageName == "" {
		packageName = "<default>"
	}
	if targets := state.resolveLexical(pending.lexicalOwner, packageName, name, filter); len(targets) != 0 {
		return classifyRelationshipTargets(targets)
	}
	if imported, bound := explicitImportNames(pending.file, name); bound {
		targets, _ := resolveNames(imported)
		return classifyRelationshipTargets(targets)
	}
	if targets, found := resolveNames([]string{qualify(packageName, name)}); found {
		return classifyRelationshipTargets(targets)
	}
	var wildcardNames []string
	for _, imported := range pending.file.imports {
		if imported.wildcard {
			wildcardNames = append(wildcardNames, strings.TrimSuffix(imported.path, ".*")+"."+name)
		}
	}
	targets, _ := resolveNames(wildcardNames)
	return classifyRelationshipTargets(targets)
}

func (state *analysis) resolveLexical(owner, packageName, name string, filter func([]relationshipTarget) []relationshipTarget) []relationshipTarget {
	for owner != "" && owner != "<default>" && owner != packageName {
		candidates := uniqueRelationshipTargets(filter(state.relationshipByQN[owner+"."+name]))
		if len(candidates) != 0 {
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

func importedAliasNames(file *parsedKotlinFile, name string) ([]string, bool) {
	first, suffix := name, ""
	if dot := strings.IndexByte(name, '.'); dot >= 0 {
		first, suffix = name[:dot], name[dot:]
	}
	var names []string
	for _, imported := range file.imports {
		if imported.alias == first {
			names = append(names, imported.path+suffix)
		}
	}
	return names, len(names) != 0
}

func explicitImportNames(file *parsedKotlinFile, name string) ([]string, bool) {
	first, suffix := name, ""
	if dot := strings.IndexByte(name, '.'); dot >= 0 {
		first, suffix = name[:dot], name[dot:]
	}
	var names []string
	for _, imported := range file.imports {
		if imported.wildcard || imported.alias != "" || simpleQualifiedName(imported.path) != first {
			continue
		}
		names = append(names, imported.path+suffix)
	}
	return names, len(names) != 0
}

func filterRelationshipTargets(targets []relationshipTarget, annotationsOnly bool) []relationshipTarget {
	var filtered []relationshipTarget
	for _, target := range targets {
		if !annotationsOnly || target.form == "annotation_class" {
			filtered = append(filtered, target)
		}
	}
	return filtered
}

func uniqueRelationshipTargets(targets []relationshipTarget) []relationshipTarget {
	seen := make(map[string]struct{}, len(targets))
	result := make([]relationshipTarget, 0, len(targets))
	for _, target := range targets {
		if _, exists := seen[target.id]; exists {
			continue
		}
		seen[target.id] = struct{}{}
		result = append(result, target)
	}
	return result
}

func classifyRelationshipTargets(targets []relationshipTarget) ([]relationshipTarget, string) {
	if len(targets) == 1 {
		return targets, ""
	}
	if len(targets) > 1 {
		return nil, "ambiguous-target"
	}
	return nil, "external-target"
}
