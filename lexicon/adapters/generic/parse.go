package main

import (
	"regexp"
	"strings"
)

var (
	typeDeclaration    = regexp.MustCompile(`(?i)^\s*((public|private|protected|internal|static|final|abstract|sealed|open|export|extern|unsafe|pub|data|value|annotation)\s+)*(class|struct|interface|trait|enum|record|protocol|union|namespace|module|type|object)\s+([A-Za-z_][A-Za-z0-9_]*)`)
	keywordFunction    = regexp.MustCompile(`(?i)^\s*((public|private|protected|internal|static|final|abstract|open|export|extern|async|unsafe|pub)\s+)*(func|function|fn|def|sub|proc|procedure|fun)\s+([A-Za-z_][A-Za-z0-9_]*)\s*[<(]`)
	cStyleFunction     = regexp.MustCompile(`^\s*((public|private|protected|internal|static|final|abstract|virtual|override|extern|async|unsafe|inline|constexpr|synchronized|native)\s+)*([A-Za-z_][A-Za-z0-9_:<>,\[\]*&?.]*\s+)+([~A-Za-z_][A-Za-z0-9_:]*)\s*\([^;{}]*\)\s*(const\b\s*)?(noexcept\b\s*)?(->\s*[^;{]+\s*)?(\{.*|=>.*)?$`)
	preprocessorImport = regexp.MustCompile(`^\s*#\s*(include|import)\s*[<"]([^>"]+)[>"]`)
	fromImport         = regexp.MustCompile(`(?i)^\s*from\s+([A-Za-z0-9_./:\-]+)\s+import\b`)
	assignedRequire    = regexp.MustCompile(`(?i)^\s*(local\s+)?[A-Za-z_][A-Za-z0-9_]*\s*=\s*require\s*\(\s*["']([^"']+)["']`)
	requireImport      = regexp.MustCompile(`(?i)^\s*(require|require_once|include|include_once)\s*\(?\s*["']([^"']+)["']`)
	usingNamespace     = regexp.MustCompile(`(?i)^\s*using\s+namespace\s+([A-Za-z0-9_:.\-]+)`)
	keywordImport      = regexp.MustCompile(`(?i)^\s*(import|use|using|require|include)\s+["']?([A-Za-z0-9_./:\\\-]+)`)
	protoImport        = regexp.MustCompile(`(?i)^\s*import\s+(public\s+|weak\s+)?["']([^"']+)["']`)
	powerShellImport   = regexp.MustCompile(`(?i)^\s*Import-Module\s+["']?([A-Za-z0-9_./:\\\-]+)`)

	luaFunction        = regexp.MustCompile(`(?i)^\s*(local\s+)?function\s+([A-Za-z_][A-Za-z0-9_.:]*)\s*\(`)
	shellFunction      = regexp.MustCompile(`(?i)^\s*(function\s+)?([A-Za-z_][A-Za-z0-9_-]*)\s*(\(\s*\))?\s*\{`)
	powerShellFunction = regexp.MustCompile(`(?i)^\s*function\s+([A-Za-z_][A-Za-z0-9_-]*)\b`)
	objectiveCType     = regexp.MustCompile(`(?i)^\s*@(interface|protocol|implementation)\s+([A-Za-z_][A-Za-z0-9_]*)`)
	objectiveCMethod   = regexp.MustCompile(`^\s*[-+]\s*\([^)]*\)\s*([A-Za-z_][A-Za-z0-9_]*)`)
	protoType          = regexp.MustCompile(`(?i)^\s*(message|enum|service)\s+([A-Za-z_][A-Za-z0-9_]*)`)
	protoRPC           = regexp.MustCompile(`(?i)^\s*rpc\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(`)
	solidityType       = regexp.MustCompile(`(?i)^\s*(abstract\s+)?(contract|interface|library|struct|enum)\s+([A-Za-z_][A-Za-z0-9_]*)`)
	sqlDeclaration     = regexp.MustCompile(`(?i)^\s*create\s+(or\s+replace\s+)?(table|view|type|procedure|function|trigger)\s+([A-Za-z_][A-Za-z0-9_$.]*)`)
	perlFunction       = regexp.MustCompile(`(?i)^\s*sub\s+([A-Za-z_][A-Za-z0-9_]*)\b`)
	rFunction          = regexp.MustCompile(`^\s*([A-Za-z.][A-Za-z0-9._]*)\s*(<-|=)\s*function\s*\(`)
)

var cStyleControls = map[string]struct{}{
	"if": {}, "for": {}, "while": {}, "switch": {}, "catch": {}, "return": {}, "sizeof": {},
}

func parseSource(facts *factSet, path, moduleID, content string) {
	language := strings.TrimPrefix(facts.language, "generic-")
	imports := strings.Split(maskComments(content, language), "\n")
	declarations := strings.Split(maskStrings(strings.Join(imports, "\n")), "\n")
	original := strings.Split(content, "\n")
	for index, line := range original {
		lineNumber := index + 1
		if keyword, target, ok := parseImport(language, imports[index]); ok {
			facts.addImport(path, moduleID, keyword, target, lineNumber, line)
		}
		if kind, name, ok := parseDeclaration(language, declarations[index]); ok {
			facts.addDeclaration(path, moduleID, kind, name, lineNumber, line)
		}
	}
}

func parseImport(language, line string) (string, string, bool) {
	if match := preprocessorImport.FindStringSubmatch(line); match != nil {
		return strings.ToLower(match[1]), match[2], true
	}
	if match := fromImport.FindStringSubmatch(line); match != nil {
		return "from", match[1], true
	}
	if match := assignedRequire.FindStringSubmatch(line); match != nil {
		return "require", match[2], true
	}
	if match := requireImport.FindStringSubmatch(line); match != nil {
		return strings.ToLower(match[1]), match[2], true
	}
	if match := usingNamespace.FindStringSubmatch(line); match != nil {
		return "using", match[1], true
	}
	if language == "proto" {
		if match := protoImport.FindStringSubmatch(line); match != nil {
			return "import", match[2], true
		}
	}
	if language == "ps1" {
		if match := powerShellImport.FindStringSubmatch(line); match != nil {
			return "import-module", match[1], true
		}
	}
	if match := keywordImport.FindStringSubmatch(line); match != nil {
		return strings.ToLower(match[1]), strings.TrimRight(match[2], ";,"), true
	}
	return "", "", false
}

func parseDeclaration(language, line string) (string, string, bool) {
	switch language {
	case "lua":
		if match := luaFunction.FindStringSubmatch(line); match != nil {
			return "function", terminalName(match[2], ".:"), true
		}
	case "sh", "bash", "fish":
		if match := shellFunction.FindStringSubmatch(line); match != nil {
			return "function", match[2], true
		}
	case "ps1":
		if match := powerShellFunction.FindStringSubmatch(line); match != nil {
			return "function", match[1], true
		}
	case "m", "mm":
		if match := objectiveCType.FindStringSubmatch(line); match != nil {
			kind := "type"
			if strings.EqualFold(match[1], "protocol") {
				kind = "interface"
			}
			return kind, match[2], true
		}
		if match := objectiveCMethod.FindStringSubmatch(line); match != nil {
			return "function", match[1], true
		}
	case "proto":
		if match := protoType.FindStringSubmatch(line); match != nil {
			kind := "type"
			if strings.EqualFold(match[1], "service") {
				kind = "interface"
			}
			return kind, match[2], true
		}
		if match := protoRPC.FindStringSubmatch(line); match != nil {
			return "function", match[1], true
		}
	case "sol":
		if match := solidityType.FindStringSubmatch(line); match != nil {
			kind := "type"
			if strings.EqualFold(match[2], "interface") {
				kind = "interface"
			}
			return kind, match[3], true
		}
	case "sql":
		if match := sqlDeclaration.FindStringSubmatch(line); match != nil {
			keyword := strings.ToLower(match[2])
			if keyword == "procedure" || keyword == "function" || keyword == "trigger" {
				return "function", match[3], true
			}
			return "type", match[3], true
		}
	case "pl", "pm":
		if match := perlFunction.FindStringSubmatch(line); match != nil {
			return "function", match[1], true
		}
	case "r":
		if match := rFunction.FindStringSubmatch(line); match != nil {
			return "function", match[1], true
		}
	}

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
		return "function", terminalName(match[4], ":"), true
	}
	return "", "", false
}

func terminalName(name, separators string) string {
	if index := strings.LastIndexAny(name, separators); index >= 0 {
		return name[index+1:]
	}
	return name
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
