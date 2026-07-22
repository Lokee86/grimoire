package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

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
