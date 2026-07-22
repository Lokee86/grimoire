import * as ts from "typescript";
import { expressionText, hasModifier, isStaticModuleSpecifier, spanFor } from "./contract";
import type { FileContext, FactStore, ImportName } from "./model";

export function recordImport(node: ts.ImportDeclaration, ownerId: string, context: FileContext, facts: FactStore): void {
  const source = isStaticModuleSpecifier(node.moduleSpecifier) ? node.moduleSpecifier.text : null;
  const names: ImportName[] = [];
  const clause = node.importClause;
  if (!clause) names.push({ imported: "*", local: "*", kind: "side-effect" });
  if (clause?.name) names.push({ imported: "default", local: clause.name.text, kind: "default" });
  if (clause?.namedBindings && ts.isNamespaceImport(clause.namedBindings)) {
    names.push({ imported: "*", local: clause.namedBindings.name.text, kind: "namespace" });
  } else if (clause?.namedBindings && ts.isNamedImports(clause.namedBindings)) {
    for (const element of clause.namedBindings.elements) {
      names.push({ imported: element.propertyName?.text ?? element.name.text, local: element.name.text, kind: "named" });
    }
  }
  const recordSpan = spanFor(node, context.sourceFile, context.relativePath);
  const identity = `${context.moduleKey}::import:${recordSpan.start_line}:${recordSpan.start_column}`;
  const id = facts.addNode("import", names.map((item) => item.imported).join(",") || "import", context.relativePath, identity, identity, recordSpan, { expression: expressionText(node, context.sourceFile) });
  facts.addEdge(ownerId, id, "defines", recordSpan);
  facts.imports.push({ nodeId: id, ownerId, moduleKey: context.moduleKey, sourceFile: context.sourceFile, source, expression: expressionText(node, context.sourceFile), names });
}

export function recordImportEquals(node: ts.ImportEqualsDeclaration, ownerId: string, context: FileContext, facts: FactStore): void {
  const recordSpan = spanFor(node, context.sourceFile, context.relativePath);
  const identity = `${context.moduleKey}::import:${recordSpan.start_line}:${recordSpan.start_column}`;
  const id = facts.addNode("import", node.name.text, context.relativePath, identity, identity, recordSpan, { expression: expressionText(node, context.sourceFile) });
  facts.addEdge(ownerId, id, "defines", recordSpan);
  const moduleReference = node.moduleReference;
  const source = ts.isExternalModuleReference(moduleReference) && isStaticModuleSpecifier(moduleReference.expression) ? moduleReference.expression.text : null;
  facts.imports.push({ nodeId: id, ownerId, moduleKey: context.moduleKey, sourceFile: context.sourceFile, source, expression: expressionText(node, context.sourceFile), names: [{ imported: "*", local: node.name.text, kind: "namespace" }] });
}

export function recordExportDeclaration(node: ts.ExportDeclaration, ownerId: string, context: FileContext, facts: FactStore): void {
  const recordSpan = spanFor(node, context.sourceFile, context.relativePath);
  const source = isStaticModuleSpecifier(node.moduleSpecifier) ? node.moduleSpecifier.text : null;
  const names = node.exportClause && ts.isNamedExports(node.exportClause)
    ? node.exportClause.elements.map((element) => element.propertyName?.text ?? element.name.text)
    : ["*"];
  const identity = `${context.moduleKey}::export:${recordSpan.start_line}:${recordSpan.start_column}`;
  const id = facts.addNode("export", names.join(","), context.relativePath, identity, identity, recordSpan, { expression: expressionText(node, context.sourceFile) });
  facts.addEdge(ownerId, id, "defines", recordSpan);
  if (source) {
    const names = node.exportClause && ts.isNamedExports(node.exportClause)
      ? node.exportClause.elements.map((element) => ({
        imported: element.propertyName?.text ?? element.name.text,
        exported: element.name.text,
      }))
      : [];
    facts.reexports.push({ ownerId, moduleKey: context.moduleKey, source, expression: expressionText(node, context.sourceFile), names, span: recordSpan });
  }
  else if (node.moduleSpecifier) facts.addUnresolved(ownerId, "imports", expressionText(node, context.sourceFile), "unsupported-form", recordSpan);
}

export function recordExportedDeclaration(ownerId: string, node: ts.Node, name: string, targetId: string, context: FileContext, facts: FactStore): void {
  const recordSpan = spanFor(node, context.sourceFile, context.relativePath);
  const identity = `${context.moduleKey}::export:${recordSpan.start_line}:${recordSpan.start_column}:${name}`;
  const id = facts.addNode("export", name, context.relativePath, identity, identity, recordSpan, { name, default: hasModifier(node, ts.SyntaxKind.DefaultKeyword) });
  facts.addEdge(ownerId, id, "defines", recordSpan);
  facts.addEdge(id, targetId, "defines", recordSpan);
}

export function recordExportAssignment(node: ts.ExportAssignment, ownerId: string, context: FileContext, facts: FactStore): void {
  if (!node.isExportEquals) facts.defaultExports.set(context.moduleKey, node.expression);
  const recordSpan = spanFor(node, context.sourceFile, context.relativePath);
  const name = node.isExportEquals ? "=" : "default";
  const identity = `${context.moduleKey}::export:${recordSpan.start_line}:${recordSpan.start_column}:${name}`;
  const id = facts.addNode("export", name, context.relativePath, identity, identity, recordSpan, { name, default: hasModifier(node, ts.SyntaxKind.DefaultKeyword) });
  facts.addEdge(ownerId, id, "defines", recordSpan);
}
