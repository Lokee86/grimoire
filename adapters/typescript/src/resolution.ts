import * as path from "node:path";
import * as ts from "typescript";
import { expressionText, spanFor, spanForNodeId, staticTarget } from "./contract";
import type { PathMapping } from "./discovery";
import type { FactStore, ImportInfo, PendingCall, PendingRelationship } from "./model";
import type { Span } from "./model";

export function resolveImports(facts: FactStore, pathMappings: PathMapping[] = []): void {
  resolveReexportBindings(facts, pathMappings);
  for (const info of facts.imports) {
    if (!info.source) {
      facts.addUnresolved(info.ownerId, "imports", info.expression, "unsupported-form", spanForNodeId(facts, info.nodeId));
      continue;
    }
    const moduleId = resolveModule(facts, info.moduleKey, info.source, pathMappings);
    const moduleBindings = facts.bindings.get(info.moduleKey) ?? new Map();
    facts.bindings.set(info.moduleKey, moduleBindings);
    for (const item of info.names) {
      const target = resolveImportTarget(facts, info.moduleKey, info.source, item, pathMappings);
      moduleBindings.set(item.local, { targetId: target, external: !target && moduleResolutionReason(facts, info.source, pathMappings) === "external-target" });
      if (target) facts.addEdge(info.ownerId, target, "imports", spanForNodeId(facts, info.nodeId));
      else facts.addUnresolved(info.ownerId, "imports", info.expression, moduleId ? "missing-target" : moduleResolutionReason(facts, info.source, pathMappings), spanForNodeId(facts, info.nodeId), `${info.source}:${item.imported}`);
    }
  }
}

function resolveReexportBindings(facts: FactStore, pathMappings: PathMapping[]): void {
  for (const reexport of facts.reexports) {
    const targetModuleKey = resolveModuleKey(facts, reexport.moduleKey, reexport.source, pathMappings);
    const target = targetModuleKey ? facts.modules.get(targetModuleKey) ?? null : null;
    if (target) facts.addEdge(reexport.ownerId, target, "imports", reexport.span);
    else facts.addUnresolved(reexport.ownerId, "imports", reexport.expression, moduleResolutionReason(facts, reexport.source, pathMappings), reexport.span, reexport.source);
    if (!targetModuleKey) continue;
    const bindings = facts.bindings.get(reexport.moduleKey) ?? new Map();
    facts.bindings.set(reexport.moduleKey, bindings);
    for (const name of reexport.names) {
      const targetId = resolveExportedSymbol(facts, targetModuleKey, name.imported, new Set<string>(), pathMappings);
      if (targetId) bindings.set(name.exported, { targetId, external: false });
    }
  }
}

export function resolveRelationships(facts: FactStore): void {
  for (const relationship of facts.relationships) {
    const targetName = staticTarget(relationship.expression);
    const recordSpan = spanFor(relationship.expression, relationship.sourceFile, relationship.sourceFile.fileName.split(path.sep).join("/"));
    if (!targetName) {
      facts.addUnresolved(relationship.source, relationship.relation, expressionText(relationship.expression, relationship.sourceFile), "unsupported-form", recordSpan);
      continue;
    }
    const target = resolveSymbol(facts, relationship.moduleKey, relationship.scope, targetName);
    if (target) facts.addEdge(relationship.source, target, relationship.relation, recordSpan);
    else facts.addUnresolved(relationship.source, relationship.relation, targetName, isImportedExternal(facts, relationship.moduleKey, targetName) ? "external-target" : "missing-target", recordSpan, targetName);
  }
}

export function resolveCalls(facts: FactStore): void {
  for (const call of facts.calls) {
    const recordSpan = spanFor(call.expression, call.sourceFile, call.sourceFile.fileName.split(path.sep).join("/"));
    const targetName = directCallTarget(call);
    if (!targetName) {
      facts.addUnresolved(call.source, "calls", expressionText(call.expression, call.sourceFile), "unsupported-form", recordSpan);
      continue;
    }
    const binding = facts.bindings.get(call.moduleKey)?.get(targetName);
    const target = resolveSymbol(facts, call.moduleKey, call.scope, targetName);
    const reason = callTargetReason(facts, call, target, binding);
    if (reason) facts.addUnresolved(call.source, "calls", expressionText(call.expression, call.sourceFile), reason, recordSpan, targetName);
    else facts.addEdge(call.source, target!, "calls", recordSpan);
  }
}

function directCallTarget(call: PendingCall): string | null {
  if (call.kind === "call") {
    const expression = call.expression as ts.CallExpression;
    if (expression.questionDotToken || !ts.isIdentifier(expression.expression)) return null;
    return expression.expression.text;
  }
  const expression = call.expression as ts.NewExpression;
  return ts.isIdentifier(expression.expression) ? expression.expression.text : null;
}

function callTargetReason(
  facts: FactStore,
  call: PendingCall,
  target: string | null,
  binding: { targetId: string | null; external: boolean } | undefined,
): "missing-target" | "ambiguous-target" | "external-target" | "unsupported-form" | null {
  if (binding && !binding.targetId) return binding.external ? "external-target" : "missing-target";
  if (!target) return "missing-target";
  const qualifiedName = findQualifiedName(facts, target);
  if (facts.ambiguousSymbols.has(qualifiedName)) return "ambiguous-target";
  const expectedKind = call.kind === "call" ? "function" : "type";
  return facts.nodes.get(target)?.kind === expectedKind ? null : "unsupported-form";
}

function resolveSymbol(facts: FactStore, moduleKey: string, scope: string[], name: string): string | null {
  const binding = facts.bindings.get(moduleKey)?.get(name.split(".")[0]);
  if (binding) {
    if (!binding.targetId) return null;
    if (name === name.split(".")[0]) return binding.targetId;
    const base = findQualifiedName(facts, binding.targetId);
    return facts.symbols.get(`${base}.${name.split(".").slice(1).join(".")}`) ?? null;
  }
  const candidates: string[] = [];
  for (let count = scope.length; count >= 0; count -= 1) candidates.push(`${moduleKey}.${scope.slice(0, count).join(".")}${scope.slice(0, count).length ? "." : ""}${name}`);
  candidates.push(`${moduleKey}.${name}`);
  for (const candidate of candidates) {
    const target = facts.symbols.get(candidate) ?? facts.modules.get(candidate);
    if (target) return target;
  }
  return null;
}

function findQualifiedName(facts: FactStore, id: string): string {
  for (const [qualifiedName, symbolId] of facts.symbols) if (symbolId === id) return qualifiedName;
  for (const [moduleKey, moduleId] of facts.modules) if (moduleId === id) return moduleKey;
  return "";
}

function resolveImportTarget(facts: FactStore, importer: string, source: string, item: ImportInfo["names"][number], pathMappings: PathMapping[]): string | null {
  const moduleId = resolveModule(facts, importer, source, pathMappings);
  if (!moduleId) return null;
  if (item.kind === "side-effect" || item.kind === "default" || item.kind === "namespace") return moduleId;
  const targetModule = resolveModuleKey(facts, importer, source, pathMappings);
  return targetModule ? resolveExportedSymbol(facts, targetModule, item.imported, new Set<string>(), pathMappings) : null;
}

function resolveExportedSymbol(facts: FactStore, moduleKey: string, name: string, seen = new Set<string>(), pathMappings: PathMapping[] = []): string | null {
  const direct = facts.symbols.get(`${moduleKey}.${name}`);
  if (direct) return direct;
  const key = `${moduleKey}:${name}`;
  if (seen.has(key)) return null;
  seen.add(key);
  const binding = facts.bindings.get(moduleKey)?.get(name);
  if (binding?.targetId) return binding.targetId;
  for (const reexport of facts.reexports) {
    if (reexport.moduleKey !== moduleKey) continue;
    const exported = reexport.names.find((item) => item.exported === name);
    if (!exported) continue;
    const targetModuleKey = resolveModuleKey(facts, moduleKey, reexport.source, pathMappings);
    if (!targetModuleKey) continue;
    const target = resolveExportedSymbol(facts, targetModuleKey, exported.imported, seen, pathMappings);
    if (target) return target;
  }
  return null;
}

export function resolveModule(facts: FactStore, importer: string, source: string, pathMappings: PathMapping[] = []): string | null {
  const key = resolveModuleKey(facts, importer, source, pathMappings);
  return key ? facts.modules.get(key) ?? null : null;
}

function resolveModuleKey(facts: FactStore, importer: string, source: string, pathMappings: PathMapping[]): string | null {
  if (!isRelative(source)) return resolvePathMappedModule(facts, source, pathMappings);
  const base = path.posix.normalize(path.posix.join(path.posix.dirname(importer), source));
  for (const candidate of [base, `${base}/index`]) if (facts.modules.has(candidate)) return candidate;
  return null;
}

function resolvePathMappedModule(facts: FactStore, source: string, pathMappings: PathMapping[]): string | null {
  const matches = matchingMappings(source, pathMappings);
  if (matches.length === 0) return facts.modules.has(source) ? source : null;
  const candidates = new Set<string>();
  for (const mapping of matches) {
    const wildcard = mapping.pattern.indexOf("*");
    const capture = wildcard < 0 ? "" : source.slice(mapping.pattern.slice(0, wildcard).length, source.length - mapping.pattern.slice(wildcard + 1).length || undefined);
    for (const target of mapping.targets) {
      const substituted = wildcard < 0 ? target : target.replace("*", capture);
      const candidate = path.posix.normalize(path.posix.join(mapping.baseUrl, substituted)).replace(/\.(?:tsx?|mts|cts)$/i, "");
      if (candidate === "." || candidate.startsWith("../") || path.posix.isAbsolute(candidate)) continue;
      if (facts.modules.has(candidate)) candidates.add(candidate);
      if (facts.modules.has(`${candidate}/index`)) candidates.add(`${candidate}/index`);
    }
  }
  return candidates.size === 1 ? [...candidates][0] : null;
}

function matchingMappings(source: string, pathMappings: PathMapping[]): PathMapping[] {
  const matches = pathMappings.filter((mapping) => {
    const wildcard = mapping.pattern.indexOf("*");
    if (wildcard < 0) return mapping.pattern === source;
    const prefix = mapping.pattern.slice(0, wildcard);
    const suffix = mapping.pattern.slice(wildcard + 1);
    return source.startsWith(prefix) && source.endsWith(suffix) && source.length >= prefix.length + suffix.length;
  });
  const exact = matches.filter((mapping) => !mapping.pattern.includes("*"));
  if (exact.length > 0) return exact;
  if (matches.length === 0) return [];
  const specificity = Math.max(...matches.map((mapping) => mapping.pattern.length - 1));
  return matches.filter((mapping) => mapping.pattern.length - 1 === specificity);
}

function moduleResolutionReason(facts: FactStore, source: string, pathMappings: PathMapping[]): "missing-target" | "ambiguous-target" | "external-target" {
  if (isRelative(source)) return "missing-target";
  const matches = matchingMappings(source, pathMappings);
  if (matches.length === 0) return "external-target";
  const candidates = new Set<string>();
  for (const mapping of matches) {
    const wildcard = mapping.pattern.indexOf("*");
    const capture = wildcard < 0 ? "" : source.slice(mapping.pattern.slice(0, wildcard).length, source.length - mapping.pattern.slice(wildcard + 1).length || undefined);
    for (const target of mapping.targets) {
      const substituted = wildcard < 0 ? target : target.replace("*", capture);
      const candidate = path.posix.normalize(path.posix.join(mapping.baseUrl, substituted)).replace(/\.(?:tsx?|mts|cts)$/i, "");
      if (facts.modules.has(candidate)) candidates.add(candidate);
      if (facts.modules.has(`${candidate}/index`)) candidates.add(`${candidate}/index`);
    }
  }
  return candidates.size > 1 ? "ambiguous-target" : "missing-target";
}

function isRelative(source: string): boolean {
  return source.startsWith(".") || source.startsWith("/");
}

function isImportedExternal(facts: FactStore, moduleKey: string, targetName: string): boolean {
  return facts.bindings.get(moduleKey)?.get(targetName.split(".")[0])?.external ?? false;
}
