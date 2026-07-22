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

func processImports(facts *factSet, pf *parsedFile) {
	ordinal := 0
	for _, stmt := range pf.statements {
		owner := ownerForStatement(pf, stmt)
		for _, ref := range findImports(stmt, pf.path) {
			ordinal++
			key := pf.path + "::import::" + fmt.Sprintf("%d", ordinal) + "::" + ref.expr
			id := nodeID("import", key)
			attrs := map[string]any{"expression": ref.expr, "loader": ref.loader, "static": ref.static}
			if ref.path != "" {
				attrs["resolved_path"] = ref.path
			}
			facts.addNode(node("import", ref.expr, pf.path, key, id, ref.span, "", attrs))
			facts.addEdge(edge(owner, id, "imports", ref.span))
			if ref.static && ref.path != "" {
				if target, ok := facts.moduleByPath[ref.path]; ok {
					facts.addEdge(edge(id, target, "references", ref.span))
				} else {
					reason := "external-target"
					if strings.EqualFold(filepath.Ext(ref.path), ".gd") {
						reason = "missing-target"
					}
					facts.addUnresolved(unresolved(owner, "imports", ref.expr, reason, ref.span))
				}
			} else if ref.static {
				facts.addUnresolved(unresolved(owner, "imports", ref.expr, "external-target", ref.span))
			} else {
				facts.addUnresolved(unresolved(owner, "imports", ref.expr, "dynamic-target", ref.span))
			}
		}
	}
}

func ownerForStatement(pf *parsedFile, stmt statement) string {
	if decl := declarationForStatement(pf, stmt); decl != nil && decl.nodeID != "" {
		return decl.nodeID
	}
	context := contextForStatement(pf, stmt)
	if context.functionID != "" {
		return context.functionID
	}
	return context.ownerID
}

func declarationForStatement(pf *parsedFile, stmt statement) *declaration {
	for i := range pf.declarations {
		decl := &pf.declarations[i]
		if spanInt(decl.span, "start_line") == stmt.start.line && spanInt(decl.span, "start_column") == stmt.start.column {
			return decl
		}
	}
	return nil
}

func findImports(stmt statement, path string) []importReference {
	var refs []importReference
	for i := 0; i+2 < len(stmt.tokens); i++ {
		current := stmt.tokens[i]
		if current.kind != tokenIdentifier || (current.text != "preload" && current.text != "load") || stmt.tokens[i+1].text != "(" {
			continue
		}
		close := matchingParen(stmt.tokens, i+1)
		if close < 0 {
			continue
		}
		args := stmt.tokens[i+2 : close]
		ref := importReference{loader: current.text, expr: joinTokens(args), span: spanFromTokens(path, current, stmt.tokens[close])}
		if len(args) == 1 && args[0].kind == tokenString {
			ref.static = true
			ref.path, _ = normalizeImportPath(args[0].text)
		}
		refs = append(refs, ref)
	}
	return refs
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
	path := filepath.ToSlash(filepath.Clean(filepath.FromSlash(strings.TrimPrefix(expression, "res://"))))
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
	case "if", "elif", "while", "for", "match", "func", "signal", "class", "class_name", "extends", "var", "const", "return", "await", "yield", "and", "or", "not":
		return true
	default:
		return false
	}
}

func isBuiltin(name string) bool {
	switch name {
	case "Node", "Node2D", "Node3D", "Object", "RefCounted", "Resource", "Control", "CanvasItem", "CharacterBody2D", "CharacterBody3D", "Area2D", "Area3D", "Sprite2D", "Sprite3D", "PackedScene", "SceneTree", "Engine", "ProjectSettings", "Input", "Time", "OS", "FileAccess", "DirAccess", "JSON", "Marshalls", "Geometry2D", "Geometry3D", "PhysicsServer2D", "PhysicsServer3D", "RenderingServer", "AudioServer", "DisplayServer", "ClassDB", "ResourceLoader", "ResourceSaver", "String", "StringName", "Vector2", "Vector2i", "Vector3", "Vector3i", "Vector4", "Vector4i", "Color", "Transform2D", "Transform3D", "Basis", "Quaternion", "Rect2", "Rect2i", "Array", "Dictionary", "Callable", "Signal", "print", "prints", "print_debug", "print_stack", "push_error", "push_warning", "str", "len", "range", "is_instance_valid", "instance_from_id", "is_same", "typeof", "type_string", "preload", "load", "abs", "min", "max", "clamp", "lerp", "inverse_lerp", "remap", "move_toward", "snapped", "wrapi", "wrapf", "floor", "ceil", "round", "sqrt", "pow", "sin", "cos", "tan", "deg_to_rad", "rad_to_deg", "randf", "randi", "randf_range", "randi_range", "assert", "int", "float", "bool", "NodePath", "PackedByteArray", "PackedInt32Array", "PackedInt64Array", "PackedFloat32Array", "PackedFloat64Array", "PackedStringArray", "weakref", "error_string", "inst_to_dict", "is_instance_id_valid", "printraw":
		return true
	default:
		return false
	}
}
