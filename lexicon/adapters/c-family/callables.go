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
	extractor.recordCallableParameters(declarator, declaration)

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
			if isFunctionPointerDeclarator(function) || context.TypeID != "" && extractor.file.Language == "c" {
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
	extractor.recordCallableParameters(declarator, declaration)
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

func (extractor *extractor) callableParameterShape(declarator *tree_sitter.Node) *callableParameterShape {
	parameterList := firstDescendant(declarator, "parameter_list")
	if parameterList == nil {
		return nil
	}
	if parameterListIsVoid(parameterList, extractor.source) {
		return &callableParameterShape{}
	}
	if extractor.file.Language == "c" && !parameterListHasNonCommentChild(parameterList) {
		return nil
	}
	parameterText := normalizeSpace(nodeText(parameterList, extractor.source))
	trailingVariadic := parameterListHasTrailingVariadic(parameterList, extractor.source)
	if strings.Contains(parameterText, "...") && !trailingVariadic {
		return nil
	}

	minimum, maximum := 0, 0
	optional := false
	variadic := false
	for _, child := range namedChildren(parameterList) {
		switch child.Kind() {
		case "comment":
			continue
		case "variadic_parameter":
			if variadic || optional {
				return nil
			}
			variadic = true
		case "parameter_declaration":
			if variadic || optional || child.ChildByFieldName("type") == nil {
				return nil
			}
			minimum++
			maximum++
		case "optional_parameter_declaration":
			if variadic || child.ChildByFieldName("type") == nil || child.ChildByFieldName("default_value") == nil {
				return nil
			}
			optional = true
			maximum++
		default:
			return nil
		}
	}
	if trailingVariadic {
		if optional {
			return nil
		}
		variadic = true
	}
	if variadic {
		return &callableParameterShape{Minimum: minimum, Maximum: -1, Variadic: true}
	}
	return &callableParameterShape{Minimum: minimum, Maximum: maximum}
}

func parameterListHasTrailingVariadic(parameterList *tree_sitter.Node, source []byte) bool {
	text := normalizeSpace(nodeText(parameterList, source))
	if len(text) < 2 || text[0] != '(' || text[len(text)-1] != ')' {
		return false
	}
	parameters := strings.TrimSpace(text[1 : len(text)-1])
	return parameters == "..." || strings.HasSuffix(parameters, ",...") || strings.HasSuffix(parameters, ", ...")
}

func parameterListHasNonCommentChild(parameterList *tree_sitter.Node) bool {
	for _, child := range namedChildren(parameterList) {
		if child.Kind() != "comment" {
			return true
		}
	}
	return false
}

func parameterListIsVoid(parameterList *tree_sitter.Node, source []byte) bool {
	var parameter *tree_sitter.Node
	for _, child := range namedChildren(parameterList) {
		if child.Kind() == "comment" {
			continue
		}
		if parameter != nil || child.Kind() != "parameter_declaration" {
			return false
		}
		parameter = child
	}
	return parameter != nil && parameter.ChildByFieldName("declarator") == nil &&
		normalizeSpace(nodeText(parameter.ChildByFieldName("type"), source)) == "void"
}
