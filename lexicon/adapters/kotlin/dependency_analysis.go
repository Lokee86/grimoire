package main

import "strings"

type dependencyEvidence struct {
	artifact      string
	configuration string
	coordinate    string
	expression    string
	group         string
	optional      bool
	optionalSet   bool
	resolved      bool
	scope         string
	span          sourceSpan
	version       string
}

func (state *analysis) emitManifestFacts(manifests []manifestFile) {
	buildModules := make(map[string]string)
	for _, manifest := range manifests {
		span := wholeFileSpan(manifest.path, manifest.content)
		fileID := state.facts.addFileNode(manifest.path, manifest.content, &span)
		state.facts.addEdge(state.parentDirectoryID(manifest.path), fileID, "contains", manifest.path, nil, nil)
		sourceID := state.manifestOwner(manifest, fileID, buildModules)

		var evidence []dependencyEvidence
		var parseExpression string
		switch manifest.format {
		case "gradle-kotlin", "gradle-groovy":
			evidence = parseGradleDependencies(manifest.path, manifest.content)
		case "maven":
			evidence, parseExpression = parseMavenDependencies(manifest.path, manifest.content)
		}
		if parseExpression != "" {
			attributes := manifestAttributes(manifest, "")
			state.facts.addUnresolved(sourceID, "depends-on", parseExpression, "unsupported-form", manifest.path, &span, attributes)
		}
		for _, dependency := range evidence {
			attributes := manifestAttributes(manifest, dependency.configuration)
			attributes["expression"] = dependency.expression
			if manifest.format == "maven" {
				attributes["optional"] = dependency.optional
				attributes["scope"] = dependency.scope
			}
			if !dependency.resolved {
				state.facts.addUnresolved(sourceID, "depends-on", dependency.expression, "unsupported-form", manifest.path, &dependency.span, attributes)
				continue
			}
			attributes["coordinate"] = dependency.coordinate
			targetID := state.externalModule(dependency)
			state.facts.addEdge(sourceID, targetID, "depends-on", manifest.path, &dependency.span, attributes)
		}
	}
}

func (state *analysis) manifestOwner(manifest manifestFile, fileID string, modules map[string]string) string {
	directory := pathDirectory(manifest.path)
	if directory == "." || directory == "" {
		return state.repositoryID
	}
	if id := modules[directory]; id != "" {
		state.facts.addEdge(fileID, id, "defines", manifest.path, nil, map[string]any{"evidence": "dependency-manifest"})
		return id
	}
	canonical := "build-module:" + directory
	attributes := map[string]any{"build_module": true}
	id := state.facts.addNode("module", canonical, baseName(directory), directory, state.repositoryName+"::"+canonical, manifest.path, nil, attributes)
	modules[directory] = id
	state.facts.addEdge(fileID, id, "defines", manifest.path, nil, map[string]any{"evidence": "dependency-manifest"})
	return id
}

func manifestAttributes(manifest manifestFile, configuration string) map[string]any {
	return map[string]any{
		"configuration": configuration,
		"evidence":      "dependency-manifest",
		"manifest":      manifest.path,
		"manifest_type": manifest.format,
	}
}

func (state *analysis) externalModule(dependency dependencyEvidence) string {
	attributes := map[string]any{
		"artifact":   dependency.artifact,
		"dependency": true,
		"ecosystem":  "maven",
		"external":   true,
		"group":      dependency.group,
	}
	if dependency.version != "" {
		attributes["version"] = dependency.version
	}
	path := "@dependencies/maven/" + strings.ReplaceAll(dependency.coordinate, ":", "/")
	return state.facts.addNode(
		"module", "dependency:maven:"+dependency.coordinate, dependency.artifact, path,
		dependency.coordinate, "", nil, attributes,
	)
}
