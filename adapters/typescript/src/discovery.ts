import * as fs from "node:fs";
import * as path from "node:path";

export const EXCLUDED_DIRECTORIES = new Set([
  ".git", ".worktrees", ".workingtrees", ".warlock", "node_modules",
  "build", "dist", "coverage", "target", "vendor", "tmp", "log",
  ".cache", ".turbo", ".next", ".nuxt", ".parcel-cache", ".pytest_cache",
  ".venv", "venv", "__pycache__", "out",
]);

export function normalizeRelative(root: string, absolutePath: string): string {
  return path.relative(root, absolutePath).split(path.sep).join("/") || ".";
}

export function moduleKeyFor(relativePath: string): string {
  return relativePath.replace(/\.(?:tsx?|mts|cts)$/i, "");
}

export type RepositoryFiles = { directories: string[]; files: string[] };

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
      } else if (entry.isFile() && /\.(?:ts|tsx)$/i.test(entry.name)) {
        files.push(normalizeRelative(root, path.join(directory, entry.name)));
      }
    }
  };
  walk(root);
  directories.sort();
  files.sort();
  return { directories, files };
}
