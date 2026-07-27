from __future__ import annotations

import argparse, json, pathlib, re, statistics, subprocess, time
from typing import Any

STOP_WORDS = {
    "about", "after", "again", "against", "before", "being", "between",
    "building", "does", "each", "every", "from", "have", "into", "new",
    "optional", "other", "then", "this", "through", "turn", "what", "when",
    "where", "which", "while", "with", "without",
}

def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Score CBM owner retrieval on an Arcana corpus")
    parser.add_argument("--cbm-command", required=True)
    parser.add_argument("--project", required=True)
    parser.add_argument("--repository-root", required=True)
    parser.add_argument("--cases", required=True)
    parser.add_argument("--output-prefix", required=True)
    return parser.parse_args()


def keywords(query: str) -> list[str]:
    result: list[str] = []
    for token in re.findall(r"[A-Za-z][A-Za-z0-9_]+", query):
        lowered = token.lower()
        if len(lowered) < 4 or lowered in STOP_WORDS or lowered in result:
            continue
        result.append(lowered)
    return result[:6]

def invoke(cbm: str, payload: dict[str, Any]) -> tuple[dict[str, Any], float]:
    started = time.perf_counter()
    completed = subprocess.run(
        [cbm, "cli", "--json", "search_graph"],
        input=json.dumps(payload), text=True, capture_output=True, timeout=120,
    )
    elapsed_ms = (time.perf_counter() - started) * 1000
    if completed.returncode != 0:
        raise RuntimeError(completed.stderr.strip() or completed.stdout.strip())
    envelope = json.loads(completed.stdout)
    return envelope.get("structuredContent", {}), elapsed_ms


def command_text(command: list[str]) -> str:
    try:
        completed = subprocess.run(command, text=True, capture_output=True, timeout=30)
    except (OSError, subprocess.SubprocessError):
        return "unknown"
    text = completed.stdout.strip() or completed.stderr.strip()
    return text.splitlines()[0] if completed.returncode == 0 and text else "unknown"


def repository_identity(root: str) -> dict[str, Any]:
    try:
        head_result = subprocess.run(
            ["git", "-C", root, "rev-parse", "HEAD"],
            text=True, capture_output=True, timeout=30,
        )
        status_result = subprocess.run(
            ["git", "-C", root, "status", "--porcelain"],
            text=True, capture_output=True, timeout=30,
        )
    except (OSError, subprocess.SubprocessError):
        return {"head": "unknown", "dirty": None}
    head = head_result.stdout.strip() if head_result.returncode == 0 else "unknown"
    dirty = bool(status_result.stdout.strip()) if status_result.returncode == 0 else None
    return {"head": head or "unknown", "dirty": dirty}


def cbm_version(command: str) -> str:
    for args in ([command, "--version"], [command, "version"]):
        version = command_text(args)
        if version != "unknown":
            return version
    return "unknown"


def normalized_path(value: str) -> str:
    return value.replace("\\", "/").lower().lstrip("./")


def seed_rank(results: list[dict[str, Any]], required: dict[str, Any]) -> int:
    expected_name = required["name"].lower()
    expected_path = normalized_path(required["path"])
    for rank, result in enumerate(results, 1):
        if str(result.get("name", "")).lower() != expected_name:
            continue
        if normalized_path(str(result.get("file_path", ""))) == expected_path:
            return rank
    return 0


def evaluate_mode(
    cbm: str,
    project: str,
    cases: list[dict[str, Any]],
    top_k: int,
    mode: str,
) -> dict[str, Any]:
    measured: list[dict[str, Any]] = []
    reciprocal_ranks: list[float] = []
    found = total = passed = 0
    latencies: list[float] = []
    for case in cases:
        query = case["query"]
        payload: dict[str, Any] = {"project": project, "limit": top_k}
        if mode == "bm25":
            payload["query"] = query
        else:
            payload["semantic_query"] = keywords(query)
        response, latency_ms = invoke(cbm, payload)
        key = "results" if mode == "bm25" else "semantic_results"
        results = list(response.get(key, []))[:top_k]
        ranks = [seed_rank(results, seed) for seed in case["required_seeds"]]
        case_found = sum(rank > 0 for rank in ranks)
        case_pass = case_found == len(ranks)
        best_rank = min((rank for rank in ranks if rank > 0), default=0)
        reciprocal_ranks.append(1 / best_rank if best_rank else 0)
        found += case_found
        total += len(ranks)
        passed += int(case_pass)
        latencies.append(latency_ms)
        measured.append({
            "id": case["id"], "query": query,
            "keywords": keywords(query) if mode == "semantic" else [],
            "pass": case_pass, "required_ranks": ranks,
            "latency_ms": latency_ms,
            "results": results,
        })
    return {
        "mode": mode,
        "pass_rate": passed / len(cases) if cases else 0,
        "required_seed_recall": found / total if total else 0,
        "mrr": statistics.fmean(reciprocal_ranks) if reciprocal_ranks else 0,
        "median_latency_ms": statistics.median(latencies) if latencies else 0,
        "cases": measured,
    }


def markdown(report: dict[str, Any]) -> str:
    lines = [
        f"# CBM owner benchmark: {report['repository']}", "",
        f"CBM: `{report['cbm_version']}`  ",
        f"Repository HEAD: `{report['repository_identity']['head']}`  ",
        f"Dirty working tree: `{report['repository_identity']['dirty']}`", "",
        "The Arcana corpus labels are human judgments used as a common benchmark, not absolute ground truth.", "",
        "| Mode | Pass | Required seed recall | MRR | Median latency |",
        "| --- | ---: | ---: | ---: | ---: |",
    ]
    for mode in report["modes"]:
        lines.append(
            f"| {mode['mode']} | {mode['pass_rate']:.1%} | "
            f"{mode['required_seed_recall']:.1%} | {mode['mrr']:.3f} | "
            f"{mode['median_latency_ms']:.1f} ms |"
        )
    for mode in report["modes"]:
        lines.extend(["", f"## {mode['mode']}", ""])
        for case in mode["cases"]:
            names = ", ".join(item.get("name", "") for item in case["results"])
            lines.append(
                f"- `{case['id']}`: ranks={case['required_ranks']}; "
                f"results={names or '(none)'}"
            )
    return "\n".join(lines) + "\n"


def main() -> int:
    args = parse_args()
    corpus_path = pathlib.Path(args.cases)
    corpus = json.loads(corpus_path.read_text(encoding="utf-8"))
    report = {
        "version": 1,
        "provider": "codebase-memory-mcp",
        "repository": corpus["repository"],
        "revision": corpus["revision"],
        "project": args.project,
        "repository_identity": repository_identity(args.repository_root),
        "cbm_version": cbm_version(args.cbm_command),
        "top_k": corpus["top_k"],
        "modes": [
            evaluate_mode(args.cbm_command, args.project, corpus["cases"], corpus["top_k"], mode)
            for mode in ("bm25", "semantic")
        ],
    }
    output = pathlib.Path("evaluation/results") / args.output_prefix
    output.parent.mkdir(parents=True, exist_ok=True)
    output.with_suffix(".json").write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")
    output.with_suffix(".md").write_text(markdown(report), encoding="utf-8")
    for mode in report["modes"]:
        print(
            mode["mode"], f"pass={mode['pass_rate']:.1%}",
            f"recall={mode['required_seed_recall']:.1%}", f"mrr={mode['mrr']:.3f}",
            f"median_ms={mode['median_latency_ms']:.1f}",
        )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
