import * as path from "node:path";
import * as ts from "typescript";
import type { FactStore, PendingCall } from "./model";

export type ParameterTargets = Map<ts.ParameterDeclaration, Set<string>>;

export function callArguments(call: PendingCall): readonly ts.Expression[] | null {
  if (call.kind === "call") return (call.expression as ts.CallExpression).arguments;
  if (call.kind === "constructor") return (call.expression as ts.NewExpression).arguments ?? [];
  return null;
}

export function callCallee(call: PendingCall): ts.Expression | null {
  if (call.kind === "call") return (call.expression as ts.CallExpression).expression;
  if (call.kind === "constructor") return (call.expression as ts.NewExpression).expression;
  if (call.kind === "tagged-template") return (call.expression as ts.TaggedTemplateExpression).tag;
  return null;
}

export function symbolLocation(expression: ts.Expression): ts.Node {
  const unwrapped = unwrapExpression(expression);
  if (ts.isPropertyAccessExpression(unwrapped)) return unwrapped.name;
  if (ts.isElementAccessExpression(unwrapped)) return unwrapped.argumentExpression;
  return unwrapped;
}

export function unwrapExpression(expression: ts.Expression): ts.Expression {
  let current = expression;
  while (
    ts.isParenthesizedExpression(current)
    || ts.isAsExpression(current)
    || ts.isTypeAssertionExpression(current)
    || ts.isNonNullExpression(current)
    || ts.isSatisfiesExpression(current)
  ) {
    current = current.expression;
  }
  return current;
}

export function callTargetName(call: PendingCall): string | undefined {
  if (call.kind === "jsx") return (call.expression as ts.JsxOpeningLikeElement).tagName.getText(call.sourceFile);
  return callCallee(call)?.getText(call.sourceFile);
}

export function relativeSourcePath(facts: FactStore, sourceFile: ts.SourceFile): string {
  return path.relative(facts.root, sourceFile.fileName).split(path.sep).join("/");
}

export function sameSet(left: Set<string>, right: Set<string> | undefined): boolean {
  return !!right && left.size === right.size && [...left].every((item) => right.has(item));
}

export function mergeSets(...sets: Set<string>[]): Set<string> {
  const result = new Set<string>();
  for (const set of sets) for (const item of set) result.add(item);
  return result;
}

export function isSignatureDeclaration(node: ts.Node): node is ts.SignatureDeclaration {
  return ts.isFunctionDeclaration(node)
    || ts.isFunctionExpression(node)
    || ts.isArrowFunction(node)
    || ts.isMethodDeclaration(node)
    || ts.isMethodSignature(node)
    || ts.isConstructorDeclaration(node)
    || ts.isCallSignatureDeclaration(node)
    || ts.isConstructSignatureDeclaration(node)
    || ts.isGetAccessorDeclaration(node)
    || ts.isSetAccessorDeclaration(node);
}
