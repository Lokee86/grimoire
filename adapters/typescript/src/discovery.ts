import * as fs from "node:fs";
import * as path from "node:path";
import * as ts from "typescript";

export const EXCLUDED_DIRECTORIES = new Set([
  ".git", ".worktrees", ".workingtrees", ".warlock", ".astro", "node_modules",
  "build", "dist", "coverage", "target", "vendor", "tmp", "log",
  ".cache", ".turbo", ".next", ".nuxt", ".parcel-cache", ".pytest_cache",
  ".venv", "venv", "__pycache__", "out",
]);

export function normalizeRelative(root: string, absolutePath: string): string {
  return path.relative(root, absolutePath).split(path.sep).join("/") || ".";
}

export function moduleKeyFor(relativePath: string): string {
  return relativePath.replace(/(?:\.d)?\.(?:[cm]?[jt]sx?)$/i, "");
}

export type RepositoryFiles = { directories: string[]; files: string[] };

export type PathMapping = {
  baseUrl: string;
  pattern: string;
  targets: string[];
};

export type TypeScriptProgram = {
  checker: ts.TypeChecker;
  program: ts.Program;
};

export function createTypeScriptProgram(root: string, files: string[]): TypeScriptProgram {
  const configPath = ["tsconfig.json", "jsconfig.json"]
    .map((name) => path.join(root, name))
    .find((candidate) => fs.existsSync(candidate) && fs.statSync(candidate).isFile());
  let options: ts.CompilerOptions = {
    allowJs: true,
    checkJs: true,
    jsx: ts.JsxEmit.Preserve,
    module: ts.ModuleKind.ESNext,
    moduleResolution: ts.ModuleResolutionKind.Node10,
    noEmit: true,
    skipLibCheck: true,
    target: ts.ScriptTarget.ESNext,
  };
  if (configPath) {
    const config = ts.readConfigFile(configPath, ts.sys.readFile);
    if (!config.error) {
      const parsed = ts.parseJsonConfigFileContent(config.config, ts.sys, path.dirname(configPath), { noEmit: true, skipLibCheck: true }, configPath);
      options = {
        ...parsed.options,
        allowJs: true,
        checkJs: true,
        jsx: parsed.options.jsx ?? ts.JsxEmit.Preserve,
        noEmit: true,
        skipLibCheck: true,
      };
    }
  }
  const rootNames = files.map((relativePath) => path.join(root, relativePath.split("/").join(path.sep)));
  const program = ts.createProgram({ rootNames, options });
  return { checker: program.getTypeChecker(), program };
}

export function readPathMappings(root: string): PathMapping[] {
  const configPath = ["tsconfig.json", "jsconfig.json"]
    .map((name) => path.join(root, name))
    .find((candidate) => fs.existsSync(candidate) && fs.statSync(candidate).isFile());
  if (!configPath) return [];

  const parsed = ts.parseConfigFileTextToJson(configPath, fs.readFileSync(configPath, "utf8"));
  const compilerOptions = parsed.config?.compilerOptions;
  const paths = compilerOptions?.paths;
  if (!paths || typeof paths !== "object" || Array.isArray(paths)) return [];
  const configDirectory = path.dirname(configPath);
  const baseUrl = normalizeRelative(root, path.resolve(configDirectory, typeof compilerOptions.baseUrl === "string" ? compilerOptions.baseUrl : "."));
  const mappings: PathMapping[] = [];
  for (const [pattern, rawTargets] of Object.entries(paths as Record<string, unknown>)) {
    const wildcardCount = (pattern.match(/\*/g) ?? []).length;
    if (wildcardCount > 1 || !Array.isArray(rawTargets)) continue;
    const targets = rawTargets.filter((target): target is string => typeof target === "string" && (target.match(/\*/g) ?? []).length <= 1);
    if (targets.length > 0) mappings.push({ baseUrl, pattern, targets });
  }
  return mappings;
}

export function scanRepository(root: string): RepositoryFiles {
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
      } else if (entry.isFile() && /\.(?:[cm]?[jt]sx?)$/i.test(entry.name)) {
        files.push(normalizeRelative(root, path.join(directory, entry.name)));
      }
    }
  };
  walk(root);
  directories.sort();
  files.sort();
  return { directories, files };
}
