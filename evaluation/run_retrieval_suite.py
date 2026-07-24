#!/usr/bin/env python3
"""Run pinned retrieval corpora and macro-average repositories."""

from __future__ import annotations

import argparse
import json
import statistics
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path

METRICS = (
    "pass_rate",
    "required_evidence_recall",
    "required_recall_at_10",
    "required_recall_at_20",
    "mean_reciprocal_rank",
    "irrelevant_selection_rate",
    "median_latency_ms",
)


def run(command: list[str], cwd: Path, dry_run: bool = False) -> None:
    print("+", subprocess.list2cmdline(command))
    if not dry_run:
        subprocess.run(command, cwd=cwd, check=True)


def output(command: list[str], cwd: Path) -> str:
    return subprocess.check_output(command, cwd=cwd, text=True).strip()


def resolve_checkout(entry: dict, workspace: Path, grimoire: Path) -> Path:
    checkout = entry["checkout"]
    return grimoire if checkout == "$GRIMOIRE" else workspace / checkout


def verify_suite(manifest: dict, entries: list[dict], workspace: Path, grimoire: Path) -> None:
    baseline = manifest["baseline_revision"]
    paths = manifest.get("baseline_paths", ["cmd", "internal", "native"])
    allowed_drift = set(manifest.get("allowed_baseline_drift", []))
    changed = output(["git", "diff", "--name-only", baseline, "--", *paths], grimoire).splitlines()
    unexpected = [path for path in changed if path and path not in allowed_drift]
    if unexpected:
        raise RuntimeError(
            "retrieval implementation differs from frozen baseline: " + ", ".join(unexpected)
        )
    for entry in entries:
        checkout = resolve_checkout(entry, workspace, grimoire)
        if not checkout.exists():
            raise FileNotFoundError(f"missing checkout for {entry['id']}: {checkout}")
        expected = entry.get("revision")
        if expected and expected != "$BASELINE":
            actual = output(["git", "rev-parse", "HEAD"], checkout)
            if actual != expected:
                raise RuntimeError(
                    f"revision mismatch for {entry['id']}: expected {expected}, got {actual}"
                )
        cases = grimoire / entry["cases"]
        if not cases.is_file():
            raise FileNotFoundError(f"missing cases for {entry['id']}: {cases}")


def aggregate(reports: list[dict], mode: str) -> dict:
    repositories = []
    for report in reports:
        row = next(item for item in report["by_mode"] if item["group"] == mode)
        repositories.append({"repository": report["repository"], **row})
    macro = {metric: statistics.fmean(row[metric] for row in repositories) for metric in METRICS}
    failure_stages: dict[str, int] = {}
    failure_total = 0
    for row in repositories:
        failure_total += row.get("required_failures", 0)
        for stage, count in row.get("required_failure_stages", {}).items():
            failure_stages[stage] = failure_stages.get(stage, 0) + count
    failure_stage_rates = {
        stage: count / failure_total if failure_total else 0.0
        for stage, count in failure_stages.items()
    }
    micro = {}
    for metric in METRICS:
        weight_key = "ranking_cases" if metric in {
            "required_recall_at_10", "required_recall_at_20", "mean_reciprocal_rank"
        } else "cases"
        total_weight = sum(row[weight_key] for row in repositories)
        micro[metric] = (
            sum(row[metric] * row[weight_key] for row in repositories) / total_weight
            if total_weight else 0.0
        )
    return {
        "repositories": repositories,
        "macro": macro,
        "micro": micro,
        "required_failures": failure_total,
        "required_failure_stages": failure_stages,
        "required_failure_stage_rates": failure_stage_rates,
    }


def write_summary(path: Path, result: dict) -> None:
    lines = [
        "# Retrieval suite summary",
        "",
        f"- Split: `{result['split']}`",
        f"- Variant: `{result['variant']}`",
        f"- Mode: `{result['mode']}`",
        f"- Baseline implementation: `{result['baseline_revision']}`",
        "",
        "| Repository | Cases | Pass | Recall | R@10 | R@20 | MRR | Irrelevant |",
        "| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |",
    ]
    for row in result["aggregate"]["repositories"]:
        lines.append(
            f"| {row['repository']} | {row['cases']} | {row['pass_rate']:.1%} | "
            f"{row['required_evidence_recall']:.1%} | {row['required_recall_at_10']:.1%} | "
            f"{row['required_recall_at_20']:.1%} | {row['mean_reciprocal_rank']:.3f} | "
            f"{row['irrelevant_selection_rate']:.1%} |"
        )
    lines.extend(["", "## Macro average", ""])
    for metric, value in result["aggregate"]["macro"].items():
        lines.append(f"- `{metric}`: {value:.4f}")
    lines.extend(["", "## Required evidence failure stages", ""])
    failure_stages = result["aggregate"].get("required_failure_stages", {})
    failure_rates = result["aggregate"].get("required_failure_stage_rates", {})
    for stage in sorted(failure_stages):
        lines.append(f"- `{stage}`: {failure_stages[stage]} ({failure_rates[stage]:.1%})")
    path.write_text("\n".join(lines) + "\n", encoding="utf-8")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--suite", default="evaluation/retrieval/suite.json")
    parser.add_argument("--workspace-root", required=True)
    parser.add_argument("--grimoire-root", default=".")
    parser.add_argument("--split", choices=("calibration", "validation", "test"), required=True)
    parser.add_argument("--allow-test", action="store_true")
    parser.add_argument("--mode", default="lexical")
    parser.add_argument("--variant", default="frozen-baseline")
    parser.add_argument("--selection-file-penalty", type=int)
    parser.add_argument("--selection-subsystem-penalty", type=int)
    parser.add_argument("--selection-adjacent-primaries", type=int)
    parser.add_argument("--assembly-strategy", choices=("legacy", "coverage"))
    parser.add_argument("--assembly-facet-depth", type=int)
    parser.add_argument("--lexical-declaration-alias-bonus", type=float)
    parser.add_argument("--skip-index", action="store_true")
    parser.add_argument("--dry-run", action="store_true")
    args = parser.parse_args()

    grimoire = Path(args.grimoire_root).resolve()
    workspace = Path(args.workspace_root).resolve()
    manifest = json.loads((grimoire / args.suite).read_text(encoding="utf-8"))
    assembly_defaults = manifest.get("assembly", {"strategy": "coverage", "facet_depth": 3})
    ranking_defaults = manifest.get("ranking", {"lexical_declaration_alias_bonus": 1.0})
    lexical_declaration_alias_bonus = (
        args.lexical_declaration_alias_bonus
        if args.lexical_declaration_alias_bonus is not None
        else ranking_defaults["lexical_declaration_alias_bonus"]
    )
    assembly_strategy = args.assembly_strategy or assembly_defaults["strategy"]
    assembly_facet_depth = (
        args.assembly_facet_depth
        if args.assembly_facet_depth is not None
        else assembly_defaults["facet_depth"]
    )
    if args.split == "test" and not args.allow_test:
        parser.error("test split is sealed; pass --allow-test only for a final frozen run")
    entries = [item for item in manifest["repositories"] if item["split"] == args.split]
    if not entries:
        parser.error(f"suite has no repositories in split {args.split}")
    verify_suite(manifest, entries, workspace, grimoire)
    stamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%SZ")
    output_dir = grimoire / "evaluation" / "validation" / f"{stamp}-{args.split}-{args.variant}"
    output_dir.mkdir(parents=True, exist_ok=True)
    binary = grimoire / ".tmp" / "retrieval-suite" / "grimoire.exe"
    binary.parent.mkdir(parents=True, exist_ok=True)
    run(["go", "build", "-o", str(binary), "./cmd/grimoire"], grimoire, args.dry_run)

    reports = []
    selection = dict(manifest["selection"])
    overrides = {
        "file_repeat_penalty": args.selection_file_penalty,
        "subsystem_repeat_penalty": args.selection_subsystem_penalty,
        "adjacent_primary_limit": args.selection_adjacent_primaries,
    }
    for key, value in overrides.items():
        if value is not None:
            if value < 0:
                parser.error("selection overrides must be non-negative")
            selection[key] = value
    if assembly_facet_depth < 0:
        parser.error("assembly facet depth must be non-negative")
    if lexical_declaration_alias_bonus < 0:
        parser.error("lexical declaration alias bonus must be non-negative")
    for entry in entries:
        checkout = resolve_checkout(entry, workspace, grimoire)
        root = (checkout / entry.get("scope", ".")).resolve()
        prefix = f"{entry['id']}-{args.variant}"
        if not args.skip_index:
            run([str(binary), "index", "--root", str(root)], grimoire, args.dry_run)
        command = [
            str(binary), "eval", "retrieval", "--root", str(root),
            "--cases", str((grimoire / entry["cases"]).resolve()),
            "--modes", args.mode, "--adaptive", "--variant", args.variant,
            "--selection-file-penalty", str(selection["file_repeat_penalty"]),
            "--selection-subsystem-penalty", str(selection["subsystem_repeat_penalty"]),
            "--selection-adjacent-primaries", str(selection["adjacent_primary_limit"]),
            "--assembly-strategy", assembly_strategy,
            "--assembly-facet-depth", str(assembly_facet_depth),
            "--lexical-declaration-alias-bonus", str(lexical_declaration_alias_bonus),
            "--output-dir", str(output_dir), "--output-prefix", prefix,
        ]
        run(command, grimoire, args.dry_run)
        if not args.dry_run:
            reports.append(json.loads((output_dir / f"{prefix}.json").read_text(encoding="utf-8")))

    if args.dry_run:
        return 0
    result = {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "split": args.split,
        "variant": args.variant,
        "mode": args.mode,
        "baseline_revision": manifest["baseline_revision"],
        "selection": selection,
        "ranking": {
            "lexical_declaration_alias_bonus": lexical_declaration_alias_bonus,
        },
        "assembly": {
            "strategy": assembly_strategy,
            "facet_depth": assembly_facet_depth,
        },
        "aggregate": aggregate(reports, args.mode),
    }
    (output_dir / "suite-summary.json").write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
    write_summary(output_dir / "suite-summary.md", result)
    print(output_dir)
    return 0


if __name__ == "__main__":
    sys.exit(main())
