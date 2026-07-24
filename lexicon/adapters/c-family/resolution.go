package main

import (
	"path/filepath"
	"sort"
	"strings"
)

type declarationIndex struct {
	byID            map[string]*declaration
	byQualified     map[string][]*declaration
	byName          map[string][]*declaration
	byContainerName map[string][]*declaration
}

type fileIndex struct {
	byPath     map[string]*sourceFile
	byBaseName map[string][]*sourceFile
}

func buildDeclarationIndex(declarations []*declaration) declarationIndex {
	index := declarationIndex{
		byID: map[string]*declaration{}, byQualified: map[string][]*declaration{},
		byName: map[string][]*declaration{}, byContainerName: map[string][]*declaration{},
	}
	for _, declaration := range declarations {
		index.byID[declaration.ID] = declaration
		qualified := normalizeQualified(declaration.QualifiedName)
		index.byQualified[qualified] = append(index.byQualified[qualified], declaration)
		index.byName[declaration.Name] = append(index.byName[declaration.Name], declaration)
		key := declaration.ContainerID + "\x00" + declaration.Name
		index.byContainerName[key] = append(index.byContainerName[key], declaration)
	}
	return index
}

func buildFileIndex(files []*sourceFile) fileIndex {
	index := fileIndex{byPath: map[string]*sourceFile{}, byBaseName: map[string][]*sourceFile{}}
	for _, file := range files {
		index.byPath[file.Path] = file
		base := strings.ToLower(filepath.Base(filepath.FromSlash(file.Path)))
		index.byBaseName[base] = append(index.byBaseName[base], file)
	}
	return index
}

func resolveDeclarations(index declarationIndex, candidate, scope string, accept func(*declaration) bool) []*declaration {
	candidate = stripTemplateArguments(normalizeQualified(candidate))
	if candidate == "" {
		return nil
	}
	if strings.Contains(candidate, "::") {
		if matches := filterDeclarations(index.byQualified[candidate], accept); len(matches) > 0 {
			return preferDefinitions(matches)
		}
	}
	for current := normalizeQualified(scope); current != ""; current = parentScope(current) {
		qualified := current + "::" + candidate
		if matches := filterDeclarations(index.byQualified[qualified], accept); len(matches) > 0 {
			return preferDefinitions(matches)
		}
	}
	if matches := filterDeclarations(index.byQualified[candidate], accept); len(matches) > 0 {
		return preferDefinitions(matches)
	}
	return preferDefinitions(filterDeclarations(index.byName[lastQualifiedPart(candidate)], accept))
}

func filterDeclarations(values []*declaration, accept func(*declaration) bool) []*declaration {
	unique := map[string]*declaration{}
	for _, value := range values {
		if accept(value) {
			unique[value.ID] = value
		}
	}
	result := make([]*declaration, 0, len(unique))
	for _, value := range unique {
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func preferDefinitions(values []*declaration) []*declaration {
	var definitions []*declaration
	for _, value := range values {
		if value.Definition {
			definitions = append(definitions, value)
		}
	}
	if len(definitions) > 0 {
		return definitions
	}
	return values
}

func unresolvedRecord(source, relation, expression, reason, owner string, span sourceSpan) map[string]any {
	return map[string]any{
		"expression": expression, "owner": owner, "reason": reason, "record": "unresolved",
		"relation": relation, "source": source, "span": span.record(),
	}
}

func resolutionReason(candidates []*declaration) string {
	if len(candidates) > 1 {
		return "ambiguous-target"
	}
	return "missing-target"
}

func parentScope(scope string) string {
	if index := strings.LastIndex(scope, "::"); index >= 0 {
		return scope[:index]
	}
	return ""
}

func stripTemplateArguments(value string) string {
	if index := strings.Index(value, "<"); index >= 0 {
		return value[:index]
	}
	return value
}
