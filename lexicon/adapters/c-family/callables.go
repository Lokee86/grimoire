package main

import (
	"strings"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

func (extractor *extractor) handleFunction(node *tree_sitter.Node, context extractionContext, definition bool) {
	declarator := firstDescendant(node.ChildByFieldName("declarator"), "function_declarator")
	if declarator == nil {
		declarator = firstDescendant(node, "function_declarator")
	}
	if declarator == nil {
		return
	}
	nameText := declaratorName(declarator, extractor.source)
	if nameText == "" {
		return
	}
	name := lastQualifiedPart(nameText)
	kind := callableKind(name, context)
	qualified := qualify(context.ContainerQualified, nameText)
	signature := normalizeSpace(nodeText(declarator, extractor.source))
	attributes := map[string]any{"definition": definition}
	if context.CallableID == "" && context.TypeID == "" && !isHeaderPath(extractor.file.Path) && hasStorageClass(node, extractor.source, "static") {
		attributes["linkage"] = "internal"
	}
	if context.Template {
		attributes["template"] = true
	}
	declaration := extractor.addDeclaration(node, context, kind, name, qualified, signature, true, definition, attributes)
	extractor.addParameters(declarator, declaration)

	callableContext := context
	callableContext.ContainerID = declaration.ID
	callableContext.ContainerQualified = declaration.QualifiedName
	callableContext.CallableID = declaration.ID
	callableContext.CallableScope = declaration.QualifiedName
	for _, child := range namedChildren(node) {
		if sameNode(child, declarator) || child.Kind() == "primitive_type" || child.Kind() == "type_identifier" || child.Kind() == "storage_class_specifier" {
			continue
		}
		extractor.walk(child, callableContext)
	}
}

func (extractor *extractor) handleDeclaration(node *tree_sitter.Node, context extractionContext) {
	for _, declarator := range topLevelDeclarators(node) {
		if function := firstDescendant(declarator, "function_declarator"); function != nil {
			if context.TypeID != "" && (extractor.file.Language == "c" || isFunctionPointerDeclarator(function)) {
				extractor.handleVariableDeclarator(node, declarator, context)
				continue
			}
			extractor.handleFunctionDeclarator(node, function, context)
			continue
		}
		extractor.handleVariableDeclarator(node, declarator, context)
	}
}

func (extractor *extractor) handleFunctionDeclarator(node, declarator *tree_sitter.Node, context extractionContext) {
	nameText := declaratorName(declarator, extractor.source)
	if nameText == "" {
		return
	}
	name := lastQualifiedPart(nameText)
	kind := callableKind(name, context)
	qualified := qualify(context.ContainerQualified, nameText)
	attributes := map[string]any{"definition": false}
	if context.CallableID == "" && context.TypeID == "" && !isHeaderPath(extractor.file.Path) && hasStorageClass(node, extractor.source, "static") {
		attributes["linkage"] = "internal"
	}
	if strings.Contains(nodeText(node, extractor.source), "virtual") {
		attributes["virtual"] = true
	}
	declaration := extractor.addDeclaration(node, context, kind, name, qualified, normalizeSpace(nodeText(declarator, extractor.source)), true, false, attributes)
	extractor.addParameters(declarator, declaration)
}

func callableKind(name string, context extractionContext) string {
	if context.TypeID == "" {
		return "function"
	}
	if name == context.TypeName {
		return "constructor"
	}
	return "method"
}
