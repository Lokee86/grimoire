from __future__ import annotations

import fnmatch
import subprocess
from dataclasses import dataclass
from pathlib import Path

from markdown import headings, linked_relative_paths, local_target, markdown_targets, normalize_heading
from model import Config, MappingRule


@dataclass(frozen=True, order=True)
class Finding:
    code: str
    path: str
    message: str


def matches(path: str, patterns: tuple[str, ...]) -> bool:
    return any(fnmatch.fnmatchcase(path, pattern) for pattern in patterns)


def all_repo_paths(repo: Path) -> list[str]:
    ignored = {".git", ".worktrees", ".workingtrees", "node_modules", "__pycache__"}
    result: list[str] = []
    for path in repo.rglob("*"):
        if any(part in ignored for part in path.relative_to(repo).parts):
            continue
        result.append(path.relative_to(repo).as_posix())
    return sorted(result)


def audit_required_paths(repo: Path, config: Config) -> list[Finding]:
    return [
        Finding("missing-required-path", path, "required documentation surface is missing")
        for path in config.required_paths
        if not (repo / path).exists()
    ]


def _expected_index(repo: Path, directory: Path, config: Config) -> Path:
    docs_root = (repo / config.docs_root).resolve()
    if directory.resolve() == docs_root:
        return repo / config.root_index
    return directory / config.folder_index


def audit_indexes(repo: Path, config: Config) -> list[Finding]:
    docs_root = repo / config.docs_root
    if not docs_root.is_dir():
        return [Finding("missing-docs-root", config.docs_root, "configured docs root does not exist")]

    findings: list[Finding] = []
    for directory in sorted(path for path in docs_root.rglob("*") if path.is_dir()):
        relative_parts = directory.relative_to(docs_root).parts
        if any(part in config.index_exempt_directories or part.startswith(".") for part in relative_parts):
            continue
        findings.extend(_audit_one_index(repo, directory, config))
    findings.extend(_audit_one_index(repo, docs_root, config))
    return findings


def _audit_one_index(repo: Path, directory: Path, config: Config) -> list[Finding]:
    index = _expected_index(repo, directory, config)
    index_rel = index.relative_to(repo).as_posix()
    if not index.is_file():
        return [Finding("missing-index", index_rel, "documentation folder lacks its configured index")]

    linked = linked_relative_paths(index, repo)
    findings: list[Finding] = []
    for child in sorted(directory.iterdir()):
        if child.name.startswith("."):
            continue
        child_rel = child.relative_to(repo).as_posix()
        if child.is_file() and child.suffix.lower() == ".md" and child.resolve() != index.resolve():
            if child_rel not in linked:
                findings.append(Finding("unindexed-document", child_rel, f"not linked from {index_rel}"))
        elif child.is_dir() and child.name not in config.index_exempt_directories:
            child_index = _expected_index(repo, child, config)
            child_index_rel = child_index.relative_to(repo).as_posix()
            if child_index_rel not in linked:
                findings.append(Finding("unindexed-folder", child_rel, f"folder index not linked from {index_rel}"))
    return findings


def audit_links(repo: Path, config: Config) -> list[Finding]:
    findings: list[Finding] = []
    docs_root = repo / config.docs_root
    for source in sorted(docs_root.rglob("*.md")):
        text = source.read_text(encoding="utf-8")
        for target in markdown_targets(text):
            candidate = local_target(source, target, repo)
            if candidate is not None and not candidate.exists():
                findings.append(
                    Finding(
                        "broken-link",
                        source.relative_to(repo).as_posix(),
                        f"target does not exist: {target}",
                    )
                )
    return findings


def audit_sections(repo: Path, config: Config) -> list[Finding]:
    findings: list[Finding] = []
    required = {normalize_heading(section): section for section in config.required_sections}
    for path in sorted((repo / config.docs_root).rglob("*.md")):
        relative = path.relative_to(repo).as_posix()
        if relative == config.root_index or path.name == config.folder_index:
            continue
        if matches(relative, config.section_exemptions):
            continue
        present = headings(path.read_text(encoding="utf-8"))
        for normalized, display in required.items():
            if normalized not in present:
                findings.append(Finding("missing-section", relative, f"missing section: {display}"))
    return findings


def _patterns_have_match(paths: list[str], patterns: tuple[str, ...]) -> bool:
    return any(matches(path, patterns) for path in paths)


def audit_coverage(repo: Path, config: Config, paths: list[str]) -> list[Finding]:
    findings: list[Finding] = []
    for index, rule in enumerate(config.coverage, start=1):
        if not _patterns_have_match(paths, rule.paths):
            continue
        if not _patterns_have_match(paths, rule.docs):
            findings.append(
                Finding(
                    "missing-coverage-owner",
                    f"coverage[{index}]",
                    f"no documentation matches configured owners: {', '.join(rule.docs)}",
                )
            )
    return findings


def changed_paths(repo: Path, base: str) -> list[str]:
    result = subprocess.run(
        ["git", "diff", "--name-only", f"{base}...HEAD"],
        cwd=repo,
        text=True,
        capture_output=True,
        check=False,
    )
    if result.returncode != 0:
        raise ValueError(result.stderr.strip() or f"git diff failed for {base}")
    return sorted(path for path in result.stdout.splitlines() if path)


def audit_change_impact(changed: list[str], rules: tuple[MappingRule, ...]) -> list[Finding]:
    findings: list[Finding] = []
    for index, rule in enumerate(rules, start=1):
        code_changes = [path for path in changed if matches(path, rule.paths)]
        if not code_changes or any(matches(path, rule.docs) for path in changed):
            continue
        findings.append(
            Finding(
                "missing-documentation-impact",
                f"change_rules[{index}]",
                f"changes to {', '.join(code_changes[:5])} require one of: {', '.join(rule.docs)}",
            )
        )
    return findings


def audit(repo: Path, config: Config, changed_from: str | None = None) -> list[Finding]:
    paths = all_repo_paths(repo)
    findings = []
    findings.extend(audit_required_paths(repo, config))
    findings.extend(audit_indexes(repo, config))
    findings.extend(audit_links(repo, config))
    findings.extend(audit_sections(repo, config))
    findings.extend(audit_coverage(repo, config, paths))
    if changed_from:
        findings.extend(audit_change_impact(changed_paths(repo, changed_from), config.change_rules))
    return sorted(set(findings))
