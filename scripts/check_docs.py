#!/usr/bin/env python3
"""Validate Grimoire's documentation surface and local links."""

from __future__ import annotations

import re
import sys
from pathlib import Path
from urllib.parse import unquote, urlsplit

ROOT = Path(__file__).resolve().parents[1]
LINK_PATTERN = re.compile(r"(?<!!)\[[^\]]*\]\(([^)]+)\)")
HTML_TARGET_PATTERN = re.compile(r"(?:src|href)=[\"']([^\"']+)[\"']", re.IGNORECASE)
EXCLUDED_PARTS = {
    ".git",
    ".worktrees",
    ".workingtrees",
    ".grimoire",
    ".lexicon",
    ".arcana",
    ".warlock",
    ".tmp",
    "bin",
    "build",
    "dist",
    "node_modules",
    "vendor",
    "__pycache__",
    ".pytest_cache",
}
REQUIRED_DOCUMENTS = (
    "README.md",
    "docs/INDEX.md",
    "docs/architecture/analysis-stack.md",
    "docs/architecture/codemap.md",
    "docs/reference/lexicon.md",
    "docs/reference/arcana.md",
    "internal/evidence/README.md",
    "internal/lexical/README.md",
    "lexicon/docs/README.md",
    "lexicon/docs/CODEMAP.md",
    "arcana/docs/README.md",
    "arcana/docs/APPLICATION.md",
    "arcana/docs/ARCHITECTURE.md",
    "arcana/docs/CODEMAP.md",
    "arcana/docs/DEVELOPMENT.md",
    "arcana/docs/STATUS.md",
)


def excluded(path: Path) -> bool:
    for part in path.relative_to(ROOT).parts:
        if part in EXCLUDED_PARTS or part.startswith("target-v"):
            return True
    return False


def documentation_files() -> list[Path]:
    return sorted(path for path in ROOT.rglob("*.md") if path.is_file() and not excluded(path))


def local_link_target(raw: str) -> str | None:
    value = raw.strip()
    if value.startswith("<") and ">" in value:
        value = value[1 : value.index(">")]
    else:
        value = value.split(maxsplit=1)[0]
    if not value or value.startswith("#"):
        return None
    parsed = urlsplit(value)
    if parsed.scheme or parsed.netloc:
        return None
    return unquote(parsed.path) or None


def validate_required_documents() -> list[str]:
    return [f"missing required documentation: {relative}" for relative in REQUIRED_DOCUMENTS if not (ROOT / relative).is_file()]


def validate_local_links() -> list[str]:
    failures: list[str] = []
    for document in documentation_files():
        text = document.read_text(encoding="utf-8")
        raw_targets = [match.group(1) for match in LINK_PATTERN.finditer(text)]
        raw_targets.extend(match.group(1) for match in HTML_TARGET_PATTERN.finditer(text))
        for raw_target in raw_targets:
            target = local_link_target(raw_target)
            if target is None:
                continue
            resolved = (document.parent / target).resolve()
            if not resolved.exists():
                failures.append(
                    f"{document.relative_to(ROOT)}: missing local link target {target!r}"
                )
    return failures


def validate_index_visibility() -> list[str]:
    failures: list[str] = []
    required_links = {
        ROOT / "README.md": (
            "docs/architecture/codemap.md",
            "docs/reference/lexicon.md",
            "docs/reference/arcana.md",
        ),
        ROOT / "lexicon" / "docs" / "README.md": ("CODEMAP.md",),
        ROOT / "arcana" / "docs" / "README.md": (
            "APPLICATION.md",
            "ARCHITECTURE.md",
            "CODEMAP.md",
            "DEVELOPMENT.md",
            "STATUS.md",
        ),
    }
    for document, targets in required_links.items():
        if not document.is_file():
            continue
        text = document.read_text(encoding="utf-8")
        for target in targets:
            if f"]({target})" not in text:
                failures.append(
                    f"{document.relative_to(ROOT)}: required documentation link is not indexed: {target}"
                )
    return failures


def validate_repository() -> list[str]:
    return validate_required_documents() + validate_local_links() + validate_index_visibility()


def main() -> int:
    failures = validate_repository()
    if failures:
        print("documentation validation failed:", file=sys.stderr)
        for failure in failures:
            print(f"- {failure}", file=sys.stderr)
        return 1
    print(f"Documentation validation passed for {len(documentation_files())} Markdown files.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
