package main

import (
	"fmt"
	"strings"
)

func (state *analysis) emitImports(file *parsedKotlinFile) {
	moduleID := state.moduleByPath[file.path]
	for index, imported := range file.imports {
		localName := imported.alias
		if localName == "" {
			localName = imported.path
			if dot := strings.LastIndexByte(localName, '.'); dot >= 0 {
				localName = localName[dot+1:]
			}
		}
		canonical := fmt.Sprintf("%s::import::%s::%s::%d", file.path, imported.path, imported.alias, index)
		attributes := map[string]any{
			"alias": imported.alias, "imported": imported.path, "wildcard": imported.wildcard,
		}
		importID := state.facts.addNode("import", canonical, localName, file.path, imported.path, file.path, &imported.span, attributes)
		state.facts.addEdge(moduleID, importID, "defines", file.path, &imported.span, nil)
		targetAttributes := map[string]any{"external": true, "resolution": "syntactic-import-target"}
		targetID := state.facts.addNode("symbol", "external-import:"+imported.path, localName, "external", imported.path, "", nil, targetAttributes)
		edgeAttributes := map[string]any{"alias": imported.alias, "wildcard": imported.wildcard}
		state.facts.addEdge(importID, targetID, "imports", file.path, &imported.span, edgeAttributes)
	}
}
