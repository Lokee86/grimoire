import * as fs from "node:fs";
import * as path from "node:path";
import * as ts from "typescript";

const SVELTE_EXTENSION = ".svelte";
const SCRIPT_BLOCK = /<script\b([^>]*)>([\s\S]*?)<\/script\s*>/gi;

type SvelteSource = {
  scriptKind: ts.ScriptKind;
  text: string;
};

export function isSvelteFile(fileName: string): boolean {
  return fileName.toLowerCase().endsWith(SVELTE_EXTENSION);
}

export function extractSvelteSource(source: string): SvelteSource {
  const output: string[] = source.split("").map((character) => character === "\n" || character === "\r" ? character : " ");
  let hasTypeScript = false;
  let match: RegExpExecArray | null;

  SCRIPT_BLOCK.lastIndex = 0;
  while ((match = SCRIPT_BLOCK.exec(source)) !== null) {
    const language = scriptLanguage(match[1]);
    if (!language) continue;

    const openingLength = match[0].indexOf(">") + 1;
    const contentStart = match.index + openingLength;
    const closingStart = match.index + match[0].lastIndexOf("</script");
    for (let index = contentStart; index < closingStart; index += 1) output[index] = source[index];
    if (language === "typescript") hasTypeScript = true;
  }

  return {
    scriptKind: hasTypeScript ? ts.ScriptKind.TS : ts.ScriptKind.JS,
    text: output.join(""),
  };
}

export function createSvelteCompilerHost(
  root: string,
  options: ts.CompilerOptions,
): ts.CompilerHost {
  const host = ts.createCompilerHost(options, true);
  const defaultGetSourceFile = host.getSourceFile.bind(host);

  host.getSourceFile = (fileName, languageVersion, onError, shouldCreateNewSourceFile) => {
    if (!isSvelteFile(fileName)) {
      return defaultGetSourceFile(fileName, languageVersion, onError, shouldCreateNewSourceFile);
    }
    try {
      const source = fs.readFileSync(fileName, "utf8");
      const extracted = extractSvelteSource(source);
      return ts.createSourceFile(fileName, extracted.text, languageVersion, true, extracted.scriptKind);
    } catch (error) {
      onError?.(error instanceof Error ? error.message : String(error));
      return undefined;
    }
  };

  host.resolveModuleNames = (moduleNames, containingFile) => moduleNames.map((moduleName) => {
    const standard = ts.resolveModuleName(moduleName, containingFile, options, host).resolvedModule;
    if (standard) return standard;
    const resolvedFileName = resolveSvelteModule(root, containingFile, moduleName, options);
    if (!resolvedFileName) return undefined;
    return {
      extension: ts.Extension.Ts,
      isExternalLibraryImport: false,
      resolvedFileName,
    };
  });

  return host;
}

function scriptLanguage(attributes: string): "javascript" | "typescript" | null {
  const match = attributes.match(/\blang\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s"'=<>`]+))/i);
  if (!match) return "javascript";
  const language = (match[1] ?? match[2] ?? match[3] ?? "").toLowerCase();
  if (language === "ts" || language === "typescript") return "typescript";
  if (language === "js" || language === "javascript") return "javascript";
  return null;
}

function resolveSvelteModule(
  root: string,
  containingFile: string,
  moduleName: string,
  options: ts.CompilerOptions,
): string | null {
  const candidates = moduleCandidates(root, containingFile, moduleName, options);
  for (const candidate of candidates) {
    for (const fileName of svelteCandidates(candidate)) {
      if (insideRoot(root, fileName) && fs.existsSync(fileName) && fs.statSync(fileName).isFile()) return fileName;
    }
  }
  return null;
}

function moduleCandidates(
  root: string,
  containingFile: string,
  moduleName: string,
  options: ts.CompilerOptions,
): string[] {
  if (moduleName.startsWith(".") || path.isAbsolute(moduleName)) {
    return [path.resolve(path.dirname(containingFile), moduleName)];
  }

  const candidates: string[] = [];
  const mappings = options.paths ?? {};
  for (const [pattern, targets] of Object.entries(mappings)) {
    const capture = matchPathPattern(pattern, moduleName);
    if (capture === null) continue;
    const baseUrl = options.baseUrl ? path.resolve(options.baseUrl) : root;
    for (const target of targets) candidates.push(path.resolve(baseUrl, target.replace("*", capture)));
  }
  if (options.baseUrl) candidates.push(path.resolve(options.baseUrl, moduleName));
  return candidates;
}

function matchPathPattern(pattern: string, moduleName: string): string | null {
  const wildcard = pattern.indexOf("*");
  if (wildcard < 0) return pattern === moduleName ? "" : null;
  if (pattern.indexOf("*", wildcard + 1) >= 0) return null;
  const prefix = pattern.slice(0, wildcard);
  const suffix = pattern.slice(wildcard + 1);
  if (!moduleName.startsWith(prefix) || !moduleName.endsWith(suffix)) return null;
  return moduleName.slice(prefix.length, moduleName.length - suffix.length || undefined);
}

function svelteCandidates(candidate: string): string[] {
  if (isSvelteFile(candidate)) return [candidate];
  return [`${candidate}${SVELTE_EXTENSION}`, path.join(candidate, `index${SVELTE_EXTENSION}`)];
}

function insideRoot(root: string, fileName: string): boolean {
  const relative = path.relative(path.resolve(root), path.resolve(fileName));
  return relative === "" || (!relative.startsWith(`..${path.sep}`) && relative !== ".." && !path.isAbsolute(relative));
}
