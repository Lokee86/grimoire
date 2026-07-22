import * as path from "node:path";
import { expressionText, spanFor, spanForNodeId, staticTarget } from "./contract";
import type { FactStore, ImportInfo, PendingRelationship } from "./model";
import type { Span } from "./model";

export function resolveImports(facts: FactStore): void {
  for (const info of facts.imports) {
    if (!info.source) {
      facts.addUnresolved(info.ownerId, "imports", info.expression, "unsupported-form", spanForNodeId(facts, info.nodeId));
      continue;
    }
    const moduleId = resolveModule(facts, info.moduleKey, info.source);
    const moduleBindings = facts.bindings.get(info.moduleKey) ?? new Map();
    facts.bindings.set(info.moduleKey, moduleBindings);
    for (const item of info.names) {
      const target = resolveImportTarget(facts, info.moduleKey, info.source, item);
      moduleBindings.set(item.local, { targetId: target, external: !target && !isRelative(info.source) });
      if (target) facts.addEdge(info.ownerId, target, "imports", spanForNodeId(facts, info.nodeId));
      else facts.addUnresolved(info.ownerId, "imports", info.expression, moduleId ? "missing-target" : (isRelative(info.source) ? "missing-target" : "external-target"), spanForNodeId(facts, info.nodeId), `${info.source}:${item.imported}`);
    }
  }
  for (const reexport of facts.reexports) {
    const target = resolveModule(facts, reexport.moduleKey, reexport.source);
    if (target) facts.addEdge(reexport.ownerId, target, "imports", reexport.span);
    else facts.addUnresolved(reexport.ownerId, "imports", reexport.expression, isRelative(reexport.source) ? "missing-target" : "external-target", reexport.span, reexport.source);
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

function resolveImportTarget(facts: FactStore, importer: string, source: string, item: ImportInfo["names"][number]): string | null {
  const moduleId = resolveModule(facts, importer, source);
  if (!moduleId) return null;
  if (item.kind === "side-effect" || item.kind === "default" || item.kind === "namespace") return moduleId;
  const targetModule = resolveModuleKey(facts, importer, source);
  return facts.symbols.get(`${targetModule}.${item.imported}`) ?? null;
}

export function resolveModule(facts: FactStore, importer: string, source: string): string | null {
  const key = resolveModuleKey(facts, importer, source);
  return key ? facts.modules.get(key) ?? null : null;
}

function resolveModuleKey(facts: FactStore, importer: string, source: string): string | null {
  if (!isRelative(source)) return facts.modules.has(source) ? source : null;
  const base = path.posix.normalize(path.posix.join(path.posix.dirname(importer), source));
  for (const candidate of [base, `${base}/index`]) if (facts.modules.has(candidate)) return candidate;
  return null;
}

function isRelative(source: string): boolean {
  return source.startsWith(".") || source.startsWith("/");
}

function isImportedExternal(facts: FactStore, moduleKey: string, targetName: string): boolean {
  return facts.bindings.get(moduleKey)?.get(targetName.split(".")[0])?.external ?? false;
}
