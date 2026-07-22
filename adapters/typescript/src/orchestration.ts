import * as fs from "node:fs";
import * as path from "node:path";
import { createFileContext, extractDeclarations } from "./ast";
import { readPathMappings, scanRepository } from "./discovery";
import { emitFacts, writeJsonl } from "./emission";
import { resolveCalls, resolveImports, resolveRelationships } from "./resolution";
import { FactStore, type Fact } from "./model";

export function buildFacts(repositoryPath: string): Fact[] {
  const root = path.resolve(repositoryPath);
  if (!fs.statSync(root).isDirectory()) throw new Error(`repository is not a directory: ${repositoryPath}`);
  const repository = path.basename(root);
  const facts = new FactStore(repository);
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
  const contexts = files.map((file) => createFileContext(root, file, directoryIds, facts));
  for (const context of contexts) extractDeclarations(context, facts);
  resolveImports(facts, readPathMappings(root));
  resolveCalls(facts);
  resolveRelationships(facts);
  return emitFacts(facts);
}

export function writeFacts(repositoryPath: string, outputPath: string): void {
  writeJsonl(buildFacts(repositoryPath), outputPath);
}
