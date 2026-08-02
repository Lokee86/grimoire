package main

import (
	"bytes"
	"encoding/xml"
	"io"
	"strings"
)

type mavenDependency struct {
	end        int
	fields     map[string]string
	invalid    bool
	seenFields map[string]bool
	start      int
}

func parseMavenDependencies(path string, content []byte) []dependencyEvidence {
	decoder := xml.NewDecoder(bytes.NewReader(content))
	var dependencies []dependencyEvidence
	var stack []string
	var current *mavenDependency
	activeField := ""
	fieldDepth := 0
	for {
		tokenValue, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			dependencies = append(dependencies, dependencyEvidence{
				category: "runtime", expression: "malformed Maven XML: " + err.Error(),
				resolved: false, source: "pom.xml:dependencies",
			})
			break
		}
		offset := int(decoder.InputOffset())
		switch token := tokenValue.(type) {
		case xml.StartElement:
			stack = append(stack, token.Name.Local)
			if current == nil && isMavenDependencyPath(stack) {
				current = &mavenDependency{
					fields: make(map[string]string), seenFields: make(map[string]bool),
					start: bytes.LastIndex(content[:offset], []byte("<")),
				}
				continue
			}
			if current == nil {
				continue
			}
			depth := len(stack)
			if depth == 4 && isMavenField(token.Name.Local) {
				activeField, fieldDepth = token.Name.Local, depth
				if current.seenFields[activeField] {
					current.invalid = true
				}
				current.seenFields[activeField] = true
			} else if activeField != "" && depth > fieldDepth {
				current.invalid = true
			}
		case xml.CharData:
			if current != nil && activeField != "" && len(stack) == fieldDepth {
				current.fields[activeField] += string(token)
			}
		case xml.EndElement:
			depth := len(stack)
			if current != nil && activeField == token.Name.Local && depth == fieldDepth {
				current.fields[activeField] = strings.TrimSpace(current.fields[activeField])
				activeField = ""
			}
			if current != nil && token.Name.Local == "dependency" && depth == 3 {
				current.end = offset
				dependencies = append(dependencies, mavenEvidence(path, content, current))
				current = nil
				activeField = ""
			}
			if depth != 0 {
				stack = stack[:depth-1]
			}
		}
	}
	return dependencies
}

func isMavenDependencyPath(stack []string) bool {
	return len(stack) == 3 && stack[0] == "project" && stack[1] == "dependencies" && stack[2] == "dependency"
}

func isMavenField(name string) bool {
	switch name {
	case "groupId", "artifactId", "version", "scope", "optional":
		return true
	default:
		return false
	}
}

func mavenEvidence(path string, content []byte, dependency *mavenDependency) dependencyEvidence {
	group := dependency.fields["groupId"]
	artifact := dependency.fields["artifactId"]
	version := dependency.fields["version"]
	scope := dependency.fields["scope"]
	optionalText := dependency.fields["optional"]
	if scope == "" {
		scope = "compile"
	}
	optional := optionalText == "true"
	literal := group != "" && artifact != "" && !strings.Contains(group, ":") && !strings.Contains(artifact, ":")
	for _, value := range []string{group, artifact, version, scope, optionalText} {
		literal = literal && !strings.Contains(value, "${")
	}
	if optionalText != "" && optionalText != "true" && optionalText != "false" {
		literal = false
	}
	category := mavenCategory(scope, optional)
	return dependencyEvidence{
		category: category, configuration: scope, constraint: version,
		coordinate: group + ":" + artifact, expression: manifestExpression(content, dependency.start, dependency.end),
		optional: optional, resolved: literal && !dependency.invalid, source: "pom.xml:dependencies",
		span: dependencySpan(path, content, dependency.start, dependency.end),
	}
}

func mavenCategory(scope string, optional bool) string {
	if optional {
		return "optional"
	}
	switch scope {
	case "test":
		return "test"
	case "provided", "system", "import":
		return "build"
	default:
		return "runtime"
	}
}

func manifestExpression(content []byte, start, end int) string {
	if start < 0 || end <= start || end > len(content) {
		return "unrecognized dependency declaration"
	}
	expression := strings.TrimSpace(string(content[start:end]))
	runes := []rune(expression)
	if len(runes) > 240 {
		expression = string(runes[:240]) + "..."
	}
	return expression
}
