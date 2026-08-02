package main

import (
	"bytes"
	"encoding/xml"
	"io"
	"strings"
	"unicode"
)

type mavenDependency struct {
	artifact, group, version string
	scope, optional          string
	start, end               int
	unsupported              bool
}

func parseMavenDependencies(path string, content []byte) ([]dependencyEvidence, string) {
	decoder := xml.NewDecoder(bytes.NewReader(content))
	var stack []string
	var current *mavenDependency
	dependencyDepth := 0
	field := ""
	var fieldValue strings.Builder
	var result []dependencyEvidence

	for {
		before := int(decoder.InputOffset())
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return result, "malformed pom.xml"
		}
		switch value := token.(type) {
		case xml.StartElement:
			name := value.Name.Local
			if current == nil && name == "dependency" && isDirectMavenDependencyContainer(stack) {
				current = &mavenDependency{start: before}
				dependencyDepth = len(stack) + 1
			} else if current != nil && len(stack) == dependencyDepth && isMavenDependencyField(name) {
				field = name
				fieldValue.Reset()
			} else if current != nil && len(stack) == dependencyDepth && isUnsupportedMavenDependencyField(name) {
				current.unsupported = true
			} else if current != nil && field != "" {
				current.unsupported = true
			}
			stack = append(stack, name)
		case xml.CharData:
			if current != nil && field != "" {
				fieldValue.Write([]byte(value))
			}
		case xml.EndElement:
			if current != nil && field == value.Name.Local && len(stack) == dependencyDepth+1 {
				setMavenDependencyField(current, field, strings.TrimSpace(fieldValue.String()))
				field = ""
				fieldValue.Reset()
			}
			if current != nil && value.Name.Local == "dependency" && len(stack) == dependencyDepth {
				current.end = int(decoder.InputOffset())
				result = append(result, mavenEvidence(path, content, *current))
				current = nil
				dependencyDepth = 0
				field = ""
			}
			if len(stack) != 0 {
				stack = stack[:len(stack)-1]
			}
		}
	}
	return result, ""
}

func isDirectMavenDependencyContainer(stack []string) bool {
	return len(stack) == 2 && stack[0] == "project" && stack[1] == "dependencies"
}

func isMavenDependencyField(name string) bool {
	switch name {
	case "groupId", "artifactId", "version", "scope", "optional":
		return true
	default:
		return false
	}
}

func isUnsupportedMavenDependencyField(name string) bool {
	switch name {
	case "classifier", "systemPath", "type":
		return true
	default:
		return false
	}
}

func setMavenDependencyField(dependency *mavenDependency, field, value string) {
	switch field {
	case "groupId":
		dependency.group = value
	case "artifactId":
		dependency.artifact = value
	case "version":
		dependency.version = value
	case "scope":
		dependency.scope = value
	case "optional":
		dependency.optional = value
	}
}

func mavenEvidence(path string, content []byte, dependency mavenDependency) dependencyEvidence {
	expression := strings.TrimSpace(string(content[dependency.start:dependency.end]))
	coordinate := dependency.group + ":" + dependency.artifact
	if dependency.version != "" {
		coordinate += ":" + dependency.version
	}
	optional, optionalSet := false, dependency.optional != ""
	if dependency.optional == "true" {
		optional = true
	}
	resolved := !dependency.unsupported && literalMavenValue(dependency.group, true) && literalMavenValue(dependency.artifact, true)
	for _, value := range []string{dependency.version, dependency.scope, dependency.optional} {
		if strings.Contains(value, "${") || hasMavenWhitespace(value) {
			resolved = false
		}
	}
	if dependency.scope == "import" {
		resolved = false
	}
	if optionalSet && dependency.optional != "true" && dependency.optional != "false" {
		resolved = false
	}
	return dependencyEvidence{
		artifact: dependency.artifact, configuration: dependency.scope, coordinate: coordinate,
		expression: expression, group: dependency.group, optional: optional, optionalSet: optionalSet,
		resolved: resolved, scope: dependency.scope,
		span: offsetSpan(path, content, dependency.start, dependency.end), version: dependency.version,
	}
}

func literalMavenValue(value string, required bool) bool {
	if value == "" {
		return !required
	}
	return !strings.Contains(value, "${") && !hasMavenWhitespace(value) && !strings.ContainsAny(value, "<>/\\")
}

func hasMavenWhitespace(value string) bool {
	return strings.IndexFunc(value, unicode.IsSpace) >= 0
}
