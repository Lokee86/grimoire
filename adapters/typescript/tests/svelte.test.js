const assert = require("node:assert/strict");
const crypto = require("node:crypto");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const { spawnSync } = require("node:child_process");
const test = require("node:test");
const ts = require("typescript");

const ADAPTER_ROOT = path.resolve(__dirname, "..");
const CLI = path.join(ADAPTER_ROOT, "dist", "cli.js");
const { extractSvelteSource } = require(path.join(ADAPTER_ROOT, "dist", "svelte.js"));

function writeFixture() {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), "lexicon-svelte-"));
  const files = {
    "tsconfig.json": JSON.stringify({
      compilerOptions: {
        baseUrl: ".",
        paths: { "@/*": ["src/*"] },
      },
    }),
    "src/helper.ts": "export function helper(): void {}\n",
    "src/Widget.svelte": [
      '<script context="module" lang="ts">',
      "  export const moduleValue = 1;",
      "</script>",
      "<h1>function fakeMarkupFunction() {}</h1>",
      '<script lang="ts">',
      '  import { helper } from "./helper";',
      "  const FLOW = 1;",
      "  class Box { field = 0; }",
      "  export function flow(value: number): number {",
      "    let local = value;",
      "    local += FLOW;",
      "    local++;",
      "    const box = new Box();",
      "    box.field = local;",
      "    return local + value;",
      "  }",
      "  export function inner(value: number): number { return value; }",
      "  export let name: string;",
      "  export function greet(): string {",
      "    helper();",
      "    return name;",
      "  }",
      "</script>",
      "<style>.fake { content: 'function fakeStyleFunction() {}'; }</style>",
      "<button on:click={greet}>{name}</button>",
      "",
    ].join("\n"),
    "src/App.svelte": [
      '<script lang="ts">',
      '  import Widget from "@/Widget.svelte";',
      "  export const component = Widget;",
      "</script>",
      "<Widget />",
      "",
    ].join("\n"),
    "src/Legacy.svelte": [
      "<script>",
      "  export function legacy() { return 1; }",
      "</script>",
      "<p>{legacy()}</p>",
      "",
    ].join("\n"),
    "src/Unsupported.svelte": [
      '<script lang="coffee">',
      "  fakeUnsupported = -> 1",
      "</script>",
      "",
    ].join("\n"),
  };
  for (const [relativePath, content] of Object.entries(files)) {
    const target = path.join(root, relativePath);
    fs.mkdirSync(path.dirname(target), { recursive: true });
    fs.writeFileSync(target, content);
  }
  return root;
}

function runAdapter(repo) {
  const output = path.join(repo, "facts.jsonl");
  const result = spawnSync(process.execPath, [CLI, "--repo", repo, "--output", output], { encoding: "utf8" });
  assert.equal(result.status, 0, `${result.stderr}\n${result.stdout}`);
  return fs.readFileSync(output, "utf8").trimEnd().split("\n").map(JSON.parse);
}

test("masks Svelte markup while preserving original source offsets", () => {
  const source = [
    '<script context="module">',
    "  export const plain = 1;",
    "</script>",
    "<div>ignored</div>",
    '<script lang="ts">',
    "  export function typed(): number { return plain; }",
    "</script>",
  ].join("\r\n");
  const extracted = extractSvelteSource(source);
  assert.equal(extracted.text.length, source.length);
  assert.equal(extracted.scriptKind, ts.ScriptKind.TS);
  assert.ok(extracted.text.includes("export const plain = 1;"));
  assert.ok(extracted.text.includes("export function typed"));
  assert.ok(!extracted.text.includes("ignored"));
  for (let index = 0; index < source.length; index += 1) {
    if (source[index] === "\r" || source[index] === "\n") assert.equal(extracted.text[index], source[index]);
  }
});

test("analyzes Svelte scripts through the TypeScript adapter", () => {
  const repo = writeFixture();
  const records = runAdapter(repo);
  const nodes = records.filter((record) => record.record === "node");
  const edges = records.filter((record) => record.record === "edge");
  const unresolved = records.filter((record) => record.record === "unresolved");
  const node = (qualifiedName) => nodes.find((record) => record.qualified_name === qualifiedName);

  const widgetModule = node("src/Widget");
  const appModule = node("src/App");
  const helper = node("src/helper.helper");
  const greet = node("src/Widget.greet");
  const flow = node("src/Widget.flow");
  const inner = node("src/Widget.inner");
  const moduleValue = node("src/Widget.moduleValue");
  const legacy = node("src/Legacy.legacy");
  assert.ok(widgetModule && appModule && helper && greet && flow && inner && moduleValue && legacy);
  assert.ok(!nodes.some((record) => record.qualified_name?.includes("fakeMarkupFunction")));
  assert.ok(!nodes.some((record) => record.qualified_name?.includes("fakeStyleFunction")));
  assert.ok(!nodes.some((record) => record.qualified_name?.includes("fakeUnsupported")));
  assert.ok(edges.some((record) => record.source === greet.id && record.target === helper.id && record.relation === "calls"));
  const byId = new Map(nodes.map((record) => [record.id, record]));
  const flowEdges = edges.filter((record) => record.source === flow.id && ["reads", "writes"].includes(record.relation));
  assert.ok(flowEdges.some((record) => byId.get(record.target)?.name === "FLOW"));
  assert.ok(flowEdges.some((record) => byId.get(record.target)?.name === "field"));
  assert.ok(flowEdges.some((record) => byId.get(record.target)?.name === "local" && record.relation === "reads"));
  assert.ok(flowEdges.some((record) => byId.get(record.target)?.name === "local" && record.relation === "writes"));
  assert.ok(edges.filter((record) => ["reads", "writes"].includes(record.relation)).every((record) => byId.has(record.target)));
  assert.ok(edges.some((record) => record.source === appModule.id && record.target === widgetModule.id && record.relation === "imports"));
  assert.equal(greet.span.path, "src/Widget.svelte");
  assert.equal(greet.span.start_line, 19);
  assert.ok(!unresolved.some((record) => record.candidate_name === "syntax-diagnostics" && record.expression === "src/Widget.svelte"));

  const file = nodes.find((record) => record.kind === "file" && record.path === "src/Widget.svelte");
  const expectedContent = `sha256:${crypto.createHash("sha256").update(fs.readFileSync(path.join(repo, "src/Widget.svelte"))).digest("hex")}`;
  assert.equal(file.content_id, expectedContent);
});
