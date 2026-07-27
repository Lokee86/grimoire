from __future__ import annotations

import argparse
import json
import re
import sqlite3
import statistics
from pathlib import Path
from typing import Any

SESSION_RE = re.compile(
    r"^srb-(?P<system>[gc])-(?P<task>auth|packet|respawn|config)-"
    r"(?P<trial>[123])(?P<replacement>r?)$"
)
CALL_RE = re.compile(
    r"DISCOVERY_CALLS_USED:\s*(\d+)\s*\nSYSTEM_CALLS_USED:\s*(\d+)\s*\n"
    r"SOURCE_READS_USED:\s*(\d+)",
    re.IGNORECASE,
)
CITATION_RE = re.compile(r"(?:[A-Za-z0-9_.@-]+/)+[A-Za-z0-9_.@-]+:\d+(?:-\d+)?")
EVIDENCE_PREFIX = "BENCHMARK_EVIDENCE_JSON:"
TASK_IDS = {
    "auth": "space-rocks-client-rails-auth",
    "packet": "space-rocks-gameplay-packet-producer-handler",
    "respawn": "space-rocks-match-respawn-lifecycle",
    "config": "space-rocks-config-key-readers",
}
SYSTEM_NAMES = {"g": "Grimoire", "c": "CBM"}


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--db", default=r"C:\Users\archa\AppData\Local\hermes\state.db")
    parser.add_argument("--cases", default="evaluation/agent_discovery/space-rocks.v1.json")
    parser.add_argument("--output-json")
    return parser.parse_args()


def normalize_path(value: str) -> str:
    normalized = value.replace("\\", "/").strip()
    return re.sub(r":\d+(?:-\d+)?$", "", normalized)


def extract_evidence(answer: str) -> tuple[dict[str, Any] | None, str | None]:
    for line in answer.splitlines():
        if not line.startswith(EVIDENCE_PREFIX):
            continue
        try:
            value = json.loads(line[len(EVIDENCE_PREFIX) :].strip())
        except json.JSONDecodeError as error:
            return None, str(error)
        return (value, None) if isinstance(value, dict) else (None, "evidence is not an object")
    return None, "missing evidence line"


def item_recovered(item: dict[str, Any], haystack: str) -> bool:
    path = normalize_path(str(item.get("path", ""))).casefold()
    if not path or path not in haystack:
        return False
    symbols = [str(value).casefold() for value in item.get("symbols", []) if str(value).strip()]
    return not symbols or any(symbol in haystack for symbol in symbols)


def score_answer(answer: str, case: dict[str, Any]) -> dict[str, Any]:
    evidence, evidence_error = extract_evidence(answer)
    evidence_text = json.dumps(evidence, ensure_ascii=False) if evidence else ""
    haystack = (answer + "\n" + evidence_text).replace("\\", "/").casefold()

    ownership = [item_recovered(item, haystack) for item in case.get("ownership_evidence", [])]
    required = [item_recovered(item, haystack) for item in case.get("required_evidence", [])]
    structural = [
        item_recovered(item, haystack) for item in case.get("required_structural_evidence", [])
    ]
    forbidden_hits = [
        str(item.get("id", item.get("pattern", "")))
        for item in case.get("forbidden_unsupported_conclusions", [])
        if str(item.get("pattern", "")).casefold() in haystack
    ]
    outcome_missing = [
        str(check["label"])
        for check in case.get("outcome_checks", [])
        if not any(str(value).casefold() in haystack for value in check.get("any", []))
    ]
    citations = CITATION_RE.findall(answer)
    calls = CALL_RE.search(answer)
    return {
        "ownership_recovered": sum(ownership),
        "ownership_total": len(ownership),
        "required_recovered": sum(required),
        "required_total": len(required),
        "structural_recovered": sum(structural),
        "structural_total": len(structural),
        "forbidden_hits": forbidden_hits,
        "outcome_missing": outcome_missing,
        "evidence_valid": evidence is not None and evidence_error is None,
        "evidence_error": evidence_error,
        "citation_count": len(citations),
        "reported_discovery_calls": int(calls.group(1)) if calls else None,
        "reported_system_calls": int(calls.group(2)) if calls else None,
        "reported_source_reads": int(calls.group(3)) if calls else None,
        "correct": not outcome_missing and not forbidden_hits and evidence is not None and bool(citations),
    }


def load_runs(db_path: str, cases: dict[str, dict[str, Any]]) -> list[dict[str, Any]]:
    connection = sqlite3.connect(f"file:{db_path}?mode=ro", uri=True)
    connection.row_factory = sqlite3.Row
    rows = connection.execute(
        "SELECT * FROM sessions WHERE title LIKE 'srb-%' ORDER BY started_at"
    ).fetchall()
    titles = {str(row["title"] or "") for row in rows}
    runs: list[dict[str, Any]] = []
    for row in rows:
        title = str(row["title"] or "")
        match = SESSION_RE.match(title)
        if not match or (not match.group("replacement") and f"{title}r" in titles):
            continue
        messages = connection.execute(
            "SELECT role, content, timestamp FROM messages WHERE session_id=? ORDER BY id",
            (row["id"],),
        ).fetchall()
        answers = [
            message
            for message in messages
            if message["role"] == "assistant" and EVIDENCE_PREFIX in (message["content"] or "")
        ]
        answer = str(answers[-1]["content"] or "") if answers else ""
        started = float(row["started_at"] or 0)
        last = max((float(message["timestamp"] or 0) for message in messages), default=started)
        task = match.group("task")
        runs.append(
            {
                "title": title,
                "session_id": row["id"],
                "system": SYSTEM_NAMES[match.group("system")],
                "task": task,
                "trial": int(match.group("trial")),
                "complete": bool(answer),
                "model": row["model"],
                "elapsed_seconds": round(max(0.0, last - started), 3),
                "tool_calls": int(row["tool_call_count"] or 0),
                "api_calls": int(row["api_call_count"] or 0),
                "input_tokens": int(row["input_tokens"] or 0),
                "output_tokens": int(row["output_tokens"] or 0),
                "cache_read_tokens": int(row["cache_read_tokens"] or 0),
                "noncached_tokens": int(row["input_tokens"] or 0) + int(row["output_tokens"] or 0),
                "score": score_answer(answer, cases[TASK_IDS[task]]) if answer else None,
            }
        )
    connection.close()
    return runs


def median(values: list[float]) -> float | None:
    return round(statistics.median(values), 3) if values else None


def aggregate(runs: list[dict[str, Any]]) -> list[dict[str, Any]]:
    output: list[dict[str, Any]] = []
    for system in ("Grimoire", "CBM"):
        for task in ("auth", "packet", "respawn", "config", "all"):
            selected = [
                run for run in runs if run["system"] == system and (task == "all" or run["task"] == task)
            ]
            completed = [run for run in selected if run["complete"]]
            scored = [run for run in completed if run["score"]]
            output.append(
                {
                    "system": system,
                    "task": task,
                    "runs": len(selected),
                    "completed": len(completed),
                    "correct": sum(1 for run in scored if run["score"]["correct"]),
                    "required_recovered": sum(run["score"]["required_recovered"] for run in scored),
                    "required_total": sum(run["score"]["required_total"] for run in scored),
                    "structural_recovered": sum(run["score"]["structural_recovered"] for run in scored),
                    "structural_total": sum(run["score"]["structural_total"] for run in scored),
                    "median_elapsed_seconds": median([run["elapsed_seconds"] for run in completed]),
                    "median_tool_calls": median([float(run["tool_calls"]) for run in completed]),
                    "median_noncached_tokens": median([float(run["noncached_tokens"]) for run in completed]),
                    "median_cache_read_tokens": median([float(run["cache_read_tokens"]) for run in completed]),
                }
            )
    return output


def main() -> None:
    args = parse_args()
    corpus = json.loads(Path(args.cases).read_text(encoding="utf-8"))
    cases = {case["id"]: case for case in corpus["cases"]}
    runs = load_runs(args.db, cases)
    data = {"runs": runs, "aggregates": aggregate(runs)}
    if args.output_json:
        Path(args.output_json).write_text(json.dumps(data, indent=2) + "\n", encoding="utf-8")
    print(json.dumps(data["aggregates"], indent=2))
    for run in runs:
        if run["complete"] and run["score"] and not run["score"]["correct"]:
            print(f"incorrect={run['title']} missing={run['score']['outcome_missing']}")
    print(f"runs={len(runs)} complete={sum(1 for run in runs if run['complete'])}")


if __name__ == "__main__":
    main()
