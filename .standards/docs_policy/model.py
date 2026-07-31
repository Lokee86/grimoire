from __future__ import annotations

import json
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any


@dataclass(frozen=True)
class MappingRule:
    paths: tuple[str, ...]
    docs: tuple[str, ...]


@dataclass(frozen=True)
class Config:
    version: int
    profile: str
    docs_root: str
    root_index: str
    folder_index: str
    required_paths: tuple[str, ...] = ()
    index_exempt_directories: tuple[str, ...] = ("stubs",)
    section_exemptions: tuple[str, ...] = ()
    required_sections: tuple[str, ...] = ("Purpose", "Overview", "Related docs", "Notes")
    baseline: str | None = None
    coverage: tuple[MappingRule, ...] = field(default_factory=tuple)
    change_rules: tuple[MappingRule, ...] = field(default_factory=tuple)


def _mapping_rules(raw: Any, field_name: str) -> tuple[MappingRule, ...]:
    if raw is None:
        return ()
    if not isinstance(raw, list):
        raise ValueError(f"{field_name} must be an array")
    rules: list[MappingRule] = []
    for index, item in enumerate(raw, start=1):
        if not isinstance(item, dict):
            raise ValueError(f"{field_name}[{index}] must be an object")
        paths = item.get("paths", [])
        docs = item.get("docs", [])
        if not isinstance(paths, list) or not all(isinstance(value, str) and value for value in paths):
            raise ValueError(f"{field_name}[{index}].paths must contain non-empty strings")
        if not isinstance(docs, list) or not all(isinstance(value, str) and value for value in docs):
            raise ValueError(f"{field_name}[{index}].docs must contain non-empty strings")
        if not paths or not docs:
            raise ValueError(f"{field_name}[{index}] requires paths and docs")
        rules.append(MappingRule(tuple(paths), tuple(docs)))
    return tuple(rules)


def _strings(raw: Any, field_name: str, default: tuple[str, ...] = ()) -> tuple[str, ...]:
    if raw is None:
        return default
    if not isinstance(raw, list) or not all(isinstance(value, str) and value for value in raw):
        raise ValueError(f"{field_name} must contain non-empty strings")
    return tuple(raw)


def load_config(repo: Path, config_path: str) -> Config:
    path = repo / config_path
    try:
        raw = json.loads(path.read_text(encoding="utf-8"))
    except FileNotFoundError as error:
        raise ValueError(f"missing configuration: {config_path}") from error
    except json.JSONDecodeError as error:
        raise ValueError(f"invalid JSON in {config_path}: {error}") from error
    if not isinstance(raw, dict):
        raise ValueError("configuration root must be an object")

    version = raw.get("version")
    if version != 1:
        raise ValueError(f"unsupported configuration version: {version!r}")

    required = ("profile", "docs_root", "root_index", "folder_index")
    for key in required:
        if not isinstance(raw.get(key), str) or not raw[key]:
            raise ValueError(f"{key} must be a non-empty string")

    baseline = raw.get("baseline")
    if baseline is not None and (not isinstance(baseline, str) or not baseline):
        raise ValueError("baseline must be a non-empty string when set")

    return Config(
        version=version,
        profile=raw["profile"],
        docs_root=raw["docs_root"],
        root_index=raw["root_index"],
        folder_index=raw["folder_index"],
        required_paths=_strings(raw.get("required_paths"), "required_paths"),
        index_exempt_directories=_strings(
            raw.get("index_exempt_directories"),
            "index_exempt_directories",
            ("stubs",),
        ),
        section_exemptions=_strings(raw.get("section_exemptions"), "section_exemptions"),
        required_sections=_strings(
            raw.get("required_sections"),
            "required_sections",
            ("Purpose", "Overview", "Related docs", "Notes"),
        ),
        baseline=baseline,
        coverage=_mapping_rules(raw.get("coverage"), "coverage"),
        change_rules=_mapping_rules(raw.get("change_rules"), "change_rules"),
    )
