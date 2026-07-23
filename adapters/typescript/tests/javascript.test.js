const assert = require("node:assert/strict");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const { spawnSync } = require("node:child_process");
const test = require("node:test");

const ADAPTER_ROOT = path.resolve(__dirname, "..");
const CLI = path.join(ADAPTER_ROOT, "dist", "cli.js");

function makeJavaScriptFixture() {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), "lexicon-javascript-"));
  const files = {
    "jsconfig.json": JSON.stringify({
      compilerOptions: {
        allowJs: true,
        checkJs: true,
        jsx: "preserve",
        module: "NodeNext",
        moduleResolution: "NodeNext",
        target: "ES2022",
      },
    }),
    "src/esm-target.js": [
      "export function esmHelper() {}",
      "export class EsmWorker { run() {} }",
      "export default function defaultHelper() {}",
      "",
    ].join("\n"),
    "src/esm-consumer.mjs": [
      'import defaultHelper, { esmHelper, EsmWorker } from "./esm-target.js";',
      "export function esmCaller() {",
      "  esmHelper();",
      "  defaultHelper();",
      "  new EsmWorker().run();",
      "}",
      "",
    ].join("\n"),
    "src/jsx-component.jsx": [
      "export function JsChild() { return null; }",
      "export function JsParent() { return <JsChild />; }",
      "",
    ].join("\n"),
    "src/cjs-target.cjs": [
      "function cjsHelper() {}",
      "class CjsWorker { run() {} }",
      "module.exports = { cjsHelper, CjsWorker };",
      "",
    ].join("\n"),
    "src/cjs-consumer.js": [
      'const { cjsHelper, CjsWorker } = require("./cjs-target.cjs");',
      'const cjs = require("./cjs-target.cjs");',
      "function cjsCaller() {",
      "  cjsHelper();",
      "  new CjsWorker();",
      "  cjs.cjsHelper();",
      "  new cjs.CjsWorker();",
      "}",
      "module.exports = cjsCaller;",
      "",
    ].join("\n"),
    "src/default-target.js": [
      "module.exports = function cjsDefault() {};",
      "",
    ].join("\n"),
    "src/default-consumer.js": [
      'const cjsDefault = require("./default-target");',
      "export function defaultCaller() { cjsDefault(); }",
      "",
    ].join("\n"),
    ".astro/generated.mjs": "export function generated() {}\n",
    "src/jsdoc-flow.js": [
      "/** @param {() => void} callback */",
      "export function invokeCallback(callback) { callback(); }",
      "export function localCallback() {}",
      "export function higherOrder() { invokeCallback(localCallback); }",
      "",
    ].join("\n"),
    "src/runtime-flow.jsx": [
      "export function first() {}",
      "export function second() {}",
      "export function invoke(callback) { callback(); }",
      "export function choose(flag) { invoke(flag ? first : second); }",
      "export function configure({ handler }) { handler(); }",
      "export function configureCaller() { configure({ handler: first }); }",
      "export function DynamicView({ Component }) { return <Component />; }",
      "export class Clocked {",
      "  constructor({ now = Date.now } = {}) { this.now = now; }",
      "  read() { return this.now(); }",
      "}",
      "export function readClock() { return new Clocked().read(); }",
      "",
    ].join("\n"),
  };
  for (const [relative, content] of Object.entries(files)) {
    const target = path.join(root, relative);
    fs.mkdirSync(path.dirname(target), { recursive: true });
    fs.writeFileSync(target, content);
  }
  return root;
}

function runAdapter(repo) {
  const output = path.join(repo, "facts.jsonl");
  const result = spawnSync(process.execPath, [CLI, "--repo", repo, "--output", output], { encoding: "utf8" });
  assert.equal(result.status, 0, `${result.stderr}\n${result.stdout}`);
  assert.equal(result.stderr, "");
  return fs.readFileSync(output, "utf8").trimEnd().split("\n").map(JSON.parse);
}

function findNode(nodes, qualifiedName) {
  const node = nodes.find((record) => record.qualified_name === qualifiedName);
  assert.ok(node, qualifiedName);
  return node;
}

function hasEdge(edges, source, target, relation, attributes) {
  return edges.some((record) => record.source === source.id
    && record.target === target.id
    && record.relation === relation
    && (!attributes || Object.entries(attributes).every(([key, value]) => record.attributes?.[key] === value)));
}

test("scans and resolves JavaScript ESM, JSX, CommonJS, and JSDoc call flow", () => {
  const records = runAdapter(makeJavaScriptFixture());
  const nodes = records.filter((record) => record.record === "node");
  const edges = records.filter((record) => record.record === "edge");
  const unresolved = records.filter((record) => record.record === "unresolved");
  const paths = new Set(nodes.filter((record) => record.kind === "file").map((record) => record.path));
  assert.ok(!paths.has(".astro/generated.mjs"));
  for (const file of [
    "src/esm-target.js",
    "src/esm-consumer.mjs",
    "src/jsx-component.jsx",
    "src/cjs-target.cjs",
    "src/cjs-consumer.js",
  ]) assert.ok(paths.has(file), file);

  const esmCaller = findNode(nodes, "src/esm-consumer.esmCaller");
  const esmHelper = findNode(nodes, "src/esm-target.esmHelper");
  const defaultHelper = findNode(nodes, "src/esm-target.defaultHelper");
  const esmWorker = findNode(nodes, "src/esm-target.EsmWorker");
  const workerRun = findNode(nodes, "src/esm-target.EsmWorker.run");
  assert.ok(hasEdge(edges, esmCaller, esmHelper, "calls"));
  assert.ok(hasEdge(edges, esmCaller, defaultHelper, "calls"));
  assert.ok(hasEdge(edges, esmCaller, esmWorker, "calls"));
  assert.ok(hasEdge(edges, esmCaller, workerRun, "calls"));

  const jsParent = findNode(nodes, "src/jsx-component.JsParent");
  const jsChild = findNode(nodes, "src/jsx-component.JsChild");
  assert.ok(hasEdge(edges, jsParent, jsChild, "calls"));

  const cjsCaller = findNode(nodes, "src/cjs-consumer.cjsCaller");
  const cjsHelper = findNode(nodes, "src/cjs-target.cjsHelper");
  const cjsWorker = findNode(nodes, "src/cjs-target.CjsWorker");
  assert.equal(edges.filter((record) => record.source === cjsCaller.id && record.target === cjsHelper.id && record.relation === "calls").length, 2);
  assert.equal(edges.filter((record) => record.source === cjsCaller.id && record.target === cjsWorker.id && record.relation === "calls").length, 2);

  const defaultCaller = findNode(nodes, "src/default-consumer.defaultCaller");
  const cjsDefault = findNode(nodes, "src/default-target.cjsDefault");
  assert.ok(hasEdge(edges, defaultCaller, cjsDefault, "calls"));

  const higherOrder = findNode(nodes, "src/jsdoc-flow.higherOrder");
  const invokeCallback = findNode(nodes, "src/jsdoc-flow.invokeCallback");
  const localCallback = findNode(nodes, "src/jsdoc-flow.localCallback");
  assert.ok(hasEdge(edges, higherOrder, invokeCallback, "calls"));
  assert.ok(hasEdge(edges, invokeCallback, localCallback, "calls"));
  assert.ok(hasEdge(edges, higherOrder, localCallback, "possible-calls", { callback: true }));

  const invoke = findNode(nodes, "src/runtime-flow.invoke");
  const first = findNode(nodes, "src/runtime-flow.first");
  const second = findNode(nodes, "src/runtime-flow.second");
  const configure = findNode(nodes, "src/runtime-flow.configure");
  const dynamicView = findNode(nodes, "src/runtime-flow.DynamicView");
  const clockRead = findNode(nodes, "src/runtime-flow.Clocked.read");
  assert.ok(hasEdge(edges, invoke, first, "possible-calls"));
  assert.ok(hasEdge(edges, invoke, second, "possible-calls"));
  assert.ok(hasEdge(edges, configure, first, "calls"));
  assert.ok(unresolved.some((record) => record.source === invoke.id && record.reason === "dynamic-target"));
  assert.ok(!unresolved.some((record) => record.source === invoke.id && record.reason === "ambiguous-target"));
  assert.ok(unresolved.some((record) => record.source === dynamicView.id && record.reason === "dynamic-target"));
  assert.ok(unresolved.some((record) => record.source === clockRead.id && record.reason === "external-target"));

  assert.ok(!unresolved.some((record) => record.relation === "calls" && [
    "missing-target",
    "ambiguous-target",
    "unsupported-form",
  ].includes(record.reason)));
});
