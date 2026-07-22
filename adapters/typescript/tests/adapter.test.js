const assert = require("node:assert/strict");
const crypto = require("node:crypto");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const { spawnSync } = require("node:child_process");
const test = require("node:test");

const ADAPTER_ROOT = path.resolve(__dirname, "..");
const REPO_ROOT = path.resolve(ADAPTER_ROOT, "../..");
const CLI = path.join(ADAPTER_ROOT, "dist", "cli.js");

function makeFixture() {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), "lexicon-typescript-"));
  const files = {
    "src/base.ts": [
      "export interface Named { id: string }",
      "export class Base { run(): void {} }",
      "export const BASE = 1;",
      "let mutable = 2;",
      "",
    ].join("\n"),
    "src/child.ts": [
      'import { Base, Named, BASE as VALUE } from "./base";',
      "export interface Child extends Named {}",
      "export class Child extends Base implements Named {",
      "  value = VALUE;",
      "  method(): void { void import(\"./dynamic\"); }",
      "}",
      "",
    ].join("\n"),
    "src/index.tsx": [
      'import React from "react";',
      'export { Child } from "./child";',
      "export default function App() { return <Child />; }",
      "",
    ].join("\n"),
    "src/dynamic.ts": "export const loaded = true;\n",
    "src/computed.ts": 'const suffix = "dynamic";\nimport("./" + suffix);\n',
    "node_modules/ignored.ts": "export class Ignored {}\n",
    "build/ignored.ts": "export class IgnoredBuild {}\n",
    ".git/ignored.ts": "export class IgnoredGit {}\n",
  };
  for (const [relative, content] of Object.entries(files)) {
    const target = path.join(root, relative);
    fs.mkdirSync(path.dirname(target), { recursive: true });
    fs.writeFileSync(target, content);
  }
  return root;
}

function runAdapter(repo, output) {
  const result = spawnSync(process.execPath, [CLI, "--repo", repo, "--output", output], { encoding: "utf8" });
  assert.equal(result.status, 0, `${result.stderr}\n${result.stdout}`);
  assert.equal(result.stderr, "");
  return fs.readFileSync(output, "utf8").trimEnd().split("\n").map(JSON.parse);
}

function recordsOf(records, kind) {
  return records.filter((record) => record.record === kind);
}

function spanKey(record) {
  const span = record.span || {};
  return [span.path || "", span.start_line || 0, span.start_column || 0, span.end_line || 0, span.end_column || 0];
}

function recordKey(record) {
  if (record.record === "node") return [0, record.id, record.kind, record.path, record.qualified_name];
  if (record.record === "edge") return [1, record.source, record.target, record.relation, ...spanKey(record)];
  return [2, record.source, record.relation, record.expression, record.reason, ...spanKey(record)];
}

function compare(left, right) {
  for (let index = 0; index < Math.max(left.length, right.length); index += 1) {
    const a = String(left[index] ?? "");
    const b = String(right[index] ?? "");
    if (a < b) return -1;
    if (a > b) return 1;
  }
  return 0;
}

function assertSortedObject(value) {
  if (Array.isArray(value)) {
    value.forEach(assertSortedObject);
    return;
  }
  if (value && typeof value === "object") {
    const keys = Object.keys(value);
    assert.deepEqual(keys, [...keys].sort());
    Object.values(value).forEach(assertSortedObject);
  }
}

test("extracts TypeScript declarations, imports, exports, inheritance, and exclusions", () => {
  const repo = makeFixture();
  const output = path.join(repo, "facts.jsonl");
  const records = runAdapter(repo, output);
  const nodes = recordsOf(records, "node");
  const edges = recordsOf(records, "edge");
  const unresolved = recordsOf(records, "unresolved");

  assert.deepEqual(records[0], {
    adapter_version: "0.1.0",
    language: "typescript",
    record: "lexicon",
    repository: path.basename(repo),
    schema_version: 1,
  });
  const kinds = new Set(nodes.map((record) => record.kind));
  for (const kind of ["repository", "directory", "file", "module", "type", "interface", "function", "method", "variable", "constant", "import", "export"]) assert.ok(kinds.has(kind), kind);
  const paths = new Set(nodes.map((record) => record.path));
  assert.ok(![...paths].some((item) => item.includes("node_modules") || item.includes("build") || item.includes(".git")));
  for (const relation of ["contains", "defines", "imports", "extends", "implements"]) assert.ok(edges.some((record) => record.relation === relation), relation);
  assert.ok(unresolved.some((record) => record.relation === "imports" && record.reason === "external-target"));
  assert.ok(unresolved.some((record) => record.relation === "imports" && record.reason === "dynamic-target"));
  assert.ok(nodes.some((record) => record.kind === "interface" && record.qualified_name === "src/child.Child"));
  assert.ok(nodes.some((record) => record.kind === "method" && record.qualified_name === "src/child.Child.method"));
  assert.ok(nodes.some((record) => record.kind === "constant" && record.qualified_name === "src/base.BASE"));
});

test("uses contract IDs, canonical ordering, and stable repeat runs", () => {
  const repo = makeFixture();
  const firstPath = path.join(repo, "first.jsonl");
  const secondPath = path.join(repo, "second.jsonl");
  const first = runAdapter(repo, firstPath);
  const second = runAdapter(repo, secondPath);
  assert.equal(fs.readFileSync(firstPath, "utf8"), fs.readFileSync(secondPath, "utf8"));

  const facts = first.slice(1);
  const nodes = recordsOf(facts, "node");
  assert.equal(new Set(nodes.map((record) => record.id)).size, nodes.length);
  for (const record of first) assertSortedObject(record);
  assert.deepEqual(facts, [...facts].sort((left, right) => compare(recordKey(left), recordKey(right))));

  const file = nodes.find((record) => record.kind === "file" && record.path === "src/base.ts");
  assert.ok(file);
  const expectedId = `sha256:${crypto.createHash("sha256").update("lexicon:v1\0typescript\0file\0src/base.ts").digest("hex")}`;
  const expectedContent = `sha256:${crypto.createHash("sha256").update(fs.readFileSync(path.join(repo, "src/base.ts"))).digest("hex")}`;
  assert.equal(file.id, expectedId);
  assert.equal(file.content_id, expectedContent);
  assert.ok(!first.some((record) => JSON.stringify(record).includes(repo)));
});

test("passes the repository JSONL validator", () => {
  const repo = makeFixture();
  const output = path.join(repo, "facts.jsonl");
  runAdapter(repo, output);
  const result = spawnSync("python", [path.join(REPO_ROOT, "tools", "validate_jsonl.py"), output], { encoding: "utf8" });
  assert.equal(result.status, 0, result.stderr);
});
