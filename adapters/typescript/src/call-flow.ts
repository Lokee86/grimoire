import * as ts from "typescript";
import { spanFor } from "./contract";
import {
  callArguments,
  isSignatureDeclaration,
  relativeSourcePath,
  type ParameterTargets,
} from "./call-shared";
import { resolveExpressionTargets } from "./call-targets";
import type { FactStore, PendingCall } from "./model";

export function propagateArguments(
  facts: FactStore,
  checker: ts.TypeChecker,
  call: PendingCall,
  targets: Set<string>,
  parameterTargets: ParameterTargets,
): boolean {
  const args = callArguments(call);
  if (!args) return false;
  let changed = false;
  for (const target of targets) {
    for (const declaration of callableDeclarations(facts, target)) {
      declaration.parameters.forEach((parameter, index) => {
        const expressions = parameter.dotDotDotToken ? args.slice(index) : args[index] ? [args[index]] : [];
        for (const expression of expressions) {
          if (ts.isSpreadElement(expression)) continue;
          const candidates = resolveExpressionTargets(facts, checker, expression, parameterTargets);
          if (candidates.size === 0) continue;
          const existing = parameterTargets.get(parameter) ?? new Set<string>();
          const before = existing.size;
          for (const candidate of candidates) existing.add(candidate);
          parameterTargets.set(parameter, existing);
          changed = changed || existing.size !== before;
        }
      });
    }
  }
  return changed;
}

function callableDeclarations(facts: FactStore, target: string): ts.SignatureDeclaration[] {
  const result: ts.SignatureDeclaration[] = [];
  for (const declaration of facts.idDeclarations.get(target) ?? []) {
    if (isSignatureDeclaration(declaration)) result.push(declaration);
    if (ts.isClassDeclaration(declaration) || ts.isClassExpression(declaration)) {
      const constructor = declaration.members.find(ts.isConstructorDeclaration);
      if (constructor) result.push(constructor);
    }
  }
  return [...new Set(result)];
}

export function emitCallableAliases(
  facts: FactStore,
  checker: ts.TypeChecker,
  parameterTargets: ParameterTargets,
): void {
  for (const alias of facts.callableAliases) {
    const targets = new Set<string>();
    for (const expression of alias.expressions) {
      for (const target of resolveExpressionTargets(facts, checker, expression, parameterTargets)) targets.add(target);
    }
    const span = facts.nodes.get(alias.source)?.span as ReturnType<typeof spanFor> | undefined;
    if (targets.size === 1) facts.addEdge(alias.source, [...targets][0], "calls", span, { wrapper: true });
    else for (const target of [...targets].sort()) facts.addEdge(alias.source, target, "possible-calls", span, { wrapper: true });
  }
}

export function emitCallbackEdges(
  facts: FactStore,
  checker: ts.TypeChecker,
  call: PendingCall,
  parameterTargets: ParameterTargets,
): void {
  const args = callArguments(call);
  if (!args) return;
  const callbackIndexes = callableArgumentIndexes(checker, call);
  if (callbackIndexes.size === 0) return;
  const span = spanFor(call.expression, call.sourceFile, relativeSourcePath(facts, call.sourceFile));
  args.forEach((argument, index) => {
    if (!callbackIndexes.has(index) || ts.isSpreadElement(argument)) return;
    const targets = resolveExpressionTargets(facts, checker, argument, parameterTargets);
    for (const target of [...targets].sort()) {
      facts.addEdge(call.source, target, "possible-calls", span, { callback: true, argument_index: index });
    }
  });
}

function callableArgumentIndexes(checker: ts.TypeChecker, call: PendingCall): Set<number> {
  const indexes = new Set<number>();
  if (call.kind !== "call" && call.kind !== "constructor") return indexes;
  const expression = call.expression as ts.CallExpression | ts.NewExpression;
  const signature = checker.getResolvedSignature(expression);
  const declaration = signature?.getDeclaration();
  const args = expression.arguments ?? [];
  if (!declaration) return indexes;
  const parameters = declaration.parameters;
  args.forEach((_argument, index) => {
    const parameter = parameters[index]
      ?? (parameters.at(-1)?.dotDotDotToken ? parameters.at(-1) : undefined);
    if (parameter && typeCanBeCalled(checker, checker.getTypeAtLocation(parameter))) indexes.add(index);
  });
  return indexes;
}

function typeCanBeCalled(checker: ts.TypeChecker, type: ts.Type): boolean {
  if (type.getCallSignatures().length > 0) return true;
  if (type.isUnionOrIntersection()) return type.types.some((member) => typeCanBeCalled(checker, member));
  if ((type.flags & ts.TypeFlags.TypeParameter) !== 0) {
    const constraint = checker.getBaseConstraintOfType(type);
    return !!constraint && constraint !== type && typeCanBeCalled(checker, constraint);
  }
  return false;
}
