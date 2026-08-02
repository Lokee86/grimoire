package main

import (
	"path/filepath"
	"sort"
	"strings"
)

type importEvidence struct {
	expression string
	id         string
	owner      string
	span       *span
	static     bool
	target     string
	wildcard   bool
}

type analysisState struct {
	callables                     []callableDeclaration
	callablesByOwnerKindName      map[callableLookupKey][]callableDeclaration
	callablesByOwnerNameSignature map[callableIndexKey][]callableDeclaration
	declarations                  map[string][]string
	parentIDs                     map[string][]string
	superclassIDs                 map[string][]string
	facts                         *factSet
	fields                        map[string]map[string][]fieldDeclaration
	imports                       []importEvidence
	namespaces                    map[string]string
	relationships                 []relationshipEvidence
	repositoryID                  string
	typeKindsByID                 map[string]string
	types                         map[string][]typeDeclaration
}

func analyzeRepository(repository string) ([]byte, error) {
	snapshot, err := discoverRepository(repository)
	if err != nil {
		return nil, err
	}
	facts := newFactSet()
	repositoryID := facts.addNode("repository", snapshot.name, ".", snapshot.name, snapshot.name, "", nil, nil, "")
	directoryIDs := make(map[string]string, len(snapshot.directories))
	for _, path := range snapshot.directories {
		id := facts.addNode("directory", filepath.Base(filepath.FromSlash(path)), path, path, path, "", nil, nil, "")
		directoryIDs[path] = id
		facts.addEdge(parentContainer(repositoryID, directoryIDs, path), id, "contains", "", nil, nil)
	}

	state := &analysisState{
		declarations: make(map[string][]string), facts: facts, namespaces: make(map[string]string),
		fields: make(map[string]map[string][]fieldDeclaration), repositoryID: repositoryID,
		typeKindsByID: make(map[string]string), types: make(map[string][]typeDeclaration),
	}
	for _, manifest := range snapshot.manifests {
		analyzeDependencyManifest(state, directoryIDs, manifest)
	}
	for _, source := range snapshot.sources {
		fileID := facts.addNode("file", filepath.Base(filepath.FromSlash(source.path)), source.path, source.path, source.path, source.path, nil, nil, contentID(source.content))
		facts.addEdge(parentContainer(repositoryID, directoryIDs, source.path), fileID, "contains", source.path, nil, nil)
		if !source.valid {
			moduleID := state.addModule(fileID, source.path, "")
			facts.addUnresolved(moduleID, "defines", "invalid UTF-8 or NUL-containing source", "unsupported-form", source.path, nil, nil)
			continue
		}
		parseJavaSource(state, fileID, source.path, string(source.content))
	}
	state.resolveImports()
	state.resolveRelationships()
	state.resolveRuntimeSemantics()
	if err := state.resolveCompilerSemantics(snapshot); err != nil {
		return nil, err
	}
	return facts.render(snapshot.name)
}

func parentContainer(repositoryID string, directories map[string]string, path string) string {
	parent := parentPath(path)
	if parent == "" {
		return repositoryID
	}
	return directories[parent]
}

func (state *analysisState) addModule(fileID, path, packageName string) string {
	qualifiedName := path
	if packageName != "" {
		qualifiedName = packageName + "::" + path
	}
	attributes := map[string]any{"package": packageName}
	moduleID := state.facts.addNode("module", strings.TrimSuffix(filepath.Base(filepath.FromSlash(path)), filepath.Ext(path)), path, qualifiedName, path, path, nil, attributes, "")
	state.facts.addEdge(fileID, moduleID, "contains", path, nil, nil)
	return moduleID
}

func (state *analysisState) ensureNamespace(packageName string) string {
	qualifiedName := packageName
	name := packageName
	path := "."
	if packageName == "" {
		qualifiedName = "<default>"
		name = "<default>"
	} else {
		path = strings.ReplaceAll(packageName, ".", "/")
		if index := strings.LastIndex(packageName, "."); index >= 0 {
			name = packageName[index+1:]
		}
	}
	if known := state.namespaces[qualifiedName]; known != "" {
		return known
	}
	id := state.facts.addNode("namespace", name, path, qualifiedName, qualifiedName, "", nil, map[string]any{"package": packageName}, "")
	state.namespaces[qualifiedName] = id
	state.declarations[qualifiedName] = appendUnique(state.declarations[qualifiedName], id)
	state.facts.addEdge(state.repositoryID, id, "contains", "", nil, nil)
	return id
}

func (state *analysisState) registerDeclaration(qualifiedName, id string) {
	state.declarations[qualifiedName] = appendUnique(state.declarations[qualifiedName], id)
}

func (state *analysisState) registerType(qualifiedName, id, declarationKind string) {
	state.registerDeclaration(qualifiedName, id)
	state.types[qualifiedName] = append(state.types[qualifiedName], typeDeclaration{id: id, declarationKind: declarationKind})
	if state.typeKindsByID == nil {
		state.typeKindsByID = make(map[string]string)
	}
	state.typeKindsByID[id] = declarationKind
}

func appendUnique(items []string, item string) []string {
	for _, existing := range items {
		if existing == item {
			return items
		}
	}
	return append(items, item)
}

func (state *analysisState) resolveImports() {
	sort.Slice(state.imports, func(left, right int) bool {
		if state.imports[left].owner != state.imports[right].owner {
			return state.imports[left].owner < state.imports[right].owner
		}
		return state.imports[left].span.StartLine < state.imports[right].span.StartLine
	})
	for _, evidence := range state.imports {
		targetName := evidence.target
		if evidence.wildcard {
			if evidence.static {
				targetName = evidence.target
			} else {
				targetName = evidence.target
			}
		}
		candidates := append([]string(nil), state.declarations[targetName]...)
		if evidence.static && !evidence.wildcard && len(candidates) == 0 {
			if split := strings.LastIndex(targetName, "."); split > 0 {
				candidates = append(candidates, state.declarations[targetName[:split]]...)
			}
		}
		sort.Strings(candidates)
		switch len(candidates) {
		case 0:
			state.facts.addUnresolved(evidence.id, "imports", evidence.expression, "external-target", evidence.owner, evidence.span, map[string]any{"static": evidence.static, "wildcard": evidence.wildcard})
		case 1:
			state.facts.addEdge(evidence.id, candidates[0], "imports", evidence.owner, evidence.span, map[string]any{"static": evidence.static, "wildcard": evidence.wildcard})
		default:
			state.facts.addUnresolved(evidence.id, "imports", evidence.expression, "ambiguous-target", evidence.owner, evidence.span, map[string]any{"candidate_count": len(candidates), "static": evidence.static, "wildcard": evidence.wildcard})
		}
	}
}

func sourceExcerpt(source string, tokens []token, start, end int) string {
	if start < 0 || start >= len(tokens) || end <= start {
		return "unrecognized Java syntax"
	}
	if end > len(tokens) {
		end = len(tokens)
	}
	text := strings.TrimSpace(source[tokens[start].offset:tokens[end-1].endOffset])
	const maximum = 240
	runes := []rune(text)
	if len(runes) > maximum {
		text = string(runes[:maximum]) + "..."
	}
	return text
}
