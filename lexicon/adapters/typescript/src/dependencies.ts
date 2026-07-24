import * as fs from "node:fs";
import * as path from "node:path";

import { readPathMappings } from "./discovery";
import { resolveModule } from "./resolution";
import type { FactStore, JsonRecord } from "./model";

function attributes(category: string, source: string, constraint = "", flags: Partial<Record<"optional" | "dev" | "build" | "peer" | "path", boolean>> = {}): JsonRecord {
  return {
    build: category === "build" || flags.build === true,
    category,
    constraint,
    dev: category === "development" || category === "test" || flags.dev === true,
    optional: category === "optional" || flags.optional === true,
    path: category === "local" || flags.path === true,
    peer: category === "peer" || flags.peer === true,
    source,
  };
}

function target(facts: FactStore, name: string, localPath = ""): string {
  const normalized = name.replaceAll("\\", "/");
  const identity = `dependency:typescript:${normalized}`;
  return facts.addNode("module", normalized, localPath || `.lexicon/dependencies/typescript/${normalized}`, identity, identity, undefined, {
    dependency: true,
    ecosystem: "javascript",
  });
}

function addPackageJson(facts: FactStore, root: string, repositoryId: string): void {
  const filename = path.join(root, "package.json");
  if (!fs.existsSync(filename)) return;
  let data: unknown;
  try { data = JSON.parse(fs.readFileSync(filename, "utf8")); } catch { return; }
  if (!data || typeof data !== "object" || Array.isArray(data)) return;
  const object = data as Record<string, unknown>;
  const sections: Array<[string, string, boolean]> = [
    ["dependencies", "runtime", false],
    ["devDependencies", "development", true],
    ["peerDependencies", "peer", false],
    ["optionalDependencies", "optional", false],
  ];
  for (const [section, category, development] of sections) {
    const values = object[section];
    if (!values || typeof values !== "object" || Array.isArray(values)) continue;
    for (const name of Object.keys(values as Record<string, unknown>).sort()) {
      const raw = (values as Record<string, unknown>)[name];
      if (typeof raw !== "string" || !raw.trim()) continue;
      const candidatePath = raw.startsWith("file:") || raw.startsWith("link:")
        ? path.posix.normalize(raw.slice(raw.indexOf(":") + 1).replaceAll("\\", "/"))
        : "";
      const local = candidatePath !== "" && !path.posix.isAbsolute(candidatePath)
        && candidatePath !== ".." && !candidatePath.startsWith("../");
      const localPath = local ? candidatePath : "";
      facts.addEdge(repositoryId, target(facts, local ? localPath : name, localPath), "depends-on", undefined,
        attributes(category, `package.json:${section}`, raw, { dev: development, path: local }));
    }
  }
}

export function addDependencyFacts(facts: FactStore, root: string): void {
  const repositoryId = [...facts.nodes.values()].find((record) => record.kind === "repository")?.id as string | undefined;
  if (!repositoryId) return;
  addPackageJson(facts, root, repositoryId);
  const pathMappings = readPathMappings(root);
  for (const info of facts.imports) {
    if (!info.source) continue;
    const moduleId = resolveModule(facts, info.moduleKey, info.source, pathMappings);
    if (!moduleId) continue;
    const source = facts.modules.get(info.moduleKey);
    if (!source) continue;
    facts.addEdge(source, moduleId, "depends-on", undefined, attributes("local", info.source, "", { path: true }));
  }
}
