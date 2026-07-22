#!/usr/bin/env node
import { writeFacts } from "./orchestration";

function usage(): string {
  return "Usage: node dist/cli.js --repo <repository> --output <jsonl path|->";
}

function main(argv: string[]): number {
  let repository: string | undefined;
  let output: string | undefined;
  let changedFiles: string[] | undefined;
  let removedFiles: string[] | undefined;
  for (let index = 0; index < argv.length; index += 1) {
    const argument = argv[index];
    if (argument === "--repo") repository = argv[++index];
    else if (argument === "--output") output = argv[++index];
    else if (argument === "--changed-file") (changedFiles ??= []).push(argv[++index]);
    else if (argument === "--removed-file") (removedFiles ??= []).push(argv[++index]);
    else if (argument === "--help" || argument === "-h") {
      process.stdout.write(`${usage()}\n`);
      return 0;
    } else {
      process.stderr.write(`Unknown argument: ${argument}\n${usage()}\n`);
      return 2;
    }
  }
  if (!repository || !output) {
    process.stderr.write(`--repo and --output are required\n${usage()}\n`);
    return 2;
  }
  try {
    writeFacts(repository, output, changedFiles, removedFiles);
    return 0;
  } catch (error) {
    process.stderr.write(`${error instanceof Error ? error.message : String(error)}\n`);
    return 1;
  }
}

process.exitCode = main(process.argv.slice(2));
