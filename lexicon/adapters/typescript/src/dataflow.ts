import * as ts from "typescript";
import { spanFor } from "./contract";
import type { FactStore } from "./model";

const DATA_KINDS = new Set(["variable", "constant", "field", "parameter"]);
const ASSIGNMENT_OPERATORS = new Set([
  ts.SyntaxKind.EqualsToken,
  ts.SyntaxKind.PlusEqualsToken,
  ts.SyntaxKind.MinusEqualsToken,
  ts.SyntaxKind.AsteriskEqualsToken,
  ts.SyntaxKind.SlashEqualsToken,
  ts.SyntaxKind.PercentEqualsToken,
  ts.SyntaxKind.AsteriskAsteriskEqualsToken,
  ts.SyntaxKind.LessThanLessThanEqualsToken,
  ts.SyntaxKind.GreaterThanGreaterThanEqualsToken,
  ts.SyntaxKind.AmpersandEqualsToken,
  ts.SyntaxKind.BarEqualsToken,
  ts.SyntaxKind.CaretEqualsToken,
  ts.SyntaxKind.QuestionQuestionEqualsToken,
]);

export function emitDataflow(facts: FactStore, checker: ts.TypeChecker): void {
  const seen = new Set<string>();
  for (const [node, source] of facts.declarationIds) {
    const kind = facts.nodes.get(source)?.kind;
    if (!kind || !["function", "method", "constructor"].includes(String(kind))) continue;
    const body = callableBody(node);
    if (!body || seen.has(source)) continue;
    seen.add(source);
    const relativePath = String(facts.nodes.get(source)?.path ?? body.getSourceFile().fileName).replace(/\\/g, "/");
    new DataflowVisitor(facts, checker, source, body.getSourceFile(), relativePath).visit(body);
  }
}

function callableBody(node: ts.Node): ts.Node | undefined {
  if (ts.isFunctionDeclaration(node) || ts.isFunctionExpression(node) || ts.isArrowFunction(node) ||
      ts.isMethodDeclaration(node) || ts.isConstructorDeclaration(node) || ts.isGetAccessorDeclaration(node) ||
      ts.isSetAccessorDeclaration(node)) return node.body ?? undefined;
  return undefined;
}

class DataflowVisitor {
  constructor(
    private readonly facts: FactStore,
    private readonly checker: ts.TypeChecker,
    private readonly source: string,
    private readonly sourceFile: ts.SourceFile,
    private readonly relativePath: string,
  ) {}

  visit(node: ts.Node): void {
    if (ts.isFunctionDeclaration(node) || ts.isFunctionExpression(node) || ts.isArrowFunction(node) ||
        ts.isMethodDeclaration(node) || ts.isConstructorDeclaration(node) || ts.isGetAccessorDeclaration(node) ||
        ts.isSetAccessorDeclaration(node)) return;
    if (ts.isVariableDeclaration(node)) {
      if (node.initializer) this.visit(node.initializer);
      this.writePattern(node.name, false);
      return;
    }
    if (ts.isBinaryExpression(node) && ASSIGNMENT_OPERATORS.has(node.operatorToken.kind)) {
      if (node.operatorToken.kind !== ts.SyntaxKind.EqualsToken) this.target(node.left, true);
      else this.target(node.left, false);
      this.visit(node.right);
      return;
    }
    if (ts.isPrefixUnaryExpression(node) || ts.isPostfixUnaryExpression(node)) {
      if (node.operator === ts.SyntaxKind.PlusPlusToken || node.operator === ts.SyntaxKind.MinusMinusToken) this.target(node.operand, true);
      else this.visit(node.operand);
      return;
    }
    if (ts.isPropertyAccessExpression(node)) {
      this.visit(node.expression);
      this.addSymbol(node.name, "reads");
      return;
    }
    if (ts.isElementAccessExpression(node)) {
      this.visit(node.expression);
      this.visit(node.argumentExpression);
      return;
    }
    if (ts.isIdentifier(node)) {
      this.addSymbol(node, "reads");
      return;
    }
    node.forEachChild((child) => this.visit(child));
  }

  private target(node: ts.Node, compound: boolean): void {
    if (ts.isIdentifier(node)) {
      if (compound) this.addSymbol(node, "reads");
      this.addSymbol(node, "writes");
      return;
    }
    if (ts.isPropertyAccessExpression(node)) {
      this.visit(node.expression);
      if (compound) this.addSymbol(node.name, "reads");
      this.addSymbol(node.name, "writes");
      return;
    }
    if (ts.isElementAccessExpression(node)) {
      this.visit(node.expression);
      this.visit(node.argumentExpression);
      return;
    }
    this.visit(node);
  }

  private writePattern(node: ts.BindingName, compound: boolean): void {
    if (ts.isIdentifier(node)) {
      this.addSymbol(node, "writes");
      return;
    }
    if (ts.isObjectBindingPattern(node) || ts.isArrayBindingPattern(node)) {
      for (const element of node.elements) {
        if (ts.isBindingElement(element)) this.writePattern(element.name, compound);
      }
    }
  }

  private addSymbol(node: ts.Node, relation: "reads" | "writes"): void {
    if (!ts.isIdentifier(node)) return;
    const symbol = this.checker.getSymbolAtLocation(node);
    if (!symbol) return;
    const id = this.symbolId(symbol);
    if (!id || !DATA_KINDS.has(String(this.facts.nodes.get(id)?.kind))) return;
    const span = spanFor(node, this.sourceFile, this.relativePath);
    this.facts.addDataflowEdge(this.source, id, relation, span);
  }

  private symbolId(symbol: ts.Symbol): string | undefined {
    for (const declaration of symbol.declarations ?? []) {
      const direct = this.facts.declarationIds.get(declaration);
      if (direct) return direct;
      const named = declaration as ts.NamedDeclaration;
      if (named.name) {
        const namedId = this.facts.declarationIds.get(named.name);
        if (namedId) return namedId;
      }
    }
    return undefined;
  }
}
