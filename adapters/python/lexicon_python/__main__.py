"""Command-line entry point for the Lexicon Python adapter."""

from __future__ import annotations

import argparse
from pathlib import Path

from .adapter import write_facts


def main() -> int:
    parser = argparse.ArgumentParser(description="Emit Lexicon facts v1 for a Python repository")
    parser.add_argument("--repo", required=True, type=Path, help="repository root to scan")
    parser.add_argument("--output", required=True, type=Path, help="JSONL output path (use - for stdout)")
    args = parser.parse_args()
    write_facts(args.repo, args.output)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
