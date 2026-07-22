import * as fs from "node:fs";
import * as path from "node:path";
import * as ts from "typescript";
import { declarationName, digest, expressionText, hasModifier, spanFor } from "./contract";
import { moduleKeyFor, normalizeRelative } from "./discovery";
import { recordExportAssignment, recordExportDeclaration, recordExportedDeclaration, recordImport, recordImportEquals } from "./ast-imports";
import type { FactStore, FileContext, PendingCall } from "./model";

export function createFileContext(root: string, relativePath: string, directoryIds: Map<string, string>, facts: FactStore): FileContext {
  const absolutePath = path.join(root, relativePath.split("/").join(path.sep));
  const sourceBytes = fs.readFileSync(absolutePath);
  const scriptKind = relativePath.toLowerCase().endsWith(".tsx") ? ts.ScriptKind.TSX : ts.ScriptKind.TS;
  const sourceFile = ts.createSourceFile(relativePath, sourceBytes.toString("utf8"), ts.ScriptTarget.Latest, true, scriptKind);
  const fileSpan = spanFor(sourceFile, sourceFile, relativePath);
  const fileId = facts.addNode("file", path.posix.basename(relativePath), relativePath, relativePath, relativePath, fileSpan, undefined, digest(sourceBytes));
  const parent = path.posix.dirname(relativePath) || ".";
  facts.addEdge(directoryIds.get(parent) ?? directoryIds.get(".")!, fileId, "contains");
  const moduleKey = moduleKeyFor(relativePath);
  const moduleId = facts.addNode("module", path.posix.basename(moduleKey), relativePath, moduleKey, moduleKey);
  facts.modules.set(moduleKey, moduleId);
  facts.addEdge(fileId, moduleId, "contains", fileSpan);
  if (((sourceFile as ts.SourceFile & { parseDiagnostics?: ts.Diagnostic[] }).parseDiagnostics ?? []).length > 0) facts.addUnresolved(moduleId, "parses", relativePath, "unsupported-form", undefined, "syntax-diagnostics");
  return { relativePath, moduleKey, sourceFile, fileId, moduleId };
}

export function extractDeclarations(context: FileContext, facts: FactStore): void {
  visit(context.sourceFile, context.moduleId, [], context, facts);
}

function visit(node: ts.Node, ownerId: string, scope: string[], context: FileContext, facts: FactStore): void {
  if (ts.isSourceFile(node)) {
    node.forEachChild((child) => visit(child, ownerId, scope, context, facts));
    return;
  }
  if (ts.isImportDeclaration(node)) return recordImport(node, ownerId, context, facts);
  if (ts.isImportEqualsDeclaration(node)) return recordImportEquals(node, ownerId, context, facts);
  if (ts.isExportDeclaration(node)) return recordExportDeclaration(node, ownerId, context, facts);
  if (ts.isExportAssignment(node)) return recordExportAssignment(node, ownerId, context, facts);
  if (ts.isClassDeclaration(node)) return visitClass(node, ownerId, scope, context, facts);
  if (ts.isInterfaceDeclaration(node)) return visitInterface(node, ownerId, scope, context, facts);
  if (ts.isTypeAliasDeclaration(node)) return visitTypeAlias(node, ownerId, scope, context, facts);
  if (ts.isFunctionDeclaration(node)) return visitFunction(node, ownerId, scope, context, facts);
  if (ts.isMethodDeclaration(node) || ts.isMethodSignature(node)) return visitMethod(node, ownerId, scope, context, facts);
  if (ts.isConstructorDeclaration(node)) return visitMethod(node, ownerId, scope, context, facts, true);
  if (ts.isPropertyDeclaration(node) || ts.isPropertySignature(node)) return visitProperty(node, ownerId, scope, context, facts);
  if (ts.isVariableStatement(node)) return visitVariableList(node.declarationList, ownerId, scope, context, facts);
  if (ts.isVariableDeclarationList(node) && !ts.isVariableStatement(node.parent)) return visitVariableList(node, ownerId, scope, context, facts);
  if (ts.isNewExpression(node)) recordCall(node, "constructor", ownerId, scope, context, facts);
  else if (ts.isCallExpression(node)) {
    if (node.expression.kind === ts.SyntaxKind.ImportKeyword) facts.addUnresolved(ownerId, "imports", expressionText(node, context.sourceFile), "dynamic-target", spanFor(node, context.sourceFile, context.relativePath));
    else recordCall(node, "call", ownerId, scope, context, facts);
  }
  node.forEachChild((child) => visit(child, ownerId, scope, context, facts));
}

function recordCall(expression: PendingCall["expression"], kind: PendingCall["kind"], source: string, scope: string[], context: FileContext, facts: FactStore): void {
  facts.calls.push({ expression, kind, moduleKey: context.moduleKey, scope, source, sourceFile: context.sourceFile });
}

function addSymbol(kind: string, name: string, scope: string[], node: ts.Node, ownerId: string, context: FileContext, facts: FactStore): string {
  const qualifiedName = `${context.moduleKey}.${[...scope, name].join(".")}`;
  const id = facts.addNode(kind, name, context.relativePath, qualifiedName, qualifiedName, spanFor(node, context.sourceFile, context.relativePath));
  if (facts.symbols.has(qualifiedName)) facts.ambiguousSymbols.add(qualifiedName);
  facts.symbols.set(qualifiedName, id);
  facts.addEdge(ownerId, id, "defines", spanFor(node, context.sourceFile, context.relativePath));
  return id;
}

function visitClass(node: ts.ClassDeclaration, ownerId: string, scope: string[], context: FileContext, facts: FactStore): void {
  const name = declarationName(node.name) ?? "<anonymous>";
  const id = addSymbol("type", name, scope, node, ownerId, context, facts);
  if (hasModifier(node, ts.SyntaxKind.ExportKeyword)) recordExportedDeclaration(ownerId, node, name, id, context, facts);
  addHeritage(node, id, scope, context, facts);
  node.members.forEach((member) => visit(member, id, [...scope, name], context, facts));
}

function visitInterface(node: ts.InterfaceDeclaration, ownerId: string, scope: string[], context: FileContext, facts: FactStore): void {
  const name = node.name.text;
  const id = addSymbol("interface", name, scope, node, ownerId, context, facts);
  if (hasModifier(node, ts.SyntaxKind.ExportKeyword)) recordExportedDeclaration(ownerId, node, name, id, context, facts);
  addHeritage(node, id, scope, context, facts);
  node.members.forEach((member) => visit(member, id, [...scope, name], context, facts));
}

function addHeritage(node: ts.ClassDeclaration | ts.InterfaceDeclaration, sourceId: string, scope: string[], context: FileContext, facts: FactStore): void {
  for (const clause of node.heritageClauses ?? []) {
    const relation = clause.token === ts.SyntaxKind.ExtendsKeyword ? "extends" : "implements";
    for (const type of clause.types) facts.relationships.push({ source: sourceId, relation, expression: type.expression, sourceFile: context.sourceFile, moduleKey: context.moduleKey, scope });
  }
}

function visitTypeAlias(node: ts.TypeAliasDeclaration, ownerId: string, scope: string[], context: FileContext, facts: FactStore): void {
  const name = declarationName(node.name) ?? "<anonymous>";
  const id = addSymbol("type", name, scope, node, ownerId, context, facts);
  if (hasModifier(node, ts.SyntaxKind.ExportKeyword)) recordExportedDeclaration(ownerId, node, name, id, context, facts);
  node.forEachChild((child) => visit(child, ownerId, scope, context, facts));
}

function visitFunction(node: ts.FunctionDeclaration, ownerId: string, scope: string[], context: FileContext, facts: FactStore): void {
  const name = declarationName(node.name) ?? "<anonymous>";
  const id = addSymbol("function", name, scope, node, ownerId, context, facts);
  if (hasModifier(node, ts.SyntaxKind.ExportKeyword)) recordExportedDeclaration(ownerId, node, name, id, context, facts);
  node.forEachChild((child) => visit(child, id, [...scope, name], context, facts));
}

function visitMethod(node: ts.MethodDeclaration | ts.MethodSignature | ts.ConstructorDeclaration, ownerId: string, scope: string[], context: FileContext, facts: FactStore, constructor = false): void {
  const name = constructor ? "constructor" : declarationName(node.name) ?? "<computed>";
  const id = addSymbol(constructor ? "constructor" : "method", name, scope, node, ownerId, context, facts);
  node.forEachChild((child) => visit(child, id, [...scope, name], context, facts));
}

function visitProperty(node: ts.PropertyDeclaration | ts.PropertySignature, ownerId: string, scope: string[], context: FileContext, facts: FactStore): void {
  const name = declarationName(node.name) ?? "<computed>";
  addSymbol("field", name, scope, node, ownerId, context, facts);
  node.forEachChild((child) => visit(child, ownerId, scope, context, facts));
}

function visitVariableList(list: ts.VariableDeclarationList, ownerId: string, scope: string[], context: FileContext, facts: FactStore): void {
  const kind = (list.flags & ts.NodeFlags.Const) !== 0 ? "constant" : "variable";
  for (const declaration of list.declarations) {
    const name = declarationName(declaration.name) ?? "<computed>";
    const id = addSymbol(kind, name, scope, declaration, ownerId, context, facts);
    if (hasModifier(list.parent, ts.SyntaxKind.ExportKeyword)) recordExportedDeclaration(ownerId, list.parent, name, id, context, facts);
    declaration.forEachChild((child) => visit(child, ownerId, [...scope, name], context, facts));
  }
}
