package main

import (
	"fmt"
	"sort"
	"strings"
)

func parseFile(path string, content []byte) (*parsedFile, error) {
	tokens, err := lex(string(content))
	if err != nil {
		return nil, err
	}
	statements := makeStatements(tokens)
	pf := &parsedFile{path: path, content: content, statements: statements}
	for _, stmt := range statements {
		if decl := parseDeclaration(stmt); decl != nil {
			decl.span["path"] = path
			pf.declarations = append(pf.declarations, *decl)
		}
	}
	pf.declarations = append(pf.declarations, parseAnonymousFunctions(path, tokens)...)
	sort.SliceStable(pf.declarations, func(i, j int) bool {
		left, right := pf.declarations[i].span, pf.declarations[j].span
		if spanInt(left, "start_line") != spanInt(right, "start_line") {
			return spanInt(left, "start_line") < spanInt(right, "start_line")
		}
		return spanInt(left, "start_column") < spanInt(right, "start_column")
	})
	return pf, nil
}

func parseAnonymousFunctions(path string, tokens []token) []declaration {
	var declarations []declaration
	for index, tok := range tokens {
		if tok.text != "func" || index+1 >= len(tokens) || tokens[index+1].text != "(" {
			continue
		}
		close := matchingParen(tokens, index+1)
		if close < 0 {
			continue
		}
		end := close
		if arrow := indexOfTokenAfter(tokens, "->", close); arrow > close {
			if colon := indexOfTokenAfter(tokens, ":", arrow); colon > arrow {
				end = colon
			}
		} else if colon := indexOfTokenAfter(tokens, ":", close); colon > close {
			end = colon
		}
		decl := declaration{
			keyword:    "lambda",
			kind:       "function",
			name:       fmt.Sprintf("<lambda@%d:%d>", tok.line, tok.column),
			nameIndex:  index,
			indent:     tok.column - 1,
			span:       spanFromTokens(path, tok, tokens[end]),
			attributes: map[string]any{},
		}
		parseFunctionSignature(&decl, tokens[index:end+1], 0)
		declarations = append(declarations, decl)
	}
	return declarations
}

func parseDeclaration(stmt statement) *declaration {
	keywordIndex := declarationKeywordIndex(stmt.tokens)
	if keywordIndex < 0 {
		return nil
	}
	keyword := stmt.tokens[keywordIndex].text
	decl := &declaration{keyword: keyword, indent: stmt.indent, span: spanFromTokens("", stmt.start, stmt.end), attributes: map[string]any{}}
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
	for _, tok := range stmt.tokens[:keywordIndex] {
		decl.static = decl.static || tok.text == "static"
		decl.async = decl.async || tok.text == "async"
	}
	nameIndex := keywordIndex + 1
	if nameIndex < len(stmt.tokens) && stmt.tokens[nameIndex].kind == tokenIdentifier {
		decl.name, decl.nameIndex = stmt.tokens[nameIndex].text, nameIndex
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
		parseFunctionSignature(decl, stmt.tokens, nameIndex)
	}
	if keyword == "class_name" || keyword == "class" {
		if ext := indexOfToken(stmt.tokens, "extends"); ext >= 0 {
			decl.extends = joinTokensUntil(stmt.tokens[ext+1:], ":")
		}
	}
	if keyword == "var" || keyword == "const" {
		parseValueDeclaration(decl, stmt.tokens, nameIndex)
	}
	return decl
}

func declarationKeywordIndex(tokens []token) int {
	for i, tok := range tokens {
		if tok.kind == tokenIdentifier && isDeclarationKeyword(tok.text) {
			return i
		}
	}
	return -1
}

func parseFunctionSignature(decl *declaration, tokens []token, nameIndex int) {
	open := nextToken(tokens, nameIndex+1, "(")
	if open < 0 {
		return
	}
	close := matchingParen(tokens, open)
	if close < 0 {
		return
	}
	parameterTokens := splitArguments(tokens[open+1 : close])
	decl.parameters = make([]string, 0, len(parameterTokens))
	for _, parameter := range parameterTokens {
		decl.parameters = append(decl.parameters, joinTokens(parameter))
	}
	decl.parameterNames, decl.parameterTypes, decl.parameterDefaults = parameterDetails(parameterTokens)
	if arrow := indexOfTokenAfter(tokens, "->", close); arrow > close {
		decl.returnType = strings.TrimSpace(joinTokensUntil(tokens[arrow+1:], ":"))
	}
}

func parseValueDeclaration(decl *declaration, tokens []token, nameIndex int) {
	equals := indexOfAssignment(tokens, nameIndex+1)
	if colon := indexOfTokenAfter(tokens, ":", nameIndex); colon > nameIndex && (equals < 0 || colon < equals) {
		end := len(tokens)
		if equals >= 0 {
			end = equals
		}
		decl.typeName = strings.TrimSpace(joinTokens(tokens[colon+1 : end]))
		if decl.typeName != "" {
			decl.attributes["type"] = decl.typeName
		}
	}
	if equals >= 0 && equals+1 < len(tokens) {
		decl.initializer = append([]token(nil), tokens[equals+1:]...)
	}
	decl.preloadPath = parseStaticLoadPath(tokens[nameIndex+1:])
}

func parameterDetails(parameters [][]token) ([]string, map[string]string, map[string][]token) {
	var names []string
	types := make(map[string]string)
	defaults := make(map[string][]token)
	for _, parameter := range parameters {
		equals := topLevelAssignment(parameter)
		left := parameter
		if equals >= 0 {
			left = parameter[:equals]
		}
		colon := topLevelToken(left, ":")
		nameTokens := left
		if colon >= 0 {
			nameTokens = left[:colon]
		}
		name := strings.TrimSpace(joinTokens(nameTokens))
		if name == "" {
			continue
		}
		names = append(names, name)
		if colon >= 0 && colon+1 < len(left) {
			types[name] = strings.TrimSpace(joinTokens(left[colon+1:]))
		}
		if equals >= 0 && equals+1 < len(parameter) {
			defaults[name] = append([]token(nil), parameter[equals+1:]...)
		}
	}
	return names, types, defaults
}

func parseStaticLoadPath(tokens []token) string {
	equals := indexOfAssignment(tokens, 0)
	if equals < 0 || equals+3 >= len(tokens) || (tokens[equals+1].text != "preload" && tokens[equals+1].text != "load") || tokens[equals+2].text != "(" {
		return ""
	}
	close := matchingParen(tokens, equals+2)
	if close != equals+4 || close != len(tokens)-1 || tokens[equals+3].kind != tokenString {
		return ""
	}
	path, ok := normalizeImportPath(tokens[equals+3].text)
	if !ok {
		return ""
	}
	return path
}

func indexOfAssignment(tokens []token, start int) int {
	for i := start; i < len(tokens); i++ {
		if tokens[i].text == "=" || tokens[i].text == ":=" {
			return i
		}
	}
	return -1
}

func nextToken(tokens []token, start int, text string) int {
	for i := start; i < len(tokens); i++ {
		if tokens[i].text == text {
			return i
		}
	}
	return -1
}

func indexOfToken(tokens []token, text string) int { return indexOfTokenAfter(tokens, text, -1) }

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
		if tok.text != "\n" {
			result.WriteString(tok.text)
		}
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
