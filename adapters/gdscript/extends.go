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
		if ids := facts.classByFileAndName[normalizeSourcePath(pf.path)][name]; len(ids) == 1 && ids[0] != source {
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
