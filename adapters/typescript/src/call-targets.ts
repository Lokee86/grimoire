import * as ts from "typescript";
import {
  callCallee,
  isSignatureDeclaration,
  mergeSets,
  symbolLocation,
  unwrapExpression,
  type ParameterTargets,
} from "./call-shared";
import type { FactStore, PendingCall } from "./model";

export function resolveCallTargets(
  facts: FactStore,
  checker: ts.TypeChecker,
  call: PendingCall,
  parameterTargets: ParameterTargets,
): Set<string> {
  const targets = new Set<string>();
  if (call.kind === "jsx") {
    const tagName = (call.expression as ts.JsxOpeningLikeElement).tagName;
    addSymbolTargets(facts, checker, checker.getSymbolAtLocation(tagName), targets, parameterTargets, true);
    if (targets.size === 0) addDefaultImportTarget(facts, checker, call, targets, parameterTargets);
    return targets;
  }

  const expression = call.expression as ts.CallExpression | ts.NewExpression | ts.TaggedTemplateExpression;
  const resolved = checker.getResolvedSignature(expression);
  if (resolved) addSignatureTarget(facts, checker, resolved, targets, parameterTargets, call.kind === "constructor");
  const callee = callCallee(call);
  if (!callee) return targets;
  const type = checker.getTypeAtLocation(callee);
  const signatures = call.kind === "constructor" ? type.getConstructSignatures() : type.getCallSignatures();
  for (const signature of signatures) {
    addSignatureTarget(facts, checker, signature, targets, parameterTargets, call.kind === "constructor");
  }
  addSymbolTargets(
    facts,
    checker,
    checker.getSymbolAtLocation(symbolLocation(callee)),
    targets,
    parameterTargets,
    call.kind === "constructor",
  );
  if (targets.size === 0) addDefaultImportTarget(facts, checker, call, targets, parameterTargets);
  normalizeConstructorTargets(facts, call, targets);
  return expandDispatchTargets(facts, checker, call, targets);
}

function normalizeConstructorTargets(facts: FactStore, call: PendingCall, targets: Set<string>): void {
  if (call.kind !== "constructor") return;
  const hasConstructor = [...targets].some((target) => facts.nodes.get(target)?.kind === "constructor");
  if (!hasConstructor) return;
  for (const target of [...targets]) if (facts.nodes.get(target)?.kind === "type") targets.delete(target);
}

function addDefaultImportTarget(
  facts: FactStore,
  checker: ts.TypeChecker,
  call: PendingCall,
  targets: Set<string>,
  parameterTargets: ParameterTargets,
): void {
  const reference = call.kind === "jsx"
    ? (call.expression as ts.JsxOpeningLikeElement).tagName
    : callCallee(call);
  if (!reference || !ts.isIdentifier(reference)) return;
  const binding = facts.bindings.get(call.moduleKey)?.get(reference.text);
  if (!binding?.targetId || facts.nodes.get(binding.targetId)?.kind !== "module") return;
  const moduleKey = [...facts.modules].find(([, id]) => id === binding.targetId)?.[0];
  const exported = moduleKey ? facts.defaultExports.get(moduleKey) : undefined;
  if (!exported) return;
  for (const target of resolveExpressionTargets(facts, checker, exported, parameterTargets)) targets.add(target);
}

function expandDispatchTargets(
  facts: FactStore,
  checker: ts.TypeChecker,
  call: PendingCall,
  targets: Set<string>,
): Set<string> {
  if (call.kind !== "call") return targets;
  const callee = unwrapExpression((call.expression as ts.CallExpression).expression);
  const access = ts.isPropertyAccessExpression(callee) || ts.isElementAccessExpression(callee) ? callee : null;
  if (!access || ts.isNewExpression(unwrapExpression(access.expression))) return targets;
  const methodName = ts.isPropertyAccessExpression(access)
    ? access.name.text
    : ts.isStringLiteralLike(access.argumentExpression)
      ? access.argumentExpression.text
      : null;
  if (!methodName) return targets;
  const receiverType = checker.getTypeAtLocation(access.expression);
  if ((receiverType.flags & (ts.TypeFlags.Any | ts.TypeFlags.Unknown)) !== 0) return targets;
  const implementations = new Set<string>();
  for (const [target, declarations] of facts.idDeclarations) {
    if (facts.nodes.get(target)?.kind !== "method") continue;
    for (const declaration of declarations) {
      if (!ts.isMethodDeclaration(declaration) || declaration.name.getText(declaration.getSourceFile()) !== methodName) continue;
      if ((ts.getModifiers(declaration) ?? []).some((modifier) => modifier.kind === ts.SyntaxKind.StaticKeyword)) continue;
      const owner = declaration.parent;
      if (!ts.isClassDeclaration(owner) || !owner.name) continue;
      const symbol = checker.getSymbolAtLocation(owner.name);
      if (!symbol) continue;
      const instanceType = checker.getDeclaredTypeOfSymbol(symbol);
      if (receiverAcceptsCandidate(checker, receiverType, instanceType)) implementations.add(target);
    }
  }
  if (implementations.size === 0) return targets;
  const result = new Set<string>();
  for (const target of targets) {
    const declarations = facts.idDeclarations.get(target) ?? [];
    if (!declarations.some(ts.isMethodSignature)) result.add(target);
  }
  for (const target of implementations) result.add(target);
  return result;
}

function receiverAcceptsCandidate(checker: ts.TypeChecker, receiverType: ts.Type, candidateType: ts.Type): boolean {
  const receiverSymbol = receiverType.getSymbol();
  if (receiverSymbol && (receiverSymbol.flags & ts.SymbolFlags.Class) !== 0) {
    return classTypeDerivesFrom(candidateType, receiverSymbol, new Set<ts.Symbol>());
  }
  return checker.isTypeAssignableTo(candidateType, receiverType);
}

function classTypeDerivesFrom(type: ts.Type, expected: ts.Symbol, seen: Set<ts.Symbol>): boolean {
  const symbol = type.getSymbol();
  if (!symbol || seen.has(symbol)) return false;
  if (symbol === expected) return true;
  seen.add(symbol);
  if ((type.flags & ts.TypeFlags.Object) === 0) return false;
  return ((type as ts.InterfaceType).getBaseTypes() ?? []).some((base) => classTypeDerivesFrom(base, expected, seen));
}

export function resolveExpressionTargets(
  facts: FactStore,
  checker: ts.TypeChecker,
  expression: ts.Expression,
  parameterTargets: ParameterTargets,
): Set<string> {
  const unwrapped = unwrapExpression(expression);
  if (ts.isConditionalExpression(unwrapped)) {
    return mergeSets(
      resolveExpressionTargets(facts, checker, unwrapped.whenTrue, parameterTargets),
      resolveExpressionTargets(facts, checker, unwrapped.whenFalse, parameterTargets),
    );
  }
  const targets = new Set<string>();
  const direct = facts.declarationIds.get(unwrapped);
  if (direct && isCallableTarget(facts, direct, false)) targets.add(direct);
  addSymbolTargets(
    facts,
    checker,
    checker.getSymbolAtLocation(symbolLocation(unwrapped)),
    targets,
    parameterTargets,
    false,
  );
  const type = checker.getTypeAtLocation(unwrapped);
  for (const signature of type.getCallSignatures()) {
    addSignatureTarget(facts, checker, signature, targets, parameterTargets, false);
  }
  return targets;
}

function addSignatureTarget(
  facts: FactStore,
  checker: ts.TypeChecker,
  signature: ts.Signature,
  targets: Set<string>,
  parameterTargets: ParameterTargets,
  includeTypes: boolean,
): void {
  const declaration = signature.getDeclaration();
  if (declaration) addDeclarationTarget(facts, checker, declaration, targets, parameterTargets, includeTypes);
}

function addSymbolTargets(
  facts: FactStore,
  checker: ts.TypeChecker,
  symbol: ts.Symbol | undefined,
  targets: Set<string>,
  parameterTargets: ParameterTargets,
  includeTypes: boolean,
): void {
  if (!symbol) return;
  const resolved = (symbol.flags & ts.SymbolFlags.Alias) !== 0 ? checker.getAliasedSymbol(symbol) : symbol;
  for (const declaration of resolved.declarations ?? []) {
    addDeclarationTarget(facts, checker, declaration, targets, parameterTargets, includeTypes);
  }
  if (resolved.valueDeclaration) {
    addDeclarationTarget(facts, checker, resolved.valueDeclaration, targets, parameterTargets, includeTypes);
  }
}

function addDeclarationTarget(
  facts: FactStore,
  checker: ts.TypeChecker,
  declaration: ts.Declaration,
  targets: Set<string>,
  parameterTargets: ParameterTargets,
  includeTypes: boolean,
): void {
  const parameter = enclosingParameter(declaration);
  if (parameter) {
    for (const target of parameterTargets.get(parameter) ?? []) targets.add(target);
    return;
  }
  if (ts.isExportAssignment(declaration)) {
    for (const target of resolveExpressionTargets(facts, checker, declaration.expression, parameterTargets)) targets.add(target);
    return;
  }
  const direct = declarationId(facts, declaration);
  if (direct && isCallableTarget(facts, direct, includeTypes)) {
    targets.add(direct);
    return;
  }
  const named = declaration as ts.NamedDeclaration;
  if (!named.name) return;
  const symbol = checker.getSymbolAtLocation(named.name);
  if (symbol && (symbol.flags & ts.SymbolFlags.Alias) !== 0) {
    addSymbolTargets(facts, checker, symbol, targets, parameterTargets, includeTypes);
  }
}

function isCallableTarget(facts: FactStore, target: string, includeTypes: boolean): boolean {
  const kind = facts.nodes.get(target)?.kind;
  if (kind === "function" || kind === "method" || kind === "constructor") return true;
  if (!includeTypes || kind !== "type") return false;
  return (facts.idDeclarations.get(target) ?? []).some(
    (declaration) => ts.isClassDeclaration(declaration) || ts.isClassExpression(declaration),
  );
}

function enclosingParameter(node: ts.Node): ts.ParameterDeclaration | null {
  let current: ts.Node | undefined = node;
  while (current && !ts.isSourceFile(current)) {
    if (ts.isParameter(current)) return current;
    if (isSignatureDeclaration(current) && current !== node) return null;
    current = current.parent;
  }
  return null;
}

function declarationId(facts: FactStore, declaration: ts.Node): string | null {
  let current: ts.Node | undefined = declaration;
  for (let depth = 0; current && depth < 6; depth += 1, current = current.parent) {
    const id = facts.declarationIds.get(current);
    if (id) return id;
    if (ts.isSourceFile(current)) break;
  }
  return null;
}
