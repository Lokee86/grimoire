package main

import (
	"strings"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

func (extractor *extractor) walk(node *tree_sitter.Node, context extractionContext) {
	if node == nil {
		return
	}
	switch node.Kind() {
	case "namespace_definition":
		extractor.handleNamespace(node, context)
		return
	case "class_specifier", "struct_specifier", "union_specifier", "enum_specifier":
		extractor.handleType(node, context)
		return
	case "function_definition":
		extractor.handleFunction(node, context, true)
		return
	case "declaration", "field_declaration":
		extractor.handleDeclaration(node, context)
		return
	case "type_definition":
		extractor.handleTypedef(node, context)
		return
	case "alias_declaration":
		extractor.handleAlias(node, context)
		return
	case "preproc_include":
		extractor.handleInclude(node)
		return
	case "preproc_def", "preproc_function_def":
		extractor.handleMacro(node, context)
		return
	case "enumerator":
		extractor.handleEnumerator(node, context)
		return
	case "template_declaration":
		context.Template = true
	case "call_expression", "assignment_expression", "update_expression", "field_expression", "subscript_expression":
		if context.CallableID != "" {
			extractor.extractExpression(node, context)
			return
		}
	case "identifier":
		if context.CallableID != "" {
			extractor.addAccess(node, context, "reads", false)
		}
		return
	default:
		if context.CallableID != "" && isExpressionKind(node.Kind()) {
			extractor.extractExpression(node, context)
			return
		}
	}
	for _, child := range namedChildren(node) {
		extractor.walk(child, context)
	}
}

func (extractor *extractor) handleNamespace(node *tree_sitter.Node, context extractionContext) {
	nameNode := node.ChildByFieldName("name")
	name := nodeText(nameNode, extractor.source)
	if name == "" {
		name = anonymousName("namespace", node, extractor.source)
	}
	qualified := qualify(context.ContainerQualified, name)
	declaration := extractor.addDeclaration(node, context, "namespace", name, qualified, "", false, true, nil)
	body := node.ChildByFieldName("body")
	childContext := context
	childContext.ContainerID = declaration.ID
	childContext.ContainerQualified = declaration.QualifiedName
	for _, child := range namedChildren(body) {
		extractor.walk(child, childContext)
	}
}

func (extractor *extractor) handleType(node *tree_sitter.Node, context extractionContext) {
	tag := strings.TrimSuffix(node.Kind(), "_specifier")
	nameNode := node.ChildByFieldName("name")
	name := nodeText(nameNode, extractor.source)
	if name == "" {
		name = anonymousName(tag, node, extractor.source)
	}
	qualified := qualify(context.ContainerQualified, name)
	attributes := map[string]any{"tag": tag}
	declaration := extractor.addDeclaration(node, context, "type", name, qualified, "", false, true, attributes)
	if baseClause := firstDescendant(node, "base_class_clause"); baseClause != nil {
		for _, child := range namedChildren(baseClause) {
			if child.Kind() == "access_specifier" {
				continue
			}
			candidate := normalizeQualified(nodeText(child, extractor.source))
			if candidate == "" {
				continue
			}
			extractor.file.Inheritance = append(extractor.file.Inheritance, inheritanceObservation{
				SourceID: declaration.ID, SourceScope: context.ContainerQualified, Path: extractor.file.Path,
				Expression: nodeText(child, extractor.source), Candidate: candidate, Span: spanForNode(extractor.file.Path, child),
			})
		}
	}
	body := node.ChildByFieldName("body")
	childContext := context
	childContext.ContainerID = declaration.ID
	childContext.ContainerQualified = declaration.QualifiedName
	childContext.TypeID = declaration.ID
	childContext.TypeName = declaration.Name
	for _, child := range namedChildren(body) {
		if child.Kind() == "access_specifier" {
			continue
		}
		extractor.walk(child, childContext)
	}
}
