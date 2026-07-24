package main

import (
	"path/filepath"
	"sort"
	"strings"
)

type declarationIndex struct {
	byID                 map[string]*declaration
	byQualified          map[string][]*declaration
	byName               map[string][]*declaration
	byContainerName      map[string][]*declaration
	byPathName           map[string][]*declaration
	byCallableParameters map[string][]*declaration
	visibility           visibilityIndex
}

type fileIndex struct {
	byPath     map[string]*sourceFile
	byBaseName map[string][]*sourceFile
}

func buildDeclarationIndex(declarations []*declaration, visibility visibilityIndex) declarationIndex {
	index := declarationIndex{
		byID: map[string]*declaration{}, byQualified: map[string][]*declaration{},
		byName: map[string][]*declaration{}, byContainerName: map[string][]*declaration{},
		byPathName: map[string][]*declaration{}, byCallableParameters: map[string][]*declaration{}, visibility: visibility,
	}
	for _, declaration := range declarations {
		index.byID[declaration.ID] = declaration
		qualified := normalizeQualified(declaration.QualifiedName)
		index.byQualified[qualified] = append(index.byQualified[qualified], declaration)
		index.byName[declaration.Name] = append(index.byName[declaration.Name], declaration)
		containerKey := declaration.ContainerID + "\x00" + declaration.Name
		index.byContainerName[containerKey] = append(index.byContainerName[containerKey], declaration)
		pathKey := declaration.Path + "\x00" + declaration.Name
		index.byPathName[pathKey] = append(index.byPathName[pathKey], declaration)
		if declaration.Kind == "parameter" {
			index.byCallableParameters[declaration.ContainerID] = append(index.byCallableParameters[declaration.ContainerID], declaration)
		}
	}
	for callableID := range index.byCallableParameters {
		sort.Slice(index.byCallableParameters[callableID], func(i, j int) bool {
			left, _ := index.byCallableParameters[callableID][i].Attributes["index"].(int)
			right, _ := index.byCallableParameters[callableID][j].Attributes["index"].(int)
			return left < right
		})
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

func resolveDeclarations(index declarationIndex, candidate, scope, path string, accept func(*declaration) bool) []*declaration {
	candidate = stripTemplateArguments(normalizeQualified(candidate))
	if candidate == "" {
		return nil
	}
	if strings.Contains(candidate, "::") {
		if matches := selectDeclarations(index, index.byQualified[candidate], path, accept); len(matches) > 0 {
			return matches
		}
	}
	for current := normalizeQualified(scope); current != ""; current = parentScope(current) {
		qualified := current + "::" + candidate
		if matches := selectDeclarations(index, index.byQualified[qualified], path, accept); len(matches) > 0 {
			return matches
		}
	}
	if matches := selectDeclarations(index, index.byQualified[candidate], path, accept); len(matches) > 0 {
		return matches
	}
	return selectDeclarations(index, index.byName[lastQualifiedPart(candidate)], path, accept)
}

func resolveUnscopedDeclarations(index declarationIndex, candidate, path string, accept func(*declaration) bool) []*declaration {
	candidate = stripTemplateArguments(normalizeQualified(candidate))
	if candidate == "" {
		return nil
	}
	if strings.Contains(candidate, "::") {
		if matches := visibleDeclarations(index, index.byQualified[candidate], path, accept); len(matches) > 0 {
			return matches
		}
	}
	return visibleDeclarations(index, index.byName[lastQualifiedPart(candidate)], path, accept)
}

func visibleDeclarations(index declarationIndex, values []*declaration, path string, accept func(*declaration) bool) []*declaration {
	return filterDeclarations(values, func(value *declaration) bool {
		return accept(value) && (!value.FileLocal || index.visibility.fileLocalVisible(path, value.Path))
	})
}

func selectDeclarations(index declarationIndex, values []*declaration, path string, accept func(*declaration) bool) []*declaration {
	visible := visibleDeclarations(index, values, path, accept)
	if len(visible) == 0 {
		return nil
	}
	var sameFile []*declaration
	for _, value := range visible {
		if value.Path == path {
			sameFile = append(sameFile, value)
		}
	}
	if len(sameFile) > 0 {
		return preferDefinitions(sameFile)
	}
	return preferDefinitions(visible)
}

func resolveMacroDeclarations(index declarationIndex, candidate, path string) []*declaration {
	values := filterDeclarations(index.byName[lastQualifiedPart(candidate)], func(value *declaration) bool {
		macro, _ := value.Attributes["macro"].(bool)
		return macro && (value.MacroFunction || value.MacroTarget != "")
	})
	bestRank := int(^uint(0) >> 1)
	var visible []*declaration
	for _, value := range values {
		rank, ok := index.visibility.includeRank(path, value.Path)
		if !ok {
			continue
		}
		if rank < bestRank {
			bestRank = rank
			visible = visible[:0]
		}
		if rank == bestRank {
			visible = append(visible, value)
		}
	}
	sort.Slice(visible, func(i, j int) bool { return visible[i].ID < visible[j].ID })
	return visible
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
