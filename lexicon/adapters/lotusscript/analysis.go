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
		basesByClass:      make(map[string][]string),
		callablesByName:   make(map[string][]declaration),
		classesByName:     make(map[string][]declaration),
		facts:             facts,
		fieldSymbols:      make(map[string]map[string]variableSymbol),
		fieldsByClass:     make(map[string]map[string]string),
		globalsByName:     make(map[string][]declaration),
		importsByPath:     make(map[string][]string),
		methodsByClass:    make(map[string]map[string][]declaration),
		moduleIDByPath:    make(map[string]string),
		modulePathsByName: make(map[string][]string),
		modulePublic:      make(map[string]bool),
		modulesByName:     make(map[string][]string),
		moduleSymbols:     make(map[string]map[string]variableSymbol),
		moduleVariables:   make(map[string]map[string]string),
		typesByCallable:   make(map[string]map[string]string),
		variableSymbols:   make(map[string]map[string]variableSymbol),
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
		moduleKey := strings.ToLower(moduleName)
		state.modulesByName[moduleKey] = append(state.modulesByName[moduleKey], file.moduleID)
		state.modulePathsByName[moduleKey] = append(state.modulePathsByName[moduleKey], file.path)
		state.moduleIDByPath[file.path] = file.moduleID
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
	state.resolveAccesses()
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
			paths := state.modulePathsByName[strings.ToLower(evidence.target)]
			if len(paths) == 1 {
				state.importsByPath[evidence.ownerPath] = appendUniqueString(state.importsByPath[evidence.ownerPath], paths[0])
			}
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
		candidates := state.visibleDeclarations(evidence.ownerPath, state.classesByName[strings.ToLower(evidence.base)])
		sortDeclarations(candidates)
		switch len(candidates) {
		case 0:
			state.facts.addUnresolved(evidence.classID, "extends", evidence.base, "external-target", evidence.ownerPath, evidence.span, nil)
		case 1:
			state.facts.addEdge(evidence.classID, candidates[0].id, "extends", evidence.ownerPath, evidence.span, nil)
			state.basesByClass[evidence.classID] = appendUniqueString(state.basesByClass[evidence.classID], candidates[0].id)
		default:
			state.facts.addUnresolved(evidence.classID, "extends", evidence.base, "ambiguous-target", evidence.ownerPath, evidence.span, map[string]any{"candidate_count": len(candidates)})
		}
	}
}

func appendUniqueString(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func (state *analysisState) visiblePaths(ownerPath string) map[string]struct{} {
	visible := map[string]struct{}{ownerPath: {}}
	queue := []string{ownerPath}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, imported := range state.importsByPath[current] {
			if _, seen := visible[imported]; seen {
				continue
			}
			visible[imported] = struct{}{}
			queue = append(queue, imported)
		}
	}
	return visible
}

func (state *analysisState) visibleDeclarations(ownerPath string, declarations []declaration) []declaration {
	visible := state.visiblePaths(ownerPath)
	result := make([]declaration, 0, len(declarations))
	for _, declaration := range declarations {
		if _, ok := visible[declaration.ownerPath]; !ok {
			continue
		}
		if declaration.ownerPath == ownerPath || declaration.public {
			result = append(result, declaration)
		}
	}
	return result
}
