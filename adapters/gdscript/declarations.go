package main

import (
	"fmt"
	"strings"
)

type declaration struct {
	kind       string
	name       string
	nameIndex  int
	indent     int
	span       sourceSpan
	extends    string
	attributes map[string]any
	parameters []string
	static     bool
	async      bool
	nodeID     string
	key        string
	ownerKey   string
}

type parsedFile struct {
	path         string
	content      []byte
	statements   []statement
	declarations []declaration
	imports      []importReference
	calls        []callReference
	moduleID     string
	classID      string
}

type scope struct {
	indent int
	id     string
	key    string
	kind   string
}

func parseFile(path string, content []byte) (*parsedFile, error) {
	tokens, err := lex(string(content))
	if err != nil {
		return nil, err
	}
	statements := makeStatements(tokens)
	pf := &parsedFile{path: path, content: content, statements: statements}
	for _, stmt := range statements {
		decl := parseDeclaration(stmt)
		if decl != nil {
			decl.span["path"] = path
			pf.declarations = append(pf.declarations, *decl)
		}
	}
	return pf, nil
}

func processDeclarations(facts *factSet, pf *parsedFile) {
	declarationOccurrences := make(map[string]int)
	scopes := []scope{}
	for i := range pf.declarations {
		decl := &pf.declarations[i]
		for len(scopes) > 0 && decl.indent <= scopes[len(scopes)-1].indent {
			scopes = scopes[:len(scopes)-1]
		}
		parentID := pf.moduleID
		parentKey := pf.path
		if len(scopes) > 0 {
			parentID = scopes[len(scopes)-1].id
			parentKey = scopes[len(scopes)-1].key
		} else if pf.classID != "" && decl.kind != "type" && decl.indent == 0 {
			parentID = pf.classID
			parentKey = pf.path + "::type::" + classNameForID(pf, pf.classID)
		}
		if decl.kind == "type" {
			parentID = pf.moduleID
			parentKey = pf.path
		}
		if decl.kind == "extends" {
			continue
		}
		if decl.kind == "type" && decl.nodeID == "" {
			decl.nodeID = nodeID("type", pf.path+"::type::"+decl.name)
		}
		if decl.nodeID == "" {
			baseKey := parentKey + "::" + decl.kind + "::" + decl.name
			occurrence := declarationOccurrences[baseKey]
			declarationOccurrences[baseKey] = occurrence + 1
			decl.key = baseKey
			if occurrence > 0 {
				decl.key += fmt.Sprintf("#%d", occurrence+1)
			}
			decl.nodeID = nodeID(decl.kind, decl.key)
		}
		attrs := cloneMap(decl.attributes)
		if decl.kind == "function" {
			attrs["parameters"] = append([]string(nil), decl.parameters...)
			if decl.static {
				attrs["static"] = true
			}
			if decl.async {
				attrs["async"] = true
			}
		}
		if decl.extends != "" {
			attrs["extends"] = decl.extends
		}
		facts.addNode(node(decl.kind, decl.name, pf.path, qualifiedDeclaration(pf.path, parentKey, decl.name), decl.nodeID, decl.span, "", attrs))
		facts.addEdge(edge(parentID, decl.nodeID, "contains", decl.span))
		facts.addEdge(edge(parentID, decl.nodeID, "defines", decl.span))
		if decl.kind == "function" || decl.kind == "class" {
			scopes = append(scopes, scope{indent: decl.indent, id: decl.nodeID, key: decl.key, kind: decl.kind})
		}
	}
}

func classNameForID(pf *parsedFile, id string) string {
	for _, decl := range pf.declarations {
		if decl.nodeID == id {
			return decl.name
		}
	}
	return "class"
}

func qualifiedDeclaration(path, parentKey, name string) string {
	if parentKey == path {
		return path + "::" + name
	}
	return parentKey + "::" + name
}

func parseDeclaration(stmt statement) *declaration {
	if len(stmt.tokens) == 0 {
		return nil
	}
	keywordIndex := -1
	for i, tok := range stmt.tokens {
		if tok.kind == tokenIdentifier && isDeclarationKeyword(tok.text) {
			keywordIndex = i
			break
		}
	}
	if keywordIndex < 0 {
		return nil
	}
	keyword := stmt.tokens[keywordIndex].text
	decl := &declaration{indent: stmt.indent, span: spanFromTokens("", stmt.start, stmt.end), attributes: map[string]any{}}
	switch keyword {
	case "class_name", "class":
		decl.kind = "type"
	case "func":
		decl.kind = "function"
	case "signal":
		decl.kind = "signal"
	case "const":
		decl.kind = "constant"
	case "var":
		decl.kind = "variable"
	case "extends":
		decl.kind = "extends"
	}
	if keyword == "func" {
		for _, tok := range stmt.tokens[:keywordIndex] {
			if tok.text == "static" {
				decl.static = true
			}
			if tok.text == "async" {
				decl.async = true
			}
		}
	}
	nameIndex := keywordIndex + 1
	if nameIndex < len(stmt.tokens) && stmt.tokens[nameIndex].kind == tokenIdentifier {
		decl.name = stmt.tokens[nameIndex].text
		decl.nameIndex = nameIndex
	}
	if keyword == "extends" {
		decl.name = "extends"
		decl.extends = joinTokensUntil(stmt.tokens[keywordIndex+1:], ":")
		return decl
	}
	if decl.name == "" {
		return nil
	}
	if keyword == "func" {
		if open := nextToken(stmt.tokens, nameIndex+1, "("); open >= 0 {
			if close := matchingParen(stmt.tokens, open); close >= 0 {
				decl.parameters = parseParameters(stmt.tokens[open+1 : close])
			}
		}
	}
	if keyword == "class_name" || keyword == "class" {
		if ext := indexOfToken(stmt.tokens, "extends"); ext >= 0 {
			decl.extends = joinTokensUntil(stmt.tokens[ext+1:], ":")
		}
	}
	if keyword == "var" {
		if colon := indexOfTokenAfter(stmt.tokens, ":", nameIndex); colon > nameIndex {
			decl.attributes["type"] = joinTokensUntil(stmt.tokens[colon+1:], "=")
		}
	}
	return decl
}

func parseParameters(tokens []token) []string {
	var parameters []string
	start := 0
	depth := 0
	for i := 0; i <= len(tokens); i++ {
		if i == len(tokens) || (tokens[i].text == "," && depth == 0) {
			part := strings.TrimSpace(joinTokens(tokens[start:i]))
			if part != "" {
				parameters = append(parameters, part)
			}
			start = i + 1
			continue
		}
		switch tokens[i].text {
		case "(", "[", "{":
			depth++
		case ")", "]", "}":
			depth--
		}
	}
	return parameters
}

func nextToken(tokens []token, start int, text string) int {
	for i := start; i < len(tokens); i++ {
		if tokens[i].text == text {
			return i
		}
	}
	return -1
}

func indexOfToken(tokens []token, text string) int {
	for i := range tokens {
		if tokens[i].text == text {
			return i
		}
	}
	return -1
}

func indexOfTokenAfter(tokens []token, text string, start int) int {
	for i := start + 1; i < len(tokens); i++ {
		if tokens[i].text == text {
			return i
		}
	}
	return -1
}

func joinTokens(tokens []token) string {
	var result strings.Builder
	for _, tok := range tokens {
		if tok.text == "\n" {
			continue
		}
		result.WriteString(tok.text)
	}
	return result.String()
}

func joinTokensUntil(tokens []token, stop string) string {
	end := len(tokens)
	for i, tok := range tokens {
		if tok.text == stop {
			end = i
			break
		}
	}
	return joinTokens(tokens[:end])
}
