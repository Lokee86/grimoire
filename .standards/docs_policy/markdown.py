from __future__ import annotations

import re
from pathlib import Path
from urllib.parse import unquote

HEADING_RE = re.compile(r"^#{1,6}\s+(.+?)\s*#*\s*$", re.MULTILINE)
LINK_RE = re.compile(r"!?\[[^\]]*\]\(([^)]+)\)")
SCHEME_RE = re.compile(r"^[a-zA-Z][a-zA-Z0-9+.-]*:")
INLINE_CODE_RE = re.compile(r"`[^`\n]*`")


def prose_text(text: str) -> str:
    lines: list[str] = []
    fence_char: str | None = None
    fence_length = 0
    for line in text.splitlines(keepends=True):
        stripped = line.lstrip()
        marker_match = re.match(r"^([`~])\1{2,}", stripped)
        if fence_char is None and marker_match:
            marker = marker_match.group(0)
            fence_char = marker[0]
            fence_length = len(marker)
            continue
        if fence_char is not None:
            closing = re.match(rf"^{re.escape(fence_char)}{{{fence_length},}}(?:\s|$)", stripped)
            if closing:
                fence_char = None
                fence_length = 0
            continue
        lines.append(line)
    return INLINE_CODE_RE.sub("", "".join(lines))


def normalize_heading(value: str) -> str:
    value = re.sub(r"[`*_~]", "", value)
    return " ".join(value.strip().lower().split())


def headings(text: str) -> set[str]:
    return {normalize_heading(match.group(1)) for match in HEADING_RE.finditer(prose_text(text))}


def markdown_targets(text: str) -> list[str]:
    targets: list[str] = []
    for match in LINK_RE.finditer(prose_text(text)):
        target = match.group(1).strip()
        if target.startswith("<") and target.endswith(">"):
            target = target[1:-1]
        elif " \"" in target or " '" in target:
            target = target.split(maxsplit=1)[0]
        targets.append(target)
    return targets


def local_target(source: Path, target: str, repo: Path) -> Path | None:
    if not target or target.startswith("#") or target.startswith("//") or SCHEME_RE.match(target):
        return None
    cleaned = unquote(target.split("#", 1)[0].split("?", 1)[0])
    if not cleaned:
        return None
    candidate = (source.parent / cleaned).resolve()
    try:
        candidate.relative_to(repo.resolve())
    except ValueError:
        return candidate
    return candidate


def linked_relative_paths(index_path: Path, repo: Path) -> set[str]:
    text = index_path.read_text(encoding="utf-8")
    result: set[str] = set()
    for target in markdown_targets(text):
        candidate = local_target(index_path, target, repo)
        if candidate is None:
            continue
        try:
            result.add(candidate.relative_to(repo.resolve()).as_posix())
        except ValueError:
            continue
    return result
