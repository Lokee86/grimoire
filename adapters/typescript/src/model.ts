import * as ts from "typescript";
import { nodeId } from "./contract";

export type JsonRecord = Record<string, unknown>;
export type Fact = JsonRecord & { record: string };

export type Span = {
  end_column: number;
  end_line: number;
  path: string;
  start_column: number;
  start_line: number;
};

export type PendingRelationship = {
  source: string;
  relation: "extends" | "implements" | "uses-trait";
  expression: ts.Expression;
  sourceFile: ts.SourceFile;
  moduleKey: string;
  scope: string[];
};

export type PendingCall = {
  expression: ts.CallExpression | ts.NewExpression | ts.TaggedTemplateExpression | ts.JsxOpeningLikeElement;
  kind: "call" | "constructor" | "tagged-template" | "jsx";
  moduleKey: string;
  scope: string[];
  source: string;
  sourceFile: ts.SourceFile;
};

export type PendingCallableAlias = {
  source: string;
  expressions: ts.Expression[];
  moduleKey: string;
  scope: string[];
  sourceFile: ts.SourceFile;
};

export type ImportName = {
  imported: string;
  local: string;
  kind: "named" | "default" | "namespace" | "side-effect";
};

export type ImportInfo = {
  nodeId: string;
  ownerId: string;
  moduleKey: string;
  sourceFile: ts.SourceFile;
  source: string | null;
  expression: string;
  names: ImportName[];
};

export type PendingReexport = {
  ownerId: string;
  moduleKey: string;
  source: string;
  expression: string;
  names: { imported: string; exported: string }[];
  span: Span;
};

export type Binding = { targetId: string | null; external: boolean };

export type FileContext = {
  relativePath: string;
  moduleKey: string;
  sourceFile: ts.SourceFile;
  fileId: string;
  moduleId: string;
};

export class FactStore {
  readonly nodes = new Map<string, Fact>();
  readonly edges = new Map<string, Fact>();
  readonly unresolved = new Map<string, Fact>();
  readonly modules = new Map<string, string>();
  readonly symbols = new Map<string, string>();
  readonly ambiguousSymbols = new Set<string>();
  readonly bindings = new Map<string, Map<string, Binding>>();
  readonly imports: ImportInfo[] = [];
  readonly reexports: PendingReexport[] = [];
  readonly relationships: PendingRelationship[] = [];
  readonly calls: PendingCall[] = [];
  readonly declarationIds = new Map<ts.Node, string>();
  readonly idDeclarations = new Map<string, ts.Node[]>();
  readonly callableAliases: PendingCallableAlias[] = [];
  readonly defaultExports = new Map<string, ts.Expression>();
  readonly defaultExportIds = new Map<string, string>();
  readonly commonJsExports = new Map<string, Map<string, ts.Expression>>();
  readonly dataflowEdges = new Set<string>();

  constructor(readonly repository: string, readonly root: string) {}

  registerDeclaration(node: ts.Node | undefined, id: string): void {
    if (!node) return;
    this.declarationIds.set(node, id);
    const declarations = this.idDeclarations.get(id) ?? [];
    if (!declarations.includes(node)) declarations.push(node);
    this.idDeclarations.set(id, declarations);
  }

  addNode(
    kind: string,
    name: string,
    relativePath: string,
    qualifiedName: string,
    identity: string = qualifiedName,
    span?: Span,
    attributes?: JsonRecord,
    contentId?: string,
  ): string {
    const id = nodeId(kind, identity);
    const record: Fact = { record: "node", id, kind, name, path: relativePath, qualified_name: qualifiedName };
    if (contentId) record.content_id = contentId;
    if (attributes && Object.keys(attributes).length > 0) record.attributes = attributes;
    if (span) record.span = span;
    this.nodes.set(id, record);
    return id;
  }

  addEdge(source: string, target: string, relation: string, span?: Span, attributes?: JsonRecord): void {
    const record: Fact = { record: "edge", relation, source, target };
    if (attributes && Object.keys(attributes).length > 0) record.attributes = attributes;
    if (span) record.span = span;
    this.edges.set(JSON.stringify(record), record);
  }

  addDataflowEdge(source: string, target: string, relation: "reads" | "writes", span?: Span): void {
    const key = `${source}\0${target}\0${relation}`;
    if (this.dataflowEdges.has(key)) return;
    this.dataflowEdges.add(key);
    this.addEdge(source, target, relation, span);
  }

  addUnresolved(source: string, relation: string, expression: string, reason: string, span?: Span, candidateName?: string): void {
    const record: Fact = { expression, reason, record: "unresolved", relation, source };
    if (candidateName) record.candidate_name = candidateName;
    if (span) record.span = span;
    this.unresolved.set(JSON.stringify(record), record);
  }
}
