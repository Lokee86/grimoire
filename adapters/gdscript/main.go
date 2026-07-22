package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

const (
	adapterVersion = "0.1.0"
	language       = "gdscript"
)

type tokenKind uint8

const (
	tokenIdentifier tokenKind = iota
	tokenString
	tokenNumber
	tokenSymbol
)

type token struct {
	kind      tokenKind
	text      string
	line      int
	column    int
	endLine   int
	endColumn int
}

type statement struct {
	tokens []token
	indent int
	start  token
	end    token
}

type sourceSpan = map[string]any

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

type importReference struct {
	loader string
	expr   string
	path   string
	static bool
	span   sourceSpan
}

type callReference struct {
	callee string
	expr   string
	span   sourceSpan
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

type factSet struct {
	nodes        []map[string]any
	edges        []map[string]any
	unresolved   []map[string]any
	nodeByID     map[string]map[string]any
	moduleByPath map[string]string
	classByName  map[string][]string
	fileByPath   map[string]string
}

func main() {
	repo := flag.String("repo", "", "repository root to analyze")
	output := flag.String("output", "", "JSONL output path")
	flag.Parse()
	if *repo == "" || *output == "" {
		flag.Usage()
		os.Exit(2)
	}
	if err := writeFacts(*repo, *output); err != nil {
		fmt.Fprintf(os.Stderr, "gdscript adapter: %v\n", err)
		os.Exit(1)
	}
}

func writeFacts(repo, output string) error {
	data, err := analyzeRepository(repo)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	if err := os.WriteFile(output, data, 0o644); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}

func analyzeRepository(repo string) ([]byte, error) {
	root, err := filepath.Abs(repo)
	if err != nil {
		return nil, fmt.Errorf("resolve repository: %w", err)
	}
	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("stat repository: %w", err)
	}
	if !info.IsDir() {
		return nil, errors.New("repository path is not a directory")
	}

	files, dirs, err := collectSources(root)
	if err != nil {
		return nil, err
	}
	repositoryName := filepath.Base(filepath.Clean(root))
	facts := &factSet{
		nodeByID:     make(map[string]map[string]any),
		moduleByPath: make(map[string]string),
		classByName:  make(map[string][]string),
		fileByPath:   make(map[string]string),
	}

	repositoryID := nodeID("repository", repositoryName)
	facts.addNode(node("repository", repositoryName, ".", repositoryName, repositoryID, nil, ""))
	rootDirectoryID := nodeID("directory", ".")
	facts.addNode(node("directory", repositoryName, ".", ".", rootDirectoryID, nil, ""))
	facts.addEdge(edge(repositoryID, rootDirectoryID, "contains", nil))

	dirIDs := map[string]string{".": rootDirectoryID}
	for _, dir := range dirs {
		if dir == "." {
			continue
		}
		id := nodeID("directory", dir)
		dirIDs[dir] = id
		facts.addNode(node("directory", filepath.Base(filepath.FromSlash(dir)), dir, dir, id, nil, ""))
	}
	for _, dir := range dirs {
		if dir == "." {
			continue
		}
		parent := filepath.ToSlash(filepath.Dir(filepath.FromSlash(dir)))
		if parent == "" {
			parent = "."
		}
		facts.addEdge(edge(dirIDs[parent], dirIDs[dir], "contains", nil))
	}

	parsed := make([]*parsedFile, 0, len(files))
	for _, path := range files {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		pf, err := parseFile(path, content)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		fileID := nodeID("file", path)
		moduleID := nodeID("module", path)
		pf.moduleID = moduleID
		facts.fileByPath[path] = fileID
		facts.moduleByPath[path] = moduleID
		facts.addNode(node("file", filepath.Base(filepath.FromSlash(path)), path, path, fileID, nil, contentID(content)))
		facts.addNode(node("module", strings.TrimSuffix(filepath.Base(filepath.FromSlash(path)), filepath.Ext(path)), path, path, moduleID, nil, ""))
		facts.addEdge(edge(dirIDs[filepath.ToSlash(filepath.Dir(filepath.FromSlash(path)))], fileID, "contains", nil))
		facts.addEdge(edge(fileID, moduleID, "contains", nil))
		parsed = append(parsed, pf)
	}

	// Discover named classes before resolving inheritance and member ownership.
	for _, pf := range parsed {
		typeOccurrences := make(map[string]int)
		for i := range pf.declarations {
			decl := &pf.declarations[i]
			if decl.kind != "type" || decl.name == "" {
				continue
			}
			baseKey := pf.path + "::type::" + decl.name
			occurrence := typeOccurrences[baseKey]
			typeOccurrences[baseKey] = occurrence + 1
			decl.key = baseKey
			if occurrence > 0 {
				decl.key += fmt.Sprintf("#%d", occurrence+1)
			}
			decl.nodeID = nodeID("type", decl.key)
			pf.classID = decl.nodeID
			facts.classByName[decl.name] = append(facts.classByName[decl.name], decl.nodeID)
		}
	}

	for _, pf := range parsed {
		processDeclarations(facts, pf)
	}
	for _, pf := range parsed {
		processImportsAndCalls(facts, pf)
	}
	for _, pf := range parsed {
		processExtends(facts, pf)
	}

	return facts.render(repositoryName), nil
}

func collectSources(root string) ([]string, []string, error) {
	var files []string
	dirs := map[string]bool{".": true}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		name := strings.ToLower(entry.Name())
		if entry.IsDir() {
			if excludedDirectory(name) {
				return filepath.SkipDir
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			dirs[filepath.ToSlash(rel)] = true
			return nil
		}
		if strings.EqualFold(filepath.Ext(entry.Name()), ".gd") {
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			files = append(files, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("scan repository: %w", err)
	}
	sort.Strings(files)
	// Keep only directories that lead to a GDScript file. This avoids facts for
	// unrelated checkout directories while retaining the complete source tree.
	needed := map[string]bool{".": true}
	for _, file := range files {
		dir := filepath.ToSlash(filepath.Dir(filepath.FromSlash(file)))
		for dir != "." && dir != "" {
			needed[dir] = true
			dir = filepath.ToSlash(filepath.Dir(filepath.FromSlash(dir)))
		}
	}
	filteredDirs := make([]string, 0, len(needed))
	for dir := range needed {
		if dirs[dir] {
			filteredDirs = append(filteredDirs, dir)
		}
	}
	sort.Slice(filteredDirs, func(i, j int) bool {
		if filteredDirs[i] == "." {
			return true
		}
		if filteredDirs[j] == "." {
			return false
		}
		return filteredDirs[i] < filteredDirs[j]
	})
	return files, filteredDirs, nil
}

func excludedDirectory(name string) bool {
	switch name {
	case ".git", ".worktrees", ".workingtrees", ".warlock", "node_modules", "target", "__pycache__", ".pytest_cache", ".bundle", "vendor", ".godot", ".import", "build", "dist", "bin", "obj":
		return true
	default:
		return false
	}
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

func processImportsAndCalls(facts *factSet, pf *parsedFile) {
	functionIDs := map[string][]string{}
	for _, decl := range pf.declarations {
		if decl.kind == "function" {
			functionIDs[decl.name] = append(functionIDs[decl.name], decl.nodeID)
		}
	}
	ordinal := 0
	for _, stmt := range pf.statements {
		owner := ownerForStatement(pf, stmt)
		if decl := declarationForStatement(pf, stmt); decl != nil && decl.nodeID != "" {
			owner = decl.nodeID
		}
		for _, ref := range findImports(stmt, pf.path) {
			ordinal++
			importKey := pf.path + "::import::" + fmt.Sprintf("%d", ordinal) + "::" + ref.expr
			importID := nodeID("import", importKey)
			attrs := map[string]any{"expression": ref.expr, "loader": ref.loader, "static": ref.static}
			if ref.path != "" {
				attrs["resolved_path"] = ref.path
			}
			facts.addNode(node("import", ref.expr, pf.path, importKey, importID, ref.span, "", attrs))
			facts.addEdge(edge(owner, importID, "imports", ref.span))
			if ref.static && ref.path != "" {
				if target, ok := facts.moduleByPath[ref.path]; ok {
					facts.addEdge(edge(importID, target, "references", ref.span))
				} else {
					facts.addUnresolved(unresolved(owner, "imports", ref.expr, "missing-target", ref.span))
				}
			} else if ref.static {
				facts.addUnresolved(unresolved(owner, "imports", ref.expr, "external-target", ref.span))
			} else {
				facts.addUnresolved(unresolved(owner, "imports", ref.expr, "dynamic-target", ref.span))
			}
		}
		for _, call := range findCalls(stmt, pf.path) {
			if ids := functionIDs[call.callee]; len(ids) == 1 && isSimpleCallee(call.callee) {
				facts.addEdge(edge(owner, ids[0], "calls", call.span))
				continue
			}
			reason := "dynamic-target"
			if isSimpleCallee(call.callee) {
				reason = "missing-target"
				if isBuiltin(call.callee) {
					reason = "builtin-target"
				}
			}
			facts.addUnresolved(unresolved(owner, "calls", call.expr, reason, call.span))
		}
	}
}

func ownerForStatement(pf *parsedFile, stmt statement) string {
	owner := pf.moduleID
	bestIndent := -1
	for _, decl := range pf.declarations {
		if decl.nodeID == "" || decl.indent >= stmt.indent || decl.indent < bestIndent || spanInt(decl.span, "start_line") > stmt.start.line {
			continue
		}
		if decl.kind == "function" || decl.kind == "class" {
			bestIndent = decl.indent
			owner = decl.nodeID
		}
	}
	if pf.classID != "" && stmt.indent == 0 {
		owner = pf.classID
	}
	return owner
}

func declarationForStatement(pf *parsedFile, stmt statement) *declaration {
	for i := range pf.declarations {
		decl := &pf.declarations[i]
		if decl.span["start_line"] == stmt.start.line && decl.span["start_column"] == stmt.start.column {
			return decl
		}
	}
	return nil
}

func processExtends(facts *factSet, pf *parsedFile) {
	for _, decl := range pf.declarations {
		if decl.extends == "" {
			continue
		}
		source := pf.moduleID
		if pf.classID != "" {
			source = pf.classID
		}
		if path, ok := normalizeImportPath(decl.extends); ok {
			if target, exists := facts.moduleByPath[path]; exists {
				facts.addEdge(edge(source, target, "extends", decl.span))
			} else {
				facts.addUnresolved(unresolved(source, "extends", decl.extends, "missing-target", decl.span))
			}
			continue
		}
		name := strings.TrimSpace(decl.extends)
		if ids := facts.classByName[name]; len(ids) == 1 {
			facts.addEdge(edge(source, ids[0], "extends", decl.span))
		} else if len(ids) > 1 {
			record := unresolved(source, "extends", decl.extends, "ambiguous-target", decl.span)
			record["candidate_name"] = name
			facts.addUnresolved(record)
		} else if isBuiltin(name) {
			facts.addUnresolved(unresolved(source, "extends", decl.extends, "builtin-target", decl.span))
		} else {
			record := unresolved(source, "extends", decl.extends, "missing-target", decl.span)
			record["candidate_name"] = name
			facts.addUnresolved(record)
		}
	}
}

func findImports(stmt statement, path string) []importReference {
	var refs []importReference
	for i := 0; i+2 < len(stmt.tokens); i++ {
		t := stmt.tokens[i]
		if t.kind != tokenIdentifier || (t.text != "preload" && t.text != "load") || stmt.tokens[i+1].text != "(" {
			continue
		}
		close := matchingParen(stmt.tokens, i+1)
		if close < 0 {
			continue
		}
		args := stmt.tokens[i+2 : close]
		expr := joinTokens(args)
		ref := importReference{loader: t.text, expr: expr, span: spanFromTokens(path, t, stmt.tokens[close])}
		if len(args) == 1 && args[0].kind == tokenString {
			ref.static = true
			ref.path, _ = normalizeImportPath(args[0].text)
		}
		refs = append(refs, ref)
		i = close
	}
	return refs
}

func findCalls(stmt statement, path string) []callReference {
	var calls []callReference
	for i := 0; i+1 < len(stmt.tokens); i++ {
		t := stmt.tokens[i]
		if t.kind != tokenIdentifier || stmt.tokens[i+1].text != "(" || t.text == "preload" || t.text == "load" || isCallKeyword(t.text) {
			continue
		}
		if i > 0 && stmt.tokens[i-1].text == "." {
			start := i - 2
			for start >= 0 && (stmt.tokens[start].kind == tokenIdentifier || stmt.tokens[start].text == ".") {
				start--
			}
			start++
			callee := joinTokens(stmt.tokens[start : i+1])
			close := matchingParen(stmt.tokens, i+1)
			if close < 0 {
				close = i + 1
			}
			calls = append(calls, callReference{callee: callee, expr: joinTokens(stmt.tokens[start : close+1]), span: spanFromTokens(path, stmt.tokens[start], stmt.tokens[close])})
			i = close
			continue
		}
		if decl := parseDeclaration(stmt); decl != nil && decl.kind == "function" && decl.nameIndex == i {
			continue
		}
		close := matchingParen(stmt.tokens, i+1)
		if close < 0 {
			close = i + 1
		}
		calls = append(calls, callReference{callee: t.text, expr: joinTokens(stmt.tokens[i : close+1]), span: spanFromTokens(path, t, stmt.tokens[close])})
		i = close
	}
	return calls
}

func matchingParen(tokens []token, open int) int {
	depth := 0
	for i := open; i < len(tokens); i++ {
		switch tokens[i].text {
		case "(":
			depth++
		case ")":
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
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
	case "class_name":
		decl.kind = "type"
	case "class":
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

func makeStatements(tokens []token) []statement {
	var statements []statement
	var current []token
	depth := 0
	flush := func() {
		if len(current) == 0 {
			return
		}
		statements = append(statements, statement{tokens: append([]token(nil), current...), indent: current[0].column - 1, start: current[0], end: current[len(current)-1]})
		current = nil
	}
	for _, tok := range tokens {
		if tok.text == "\n" {
			if depth == 0 {
				flush()
			}
			continue
		}
		current = append(current, tok)
		switch tok.text {
		case "(", "[", "{":
			depth++
		case ")", "]", "}":
			if depth > 0 {
				depth--
			}
		}
	}
	flush()
	return statements
}

func lex(source string) ([]token, error) {
	var tokens []token
	line, column := 1, 1
	for i := 0; i < len(source); {
		ch := source[i]
		if ch == '\r' {
			i++
			continue
		}
		if ch == '\n' {
			tokens = append(tokens, token{text: "\n", line: line, column: column, endLine: line, endColumn: column + 1})
			i++
			line++
			column = 1
			continue
		}
		if ch == ' ' || ch == '\t' {
			i++
			if ch == '\t' {
				column += 4
			} else {
				column++
			}
			continue
		}
		if ch == '#' {
			for i < len(source) && source[i] != '\n' {
				i++
				column++
			}
			continue
		}
		startLine, startColumn, start := line, column, i
		if ch == '\'' || ch == '"' {
			quote := ch
			triple := i+2 < len(source) && source[i+1] == quote && source[i+2] == quote
			if triple {
				i += 3
				column += 3
			} else {
				i++
				column++
			}
			closed := false
			for i < len(source) {
				if triple && i+2 < len(source) && source[i] == quote && source[i+1] == quote && source[i+2] == quote {
					i += 3
					column += 3
					closed = true
					break
				}
				if !triple && source[i] == quote {
					i++
					column++
					closed = true
					break
				}
				if source[i] == '\\' && i+1 < len(source) {
					i += 2
					column += 2
					continue
				}
				if source[i] == '\n' {
					i++
					line++
					column = 1
					continue
				}
				i++
				column++
			}
			if !closed {
				return nil, fmt.Errorf("unterminated string at %d:%d", startLine, startColumn)
			}
			tokens = append(tokens, token{kind: tokenString, text: source[start:i], line: startLine, column: startColumn, endLine: line, endColumn: column})
			continue
		}
		if isIdentifierStart(ch) {
			i++
			column++
			for i < len(source) && isIdentifierPart(source[i]) {
				i++
				column++
			}
			tokens = append(tokens, token{kind: tokenIdentifier, text: source[start:i], line: startLine, column: startColumn, endLine: line, endColumn: column})
			continue
		}
		if unicode.IsDigit(rune(ch)) {
			i++
			column++
			for i < len(source) && (isIdentifierPart(source[i]) || source[i] == '.') {
				i++
				column++
			}
			tokens = append(tokens, token{kind: tokenNumber, text: source[start:i], line: startLine, column: startColumn, endLine: line, endColumn: column})
			continue
		}
		symbol := string(ch)
		if i+1 < len(source) {
			candidate := source[i : i+2]
			switch candidate {
			case "->", ":=", "==", "!=", "<=", ">=", "&&", "||", "+=", "-=", "*=", "/=":
				symbol = candidate
			}
		}
		i += len(symbol)
		column += len(symbol)
		tokens = append(tokens, token{kind: tokenSymbol, text: symbol, line: startLine, column: startColumn, endLine: line, endColumn: column})
	}
	return tokens, nil
}

func spanFromTokens(path string, start, end token) sourceSpan {
	span := sourceSpan{"end_column": end.endColumn, "end_line": end.endLine, "start_column": start.column, "start_line": start.line}
	if path != "" {
		span["path"] = path
	}
	return span
}

func node(kind, name, path, qualified, id string, span sourceSpan, content string, attributes ...map[string]any) map[string]any {
	record := map[string]any{"id": id, "kind": kind, "name": name, "path": path, "qualified_name": qualified, "record": "node"}
	if content != "" {
		record["content_id"] = content
	}
	if span != nil {
		record["span"] = span
	}
	if len(attributes) > 0 && len(attributes[0]) > 0 {
		record["attributes"] = attributes[0]
	}
	return record
}

func edge(source, target, relation string, span sourceSpan) map[string]any {
	record := map[string]any{"record": "edge", "relation": relation, "source": source, "target": target}
	if span != nil {
		record["span"] = span
	}
	return record
}

func unresolved(source, relation, expression, reason string, span sourceSpan) map[string]any {
	record := map[string]any{"expression": expression, "reason": reason, "record": "unresolved", "relation": relation, "source": source}
	if span != nil {
		record["span"] = span
	}
	return record
}

func (f *factSet) addNode(record map[string]any) {
	id := record["id"].(string)
	if _, exists := f.nodeByID[id]; exists {
		return
	}
	f.nodeByID[id] = record
	f.nodes = append(f.nodes, record)
}

func (f *factSet) addEdge(record map[string]any) {
	for _, existing := range f.edges {
		if recordKey(existing) == recordKey(record) {
			return
		}
	}
	f.edges = append(f.edges, record)
}

func (f *factSet) addUnresolved(record map[string]any) {
	for _, existing := range f.unresolved {
		if recordKey(existing) == recordKey(record) {
			return
		}
	}
	f.unresolved = append(f.unresolved, record)
}

func (f *factSet) render(repositoryName string) []byte {
	sort.Slice(f.nodes, func(i, j int) bool { return nodeSortKey(f.nodes[i]) < nodeSortKey(f.nodes[j]) })
	sort.Slice(f.edges, func(i, j int) bool { return edgeSortKey(f.edges[i]) < edgeSortKey(f.edges[j]) })
	sort.Slice(f.unresolved, func(i, j int) bool { return unresolvedSortKey(f.unresolved[i]) < unresolvedSortKey(f.unresolved[j]) })
	records := make([]map[string]any, 0, 1+len(f.nodes)+len(f.edges)+len(f.unresolved))
	records = append(records, map[string]any{"adapter_version": adapterVersion, "language": language, "record": "lexicon", "repository": repositoryName, "schema_version": 1})
	records = append(records, f.nodes...)
	records = append(records, f.edges...)
	records = append(records, f.unresolved...)
	var output strings.Builder
	for _, record := range records {
		data, _ := json.Marshal(record)
		output.Write(data)
		output.WriteByte('\n')
	}
	return []byte(output.String())
}

func nodeSortKey(record map[string]any) string {
	return fmt.Sprintf("%s\x00%s\x00%s\x00%s", record["id"], record["kind"], record["path"], record["qualified_name"])
}

func edgeSortKey(record map[string]any) string {
	return fmt.Sprintf("%s\x00%s\x00%s\x00%s", record["source"], record["target"], record["relation"], spanSortKey(record))
}

func unresolvedSortKey(record map[string]any) string {
	return fmt.Sprintf("%s\x00%s\x00%s\x00%s\x00%s", record["source"], record["relation"], record["expression"], record["reason"], spanSortKey(record))
}

func spanSortKey(record map[string]any) string {
	span, _ := record["span"].(map[string]any)
	if span == nil {
		return ""
	}
	return fmt.Sprintf("%s\x00%08d\x00%08d\x00%08d\x00%08d", spanString(span, "path"), spanInt(span, "start_line"), spanInt(span, "start_column"), spanInt(span, "end_line"), spanInt(span, "end_column"))
}

func recordKey(record map[string]any) string {
	data, _ := json.Marshal(record)
	return string(data)
}

func nodeID(kind, canonical string) string {
	return digest("lexicon:v1\x00" + language + "\x00" + kind + "\x00" + canonical)
}

func digest(value string) string {
	hash := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(hash[:])
}

func contentID(content []byte) string {
	hash := sha256.Sum256(content)
	return "sha256:" + hex.EncodeToString(hash[:])
}

func normalizeImportPath(expression string) (string, bool) {
	expression = strings.TrimSpace(expression)
	if len(expression) >= 2 && ((expression[0] == '"' && expression[len(expression)-1] == '"') || (expression[0] == '\'' && expression[len(expression)-1] == '\'')) {
		expression = expression[1 : len(expression)-1]
	}
	if !strings.HasPrefix(expression, "res://") {
		return "", false
	}
	path := strings.TrimPrefix(expression, "res://")
	path = filepath.ToSlash(filepath.Clean(filepath.FromSlash(path)))
	if path == "." || strings.HasPrefix(path, "../") {
		return "", false
	}
	return path, true
}

func isDeclarationKeyword(text string) bool {
	switch text {
	case "class_name", "class", "func", "signal", "const", "var", "extends":
		return true
	default:
		return false
	}
}

func isCallKeyword(text string) bool {
	switch text {
	case "if", "elif", "while", "for", "match", "func", "signal", "class", "class_name", "extends", "var", "const", "return", "await", "yield":
		return true
	default:
		return false
	}
}

func isSimpleCallee(callee string) bool { return !strings.Contains(callee, ".") && callee != "" }

func isBuiltin(name string) bool {
	switch name {
	case "Node", "Node2D", "Node3D", "Object", "RefCounted", "Resource", "Control", "CanvasItem", "CharacterBody2D", "CharacterBody3D", "Area2D", "Area3D", "Sprite2D", "Sprite3D", "PackedScene", "String", "StringName", "Vector2", "Vector3", "Color", "Transform2D", "Transform3D", "print", "print_debug", "push_error", "push_warning", "str", "len", "range", "is_instance_valid", "typeof", "preload", "load", "Callable", "Signal":
		return true
	default:
		return false
	}
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

func cloneMap(input map[string]any) map[string]any {
	output := make(map[string]any, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

func spanString(span map[string]any, key string) string {
	value, _ := span[key].(string)
	return value
}

func spanInt(span map[string]any, key string) int {
	value, _ := span[key].(int)
	return value
}

func isIdentifierStart(ch byte) bool {
	return ch == '_' || ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z'
}
func isIdentifierPart(ch byte) bool { return isIdentifierStart(ch) || ch >= '0' && ch <= '9' }
