package main

import (
	"path/filepath"
	"sort"
	"strings"
)

type analysis struct {
	facts            *factSet
	moduleByPath     map[string]string
	namespaceByQN    map[string]string
	pendingRelations []pendingRelationship
	repositoryID     string
	repositoryName   string
	relationshipByQN map[string][]relationshipTarget
	runtime          *runtimeIndex
}

func analyzeRepository(repository string) ([]byte, error) {
	_, repositoryName, sources, manifests, err := discoverRepository(repository)
	if err != nil {
		return nil, err
	}
	facts := newFactSet()
	state := &analysis{
		facts: facts, moduleByPath: make(map[string]string), namespaceByQN: make(map[string]string),
		relationshipByQN: make(map[string][]relationshipTarget), repositoryName: repositoryName,
		runtime: newRuntimeIndex(),
	}
	state.repositoryID = facts.addNode(
		"repository", repositoryName, repositoryName, ".", repositoryName, "", nil,
		map[string]any{
			"analysis_mode": "structural", "dependency_manifest_count": len(manifests),
			"source_file_count": len(sources),
		},
	)
	paths := make([]string, 0, len(sources)+len(manifests))
	for _, source := range sources {
		paths = append(paths, source.path)
	}
	for _, manifest := range manifests {
		paths = append(paths, manifest.path)
	}
	state.emitDirectories(paths)
	state.emitManifestFacts(manifests)

	parsed := make([]*parsedKotlinFile, 0, len(sources))
	for _, source := range sources {
		file := parseKotlinFile(source.path, source.content)
		if err := validateParsedFile(file); err != nil {
			return nil, err
		}
		parsed = append(parsed, file)
		fileSpan := wholeFileSpan(source.path, source.content)
		fileID := facts.addFileNode(source.path, source.content, &fileSpan)
		parentID := state.parentDirectoryID(source.path)
		facts.addEdge(parentID, fileID, "contains", source.path, nil, nil)

		packageName := file.packageName
		if packageName == "" {
			packageName = "<default>"
		}
		moduleQN := sourceModuleQualifiedName(packageName, source.path)
		moduleAttributes := map[string]any{"package": packageName, "script": strings.HasSuffix(strings.ToLower(source.path), ".kts")}
		moduleID := facts.addNode("module", "source:"+source.path, strings.TrimSuffix(baseName(source.path), filepath.Ext(source.path)), source.path, moduleQN, source.path, &fileSpan, moduleAttributes)
		state.moduleByPath[source.path] = moduleID
		facts.addEdge(fileID, moduleID, "contains", source.path, nil, nil)

		namespaceID := state.ensureNamespace(packageName)
		packageSpan := file.packageSpan
		facts.addEdge(moduleID, namespaceID, "defines", source.path, packageSpan, map[string]any{"evidence": "package-directive"})
	}

	for _, file := range parsed {
		state.emitImports(file)
		ownerID := state.moduleByPath[file.path]
		packageName := file.packageName
		if packageName == "" {
			packageName = "<default>"
		}
		occurrences := make(map[string]int)
		for _, declaration := range file.declarations {
			state.emitDeclaration(file, declaration, ownerID, packageName, "source:"+file.path, "module", occurrences)
		}
		for _, diagnostic := range file.diagnostics {
			span := sourceSpan{
				EndColumn: diagnostic.token.endColumn, EndLine: diagnostic.token.endLine, Path: file.path,
				StartColumn: diagnostic.token.startColumn, StartLine: diagnostic.token.startLine,
			}
			facts.addUnresolved(
				ownerID, "defines", diagnosticExpression(file.content, diagnostic), "unsupported-form", file.path, &span,
				diagnosticAttributes(diagnostic),
			)
		}
	}
	state.emitRelationships()
	state.emitRuntimeSemantics()
	return facts.render(repositoryName)
}

func (state *analysis) emitDirectories(paths []string) {
	directories := map[string]struct{}{}
	for _, path := range paths {
		directory := pathDirectory(path)
		for directory != "." && directory != "" {
			directories[directory] = struct{}{}
			directory = pathDirectory(directory)
		}
	}
	ordered := make([]string, 0, len(directories))
	for directory := range directories {
		ordered = append(ordered, directory)
	}
	sort.Slice(ordered, func(i, j int) bool {
		leftDepth := strings.Count(ordered[i], "/")
		rightDepth := strings.Count(ordered[j], "/")
		if leftDepth != rightDepth {
			return leftDepth < rightDepth
		}
		return ordered[i] < ordered[j]
	})
	for _, directory := range ordered {
		id := state.facts.addNode("directory", directory, baseName(directory), directory, directory, "", nil, nil)
		parent := pathDirectory(directory)
		parentID := state.repositoryID
		if parent != "." && parent != "" {
			parentID = stableID("directory", parent)
		}
		state.facts.addEdge(parentID, id, "contains", "", nil, nil)
	}
}

func (state *analysis) parentDirectoryID(path string) string {
	directory := pathDirectory(path)
	if directory == "." || directory == "" {
		return state.repositoryID
	}
	return stableID("directory", directory)
}

func (state *analysis) ensureNamespace(qualifiedName string) string {
	if existing := state.namespaceByQN[qualifiedName]; existing != "" {
		return existing
	}
	name := qualifiedName
	if index := strings.LastIndexByte(name, '.'); index >= 0 {
		name = name[index+1:]
	}
	id := state.facts.addNode(
		"namespace", "package:"+qualifiedName, name, ".", qualifiedName, "", nil,
		map[string]any{"language_construct": "package"},
	)
	state.namespaceByQN[qualifiedName] = id
	state.facts.addEdge(state.repositoryID, id, "contains", "", nil, nil)
	return id
}
