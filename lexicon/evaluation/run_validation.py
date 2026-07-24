from __future__ import annotations

import argparse
import hashlib
import json
import os
import shutil
import subprocess
import sys
import time
from collections import Counter
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path
from typing import Any

SAMPLED_RELATIONS = (
    "calls",
    "possible-calls",
    "reads",
    "writes",
    "depends-on",
    "extends",
    "implements",
    "overrides",
)


def executable(name: str, *fallbacks: Path) -> str:
    candidates = [shutil.which(name), *(str(path) for path in fallbacks)]
    for candidate in candidates:
        if candidate and Path(candidate).is_file():
            return candidate
    raise FileNotFoundError(f"cannot locate executable: {name}")


def npm_command(*args: str) -> list[str]:
    node = Path(executable("node.exe", Path("C:/Program Files/nodejs/node.exe")))
    npm_cli = node.parent / "node_modules" / "npm" / "bin" / "npm-cli.js"
    if not npm_cli.is_file():
        raise FileNotFoundError(f"cannot locate npm CLI: {npm_cli}")
    return [str(node), str(npm_cli), *args]


def run(command: list[str], cwd: Path, env: dict[str, str] | None = None) -> str:
    rendered = " ".join(command)
    print(f"[{cwd.name}] {rendered}", flush=True)
    completed = subprocess.run(
        command,
        cwd=cwd,
        env=env,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
    )
    if completed.returncode != 0:
        raise RuntimeError(f"command failed ({completed.returncode}): {rendered}\n{completed.stdout}")
    return completed.stdout


def build_adapters(root: Path, adapters: set[str]) -> None:
    bin_dir = root / "evaluation" / "bin"
    bin_dir.mkdir(parents=True, exist_ok=True)
    if "c-family" in adapters:
        output = bin_dir / ("lexicon-c-family.exe" if os.name == "nt" else "lexicon-c-family")
        go = executable("go.exe", Path("C:/Program Files/Go/bin/go.exe"))
        run([go, "build", "-o", str(output), "."], root / "adapters" / "c-family")
    if "typescript" in adapters:
        run(npm_command("run", "build"), root / "adapters" / "typescript")
    if "gdscript" in adapters:
        output = bin_dir / ("lexicon-gdscript.exe" if os.name == "nt" else "lexicon-gdscript")
        go = executable("go.exe", Path("C:/Program Files/Go/bin/go.exe"))
        run([go, "build", "-o", str(output), "."], root / "adapters" / "gdscript")
    if "rust" in adapters:
        cargo = executable("cargo.exe", Path.home() / ".cargo" / "bin" / "cargo.exe")
        run(
            [cargo, "build", "--release", "--manifest-path", "adapters/rust/Cargo.toml"],
            root,
        )


def adapter_command(root: Path, adapter: str, repository: Path, output: Path) -> tuple[list[str], dict[str, str]]:
    env = os.environ.copy()
    if adapter == "c-family":
        binary = root / "evaluation" / "bin" / ("lexicon-c-family.exe" if os.name == "nt" else "lexicon-c-family")
        return [str(binary), "--repo", str(repository), "--output", str(output)], env
    if adapter == "python":
        env["PYTHONPATH"] = str(root / "adapters" / "python")
        return [sys.executable, "-m", "lexicon_python", "--repo", str(repository), "--output", str(output)], env
    if adapter == "ruby":
        ruby = executable("ruby.exe", Path("C:/Ruby34-x64/bin/ruby.exe"))
        return [ruby, str(root / "adapters" / "ruby" / "lexicon_ruby.rb"), "--repo", str(repository), "--output", str(output)], env
    if adapter == "typescript":
        node = executable("node.exe", Path("C:/Program Files/nodejs/node.exe"))
        return [node, str(root / "adapters" / "typescript" / "dist" / "cli.js"), "--repo", str(repository), "--output", str(output)], env
    if adapter == "gdscript":
        binary = root / "evaluation" / "bin" / ("lexicon-gdscript.exe" if os.name == "nt" else "lexicon-gdscript")
        return [str(binary), "--repo", str(repository), "--output", str(output)], env
    if adapter == "rust":
        cargo = Path(executable("cargo.exe", Path.home() / ".cargo" / "bin" / "cargo.exe"))
        env["PATH"] = str(cargo.parent) + os.pathsep + env.get("PATH", "")
        binary = root / "adapters" / "rust" / "target" / "release" / ("lexicon-rust-adapter.exe" if os.name == "nt" else "lexicon-rust-adapter")
        return [str(binary), "--repo", str(repository), "--output", str(output)], env
    raise ValueError(f"unsupported adapter: {adapter}")


def describe_node(node: dict[str, Any] | None) -> dict[str, Any] | None:
    if node is None:
        return None
    return {
        "kind": node.get("kind"),
        "name": node.get("name"),
        "owner": node.get("owner"),
        "path": node.get("path"),
        "qualified_name": node.get("qualified_name"),
        "span": node.get("span"),
    }


def summarize(path: Path) -> dict[str, Any]:
    records = [json.loads(line) for line in path.read_text(encoding="utf-8").splitlines()]
    header = records[0]
    nodes = {record["id"]: record for record in records[1:] if record["record"] == "node"}
    node_counts: Counter[str] = Counter()
    edge_counts: Counter[str] = Counter()
    unresolved_counts: Counter[str] = Counter()
    unresolved_reasons: Counter[str] = Counter()
    samples: dict[str, list[dict[str, Any]]] = {relation: [] for relation in SAMPLED_RELATIONS}
    samples["unresolved-calls"] = []

    for record in records[1:]:
        if record["record"] == "node":
            node_counts[record["kind"]] += 1
        elif record["record"] == "edge":
            relation = record["relation"]
            edge_counts[relation] += 1
            if relation in samples and len(samples[relation]) < 12:
                samples[relation].append(
                    {
                        "owner": record.get("owner"),
                        "source": describe_node(nodes.get(record["source"])),
                        "span": record.get("span"),
                        "target": describe_node(nodes.get(record["target"])),
                    }
                )
        else:
            unresolved_counts[record["relation"]] += 1
            unresolved_reasons[record["reason"]] += 1
            if record["relation"] == "calls" and len(samples["unresolved-calls"]) < 20:
                samples["unresolved-calls"].append(
                    {
                        "candidate_name": record.get("candidate_name"),
                        "expression": record.get("expression"),
                        "owner": record.get("owner"),
                        "reason": record.get("reason"),
                        "source": describe_node(nodes.get(record["source"])),
                        "span": record.get("span"),
                    }
                )

    return {
        "adapter_version": header["adapter_version"],
        "edge_relations": dict(sorted(edge_counts.items())),
        "language": header["language"],
        "node_kinds": dict(sorted(node_counts.items())),
        "repository": header["repository"],
        "samples": samples,
        "unresolved_reasons": dict(sorted(unresolved_reasons.items())),
        "unresolved_relations": dict(sorted(unresolved_counts.items())),
    }


def validate_case(root: Path, workspace: Path, output_root: Path, case: dict[str, Any]) -> dict[str, Any]:
    case_dir = output_root / case["id"]
    if case_dir.exists():
        shutil.rmtree(case_dir)
    case_dir.mkdir(parents=True)
    repository = workspace / case["repository"]
    if not repository.exists():
        raise FileNotFoundError(f"missing corpus repository: {repository}")

    durations: list[float] = []
    outputs: list[Path] = []
    for run_number in (1, 2):
        output = case_dir / f"run{run_number}.jsonl"
        command, env = adapter_command(root, case["adapter"], repository, output)
        started = time.perf_counter()
        run(command, root, env)
        durations.append(round(time.perf_counter() - started, 6))
        run([sys.executable, "tools/validate_jsonl.py", str(output)], root)
        outputs.append(output)

    first_bytes = outputs[0].read_bytes()
    deterministic = first_bytes == outputs[1].read_bytes()
    semantic = summarize(outputs[0])
    missing = [
        relation
        for relation in case.get("required_relations", [])
        if semantic["edge_relations"].get(relation, 0) == 0
    ]
    unexpected = [
        relation
        for relation in case.get("expected_zero_relations", [])
        if semantic["edge_relations"].get(relation, 0) != 0
    ]
    sample_path = case_dir / "audit_samples.json"
    sample_path.write_text(json.dumps(semantic["samples"], indent=2, sort_keys=True) + "\n", encoding="utf-8")
    del semantic["samples"]
    result = {
        "adapter": case["adapter"],
        "deterministic": deterministic,
        "durations_seconds": durations,
        "id": case["id"],
        "missing_required_relations": missing,
        "unexpected_nonzero_relations": unexpected,
        "output_bytes": len(first_bytes),
        "output_sha256": hashlib.sha256(first_bytes).hexdigest(),
        "repository": case["repository"],
        "revision": case.get("revision"),
        "split": case["split"],
        **semantic,
    }
    (case_dir / "summary.json").write_text(json.dumps(result, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    return result


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--adapter", action="append", choices=("c-family", "python", "ruby", "typescript", "gdscript", "rust"))
    parser.add_argument("--case", action="append")
    parser.add_argument("--jobs", type=int, default=3)
    args = parser.parse_args()

    root = Path(__file__).resolve().parents[1]
    workspace = root.parent
    manifest = json.loads((Path(__file__).with_name("corpus.json")).read_text(encoding="utf-8"))
    cases = manifest["cases"]
    if args.adapter:
        cases = [case for case in cases if case["adapter"] in set(args.adapter)]
    if args.case:
        cases = [case for case in cases if case["id"] in set(args.case)]
    if not cases:
        parser.error("no corpus cases selected")

    selected_adapters = {case["adapter"] for case in cases}
    build_adapters(root, selected_adapters)
    output_root = root / "evaluation" / "validation" / "generated"
    output_root.mkdir(parents=True, exist_ok=True)

    results: list[dict[str, Any]] = []
    failures: list[dict[str, str]] = []
    with ThreadPoolExecutor(max_workers=max(1, args.jobs)) as executor:
        future_cases = {
            executor.submit(validate_case, root, workspace, output_root, case): case
            for case in cases
        }
        for future in as_completed(future_cases):
            case = future_cases[future]
            try:
                results.append(future.result())
            except Exception as error:
                failures.append({"id": case["id"], "error": str(error)})

    results.sort(key=lambda item: item["id"])
    summary = {"failures": failures, "results": results, "version": manifest["version"]}
    summary_path = output_root / "summary.json"
    summary_path.write_text(json.dumps(summary, indent=2, sort_keys=True) + "\n", encoding="utf-8")

    run1_paths = [str(output_root / result["id"] / "run1.jsonl") for result in results]
    if run1_paths:
        run([sys.executable, "tools/semantic_report.py", *run1_paths, "--output", str(output_root / "semantic_report.json")], root)

    failed_gates = failures or any(
        not result["deterministic"]
        or result["missing_required_relations"]
        or result["unexpected_nonzero_relations"]
        for result in results
    )
    complete_selection = not args.adapter and not args.case and len(cases) == len(manifest["cases"])
    if not failed_gates and complete_selection:
        stable_results = []
        for result in results:
            stable = dict(result)
            stable.pop("durations_seconds", None)
            stable_results.append(stable)
        baseline = {"results": stable_results, "version": manifest["version"]}
        (root / "evaluation" / "validation" / "baseline.json").write_text(
            json.dumps(baseline, indent=2, sort_keys=True) + "\n",
            encoding="utf-8",
        )
    print(json.dumps(summary, indent=2, sort_keys=True))
    return 1 if failed_gates else 0


if __name__ == "__main__":
    raise SystemExit(main())
