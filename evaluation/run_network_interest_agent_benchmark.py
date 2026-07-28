from __future__ import annotations

import json
import os
import shutil
import subprocess
import time
from pathlib import Path

ROOT = Path(r"C:\!bin\workspace")
OUT = ROOT / "grimoire" / "evaluation" / "results" / "network-interest-agent-benchmark-2026-07-27-v4"
OUT.mkdir(parents=True, exist_ok=True)
HERMES = shutil.which("hermes") or "hermes"
CBM_BIN = ROOT / "cbm-bin"
GRIMOIRE_BIN = ROOT / "space-rocks-quarantine" / "lexicon-arcana-investigation-2026-07-26"
CBM_CACHE = ROOT / "grimoire" / "evaluation" / "results" / "network-interest-agent-benchmark-2026-07-27-v2" / "cbm-cache"

PROMPT = r'''Read-only repository investigation. Do not modify files and do not use the active network-interest worktree.

Task: Design a concrete implementation plan for per-client network interest in Space Rocks at the checked-out revision.

The server should send each client only the hot entity state relevant to that client's current camera/view target while preserving authoritative simulation. Reuse the existing camera visibility/region math. Prevent boundary thrashing. Preserve self and spectated-target state. Entities leaving interest must disappear correctly through the existing baseline/delta/lifecycle system. Distant players must still support offscreen indicators without receiving full hot movement updates; identify the appropriate low-cadence data path. Account for packet budgets and prioritization. Trace the required generated/shared contract changes through the Go server and Godot client.

Deliver:
1. Current ownership and data-flow map.
2. Exact filtering boundary and why it belongs there rather than in simulation.
3. Entity-class policy, hysteresis, self/spectate rules, and leaving-interest behavior.
4. Low-cadence offscreen-player locator design.
5. Server lane/candidate/baseline/budget changes.
6. Shared generated packet/wire changes and client routing/application changes.
7. Concrete test plan and important failure modes.
8. A minimal ordered implementation sequence.

Cite every material conclusion as repository-relative `path:line` or `path:start-end`, and name concrete symbols. Distinguish verified current behavior from proposed design. Do not claim completeness without checking the relevant scope.

End with exactly one line beginning `BENCHMARK_EVIDENCE_JSON:` followed by a compact JSON object. The object must contain arrays named `visibility_seam`, `projection_boundary`, `baseline_membership`, `hysteresis`, `self_spectate`, `locator_lane`, `generated_contracts`, `client_application`, and `tests`. Each item must contain `path`, `symbol`, `lines`, and `claim`.
'''


def isolated_path(selected: Path | None) -> str:
    blocked = {str(CBM_BIN).casefold(), str(GRIMOIRE_BIN).casefold()}
    entries = [entry for entry in os.environ.get("PATH", "").split(os.pathsep) if entry]
    kept = [entry for entry in entries if str(Path(entry)).casefold() not in blocked]
    if selected is not None:
        kept.insert(0, str(selected))
    return os.pathsep.join(kept)


runs = {
    "plain": {
        "cwd": ROOT / "space-rocks-bench-plain",
        "profile": "benchplain",
        "skill": None,
        "path": isolated_path(None),
    },
    "cbm": {
        "cwd": ROOT / "space-rocks-bench-cbm",
        "profile": "benchcbm",
        "skill": "codebase-memory",
        "path": isolated_path(CBM_BIN),
    },
    "grimoire": {
        "cwd": ROOT / "space-rocks-bench-grimoire",
        "profile": "benchgrim",
        "skill": "grimoire",
        "path": isolated_path(GRIMOIRE_BIN),
    },
}

processes: dict[str, tuple[subprocess.Popen[str], float, Path, Path, Path]] = {}
for name, spec in runs.items():
    usage = OUT / f"{name}.usage.json"
    stdout_path = OUT / f"{name}.stdout.txt"
    stderr_path = OUT / f"{name}.stderr.txt"
    command = [
        HERMES,
        "-p", str(spec["profile"]),
        "-m", "gpt-5.6-sol",
        "--provider", "openai-codex",
        "--usage-file", str(usage),
    ]
    if spec["skill"] is not None:
        command.extend(["--skills", str(spec["skill"])])
    command.extend(["-z", PROMPT])
    env = os.environ.copy()
    env["PATH"] = str(spec["path"])
    env["HERMES_ACCEPT_HOOKS"] = "1"
    if name == "cbm":
        env["CBM_CACHE_DIR"] = str(CBM_CACHE)
    stdout_file = stdout_path.open("w", encoding="utf-8", newline="")
    stderr_file = stderr_path.open("w", encoding="utf-8", newline="")
    process = subprocess.Popen(
        command,
        cwd=spec["cwd"],
        env=env,
        stdout=stdout_file,
        stderr=stderr_file,
        text=True,
    )
    processes[name] = (process, time.perf_counter(), usage, stdout_path, stderr_path)

summary: dict[str, object] = {
    "task": "network-interest architecture and implementation plan",
    "commit": "460da4af05c44d1835401fa853f5fc6b718262c8",
    "model": "gpt-5.6-sol",
    "parallel": True,
    "runs": {},
}
for name, (process, started, usage, stdout_path, stderr_path) in processes.items():
    try:
        exit_code = process.wait(timeout=1800)
    except subprocess.TimeoutExpired:
        process.kill()
        exit_code = -9
    elapsed = time.perf_counter() - started
    usage_data = json.loads(usage.read_text(encoding="utf-8")) if usage.is_file() else None
    summary["runs"][name] = {
        "exit_code": exit_code,
        "elapsed_seconds": round(elapsed, 3),
        "usage": usage_data,
        "stdout": str(stdout_path),
        "stderr": str(stderr_path),
    }

(OUT / "summary.json").write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")
print(json.dumps(summary, indent=2))
