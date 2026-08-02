package main

var javaModifiers = map[string]struct{}{
	"abstract": {}, "default": {}, "final": {}, "native": {}, "private": {}, "protected": {},
	"public": {}, "sealed": {}, "static": {}, "strictfp": {}, "synchronized": {}, "transient": {},
	"volatile": {},
}

type javaParser struct {
	facts                *factSet
	fileID               string
	moduleID             string
	packageName          string
	path                 string
	recordComponentTypes map[string][]string
	resolution           resolutionContext
	source               string
	state                *analysisState
	tokens               []token
}

func parseJavaSource(state *analysisState, fileID, path, source string) {
	tokens, lexErrors := lexJava(source)
	packageName, headerEnd, packageProblems := parsePackageHeader(tokens)
	moduleID := state.addModule(fileID, path, packageName)
	namespaceID := state.ensureNamespace(packageName)
	state.facts.addEdge(namespaceID, moduleID, "contains", path, nil, nil)
	parser := &javaParser{
		facts: state.facts, fileID: fileID, moduleID: moduleID, packageName: packageName,
		path: path, recordComponentTypes: make(map[string][]string),
		resolution: resolutionContext{packageName: packageName}, source: source, state: state, tokens: tokens,
	}
	for _, problem := range lexErrors {
		problemSpan := &span{EndColumn: problem.column + 1, EndLine: problem.line, Path: path, StartColumn: problem.column, StartLine: problem.line}
		state.facts.addUnresolved(moduleID, "defines", problem.String(), "unsupported-form", path, problemSpan, nil)
	}
	for _, problem := range packageProblems {
		state.facts.addUnresolved(moduleID, "defines", problem, "unsupported-form", path, nil, nil)
	}
	parser.parseImports()
	parser.parseTopLevel(headerEnd)
}

func parsePackageHeader(tokens []token) (string, int, []string) {
	packageName := ""
	headerEnd := 0
	var problems []string
	for index := 0; index < len(tokens); {
		if tokens[index].text == "package" {
			end := findToken(tokens, index+1, len(tokens), ";")
			if end < 0 {
				problems = append(problems, "package declaration has no terminating semicolon")
				return packageName, index + 1, problems
			}
			name, valid := qualifiedTokenName(tokens[index+1 : end])
			if !valid {
				problems = append(problems, "package declaration is not a qualified identifier")
			} else {
				packageName = name
			}
			headerEnd = end + 1
			index = end + 1
			continue
		}
		if tokens[index].text == "import" {
			end := findToken(tokens, index+1, len(tokens), ";")
			if end < 0 {
				problems = append(problems, "import declaration has no terminating semicolon")
				return packageName, index + 1, problems
			}
			headerEnd = end + 1
			index = end + 1
			continue
		}
		if _, _, ok := typeKeywordAt(tokens, skipPrefix(tokens, index, len(tokens))); ok {
			break
		}
		index++
	}
	return packageName, headerEnd, problems
}

func (parser *javaParser) parseImports() {
	for index := 0; index < len(parser.tokens); {
		if parser.tokens[index].text != "import" {
			index++
			continue
		}
		end := findToken(parser.tokens, index+1, len(parser.tokens), ";")
		if end < 0 {
			return
		}
		cursor := index + 1
		static := cursor < end && parser.tokens[cursor].text == "static"
		if static {
			cursor++
		}
		parts := parser.tokens[cursor:end]
		wildcard := len(parts) >= 2 && parts[len(parts)-2].text == "." && parts[len(parts)-1].text == "*"
		nameParts := parts
		if wildcard {
			nameParts = parts[:len(parts)-2]
		}
		target, valid := qualifiedTokenName(nameParts)
		statementSpan := parser.tokenSpan(index, end+1)
		expression := sourceExcerpt(parser.source, parser.tokens, index, end+1)
		if !valid {
			parser.facts.addUnresolved(parser.moduleID, "defines", expression, "unsupported-form", parser.path, statementSpan, nil)
			index = end + 1
			continue
		}
		identity := parser.path + "::import:" + decimal(parser.tokens[index].line) + ":" + decimal(parser.tokens[index].column) + ":" + target
		id := parser.facts.addNode("import", target, parser.path, identity, identity, parser.path, statementSpan, map[string]any{
			"expression": expression, "static": static, "target": target, "wildcard": wildcard,
		}, "")
		parser.facts.addEdge(parser.moduleID, id, "defines", parser.path, statementSpan, nil)
		parser.state.imports = append(parser.state.imports, importEvidence{
			expression: expression, id: id, owner: parser.path, span: statementSpan,
			static: static, target: target, wildcard: wildcard,
		})
		if !static {
			if wildcard {
				parser.resolution.wildcardImports = append(parser.resolution.wildcardImports, target)
			} else {
				parser.resolution.explicitImports = append(parser.resolution.explicitImports, target)
			}
		}
		index = end + 1
	}
}

func (parser *javaParser) parseTopLevel(start int) {
	for index := start; index < len(parser.tokens); {
		if parser.tokens[index].text == ";" {
			index++
			continue
		}
		core := skipPrefix(parser.tokens, index, len(parser.tokens))
		if _, _, ok := typeKeywordAt(parser.tokens, core); ok {
			next := parser.parseType(index, core, len(parser.tokens), parser.moduleID, "", "defines")
			if next > index {
				index = next
				continue
			}
		}
		end := nextTopLevelTerminator(parser.tokens, index, len(parser.tokens))
		if end <= index {
			end = index + 1
		}
		parser.facts.addUnresolved(parser.moduleID, "defines", sourceExcerpt(parser.source, parser.tokens, index, end), "unsupported-form", parser.path, parser.tokenSpan(index, end), nil)
		index = end
	}
}

func (parser *javaParser) parseType(start, keywordIndex, limit int, ownerID, parentQualifiedName, relation string) int {
	keyword, declarationKind, ok := typeKeywordAt(parser.tokens, keywordIndex)
	if !ok {
		return start
	}
	nameIndex := keywordIndex + 1
	if keyword == "annotation" {
		nameIndex = keywordIndex + 2
	}
	if nameIndex >= limit || !identifierToken(parser.tokens[nameIndex].text) {
		parser.facts.addUnresolved(ownerID, "defines", sourceExcerpt(parser.source, parser.tokens, start, min(limit, nameIndex+1)), "unsupported-form", parser.path, parser.tokenSpan(start, min(limit, nameIndex+1)), nil)
		return min(limit, nameIndex+1)
	}
	bodyOpen := findTypeBody(parser.tokens, nameIndex+1, limit)
	if bodyOpen < 0 {
		end := nextTopLevelTerminator(parser.tokens, nameIndex+1, limit)
		if end <= nameIndex {
			end = min(limit, nameIndex+1)
		}
		parser.facts.addUnresolved(ownerID, "defines", sourceExcerpt(parser.source, parser.tokens, start, end), "unsupported-form", parser.path, parser.tokenSpan(start, end), nil)
		return end
	}
	bodyClose := matchingToken(parser.tokens, bodyOpen, limit, "{", "}")
	if bodyClose < 0 {
		parser.facts.addUnresolved(ownerID, "defines", "unclosed "+declarationKind+" "+parser.tokens[nameIndex].text, "unsupported-form", parser.path, parser.tokenSpan(keywordIndex, min(limit, bodyOpen+1)), nil)
		bodyClose = limit - 1
		if bodyClose < bodyOpen {
			return limit
		}
	}

	name := parser.tokens[nameIndex].text
	qualifiedName := name
	if parentQualifiedName != "" {
		qualifiedName = parentQualifiedName + "." + name
	} else if parser.packageName != "" {
		qualifiedName = parser.packageName + "." + name
	}
	kind := "type"
	if declarationKind == "interface" || declarationKind == "annotation" {
		kind = "interface"
	}
	modifiers := modifierList(parser.tokens, start, keywordIndex)
	attributes := map[string]any{"declaration_kind": declarationKind}
	if len(modifiers) != 0 {
		attributes["modifiers"] = modifiers
	}
	typeSpan := parser.tokenSpan(keywordIndex, bodyClose+1)
	id := parser.facts.addNode(kind, name, parser.path, qualifiedName, qualifiedName, parser.path, typeSpan, attributes, "")
	parser.state.registerType(qualifiedName, id, declarationKind)
	parser.facts.addEdge(ownerID, id, relation, parser.path, typeSpan, nil)
	parser.queueAnnotations(id, parentQualifiedName, start, keywordIndex)
	parser.parseTypeRelationships(id, parentQualifiedName, nameIndex+1, bodyOpen)
	if declarationKind == "record" {
		parser.recordComponentTypes[qualifiedName] = parser.parseRecordComponents(id, qualifiedName, nameIndex+1, bodyOpen)
	}
	memberStart := bodyOpen + 1
	if declarationKind == "enum" {
		memberStart = enumMemberStart(parser.tokens, memberStart, bodyClose)
	}
	parser.parseMembers(id, qualifiedName, name, memberStart, bodyClose)
	return bodyClose + 1
}
