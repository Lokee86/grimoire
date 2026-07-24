package main

import (
	"fmt"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

func (extractor *extractor) handleTypedef(node *tree_sitter.Node, context extractionContext) {
	declarator := node.ChildByFieldName("declarator")
	name := declaratorName(declarator, extractor.source)
	if name != "" {
		qualified := qualify(context.ContainerQualified, name)
		target := nodeText(node.ChildByFieldName("type"), extractor.source)
		extractor.addDeclaration(node, context, "type", lastQualifiedPart(name), qualified, "", false, true, map[string]any{"alias": true, "target": normalizeSpace(target)})
	}
	if typeNode := node.ChildByFieldName("type"); typeNode != nil {
		extractor.walk(typeNode, context)
	}
}

func (extractor *extractor) handleAlias(node *tree_sitter.Node, context extractionContext) {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		nameNode = firstDescendant(node, "type_identifier")
	}
	name := nodeText(nameNode, extractor.source)
	if name == "" {
		return
	}
	extractor.addDeclaration(node, context, "type", name, qualify(context.ContainerQualified, name), "", false, true, map[string]any{"alias": true})
}

func (extractor *extractor) handleMacro(node *tree_sitter.Node, context extractionContext) {
	nameNode := node.ChildByFieldName("name")
	name := nodeText(nameNode, extractor.source)
	if name == "" {
		return
	}
	attributes := map[string]any{"macro": true, "function_like": node.Kind() == "preproc_function_def"}
	extractor.addDeclaration(node, context, "symbol", name, qualify(context.ContainerQualified, name), "", false, true, attributes)
}

func (extractor *extractor) addDeclaration(node *tree_sitter.Node, context extractionContext, kind, name, qualified, signature string, callable, definition bool, attributes map[string]any) *declaration {
	canonical := extractor.file.Path + "::" + kind + "::" + qualified
	if signature != "" {
		canonical += "::" + signature
	}
	if attributes == nil {
		attributes = map[string]any{}
	}
	attributes["language"] = extractor.file.Language
	declaration := &declaration{
		ID: nodeID(kind, canonical), Kind: kind, Name: name, QualifiedName: qualified, Path: extractor.file.Path,
		ContainerID: context.ContainerID, ContainerQualified: context.ContainerQualified, ParentTypeID: context.TypeID,
		Signature: signature, FileLanguage: extractor.file.Language, Span: spanForNode(extractor.file.Path, node),
		Attributes: attributes, Callable: callable, Definition: definition,
	}
	extractor.file.Declarations = append(extractor.file.Declarations, declaration)
	return declaration
}

func (extractor *extractor) handleInclude(node *tree_sitter.Node) {
	pathNode := node.ChildByFieldName("path")
	expression := nodeText(pathNode, extractor.source)
	target := stripIncludeTarget(expression)
	if target == "" {
		return
	}
	canonical := fmt.Sprintf("%s::include::%s::%d", extractor.file.Path, target, node.StartByte())
	extractor.file.Includes = append(extractor.file.Includes, includeObservation{
		ID: nodeID("import", canonical), ModuleID: extractor.file.ModuleID, Path: extractor.file.Path,
		Target: target, Expression: expression, System: pathNode != nil && pathNode.Kind() == "system_lib_string",
		Span: spanForNode(extractor.file.Path, node),
	})
}
