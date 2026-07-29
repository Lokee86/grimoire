from __future__ import annotations

from dataclasses import asdict, dataclass, field
import json
from pathlib import Path, PurePosixPath
import re
from typing import Any, Iterable

EVIDENCE_MARKER = "BENCHMARK_EVIDENCE_JSON:"
CITATION = re.compile(
    r"`?((?:(?:[A-Za-z0-9_.@+\-]+[/\\])+[A-Za-z0-9_.@+\-]+)|(?:[A-Za-z0-9_.@+\-]+\.[A-Za-z0-9_.@+\-]+)):(\d+)(?:-(\d+))?`?"
)
REFUSAL_PATTERNS = (
    "i cannot complete",
    "i can't complete",
    "unable to complete",
    "cannot provide the requested",
    "unable to provide the requested",
)


@dataclass(frozen=True)
class CanonicalRange:
    path: str
    start_line: int
    end_line: int


@dataclass
class Finding:
    code: str
    message: str
    path: str = ""
    start_line: int = 0
    end_line: int = 0


@dataclass
class GroundingReport:
    valid: bool = False
    process_completed: bool = False
    refusal_detected: bool = False
    evidence_json_found: bool = False
    evidence_json_valid: bool = False
    inline_citations: int = 0
    structured_items: int = 0
    handle_items: int = 0
    canonical_handle_items: int = 0
    required_prefixes_found: list[str] = field(default_factory=list)
    missing_required_prefixes: list[str] = field(default_factory=list)
    findings: list[Finding] = field(default_factory=list)

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


def validate_answer(
    repository: Path,
    answer: str,
    *,
    exit_code: int,
    expected_sections: Iterable[str],
    audit_log: Path | None = None,
    require_grimoire_handles: bool = False,
    required_path_prefixes: Iterable[str] = (),
) -> GroundingReport:
    root = repository.resolve()
    report = GroundingReport(process_completed=exit_code == 0)
    lowered = answer.casefold()
    report.refusal_detected = any(pattern in lowered for pattern in REFUSAL_PATTERNS)
    if exit_code != 0:
        report.findings.append(Finding("process_failed", f"agent process exited with code {exit_code}"))
    if report.refusal_detected:
        report.findings.append(Finding("refusal", "answer contains a refusal marker"))

    try:
        handle_ranges = load_handle_ranges(audit_log) if audit_log else {}
    except (OSError, ValueError, json.JSONDecodeError) as error:
        report.findings.append(Finding("invalid_audit_log", str(error)))
        handle_ranges = {}
    for match in CITATION.finditer(answer):
        report.inline_citations += 1
        path, start, end = match.group(1), int(match.group(2)), int(match.group(3) or match.group(2))
        validate_range(root, path, start, end, report, "inline_citation")

    evidence = extract_evidence(answer, report)
    if evidence is not None:
        expected = list(expected_sections)
        for section in expected:
            items = evidence.get(section)
            if not isinstance(items, list) or not items:
                report.findings.append(Finding("missing_evidence_section", f"section {section!r} is missing or empty"))
                continue
            for index, item in enumerate(items):
                validate_item(
                    root,
                    section,
                    index,
                    item,
                    report,
                    handle_ranges,
                    require_grimoire_handles,
                )
        unknown = sorted(set(evidence) - set(expected))
        if unknown:
            report.findings.append(Finding("unknown_evidence_sections", f"unexpected sections: {', '.join(unknown)}"))
        evidence_paths = {
            canonical_evidence_path(item, handle_ranges)
            for section in expected
            for item in evidence.get(section, [])
            if isinstance(item, dict) and canonical_evidence_path(item, handle_ranges)
        }
        for raw_prefix in required_path_prefixes:
            raw = str(raw_prefix)
            directory_prefix = raw.replace("\\", "/").endswith("/")
            prefix = normalize_path(raw).rstrip("/")
            if directory_prefix:
                covered = any(path == prefix or path.startswith(prefix + "/") for path in evidence_paths)
            else:
                covered = any(path.startswith(prefix) for path in evidence_paths)
            if covered:
                report.required_prefixes_found.append(prefix)
            else:
                report.missing_required_prefixes.append(prefix)

    report.valid = (
        report.process_completed
        and not report.refusal_detected
        and report.evidence_json_valid
        and not report.findings
    )
    return report


def extract_evidence(answer: str, report: GroundingReport) -> dict[str, Any] | None:
    lines = [line for line in answer.splitlines() if line.startswith(EVIDENCE_MARKER)]
    if not lines:
        report.findings.append(Finding("missing_evidence_json", f"missing {EVIDENCE_MARKER} line"))
        return None
    report.evidence_json_found = True
    payload = lines[-1][len(EVIDENCE_MARKER):].strip()
    try:
        evidence = json.loads(payload)
    except json.JSONDecodeError as error:
        report.findings.append(Finding("invalid_evidence_json", str(error)))
        return None
    if not isinstance(evidence, dict):
        report.findings.append(Finding("invalid_evidence_shape", "evidence JSON must be an object"))
        return None
    report.evidence_json_valid = True
    return evidence


def validate_item(
    root: Path,
    section: str,
    index: int,
    item: Any,
    report: GroundingReport,
    handle_ranges: dict[str, CanonicalRange],
    require_handle: bool,
) -> None:
    label = f"{section}[{index}]"
    if not isinstance(item, dict):
        report.findings.append(Finding("invalid_evidence_item", f"{label} must be an object"))
        return
    report.structured_items += 1
    if not item.get("claim"):
        report.findings.append(Finding("missing_evidence_fields", f"{label} missing: claim"))
        return

    handle = item.get("handle")
    if handle:
        report.handle_items += 1
        canonical = handle_ranges.get(str(handle))
        if canonical is None:
            report.findings.append(Finding("unknown_handle", f"{label} references an unrecorded handle"))
            return
        validate_range(
            root,
            canonical.path,
            canonical.start_line,
            canonical.end_line,
            report,
            "structured_evidence",
        )
        report.canonical_handle_items += 1
        return

    missing = [name for name in ("path", "symbol", "lines") if not item.get(name)]
    if missing:
        report.findings.append(Finding("missing_evidence_fields", f"{label} missing: {', '.join(missing)}"))
        return
    try:
        ranges = parse_line_ranges(item["lines"])
    except ValueError as error:
        report.findings.append(Finding("invalid_line_range", f"{label}: {error}", str(item.get("path", ""))))
        return
    path = str(item["path"])
    for start, end in ranges:
        validate_range(root, path, start, end, report, "structured_evidence")
    if require_handle:
        start, end = ranges[0]
        report.findings.append(Finding("missing_grimoire_handle", f"{label} must include an inspected source-range handle", path, start, end))


def canonical_evidence_path(item: dict[str, Any], handle_ranges: dict[str, CanonicalRange]) -> str:
    handle = item.get("handle")
    if handle and str(handle) in handle_ranges:
        return handle_ranges[str(handle)].path
    path = item.get("path")
    return normalize_path(str(path)) if path else ""


def parse_line_ranges(value: Any) -> list[tuple[int, int]]:
    if isinstance(value, int):
        return [(value, value)]
    if isinstance(value, str):
        parts = [part.strip() for part in value.split(",")]
        if parts and all(parts):
            ranges: list[tuple[int, int]] = []
            for part in parts:
                match = re.fullmatch(r"(\d+)(?:\s*-\s*(\d+))?", part)
                if not match:
                    break
                ranges.append((int(match.group(1)), int(match.group(2) or match.group(1))))
            else:
                return ranges
    if isinstance(value, list):
        if len(value) == 2 and all(isinstance(item, int) for item in value):
            return [(value[0], value[1])]
        ranges = []
        for item in value:
            ranges.extend(parse_line_ranges(item))
        if ranges:
            return ranges
    if isinstance(value, dict):
        start, end = value.get("start"), value.get("end")
        if isinstance(start, int) and isinstance(end, int):
            return [(start, end)]
    raise ValueError(f"unsupported lines value {value!r}")


def validate_range(root: Path, path: str, start: int, end: int, report: GroundingReport, code: str) -> None:
    normalized = normalize_path(path)
    pure = PurePosixPath(normalized)
    if not normalized or pure.is_absolute() or ".." in pure.parts:
        report.findings.append(Finding(f"{code}_invalid_path", f"path is not repository-relative: {path!r}", path, start, end))
        return
    target = (root / Path(*pure.parts)).resolve()
    try:
        target.relative_to(root)
    except ValueError:
        report.findings.append(Finding(f"{code}_path_escape", f"path escapes repository: {path!r}", path, start, end))
        return
    if not target.is_file():
        report.findings.append(Finding(f"{code}_missing_path", f"file does not exist: {normalized}", normalized, start, end))
        return
    line_count = len(target.read_text(encoding="utf-8", errors="replace").splitlines())
    if start < 1 or end < start or end > line_count:
        report.findings.append(Finding(
            f"{code}_out_of_range",
            f"{normalized}:{start}-{end} is outside 1-{line_count}",
            normalized,
            start,
            end,
        ))


def summarize_audit(path: Path | None) -> dict[str, Any]:
    summary: dict[str, Any] = {
        "calls": 0,
        "response_bytes": 0,
        "modes": {},
        "new_nodes": 0,
        "new_source_ranges": 0,
        "new_documents": 0,
        "new_relationships": 0,
        "new_graph_paths": 0,
        "tool_errors": 0,
        "invalid_records": 0,
    }
    if path is None or not path.is_file():
        return summary
    for line in path.read_text(encoding="utf-8").splitlines():
        if not line.strip():
            continue
        try:
            record = json.loads(line)
        except json.JSONDecodeError:
            summary["invalid_records"] += 1
            continue
        response = record.get("response") or {}
        if record.get("error"):
            summary["tool_errors"] += 1
        mode = str((record.get("request") or {}).get("mode") or "unknown")
        summary["calls"] += 1
        summary["response_bytes"] += len(json.dumps(response, separators=(",", ":")).encode("utf-8"))
        summary["modes"][mode] = summary["modes"].get(mode, 0) + 1
        delta = response.get("delta") or {}
        for name in (
            "new_nodes",
            "new_source_ranges",
            "new_documents",
            "new_relationships",
            "new_graph_paths",
        ):
            summary[name] += len(delta.get(name) or [])
    return summary


def load_handle_ranges(path: Path | None) -> dict[str, CanonicalRange]:
    if path is None or not path.is_file():
        return {}
    ranges: dict[str, CanonicalRange] = {}
    for line in path.read_text(encoding="utf-8").splitlines():
        if not line.strip():
            continue
        record = json.loads(line)
        response = record.get("response") or {}
        delta = response.get("delta") or {}
        for item in delta.get("new_source_ranges") or []:
            evidence = item.get("evidence") or {}
            add_handle_range(ranges, item.get("handle"), evidence)
        for inspection in response.get("inspections") or []:
            containing = inspection.get("containing_span") or {}
            handle = (containing.get("handle") or {}).get("value")
            add_handle_range(ranges, handle, containing)
    return ranges


def add_handle_range(ranges: dict[str, CanonicalRange], handle: Any, evidence: dict[str, Any]) -> None:
    if handle and evidence.get("path") and evidence.get("start_line"):
        ranges[str(handle)] = CanonicalRange(
            normalize_path(str(evidence["path"])),
            int(evidence["start_line"]),
            int(evidence.get("end_line") or evidence["start_line"]),
        )


def normalize_path(path: str) -> str:
    return str(PurePosixPath(path.strip().replace("\\", "/"))).removeprefix("./")
