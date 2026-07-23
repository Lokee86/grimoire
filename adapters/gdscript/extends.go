package main

import "strings"

func processExtends(facts *factSet, pf *parsedFile) {
	for i := range pf.declarations {
		decl := &pf.declarations[i]
		if decl.extends == "" {
			continue
		}
		source := decl.nodeID
		if decl.kind == "extends" || source == "" {
			source = pf.scriptOwnerID
		}
		if path, ok := normalizeImportPath(decl.extends); ok {
			path = projectResourcePath(pf.projectRoot, path)
			target := facts.scriptOwnerByPath[path]
			if target == "" {
				target = facts.moduleByPath[path]
			}
			if target != "" {
				facts.addEdge(edge(source, target, "extends", decl.span))
				facts.indexParent(source, target)
			} else {
				facts.addUnresolved(unresolved(source, "extends", decl.extends, "missing-target", decl.span))
			}
			continue
		}
		name := strings.TrimSpace(decl.extends)
		if owners, known := preloadTypeOwners(facts, pf.path, name); known {
			if len(owners) == 1 && owners[0] != source {
				facts.addEdge(edge(source, owners[0], "extends", decl.span))
				facts.indexParent(source, owners[0])
			} else {
				reason := "missing-target"
				if len(owners) > 1 {
					reason = "ambiguous-target"
				}
				record := unresolved(source, "extends", decl.extends, reason, decl.span)
				record["candidate_name"] = name
				facts.addUnresolved(record)
			}
		} else if ids := facts.classByFileAndName[normalizeSourcePath(pf.path)][name]; len(ids) == 1 && ids[0] != source {
			facts.addEdge(edge(source, ids[0], "extends", decl.span))
			facts.indexParent(source, ids[0])
		} else if ids := facts.classByName[name]; len(ids) == 1 {
			facts.addEdge(edge(source, ids[0], "extends", decl.span))
			facts.indexParent(source, ids[0])
		} else if len(facts.classByName[name]) > 1 {
			record := unresolved(source, "extends", decl.extends, "ambiguous-target", decl.span)
			record["candidate_name"] = name
			facts.addUnresolved(record)
		} else {
			if facts.externalParentByOwnerID == nil {
				facts.externalParentByOwnerID = make(map[string]bool)
			}
			facts.externalParentByOwnerID[source] = true
			reason := "external-target"
			if isBuiltin(name) {
				reason = "builtin-target"
			}
			record := unresolved(source, "extends", decl.extends, reason, decl.span)
			record["candidate_name"] = name
			facts.addUnresolved(record)
		}
	}
}

func processOverrides(facts *factSet) {
	for methodID := range facts.ownerByFunctionID {
		method := facts.declarationByID[methodID]
		if method == nil || method.ownerClassID == "" {
			continue
		}
		for _, parent := range facts.parentByOwnerID[method.ownerClassID] {
			for _, target := range inheritedMethods(facts, parent, method.name, map[string]bool{}) {
				if target != methodID {
					facts.addEdge(edge(methodID, target, "overrides", method.span))
				}
			}
		}
	}
}

func inheritedMethods(facts *factSet, owner, name string, seen map[string]bool) []string {
	if owner == "" || seen[owner] {
		return nil
	}
	seen[owner] = true
	if methods := facts.methodByOwnerID[owner][name]; len(methods) > 0 {
		return uniqueSorted(methods)
	}
	var result []string
	for _, parent := range facts.parentByOwnerID[owner] {
		result = append(result, inheritedMethods(facts, parent, name, seen)...)
	}
	return uniqueSorted(result)
}

func preloadTypeOwners(facts *factSet, sourcePath, name string) ([]string, bool) {
	parts := strings.Split(name, ".")
	if len(parts) == 0 {
		return nil, false
	}
	paths, known := facts.preloadAliasByFileAndName[normalizeSourcePath(sourcePath)][parts[0]]
	if !known {
		return nil, false
	}
	var owners []string
	for _, path := range paths {
		if owner := facts.scriptOwnerByPath[normalizeSourcePath(path)]; owner != "" {
			owners = append(owners, owner)
		}
	}
	owners = uniqueSorted(owners)
	for _, nested := range parts[1:] {
		var next []string
		for _, owner := range owners {
			next = append(next, facts.typeByOwnerID[owner][nested]...)
		}
		owners = uniqueSorted(next)
		if len(owners) == 0 {
			break
		}
	}
	return owners, true
}
