import * as crypto from "node:crypto";
import * as fs from "node:fs";
import * as path from "node:path";
import * as ts from "typescript";

export const VERSION = "0.1.0";
export const LANGUAGE = "typescript";
export const SCHEMA_VERSION = 1;

const EXCLUDED_DIRECTORIES = new Set([
  ".git", ".worktrees", ".workingtrees", ".warlock", "node_modules",
  "build", "dist", "coverage", "target", "vendor", "tmp", "log",
  ".cache", ".turbo", ".next", ".nuxt", ".parcel-cache", ".pytest_cache",
  ".venv", "venv", "__pycache__", "out",
]);

type JsonRecord = Record<string, unknown>;
type Fact = JsonRecord & { record: string };
type Span = {
  end_column: number;
  end_line: number;
  path: string;
  start_column: number;
  start_line: number;
};

type PendingRelationship = {
  source: string;
  relation: "extends" | "implements";
  expression: ts.Expression;
  sourceFile: ts.SourceFile;
  moduleKey: string;
  scope: string[];
};

type ImportInfo = {
  nodeId: string;
  ownerId: string;
  moduleKey: string;
  sourceFile: ts.SourceFile;
  source: string | null;
  expression: string;
  names: Array<{ imported: string; local: string; kind: "named" | "default" | "namespace" | "side-effect" }>;
};

type Binding = { targetId: string | null; external: boolean };

type FileContext = {
  absolutePath: string;
  relativePath: string;
  moduleKey: string;
  sourceFile: ts.SourceFile;
  fileId: string;
  moduleId: string;
};

function digest(value: string | Buffer): string {
  const input = typeof value === "string" ? Buffer.from(value, "utf8") : value;
  return `sha256:${crypto.createHash("sha256").update(input).digest("hex")}`;
}

function nodeId(kind: string, identity: string): string {
  return digest(`lexicon:v1\0${LANGUAGE}\0${kind}\0${identity}`);
}

function sortedObject(value: unknown): unknown {
  if (Array.isArray(value)) return value.map(sortedObject);
  if (value !== null && typeof value === "object") {
    const object = value as JsonRecord;
    return Object.fromEntries(Object.keys(object).sort().map((key) => [key, sortedObject(object[key])]));
  }
  return value;
}

function jsonLine(record: unknown): string {
  return JSON.stringify(sortedObject(record), undefined, undefined);
}

function spanFor(node: ts.Node, sourceFile: ts.SourceFile, relativePath: string): Span {
  const start = sourceFile.getLineAndCharacterOfPosition(node.getStart(sourceFile));
  const end = sourceFile.getLineAndCharacterOfPosition(node.getEnd());
  return {
    end_column: end.character + 1,
    end_line: end.line + 1,
    path: relativePath,
    start_column: start.character + 1,
    start_line: start.line + 1,
  };
}

function spanKey(record: JsonRecord): unknown[] {
  const value = (record.span ?? {}) as JsonRecord;
  return [value.path ?? "", value.start_line ?? 0, value.start_column ?? 0, value.end_line ?? 0, value.end_column ?? 0];
}

function factSortKey(record: Fact): unknown[] {
  if (record.record === "node") return [0, record.id, record.kind, record.path, record.qualified_name];
  if (record.record === "edge") return [1, record.source, record.target, record.relation, ...spanKey(record)];
  return [2, record.source, record.relation, record.expression, record.reason, ...spanKey(record)];
}

function compareKeys(left: unknown[], right: unknown[]): number {
  for (let index = 0; index < Math.max(left.length, right.length); index += 1) {
    const a = left[index] ?? "";
    const b = right[index] ?? "";
    if (typeof a === "number" && typeof b === "number") {
      if (a < b) return -1;
      if (a > b) return 1;
      continue;
    }
    const leftText = String(a);
    const rightText = String(b);
    if (leftText < rightText) return -1;
    if (leftText > rightText) return 1;
  }
  return 0;
}

function normalizeRelative(root: string, absolutePath: string): string {
  return path.relative(root, absolutePath).split(path.sep).join("/") || ".";
}

function withoutExtension(relativePath: string): string {
  return relativePath.replace(/\.(?:tsx?|mts|cts)$/i, "");
}

function moduleKeyFor(relativePath: string): string {
  return withoutExtension(relativePath);
}

function declarationName(name: ts.PropertyName | ts.BindingName | undefined): string | null {
  if (!name) return null;
  if (ts.isIdentifier(name) || ts.isStringLiteral(name) || ts.isNumericLiteral(name)) return name.text;
  return null;
}

function expressionText(node: ts.Node, sourceFile: ts.SourceFile): string {
  return node.getText(sourceFile).trim();
}

function staticTarget(node: ts.Expression): string | null {
  if (ts.isIdentifier(node)) return node.text;
  if (ts.isPropertyAccessExpression(node)) {
    const parent = staticTarget(node.expression);
    return parent ? `${parent}.${node.name.text}` : null;
  }
  if (ts.isParenthesizedExpression(node)) return staticTarget(node.expression);
  return null;
}

function hasModifier(node: ts.Node, kind: ts.SyntaxKind): boolean {
  return (ts.canHaveModifiers(node) ? ts.getModifiers(node) ?? [] : []).some((modifier) => modifier.kind === kind);
}

function isStaticModuleSpecifier(node: ts.Node | undefined): node is ts.StringLiteralLike {
  return !!node && (ts.isStringLiteral(node) || ts.isNoSubstitutionTemplateLiteral(node));
}

class Facts {
  readonly nodes = new Map<string, Fact>();
  readonly edges = new Map<string, Fact>();
  readonly unresolved = new Map<string, Fact>();
  readonly modules = new Map<string, string>();
  readonly symbols = new Map<string, string>();
  readonly bindings = new Map<string, Map<string, Binding>>();
  readonly imports: ImportInfo[] = [];
  readonly relationships: PendingRelationship[] = [];

  constructor(readonly repository: string) {}

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
    this.edges.set(jsonLine(record), record);
  }

  addUnresolved(source: string, relation: string, expression: string, reason: string, span?: Span, candidateName?: string): void {
    const record: Fact = { expression, reason, record: "unresolved", relation, source };
    if (candidateName) record.candidate_name = candidateName;
    if (span) record.span = span;
    this.unresolved.set(jsonLine(record), record);
  }
}

class Extractor {
  private readonly contexts: FileContext[] = [];

  constructor(private readonly root: string, private readonly facts: Facts, private readonly directoryIds: Map<string, string>) {}

  scan(files: string[]): void {
    for (const absolutePath of files) this.scanFile(absolutePath);
    for (const context of this.contexts) this.visit(context.sourceFile, context.moduleId, [], context);
    this.resolveImports();
    this.resolveRelationships();
  }

  private scanFile(absolutePath: string): void {
    const relativePath = normalizeRelative(this.root, absolutePath);
    const sourceBytes = fs.readFileSync(absolutePath);
    const source = sourceBytes.toString("utf8");
    const scriptKind = relativePath.toLowerCase().endsWith(".tsx") ? ts.ScriptKind.TSX : ts.ScriptKind.TS;
    const sourceFile = ts.createSourceFile(relativePath, source, ts.ScriptTarget.Latest, true, scriptKind);
    const fileId = this.facts.addNode("file", path.posix.basename(relativePath), relativePath, relativePath, relativePath, spanFor(sourceFile, sourceFile, relativePath), undefined, digest(sourceBytes));
    const parent = path.posix.dirname(relativePath) || ".";
    this.facts.addEdge(this.directoryIds.get(parent) ?? this.directoryIds.get(".")!, fileId, "contains");
    const moduleKey = moduleKeyFor(relativePath);
    const moduleId = this.facts.addNode("module", path.posix.basename(moduleKey), relativePath, moduleKey, moduleKey);
    this.facts.modules.set(moduleKey, moduleId);
    this.facts.addEdge(fileId, moduleId, "contains", spanFor(sourceFile, sourceFile, relativePath));
    if (((sourceFile as ts.SourceFile & { parseDiagnostics?: ts.Diagnostic[] }).parseDiagnostics ?? []).length > 0) {
      this.facts.addUnresolved(moduleId, "parses", relativePath, "unsupported-form", undefined, "syntax-diagnostics");
    }
    this.contexts.push({ absolutePath, relativePath, moduleKey, sourceFile, fileId, moduleId });
  }

  private visit(node: ts.Node, ownerId: string, scope: string[], context: FileContext): void {
    if (ts.isSourceFile(node)) {
      node.forEachChild((child) => this.visit(child, ownerId, scope, context));
      return;
    }
    if (ts.isImportDeclaration(node)) {
      this.visitImport(node, ownerId, context);
      return;
    }
    if (ts.isImportEqualsDeclaration(node)) {
      this.visitImportEquals(node, ownerId, context);
      return;
    }
    if (ts.isExportDeclaration(node)) {
      this.visitExportDeclaration(node, ownerId, context);
      return;
    }
    if (ts.isExportAssignment(node)) {
      this.addExport(ownerId, node, context, node.isExportEquals ? "=" : "default");
      return;
    }
    if (ts.isClassDeclaration(node)) {
      this.visitClass(node, ownerId, scope, context);
      return;
    }
    if (ts.isInterfaceDeclaration(node)) {
      this.visitInterface(node, ownerId, scope, context);
      return;
    }
    if (ts.isTypeAliasDeclaration(node)) {
      const name = declarationName(node.name) ?? "<anonymous>";
      const id = this.addSymbol("type", name, scope, node, ownerId, context);
      if (hasModifier(node, ts.SyntaxKind.ExportKeyword)) this.addExport(ownerId, node, context, name, id);
      node.forEachChild((child) => this.visit(child, ownerId, scope, context));
      return;
    }
    if (ts.isFunctionDeclaration(node)) {
      this.visitFunction(node, ownerId, scope, context);
      return;
    }
    if (ts.isMethodDeclaration(node) || ts.isMethodSignature(node)) {
      this.visitMethod(node, ownerId, scope, context);
      return;
    }
    if (ts.isConstructorDeclaration(node)) {
      this.visitMethod(node, ownerId, scope, context, true);
      return;
    }
    if (ts.isPropertyDeclaration(node) || ts.isPropertySignature(node)) {
      this.visitProperty(node, ownerId, scope, context);
      return;
    }
    if (ts.isVariableStatement(node)) {
      this.visitVariableList(node.declarationList, ownerId, scope, context);
      return;
    }
    if (ts.isVariableDeclarationList(node) && !ts.isVariableStatement(node.parent)) {
      this.visitVariableList(node, ownerId, scope, context);
      return;
    }
    if (ts.isCallExpression(node) && node.expression.kind === ts.SyntaxKind.ImportKeyword) {
      this.facts.addUnresolved(ownerId, "imports", expressionText(node, context.sourceFile), "dynamic-target", spanFor(node, context.sourceFile, context.relativePath));
    }
    node.forEachChild((child) => this.visit(child, ownerId, scope, context));
  }

  private addSymbol(kind: string, name: string, scope: string[], node: ts.Node, ownerId: string, context: FileContext, attributes?: JsonRecord): string {
    const qualifiedName = `${context.moduleKey}.${[...scope, name].join(".")}`;
    const id = this.facts.addNode(kind, name, context.relativePath, qualifiedName, qualifiedName, spanFor(node, context.sourceFile, context.relativePath), attributes);
    this.facts.symbols.set(qualifiedName, id);
    this.facts.addEdge(ownerId, id, "defines", spanFor(node, context.sourceFile, context.relativePath));
    return id;
  }

  private visitClass(node: ts.ClassDeclaration, ownerId: string, scope: string[], context: FileContext): void {
    const name = declarationName(node.name) ?? "<anonymous>";
    const id = this.addSymbol("type", name, scope, node, ownerId, context);
    if (hasModifier(node, ts.SyntaxKind.ExportKeyword)) this.addExport(ownerId, node, context, name, id);
    this.addHeritage(node, id, context, scope);
    const childScope = [...scope, name];
    node.members.forEach((member) => this.visit(member, id, childScope, context));
  }

  private visitInterface(node: ts.InterfaceDeclaration, ownerId: string, scope: string[], context: FileContext): void {
    const name = node.name.text;
    const id = this.addSymbol("interface", name, scope, node, ownerId, context);
    if (hasModifier(node, ts.SyntaxKind.ExportKeyword)) this.addExport(ownerId, node, context, name, id);
    this.addHeritage(node, id, context, scope);
    const childScope = [...scope, name];
    node.members.forEach((member) => this.visit(member, id, childScope, context));
  }

  private addHeritage(node: ts.ClassDeclaration | ts.InterfaceDeclaration, sourceId: string, context: FileContext, scope: string[]): void {
    for (const clause of node.heritageClauses ?? []) {
      const relation = clause.token === ts.SyntaxKind.ExtendsKeyword ? "extends" : "implements";
      for (const type of clause.types) {
        this.facts.relationships.push({ source: sourceId, relation, expression: type.expression, sourceFile: context.sourceFile, moduleKey: context.moduleKey, scope });
      }
    }
  }

  private visitFunction(node: ts.FunctionDeclaration, ownerId: string, scope: string[], context: FileContext): void {
    const name = declarationName(node.name) ?? "<anonymous>";
    const id = this.addSymbol("function", name, scope, node, ownerId, context);
    if (hasModifier(node, ts.SyntaxKind.ExportKeyword)) this.addExport(ownerId, node, context, name, id);
    const childScope = [...scope, name];
    node.forEachChild((child) => this.visit(child, id, childScope, context));
  }

  private visitMethod(node: ts.MethodDeclaration | ts.MethodSignature | ts.ConstructorDeclaration, ownerId: string, scope: string[], context: FileContext, constructor = false): void {
    const name = constructor ? "constructor" : declarationName(node.name) ?? "<computed>";
    const kind = constructor ? "constructor" : "method";
    const id = this.addSymbol(kind, name, scope, node, ownerId, context);
    const childScope = [...scope, name];
    node.forEachChild((child) => this.visit(child, id, childScope, context));
  }

  private visitProperty(node: ts.PropertyDeclaration | ts.PropertySignature, ownerId: string, scope: string[], context: FileContext): void {
    const name = declarationName(node.name) ?? "<computed>";
    const id = this.addSymbol("field", name, scope, node, ownerId, context);
    node.forEachChild((child) => this.visit(child, ownerId, scope, context));
    void id;
  }

  private visitVariableList(list: ts.VariableDeclarationList, ownerId: string, scope: string[], context: FileContext): void {
    const kind = (list.flags & ts.NodeFlags.Const) !== 0 ? "constant" : "variable";
    for (const declaration of list.declarations) {
      const name = declarationName(declaration.name) ?? "<computed>";
      const id = this.addSymbol(kind, name, scope, declaration, ownerId, context);
      if (hasModifier(list.parent, ts.SyntaxKind.ExportKeyword)) this.addExport(ownerId, list.parent, context, name, id);
      declaration.forEachChild((child) => this.visit(child, id, [...scope, name], context));
    }
  }

  private visitImport(node: ts.ImportDeclaration, ownerId: string, context: FileContext): void {
    const source = isStaticModuleSpecifier(node.moduleSpecifier) ? node.moduleSpecifier.text : null;
    const names: ImportInfo["names"] = [];
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
    const id = this.facts.addNode("import", names.map((item) => item.imported).join(",") || "import", context.relativePath, identity, identity, recordSpan, { expression: expressionText(node, context.sourceFile) });
    this.facts.addEdge(ownerId, id, "defines", recordSpan);
    this.facts.imports.push({ nodeId: id, ownerId, moduleKey: context.moduleKey, sourceFile: context.sourceFile, source, expression: expressionText(node, context.sourceFile), names });
  }

  private visitImportEquals(node: ts.ImportEqualsDeclaration, ownerId: string, context: FileContext): void {
    const recordSpan = spanFor(node, context.sourceFile, context.relativePath);
    const identity = `${context.moduleKey}::import:${recordSpan.start_line}:${recordSpan.start_column}`;
    const id = this.facts.addNode("import", node.name.text, context.relativePath, identity, identity, recordSpan, { expression: expressionText(node, context.sourceFile) });
    this.facts.addEdge(ownerId, id, "defines", recordSpan);
    const moduleReference = node.moduleReference;
    const source = ts.isExternalModuleReference(moduleReference) && isStaticModuleSpecifier(moduleReference.expression) ? moduleReference.expression.text : null;
    this.facts.imports.push({ nodeId: id, ownerId, moduleKey: context.moduleKey, sourceFile: context.sourceFile, source, expression: expressionText(node, context.sourceFile), names: [{ imported: "*", local: node.name.text, kind: "namespace" }] });
  }

  private visitExportDeclaration(node: ts.ExportDeclaration, ownerId: string, context: FileContext): void {
    const recordSpan = spanFor(node, context.sourceFile, context.relativePath);
    const source = isStaticModuleSpecifier(node.moduleSpecifier) ? node.moduleSpecifier.text : null;
    const names = node.exportClause && ts.isNamedExports(node.exportClause)
      ? node.exportClause.elements.map((element) => element.propertyName?.text ?? element.name.text)
      : ["*"];
    const identity = `${context.moduleKey}::export:${recordSpan.start_line}:${recordSpan.start_column}`;
    const id = this.facts.addNode("export", names.join(","), context.relativePath, identity, identity, recordSpan, { expression: expressionText(node, context.sourceFile) });
    this.facts.addEdge(ownerId, id, "defines", recordSpan);
    if (source) {
      const target = this.resolveModule(context.moduleKey, source);
      if (target) this.facts.addEdge(ownerId, target, "imports", recordSpan);
      else this.facts.addUnresolved(ownerId, "imports", expressionText(node, context.sourceFile), this.isRelative(source) ? "missing-target" : "external-target", recordSpan, source);
    } else if (node.moduleSpecifier) {
      this.facts.addUnresolved(ownerId, "imports", expressionText(node, context.sourceFile), "unsupported-form", recordSpan);
    }
  }

  private addExport(ownerId: string, node: ts.Node, context: FileContext, name: string, targetId?: string): void {
    const recordSpan = spanFor(node, context.sourceFile, context.relativePath);
    const identity = `${context.moduleKey}::export:${recordSpan.start_line}:${recordSpan.start_column}:${name}`;
    const id = this.facts.addNode("export", name, context.relativePath, identity, identity, recordSpan, { name, default: hasModifier(node, ts.SyntaxKind.DefaultKeyword) });
    this.facts.addEdge(ownerId, id, "defines", recordSpan);
    if (targetId) this.facts.addEdge(id, targetId, "defines", recordSpan);
  }

  private resolveImports(): void {
    for (const info of this.facts.imports) {
      if (!info.source) {
        this.facts.addUnresolved(info.ownerId, "imports", info.expression, "unsupported-form", spanForNodeId(this.facts, info.nodeId));
        continue;
      }
      const moduleId = this.resolveModule(info.moduleKey, info.source);
      const moduleBindings = this.facts.bindings.get(info.moduleKey) ?? new Map<string, Binding>();
      this.facts.bindings.set(info.moduleKey, moduleBindings);
      for (const item of info.names) {
        const target = this.resolveImportTarget(info.moduleKey, info.source, item);
        moduleBindings.set(item.local, { targetId: target, external: !target && !this.isRelative(info.source) });
        if (target) this.facts.addEdge(info.ownerId, target, "imports", spanForNodeId(this.facts, info.nodeId));
        else this.facts.addUnresolved(info.ownerId, "imports", info.expression, moduleId ? "missing-target" : (this.isRelative(info.source) ? "missing-target" : "external-target"), spanForNodeId(this.facts, info.nodeId), `${info.source}:${item.imported}`);
      }
    }
  }

  private resolveRelationships(): void {
    for (const relationship of this.facts.relationships) {
      const targetName = staticTarget(relationship.expression);
      const recordSpan = spanFor(relationship.expression, relationship.sourceFile, relativeFromSourceFile(relationship.sourceFile));
      if (!targetName) {
        this.facts.addUnresolved(relationship.source, relationship.relation, expressionText(relationship.expression, relationship.sourceFile), "unsupported-form", recordSpan);
        continue;
      }
      const target = this.resolveSymbol(relationship.moduleKey, relationship.scope, targetName);
      if (target) this.facts.addEdge(relationship.source, target, relationship.relation, recordSpan);
      else this.facts.addUnresolved(relationship.source, relationship.relation, targetName, this.isImportedExternal(relationship.moduleKey, targetName) ? "external-target" : "missing-target", recordSpan, targetName);
    }
  }

  private resolveSymbol(moduleKey: string, scope: string[], name: string): string | null {
    const bindings = this.facts.bindings.get(moduleKey);
    const first = name.split(".")[0];
    const binding = bindings?.get(first);
    if (binding) {
      if (binding.targetId) {
        if (name === first) return binding.targetId;
        const base = this.findQualifiedName(binding.targetId);
        return this.facts.symbols.get(`${base}.${name.split(".").slice(1).join(".")}`) ?? null;
      }
      return null;
    }
    const candidates = [];
    for (let count = scope.length; count >= 0; count -= 1) candidates.push(`${moduleKey}.${scope.slice(0, count).join(".")}${scope.slice(0, count).length ? "." : ""}${name}`);
    candidates.push(`${moduleKey}.${name}`);
    for (const candidate of candidates) {
      const target = this.facts.symbols.get(candidate) ?? this.facts.modules.get(candidate);
      if (target) return target;
    }
    return null;
  }

  private findQualifiedName(id: string): string {
    for (const [qualifiedName, symbolId] of this.facts.symbols) if (symbolId === id) return qualifiedName;
    for (const [moduleKey, moduleId] of this.facts.modules) if (moduleId === id) return moduleKey;
    return "";
  }

  private resolveImportTarget(importer: string, source: string, item: ImportInfo["names"][number]): string | null {
    const moduleId = this.resolveModule(importer, source);
    if (!moduleId) return null;
    if (item.kind === "side-effect" || item.kind === "default" || item.kind === "namespace") return moduleId;
    const targetModule = this.resolveModuleKey(importer, source);
    return this.facts.symbols.get(`${targetModule}.${item.imported}`) ?? null;
  }

  private resolveModule(importer: string, source: string): string | null {
    const key = this.resolveModuleKey(importer, source);
    return key ? this.facts.modules.get(key) ?? null : null;
  }

  private resolveModuleKey(importer: string, source: string): string | null {
    if (!this.isRelative(source)) return this.facts.modules.has(source) ? source : null;
    const base = path.posix.normalize(path.posix.join(path.posix.dirname(importer), source));
    for (const candidate of [base, `${base}/index`]) if (this.facts.modules.has(candidate)) return candidate;
    return null;
  }

  private isRelative(source: string): boolean {
    return source.startsWith(".") || source.startsWith("/");
  }

  private isImportedExternal(moduleKey: string, targetName: string): boolean {
    return this.facts.bindings.get(moduleKey)?.get(targetName.split(".")[0])?.external ?? false;
  }
}

function relativeFromSourceFile(sourceFile: ts.SourceFile): string {
  return sourceFile.fileName.split(path.sep).join("/");
}

function spanForNodeId(facts: Facts, id: string): Span | undefined {
  return facts.nodes.get(id)?.span as Span | undefined;
}

function scanRepository(root: string): { directories: string[]; files: string[] } {
  const directories = ["."];
  const files: string[] = [];
  const walk = (directory: string): void => {
    const entries = fs.readdirSync(directory, { withFileTypes: true }).sort((left, right) => left.name.localeCompare(right.name));
    for (const entry of entries) {
      if (entry.isDirectory()) {
        if (EXCLUDED_DIRECTORIES.has(entry.name)) continue;
        const absolute = path.join(directory, entry.name);
        directories.push(normalizeRelative(root, absolute));
        walk(absolute);
      } else if (entry.isFile() && /\.(?:ts|tsx)$/i.test(entry.name)) {
        files.push(normalizeRelative(root, path.join(directory, entry.name)));
      }
    }
  };
  walk(root);
  directories.sort();
  files.sort();
  return { directories, files };
}

export function buildFacts(repositoryPath: string): Fact[] {
  const root = path.resolve(repositoryPath);
  if (!fs.statSync(root).isDirectory()) throw new Error(`repository is not a directory: ${repositoryPath}`);
  const repository = path.basename(root);
  const facts = new Facts(repository);
  const repositoryId = facts.addNode("repository", repository, ".", repository, repository);
  const { directories, files } = scanRepository(root);
  const directoryIds = new Map<string, string>();
  for (const directory of directories) {
    const name = directory === "." ? repository : path.posix.basename(directory);
    directoryIds.set(directory, facts.addNode("directory", name, directory, directory, directory));
  }
  facts.addEdge(repositoryId, directoryIds.get(".")!, "contains");
  for (const directory of directories.filter((item) => item !== ".")) {
    const parent = path.posix.dirname(directory) || ".";
    facts.addEdge(directoryIds.get(parent) ?? directoryIds.get(".")!, directoryIds.get(directory)!, "contains");
  }
  const extractor = new Extractor(root, facts, directoryIds);
  extractor.scan(files.map((file) => path.join(root, file.split("/").join(path.sep))));
  const header: Fact = { adapter_version: VERSION, language: LANGUAGE, record: "lexicon", repository, schema_version: SCHEMA_VERSION };
  return [header, ...Array.from(facts.nodes.values()).sort((a, b) => compareKeys(factSortKey(a), factSortKey(b))), ...Array.from(facts.edges.values()).sort((a, b) => compareKeys(factSortKey(a), factSortKey(b))), ...Array.from(facts.unresolved.values()).sort((a, b) => compareKeys(factSortKey(a), factSortKey(b)))];
}

export function writeFacts(repositoryPath: string, outputPath: string): void {
  const lines = buildFacts(repositoryPath).map(jsonLine).join("\n") + "\n";
  if (outputPath === "-") {
    process.stdout.write(lines);
    return;
  }
  const destination = path.resolve(outputPath);
  fs.mkdirSync(path.dirname(destination), { recursive: true });
  fs.writeFileSync(destination, lines, { encoding: "utf8" });
}
