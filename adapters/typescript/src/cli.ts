#!/usr/bin/env node
import { writeFacts } from "./adapter";

function usage(): string {
  return "Usage: node dist/cli.js --repo <repository> --output <jsonl path|->";
}

function main(argv: string[]): number {
  let repository: string | undefined;
  let output: string | undefined;
  for (let index = 0; index < argv.length; index += 1) {
    const argument = argv[index];
    if (argument === "--repo") repository = argv[++index];
    else if (argument === "--output") output = argv[++index];
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
    writeFacts(repository, output);
    return 0;
  } catch (error) {
    process.stderr.write(`${error instanceof Error ? error.message : String(error)}\n`);
    return 1;
  }
}

process.exitCode = main(process.argv.slice(2));
