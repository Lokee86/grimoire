import * as fs from "node:fs";
import * as path from "node:path";
import { compareKeys, factSortKey, jsonLine, LANGUAGE, SCHEMA_VERSION, VERSION } from "./contract";
import type { Fact, FactStore } from "./model";

export function emitFacts(facts: FactStore): Fact[] {
  const header: Fact = {
    adapter_version: VERSION,
    language: LANGUAGE,
    record: "lexicon",
    repository: facts.repository,
    schema_version: SCHEMA_VERSION,
  };
  const nodes = Array.from(facts.nodes.values()).sort((a, b) => compareKeys(factSortKey(a), factSortKey(b)));
  const edges = Array.from(facts.edges.values()).sort((a, b) => compareKeys(factSortKey(a), factSortKey(b)));
  const unresolved = Array.from(facts.unresolved.values()).sort((a, b) => compareKeys(factSortKey(a), factSortKey(b)));
  return [header, ...nodes, ...edges, ...unresolved];
}

export function writeJsonl(records: Fact[], outputPath: string): void {
  const lines = records.map(jsonLine).join("\n") + "\n";
  if (outputPath === "-") {
    process.stdout.write(lines);
    return;
  }
  const destination = path.resolve(outputPath);
  fs.mkdirSync(path.dirname(destination), { recursive: true });
  fs.writeFileSync(destination, lines, { encoding: "utf8" });
}
