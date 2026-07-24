package main

import (
	"fmt"
	"strings"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

func topLevelDeclarators(node *tree_sitter.Node) []*tree_sitter.Node {
	var result []*tree_sitter.Node
	for _, child := range namedChildren(node) {
		switch child.Kind() {
		case "primitive_type", "type_identifier", "sized_type_specifier", "struct_specifier", "union_specifier", "enum_specifier", "class_specifier", "storage_class_specifier", "type_qualifier", "attribute_specifier", "attribute_declaration", "access_specifier":
			continue
		}
		if isDeclaratorNode(child.Kind()) {
			result = append(result, child)
		}
	}
	return result
}

func isDeclaratorNode(kind string) bool {
	return kind == "identifier" || kind == "field_identifier" || kind == "type_identifier" ||
		strings.HasSuffix(kind, "_declarator") || kind == "init_declarator"
}

func declaratorName(node *tree_sitter.Node, source []byte) string {
	if node == nil {
		return ""
	}
	switch node.Kind() {
	case "identifier", "field_identifier", "type_identifier", "namespace_identifier", "operator_name", "destructor_name":
		return nodeText(node, source)
	case "qualified_identifier", "scoped_identifier":
		return normalizeQualified(nodeText(node, source))
	}
	if child := node.ChildByFieldName("declarator"); child != nil {
		if name := declaratorName(child, source); name != "" {
			return name
		}
	}
	if child := node.ChildByFieldName("name"); child != nil {
		if name := declaratorName(child, source); name != "" {
			return name
		}
	}
	for _, child := range namedChildren(node) {
		if name := declaratorName(child, source); name != "" {
			return name
		}
	}
	return ""
}

func (extractor *extractor) handleVariableDeclarator(node, declarator *tree_sitter.Node, context extractionContext) {
	name := declaratorName(declarator, extractor.source)
	if name == "" {
		return
	}
	name = lastQualifiedPart(name)
	kind := "variable"
	if context.CallableID == "" && context.TypeID != "" {
		kind = "field"
	} else if strings.Contains(nodeText(node, extractor.source), "const") || strings.Contains(nodeText(node, extractor.source), "constexpr") {
		kind = "constant"
	}
	qualified := qualify(context.ContainerQualified, name)
	signature := ""
	if context.CallableID != "" {
		signature = fmt.Sprintf("local@%d", declarator.StartByte())
	}
	attributes := map[string]any{"type": declarationType(node, extractor.source)}
	declaration := extractor.addDeclaration(declarator, context, kind, name, qualified, signature, false, true, attributes)
	if context.CallableID != "" && declarator.Kind() == "init_declarator" {
		extractor.file.Accesses = append(extractor.file.Accesses, accessObservation{
			SourceID: context.CallableID, SourceScope: context.CallableScope, ParentType: context.TypeID,
			Path: extractor.file.Path, Expression: name, Candidate: name, Relation: "writes",
			Span: declaration.Span,
		})
	}
	for _, child := range namedChildren(declarator) {
		if child.Kind() == "identifier" || child.Kind() == "field_identifier" || child.Kind() == "type_identifier" {
			continue
		}
		if isExpressionKind(child.Kind()) || child.Kind() == "initializer_list" {
			extractor.extractExpression(child, context)
		}
	}
}

func declarationType(node *tree_sitter.Node, source []byte) string {
	if typeNode := node.ChildByFieldName("type"); typeNode != nil {
		return normalizeSpace(nodeText(typeNode, source))
	}
	for _, child := range namedChildren(node) {
		switch child.Kind() {
		case "primitive_type", "type_identifier", "sized_type_specifier", "struct_specifier", "union_specifier", "enum_specifier", "class_specifier":
			return normalizeSpace(nodeText(child, source))
		}
	}
	return ""
}

func (extractor *extractor) addParameters(declarator *tree_sitter.Node, callable *declaration) {
	parameterList := firstDescendant(declarator, "parameter_list")
	if parameterList == nil {
		return
	}
	index := 0
	for _, child := range namedChildren(parameterList) {
		if child.Kind() != "parameter_declaration" && child.Kind() != "optional_parameter_declaration" {
			continue
		}
		name := declaratorName(child.ChildByFieldName("declarator"), extractor.source)
		if name == "" {
			index++
			continue
		}
		name = lastQualifiedPart(name)
		context := extractionContext{
			ContainerID: callable.ID, ContainerQualified: callable.QualifiedName,
			TypeID: callable.ParentTypeID, CallableID: callable.ID, CallableScope: callable.QualifiedName,
		}
		attributes := map[string]any{"index": index, "type": declarationType(child, extractor.source)}
		extractor.addDeclaration(child, context, "parameter", name, callable.QualifiedName+"::"+name, fmt.Sprintf("parameter#%d", index), false, true, attributes)
		index++
	}
	callable.Attributes["parameter_count"] = index
}

func (extractor *extractor) handleEnumerator(node *tree_sitter.Node, context extractionContext) {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		nameNode = firstDescendant(node, "identifier")
	}
	name := nodeText(nameNode, extractor.source)
	if name == "" {
		return
	}
	extractor.addDeclaration(node, context, "constant", name, qualify(context.ContainerQualified, name), "", false, true, map[string]any{"enum_member": true})
}

func sameNode(left, right *tree_sitter.Node) bool {
	if left == nil || right == nil {
		return false
	}
	return left.StartByte() == right.StartByte() && left.EndByte() == right.EndByte() && left.Kind() == right.Kind()
}
