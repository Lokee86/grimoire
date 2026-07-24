import * as path from "node:path";
import * as ts from "typescript";
import {
  callCallee,
  callTargetName,
  symbolLocation,
  unwrapExpression,
  type ParameterTargets,
} from "./call-shared";
import type { FactStore, PendingCall } from "./model";

type UnresolvedCallReason =
  | "missing-target"
  | "external-target"
  | "unsupported-form"
  | "dynamic-target"
  | "builtin-target";

export function multiTargetCallReason(
  checker: ts.TypeChecker,
  call: PendingCall,
): "ambiguous-target" | "dynamic-target" {
  const declarations = callSymbolDeclarations(checker, call);
  return declarations.some(dynamicValueDeclaration) ? "dynamic-target" : "ambiguous-target";
}

export function unresolvedCallReason(
  facts: FactStore,
  checker: ts.TypeChecker,
  call: PendingCall,
  parameterTargets: ParameterTargets,
): UnresolvedCallReason {
  if (call.kind === "jsx") {
    const tagName = (call.expression as ts.JsxOpeningLikeElement).tagName;
    if (ts.isIdentifier(tagName) && tagName.text[0] === tagName.text[0]?.toLowerCase()) return "builtin-target";
  }
  const callee = call.kind === "jsx" ? (call.expression as ts.JsxOpeningLikeElement).tagName : callCallee(call);
  if (!callee) return "unsupported-form";
  const symbol = checker.getSymbolAtLocation(symbolLocation(callee as ts.Expression));
  if (symbol) {
    const resolved = (symbol.flags & ts.SymbolFlags.Alias) !== 0 ? checker.getAliasedSymbol(symbol) : symbol;
    const declarations = resolved.declarations ?? [];
    if (declarations.some((declaration) => !isRepositoryDeclaration(facts, declaration))) return "external-target";
    if (
      call.kind === "constructor"
      && declarations.length > 0
      && !declarations.some((declaration) => ts.isClassDeclaration(declaration) || ts.isClassExpression(declaration))
    ) return "dynamic-target";
    if (declarations.some((declaration) => declarationHasExternalOrigin(facts, checker, declaration))) {
      return "external-target";
    }
    const parameterBindings = declarations
      .map(enclosingParameterBinding)
      .filter((binding): binding is ts.ParameterDeclaration | ts.BindingElement => !!binding);
    if (parameterBindings.length > 0) {
      if (parameterBindings.some((binding) => parameterTargets.get(binding)?.size)) return "dynamic-target";
      const parameters = parameterBindings
        .map((binding) => ts.isParameter(binding) ? binding : enclosingParameter(binding))
        .filter((parameter): parameter is ts.ParameterDeclaration => !!parameter);
      if (parameters.some((parameter) => parameterHasExternalCallableType(checker, parameter))) return "external-target";
      return "dynamic-target";
    }
    if (declarations.some(dynamicValueDeclaration)) return "dynamic-target";
    if (declarations.some((declaration) => isRepositoryDeclaration(facts, declaration))) return "missing-target";
  }
  const name = callTargetName(call)?.split(/[.[]/, 1)[0];
  const binding = name ? facts.bindings.get(call.moduleKey)?.get(name) : undefined;
  if (binding?.external) return "external-target";
  if (name && binding && !binding.targetId && importSourceIsMissing(facts, call.moduleKey, name)) return "missing-target";
  const type = checker.getTypeAtLocation(callee);
  if ((type.flags & (ts.TypeFlags.Any | ts.TypeFlags.Unknown)) !== 0) return "dynamic-target";
  if (ts.isElementAccessExpression(callee as ts.Expression)) return "dynamic-target";
  return binding && !binding.targetId ? "missing-target" : "unsupported-form";
}

function callSymbolDeclarations(checker: ts.TypeChecker, call: PendingCall): ts.Declaration[] {
  const reference = call.kind === "jsx"
    ? (call.expression as ts.JsxOpeningLikeElement).tagName
    : callCallee(call);
  if (!reference) return [];
  const symbol = checker.getSymbolAtLocation(symbolLocation(reference as ts.Expression));
  if (!symbol) return [];
  const resolved = (symbol.flags & ts.SymbolFlags.Alias) !== 0 ? checker.getAliasedSymbol(symbol) : symbol;
  return resolved.declarations ?? [];
}

function importSourceIsMissing(facts: FactStore, moduleKey: string, localName: string): boolean {
  const info = facts.imports.find(
    (candidate) => candidate.moduleKey === moduleKey && candidate.names.some((name) => name.local === localName),
  );
  if (!info?.source || !info.source.startsWith(".")) return false;
  const base = path.posix.normalize(path.posix.join(path.posix.dirname(moduleKey), info.source));
  return !facts.modules.has(base) && !facts.modules.has(`${base}/index`);
}

function declarationHasExternalOrigin(
  facts: FactStore,
  checker: ts.TypeChecker,
  declaration: ts.Node,
  seen = new Set<ts.Node>(),
): boolean {
  if (seen.has(declaration)) return false;
  const expression = declarationValueExpression(declaration);
  if (!expression) return false;
  const nextSeen = new Set(seen);
  nextSeen.add(declaration);
  return expressionHasExternalOrigin(facts, checker, unwrapExpression(expression), nextSeen);
}

function declarationValueExpression(declaration: ts.Node): ts.Expression | null {
  if (
    ts.isBindingElement(declaration)
    || ts.isVariableDeclaration(declaration)
    || ts.isPropertyDeclaration(declaration)
    || ts.isParameter(declaration)
  ) return declaration.initializer ?? null;
  if (ts.isPropertyAssignment(declaration)) return declaration.initializer;
  if (ts.isShorthandPropertyAssignment(declaration)) return declaration.name;
  if (ts.isBinaryExpression(declaration) && declaration.operatorToken.kind === ts.SyntaxKind.EqualsToken) {
    return declaration.right;
  }
  return null;
}

function expressionHasExternalOrigin(
  facts: FactStore,
  checker: ts.TypeChecker,
  expression: ts.Expression,
  seen: Set<ts.Node>,
): boolean {
  const symbol = checker.getSymbolAtLocation(symbolLocation(expression));
  if (symbolIsExternal(facts, checker, symbol)) return true;
  if (symbol) {
    const resolved = (symbol.flags & ts.SymbolFlags.Alias) !== 0 ? checker.getAliasedSymbol(symbol) : symbol;
    if ((resolved.declarations ?? []).some((declaration) => declarationHasExternalOrigin(facts, checker, declaration, seen))) {
      return true;
    }
  }
  if (ts.isPropertyAccessExpression(expression) || ts.isElementAccessExpression(expression)) {
    return expressionHasExternalOrigin(facts, checker, expression.expression, seen);
  }
  if (ts.isCallExpression(expression) || ts.isNewExpression(expression)) {
    return expressionHasExternalOrigin(facts, checker, expression.expression, seen);
  }
  return false;
}

function dynamicValueDeclaration(declaration: ts.Node): boolean {
  return ts.isParameter(declaration)
    || ts.isBindingElement(declaration)
    || ts.isVariableDeclaration(declaration)
    || ts.isPropertyDeclaration(declaration)
    || ts.isPropertyAssignment(declaration)
    || ts.isShorthandPropertyAssignment(declaration)
    || (ts.isBinaryExpression(declaration) && declaration.operatorToken.kind === ts.SyntaxKind.EqualsToken);
}

function symbolIsExternal(facts: FactStore, checker: ts.TypeChecker, symbol: ts.Symbol | undefined): boolean {
  if (!symbol) return false;
  if ((symbol.flags & ts.SymbolFlags.Alias) !== 0) {
    const aliased = checker.getAliasedSymbol(symbol);
    const aliasedDeclarations = aliased.declarations ?? [];
    if (aliasedDeclarations.some((declaration) => !isRepositoryDeclaration(facts, declaration))) return true;
    if (aliasedDeclarations.length === 0) {
      for (const declaration of symbol.declarations ?? []) {
        const importDeclaration = findImportDeclaration(declaration);
        const source = importDeclaration && ts.isStringLiteralLike(importDeclaration.moduleSpecifier)
          ? importDeclaration.moduleSpecifier.text
          : null;
        if (source && !source.startsWith(".")) return true;
      }
    }
  }
  return (symbol.declarations ?? []).some((declaration) => !isRepositoryDeclaration(facts, declaration));
}

function findImportDeclaration(node: ts.Node): ts.ImportDeclaration | null {
  let current: ts.Node | undefined = node;
  while (current && !ts.isSourceFile(current)) {
    if (ts.isImportDeclaration(current)) return current;
    current = current.parent;
  }
  return null;
}

function enclosingParameterBinding(node: ts.Node): ts.ParameterDeclaration | ts.BindingElement | null {
  let current: ts.Node | undefined = node;
  while (current && !ts.isSourceFile(current)) {
    if (ts.isBindingElement(current) || ts.isParameter(current)) return current;
    current = current.parent;
  }
  return null;
}

function enclosingParameter(node: ts.Node): ts.ParameterDeclaration | null {
  let current: ts.Node | undefined = node;
  while (current && !ts.isSourceFile(current)) {
    if (ts.isParameter(current)) return current;
    current = current.parent;
  }
  return null;
}

function parameterHasExternalCallableType(checker: ts.TypeChecker, parameter: ts.ParameterDeclaration): boolean {
  const signatures = checker.getTypeAtLocation(parameter).getCallSignatures();
  return signatures.length > 0 && signatures.every((signature) => {
    const declaration = signature.getDeclaration();
    return declaration ? declaration.getSourceFile().isDeclarationFile : false;
  });
}

function isRepositoryDeclaration(facts: FactStore, declaration: ts.Node): boolean {
  const fileName = path.resolve(declaration.getSourceFile().fileName);
  const relative = path.relative(facts.root, fileName);
  return relative !== ""
    && !relative.startsWith(`..${path.sep}`)
    && !path.isAbsolute(relative)
    && !relative.split(path.sep).includes("node_modules");
}
