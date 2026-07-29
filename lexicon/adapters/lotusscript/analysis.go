package main

import (
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

func analyzeRepository(repository string) ([]byte, error) {
	snapshot, err := discoverRepository(repository)
	if err != nil {
		return nil, err
	}
	facts := newFactSet()
	state := &analysisState{
		basesByClass:    make(map[string][]string),
		callablesByName: make(map[string][]declaration),
		classesByName:   make(map[string][]declaration),
		facts:           facts,
		fieldsByClass:   make(map[string]map[string]string),
		globalsByName:   make(map[string][]declaration),
		methodsByClass:  make(map[string]map[string][]declaration),
		modulesByName:   make(map[string][]string),
		moduleVariables: make(map[string]map[string]string),
		typesByCallable: make(map[string]map[string]string),
	}
	repositoryID := facts.addNode("repository", snapshot.name, ".", snapshot.name, snapshot.name, "", nil, nil, "")
	directoryIDs := addDirectories(facts, repositoryID, snapshot.name, snapshot.directories)

	for index := range snapshot.sources {
		file := &snapshot.sources[index]
		fileID := facts.addNode("file", filepath.Base(filepath.FromSlash(file.path)), file.path, file.path, file.path, file.path, nil, nil, file.contentHash)
		facts.addEdge(parentDirectoryID(directoryIDs, file.path), fileID, "contains", file.path, nil, nil)
		moduleName := strings.TrimSuffix(filepath.Base(filepath.FromSlash(file.path)), filepath.Ext(file.path))
		file.moduleID = facts.addNode("module", moduleName, file.path, file.path, file.path, file.path, nil, map[string]any{"script_library": moduleName}, "")
		facts.addEdge(fileID, file.moduleID, "contains", file.path, nil, nil)
		state.modulesByName[strings.ToLower(moduleName)] = append(state.modulesByName[strings.ToLower(moduleName)], file.moduleID)
		file.invalid = !utf8.Valid(file.content) || strings.IndexByte(string(file.content), 0) >= 0
		if file.invalid {
			facts.addUnresolved(file.moduleID, "defines", "invalid UTF-8 or NUL-containing source", "unsupported-form", file.path, nil, nil)
			continue
		}
		parseLotusScript(state, file)
	}
	state.resolveUses()
	state.resolveExtends()
	state.resolveCalls()
	return facts.render(snapshot.name)
}

func addDirectories(facts *factSet, repositoryID string, repositoryName string, directories []string) map[string]string {
	ids := make(map[string]string, len(directories))
	for _, directory := range directories {
		name := filepath.Base(filepath.FromSlash(directory))
		if directory == "." {
			name = repositoryName
		}
		id := facts.addNode("directory", name, directory, directory, directory, "", nil, nil, "")
		ids[directory] = id
		if directory == "." {
			facts.addEdge(repositoryID, id, "contains", "", nil, nil)
		}
	}
	for _, directory := range directories {
		if directory == "." {
			continue
		}
		parent := filepath.ToSlash(filepath.Dir(filepath.FromSlash(directory)))
		if parent == "" {
			parent = "."
		}
		facts.addEdge(ids[parent], ids[directory], "contains", "", nil, nil)
	}
	return ids
}

func parentDirectoryID(directories map[string]string, path string) string {
	parent := filepath.ToSlash(filepath.Dir(filepath.FromSlash(path)))
	if parent == "" {
		parent = "."
	}
	return directories[parent]
}

func (state *analysisState) resolveUses() {
	sort.Slice(state.uses, func(left, right int) bool {
		if state.uses[left].ownerPath != state.uses[right].ownerPath {
			return state.uses[left].ownerPath < state.uses[right].ownerPath
		}
		return state.uses[left].span.StartLine < state.uses[right].span.StartLine
	})
	for _, evidence := range state.uses {
		if evidence.dynamic {
			state.facts.addUnresolved(evidence.importID, "imports", evidence.expression, "dynamic-target", evidence.ownerPath, evidence.span, nil)
			continue
		}
		if strings.EqualFold(evidence.keyword, "UseLSX") {
			state.facts.addUnresolved(evidence.importID, "imports", evidence.target, "external-target", evidence.ownerPath, evidence.span, map[string]any{"extension": true})
			continue
		}
		candidates := append([]string(nil), state.modulesByName[strings.ToLower(evidence.target)]...)
		sort.Strings(candidates)
		switch len(candidates) {
		case 0:
			state.facts.addUnresolved(evidence.importID, "imports", evidence.target, "external-target", evidence.ownerPath, evidence.span, nil)
		case 1:
			state.facts.addEdge(evidence.importID, candidates[0], "imports", evidence.ownerPath, evidence.span, nil)
		default:
			state.facts.addUnresolved(evidence.importID, "imports", evidence.target, "ambiguous-target", evidence.ownerPath, evidence.span, map[string]any{"candidate_count": len(candidates)})
		}
	}
}

func (state *analysisState) resolveExtends() {
	sort.Slice(state.extends, func(left, right int) bool {
		if state.extends[left].ownerPath != state.extends[right].ownerPath {
			return state.extends[left].ownerPath < state.extends[right].ownerPath
		}
		return state.extends[left].span.StartLine < state.extends[right].span.StartLine
	})
	for _, evidence := range state.extends {
		candidates := append([]declaration(nil), state.classesByName[strings.ToLower(evidence.base)]...)
		sortDeclarations(candidates)
		switch len(candidates) {
		case 0:
			state.facts.addUnresolved(evidence.classID, "extends", evidence.base, "external-target", evidence.ownerPath, evidence.span, nil)
		case 1:
			state.facts.addEdge(evidence.classID, candidates[0].id, "extends", evidence.ownerPath, evidence.span, nil)
		default:
			state.facts.addUnresolved(evidence.classID, "extends", evidence.base, "ambiguous-target", evidence.ownerPath, evidence.span, map[string]any{"candidate_count": len(candidates)})
		}
	}
}
