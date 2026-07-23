package main

import (
	"regexp"
	"strings"
)

var (
	typeDeclaration = regexp.MustCompile(`(?i)^\s*((public|private|protected|internal|static|final|abstract|sealed|open|export|extern|unsafe|pub)\s+)*(class|struct|interface|trait|enum|record|protocol|union|namespace|module|type)\s+([A-Za-z_][A-Za-z0-9_]*)`)
	keywordFunction = regexp.MustCompile(`(?i)^\s*((public|private|protected|internal|static|final|abstract|open|export|extern|async|unsafe|pub)\s+)*(func|function|fn|def|sub|proc|procedure|fun)\s+([A-Za-z_][A-Za-z0-9_]*)\s*[<(]`)
	cStyleFunction  = regexp.MustCompile(`^\s*((public|private|protected|internal|static|final|abstract|virtual|override|extern|async|unsafe|inline|constexpr)\s+)*([A-Za-z_][A-Za-z0-9_:<>,\[\]*&?]*\s+)+([A-Za-z_][A-Za-z0-9_]*)\s*\([^;{}]*\)\s*(\{|=>)\s*$`)
	includeImport   = regexp.MustCompile(`^\s*#\s*include\s*[<"]([^>"]+)[>"]`)
	fromImport      = regexp.MustCompile(`(?i)^\s*from\s+([A-Za-z0-9_./:\-]+)\s+import\b`)
	requireImport   = regexp.MustCompile(`(?i)^\s*require\s*\(\s*["']([^"']+)["']\s*\)`)
	usingNamespace  = regexp.MustCompile(`(?i)^\s*using\s+namespace\s+([A-Za-z0-9_:.\-]+)`)
	keywordImport   = regexp.MustCompile(`(?i)^\s*(import|use|using|require|include)\s+["']?([A-Za-z0-9_./:\-]+)`)
)

var cStyleControls = map[string]struct{}{
	"if": {}, "for": {}, "while": {}, "switch": {}, "catch": {}, "return": {}, "sizeof": {},
}

func parseSource(facts *factSet, path, moduleID, content string) {
	for index, line := range strings.Split(content, "\n") {
		lineNumber := index + 1
		if keyword, target, ok := parseImport(line); ok {
			facts.addImport(path, moduleID, keyword, target, lineNumber, line)
		}
		if kind, name, ok := parseDeclaration(line); ok {
			facts.addDeclaration(path, moduleID, kind, name, lineNumber, line)
		}
	}
}

func parseImport(line string) (string, string, bool) {
	if match := includeImport.FindStringSubmatch(line); match != nil {
		return "include", match[1], true
	}
	if match := fromImport.FindStringSubmatch(line); match != nil {
		return "from", match[1], true
	}
	if match := requireImport.FindStringSubmatch(line); match != nil {
		return "require", match[1], true
	}
	if match := usingNamespace.FindStringSubmatch(line); match != nil {
		return "using", match[1], true
	}
	if match := keywordImport.FindStringSubmatch(line); match != nil {
		return strings.ToLower(match[1]), strings.TrimRight(match[2], ";,"), true
	}
	return "", "", false
}

func parseDeclaration(line string) (string, string, bool) {
	if match := typeDeclaration.FindStringSubmatch(line); match != nil {
		return declarationKind(strings.ToLower(match[3])), match[4], true
	}
	if match := keywordFunction.FindStringSubmatch(line); match != nil {
		return "function", match[4], true
	}
	trimmed := strings.TrimSpace(line)
	first := trimmed
	if index := strings.IndexAny(first, " (\t"); index >= 0 {
		first = first[:index]
	}
	if _, control := cStyleControls[strings.ToLower(first)]; control {
		return "", "", false
	}
	if match := cStyleFunction.FindStringSubmatch(line); match != nil {
		return "function", match[4], true
	}
	return "", "", false
}

func declarationKind(keyword string) string {
	switch keyword {
	case "interface", "protocol":
		return "interface"
	case "trait":
		return "trait"
	case "namespace":
		return "namespace"
	case "module":
		return "module"
	default:
		return "type"
	}
}
