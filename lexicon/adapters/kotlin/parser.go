package main

import "strings"

type importDecl struct {
	alias    string
	path     string
	span     sourceSpan
	wildcard bool
}

type parameterDecl struct {
	annotations []string
	hasDefault  bool
	modifiers   []string
	mutable     bool
	name        string
	property    bool
	span        sourceSpan
	typeName    string
}

type supertypeDecl struct {
	delegateExpression string
	delegated          bool
	expression         string
	span               sourceSpan
	targetName         string
}

type tokenRange struct {
	end   int
	start int
}

type declaration struct {
	annotations []string
	body        tokenRange
	children    []*declaration
	delegation  tokenRange
	delegated   bool
	form        string
	kind        string
	modifiers   []string
	mutable     bool
	name        string
	parameters  []parameterDecl
	primary     bool
	receiver    string
	returnType  string
	span        sourceSpan
	supertypes  []supertypeDecl
	typeName    string
}

type parsedKotlinFile struct {
	content      []byte
	declarations []*declaration
	diagnostics  []syntaxDiagnostic
	imports      []importDecl
	packageName  string
	packageSpan  *sourceSpan
	path         string
	tokens       []token
}

type parser struct {
	content     []byte
	diagnostics []syntaxDiagnostic
	index       int
	path        string
	tokens      []token
}

var declarationModifiers = map[string]struct{}{
	"abstract": {}, "actual": {}, "annotation": {}, "companion": {}, "const": {},
	"crossinline": {}, "data": {}, "enum": {}, "expect": {}, "external": {},
	"final": {}, "infix": {}, "inline": {}, "inner": {}, "internal": {},
	"lateinit": {}, "noinline": {}, "open": {}, "operator": {}, "out": {},
	"override": {}, "private": {}, "protected": {}, "public": {}, "reified": {},
	"sealed": {}, "suspend": {}, "tailrec": {}, "value": {}, "vararg": {},
}

var parameterModifiers = map[string]struct{}{
	"crossinline": {}, "noinline": {}, "vararg": {},
}

type declarationPrefix struct {
	annotations []string
	modifiers   []string
	start       int
}

func parseKotlinFile(path string, content []byte) *parsedKotlinFile {
	tokens, lexerDiagnostics := lex(content)
	state := &parser{content: content, diagnostics: append([]syntaxDiagnostic(nil), lexerDiagnostics...), path: path, tokens: tokens}
	result := &parsedKotlinFile{content: content, path: path, tokens: tokens}

	state.skipSeparators()
	for state.at("@") {
		state.parseAnnotation()
		state.skipSeparators()
	}
	if state.at("package") {
		start := state.index
		state.index++
		name, end, ok := state.parseQualifiedDirective(false)
		if ok {
			result.packageName = name
			span := state.span(start, end)
			result.packageSpan = &span
		} else {
			state.addDiagnostic(start, "malformed package directive")
		}
	}
	state.skipSeparators()
	for state.at("import") {
		start := state.index
		state.index++
		name, end, ok := state.parseQualifiedDirective(true)
		alias := ""
		if state.at("as") {
			state.index++
			if state.current().kind == tokenIdentifier {
				alias = identifierText(state.current())
				end = state.index
				state.index++
			} else {
				ok = false
			}
		}
		if ok {
			result.imports = append(result.imports, importDecl{
				alias: alias, path: name, span: state.span(start, end), wildcard: strings.HasSuffix(name, ".*"),
			})
		} else {
			state.addDiagnostic(start, "malformed import directive")
		}
		state.skipToLineEnd()
		state.skipSeparators()
	}

	result.declarations = state.parseScope(false)
	result.diagnostics = state.diagnostics
	return result
}

func (state *parser) parseScope(stopAtBrace bool) []*declaration {
	var declarations []*declaration
	for state.current().kind != tokenEOF {
		state.skipSeparators()
		if stopAtBrace && state.at("}") {
			state.index++
			return declarations
		}
		if state.current().kind == tokenEOF {
			break
		}
		if state.at(")") || state.at("]") || state.at("}") {
			state.addDiagnostic(state.index, "unmatched closing delimiter")
			state.index++
			continue
		}

		prefix := state.parsePrefix()
		var declaration *declaration
		switch state.current().text {
		case "class", "interface", "object":
			declaration = state.parseType(prefix)
		case "fun":
			declaration = state.parseFunction(prefix)
		case "val", "var":
			declaration = state.parseProperty(prefix)
		case "constructor":
			declaration = state.parseConstructor(prefix, false)
		default:
			if len(prefix.annotations) != 0 || len(prefix.modifiers) != 0 {
				state.addDiagnostic(prefix.start, "unsupported or malformed declaration")
			}
			state.skipStatement()
		}
		if declaration != nil {
			declarations = append(declarations, declaration)
		}
	}
	if stopAtBrace {
		state.addDiagnostic(maxInt(0, state.index-1), "unclosed declaration body")
	}
	return declarations
}

func (state *parser) parsePrefix() declarationPrefix {
	prefix := declarationPrefix{start: state.index}
	for {
		if state.at("@") {
			if annotation := state.parseAnnotation(); annotation != "" {
				prefix.annotations = append(prefix.annotations, annotation)
			}
			state.skipNewlines()
			continue
		}
		if _, ok := declarationModifiers[state.current().text]; ok {
			prefix.modifiers = append(prefix.modifiers, state.current().text)
			state.index++
			state.skipNewlines()
			continue
		}
		break
	}
	prefix.annotations = uniqueSorted(prefix.annotations)
	prefix.modifiers = uniqueSorted(prefix.modifiers)
	return prefix
}

func (state *parser) parseAnnotation() string {
	if !state.at("@") {
		return ""
	}
	state.index++
	var parts []string
	for state.current().kind == tokenIdentifier || state.at(".") || state.at(":") {
		parts = append(parts, state.current().text)
		state.index++
	}
	if state.at("[") {
		state.skipBalanced("[", "]")
	} else if state.at("(") {
		state.skipBalanced("(", ")")
	}
	return compactTokens(parts)
}
