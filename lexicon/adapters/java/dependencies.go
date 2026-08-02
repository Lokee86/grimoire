package main

import (
	"path/filepath"
	"strings"
	"unicode/utf8"
)

type dependencyEvidence struct {
	category      string
	configuration string
	constraint    string
	coordinate    string
	expression    string
	optional      bool
	resolved      bool
	source        string
	span          *span
}

func analyzeDependencyManifest(state *analysisState, directories map[string]string, manifest repositoryManifest) {
	facts := state.facts
	fileID := facts.addNode(
		"file", filepath.Base(filepath.FromSlash(manifest.path)), manifest.path, manifest.path,
		manifest.path, manifest.path, nil, map[string]any{"manifest": true}, contentID(manifest.content),
	)
	facts.addEdge(parentContainer(state.repositoryID, directories, manifest.path), fileID, "contains", manifest.path, nil, nil)
	identity := "build-module:" + manifest.path
	moduleID := facts.addNode(
		"module", strings.TrimSuffix(filepath.Base(filepath.FromSlash(manifest.path)), filepath.Ext(manifest.path)),
		manifest.path, identity, identity, manifest.path, nil,
		map[string]any{"build_system": buildSystem(manifest.kind), "manifest": manifest.path}, "",
	)
	facts.addEdge(fileID, moduleID, "contains", manifest.path, nil, nil)
	if !manifest.valid {
		facts.addUnresolved(moduleID, "depends-on", "invalid UTF-8 or NUL-containing manifest", "unsupported-form", manifest.path, nil, manifestAttributes(manifest.path, "", "manifest"))
		return
	}

	var dependencies []dependencyEvidence
	switch manifest.kind {
	case "maven":
		dependencies = parseMavenDependencies(manifest.path, manifest.content)
	case "gradle-groovy", "gradle-kotlin":
		dependencies = parseGradleDependencies(manifest.path, manifest.kind, manifest.content)
	}
	for _, dependency := range dependencies {
		attributes := dependencyAttributes(manifest.path, dependency)
		if !dependency.resolved {
			facts.addUnresolved(moduleID, "depends-on", dependency.expression, "unsupported-form", manifest.path, dependency.span, attributes)
			continue
		}
		targetID := addExternalDependency(facts, dependency.coordinate)
		facts.addEdge(moduleID, targetID, "depends-on", manifest.path, dependency.span, attributes)
	}
}

func buildSystem(kind string) string {
	if kind == "maven" {
		return "maven"
	}
	return "gradle"
}

func addExternalDependency(facts *factSet, coordinate string) string {
	identity := "dependency:maven:" + coordinate
	path := "@dependencies/maven/" + strings.ReplaceAll(coordinate, ":", "/")
	return facts.addNode(
		"module", coordinate, path, identity, identity, "", nil,
		map[string]any{"dependency": true, "ecosystem": "maven"}, "",
	)
}

func dependencyAttributes(manifest string, dependency dependencyEvidence) map[string]any {
	attributes := manifestAttributes(manifest, dependency.configuration, dependency.source)
	attributes["build"] = dependency.category == "build"
	attributes["category"] = dependency.category
	attributes["constraint"] = dependency.constraint
	attributes["dev"] = dependency.category == "development"
	attributes["optional"] = dependency.optional
	attributes["path"] = false
	attributes["peer"] = false
	return attributes
}

func manifestAttributes(manifest, configuration, source string) map[string]any {
	return map[string]any{
		"configuration": configuration,
		"manifest":      manifest,
		"source":        source,
	}
}

func dependencySpan(path string, content []byte, start, end int) *span {
	startLine, startColumn := offsetPosition(content, start)
	endLine, endColumn := offsetPosition(content, end)
	return &span{
		EndColumn: endColumn, EndLine: endLine, Path: path,
		StartColumn: startColumn, StartLine: startLine,
	}
}

func offsetPosition(content []byte, offset int) (int, int) {
	if offset < 0 {
		offset = 0
	}
	if offset > len(content) {
		offset = len(content)
	}
	line, column := 1, 1
	for index := 0; index < offset; {
		value, size := utf8.DecodeRune(content[index:offset])
		if size == 0 {
			break
		}
		index += size
		if value == '\n' {
			line, column = line+1, 1
		} else {
			column++
		}
	}
	return line, column
}
