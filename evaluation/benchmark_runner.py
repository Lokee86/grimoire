from __future__ import annotations

import json
import os
from pathlib import Path
import shutil
import subprocess
import time
from datetime import datetime, timezone
from typing import Any

import yaml

from benchmark_grounding import summarize_audit, validate_answer

COMMON_PROMPT = r'''Read-only repository investigation. Do not modify repository files, run generators, or implement the change. Normal shell, Git, and direct file inspection are allowed. Use only the optional discovery tool available in this condition, and stop using it when direct inspection is cheaper.

Produce an implementation-grade investigation grounded in the checked-out revision. Cite every material current-behavior claim as repository-relative `path:line` or `path:start-end`. Distinguish verified current behavior from proposed design and from documentation rationale.

End with exactly one line beginning `BENCHMARK_EVIDENCE_JSON:` followed by compact JSON with one array named `evidence`. Every item must contain `path`, `symbol`, `lines`, and `claim`. When the available discovery system returns an opaque inspected source-range handle, also include that exact value as `handle`; do not invent handles.
'''


def run_checked(command: list[str], *, cwd: Path | None = None, env: dict[str, str] | None = None, timeout: int = 1200) -> subprocess.CompletedProcess[str]:
    result = subprocess.run(command, cwd=cwd, env=env, text=True, capture_output=True, timeout=timeout)
    if result.returncode != 0:
        raise RuntimeError(
            f"command failed ({result.returncode}): {command}\nstdout:\n{result.stdout}\nstderr:\n{result.stderr}"
        )
    return result


def resolve_revision(repository: Path, revision: str) -> str:
    return run_checked(["git", "rev-parse", revision], cwd=repository, timeout=300).stdout.strip()


def prepare_worktree(repository: Path, destination: Path, revision: str) -> str:
    commit = resolve_revision(repository, revision)
    destination.parent.mkdir(parents=True, exist_ok=True)
    if destination.exists():
        try:
            run_checked(["git", "worktree", "remove", "--force", str(destination)], cwd=repository, timeout=300)
        except Exception:
            shutil.rmtree(destination, ignore_errors=True)
    run_checked(["git", "worktree", "prune"], cwd=repository, timeout=300)
    run_checked(["git", "worktree", "add", "--detach", str(destination), commit], cwd=repository, timeout=600)
    return commit


def isolated_path(selected: Path | None, blocked: list[Path]) -> str:
    denied = {str(path).casefold() for path in blocked}
    kept = [
        entry for entry in os.environ.get("PATH", "").split(os.pathsep)
        if entry and str(Path(entry)).casefold() not in denied
    ]
    if selected is not None:
        kept.insert(0, str(selected))
    return os.pathsep.join(kept)


class BenchmarkEnvironment:
    def __init__(self, workspace: Path, grimoire_build: Path, model: str, provider: str):
        self.workspace = workspace
        self.grimoire_repo = workspace / "grimoire"
        self.grimoire_build = grimoire_build
        self.model = model
        self.provider = provider
        self.hermes = Path(shutil.which("hermes") or "hermes")
        self.profile_root = Path.home() / "AppData" / "Local" / "hermes" / "profiles"
        self.base_profile = self.profile_root / "benchplain"
        self.cbm_binary = workspace / "cbm-bin" / "codebase-memory-mcp.exe"
        self.cbm_proxy = self.grimoire_repo / "evaluation" / "cbm_mcp_unpaged_proxy.py"
        self.grimoire_binary = grimoire_build / "bin" / "grimoire.exe"
        self.lexicon_binary = grimoire_build / "bin" / "lexicon.exe"
        self.arcana_binary = grimoire_build / "bin" / "arcana.exe"
        self.cbm_skill = self.grimoire_repo / "evaluation" / "results" / "cbm-0.9.0-SKILL.md"
        self.grimoire_skill = grimoire_build / "skills" / "grimoire" / "SKILL.md"

    def require_dependencies(self, conditions: tuple[str, ...]) -> None:
        required = [self.hermes, self.base_profile / "config.yaml", self.base_profile / ".env"]
        if "cbm" in conditions:
            required.extend([self.cbm_binary, self.cbm_proxy, self.cbm_skill])
        if "grimoire" in conditions:
            required.extend([
                self.grimoire_binary, self.lexicon_binary, self.arcana_binary, self.grimoire_skill,
            ])
        for path in required:
            if not path.exists():
                raise RuntimeError(f"required benchmark dependency missing: {path}")

    def base_config(self) -> dict[str, Any]:
        return yaml.safe_load((self.base_profile / "config.yaml").read_text(encoding="utf-8"))

    def prepare_profile(
        self,
        name: str,
        condition: str,
        checkout: Path,
        *,
        cbm_cache: Path | None,
        audit_log: Path | None,
    ) -> None:
        profile = self.profile_root / name
        shutil.rmtree(profile, ignore_errors=True)
        profile.mkdir(parents=True, exist_ok=True)
        shutil.copy2(self.base_profile / ".env", profile / ".env")
        config = self.base_config()
        config["model"] = {
            "provider": self.provider,
            "default": self.model,
            "base_url": "https://chatgpt.com/backend-api/codex",
        }
        config.setdefault("memory", {})["memory_enabled"] = False
        config.setdefault("memory", {})["user_profile_enabled"] = False
        config["mcp_servers"] = {}
        if condition == "cbm":
            if cbm_cache is None:
                raise ValueError("CBM profile requires a cache")
            config["mcp_servers"]["codebase-memory"] = {
                "command": str(self.hermes.parent / "python.exe"),
                "args": [
                    str(self.cbm_proxy), "--binary", str(self.cbm_binary),
                    "--cwd", str(checkout), "--cache-dir", str(cbm_cache),
                ],
                "connect_timeout": 120.0,
                "enabled": True,
            }
        elif condition == "grimoire":
            arguments = ["mcp", "--root", str(checkout), "--state-mode", "refresh-if-needed"]
            if audit_log is not None:
                arguments.extend(["--audit-log", str(audit_log)])
            config["mcp_servers"]["grimoire"] = {
                "command": str(self.grimoire_binary),
                "args": arguments,
                "connect_timeout": 300.0,
                "enabled": True,
            }
        (profile / "config.yaml").write_text(yaml.safe_dump(config, sort_keys=False), encoding="utf-8")
        if condition == "cbm":
            target = profile / "skills" / "codebase-memory"
            target.mkdir(parents=True, exist_ok=True)
            shutil.copy2(self.cbm_skill, target / "SKILL.md")
        elif condition == "grimoire":
            target = profile / "skills" / "grimoire"
            target.mkdir(parents=True, exist_ok=True)
            shutil.copy2(self.grimoire_skill, target / "SKILL.md")

    def prewarm_cbm(self, checkout: Path, cache: Path, output: Path) -> dict[str, Any]:
        shutil.rmtree(cache, ignore_errors=True)
        cache.mkdir(parents=True, exist_ok=True)
        environment = os.environ.copy()
        environment["CBM_CACHE_DIR"] = str(cache)
        started = time.perf_counter()
        result = run_checked(
            [str(self.cbm_binary), "cli", "index_repository", "--repo-path", str(checkout)],
            env=environment,
            timeout=1800,
        )
        elapsed = time.perf_counter() - started
        (output / "cbm-index.stdout.txt").write_text(result.stdout, encoding="utf-8")
        (output / "cbm-index.stderr.txt").write_text(result.stderr, encoding="utf-8")
        payload = json.loads(result.stdout)
        project = payload.get("project") or payload.get("name")
        if not project:
            raise RuntimeError("CBM index result omitted project name")
        return {"project": str(project), "elapsed_seconds": round(elapsed, 3)}

    def prewarm_grimoire(self, checkout: Path, output: Path) -> dict[str, Any]:
        environment = os.environ.copy()
        environment["PATH"] = isolated_path(
            self.grimoire_build / "bin",
            [self.cbm_binary.parent, self.grimoire_build / "bin"],
        )
        started = time.perf_counter()
        result = run_checked([
            str(self.grimoire_binary), "status", "--root", str(checkout), "--force",
            "--lexicon-command", str(self.lexicon_binary),
            "--arcana-command", str(self.arcana_binary),
        ], env=environment, timeout=3600)
        elapsed = time.perf_counter() - started
        (output / "grimoire-prewarm.stdout.txt").write_text(result.stdout, encoding="utf-8")
        (output / "grimoire-prewarm.stderr.txt").write_text(result.stderr, encoding="utf-8")
        payload = json.loads(result.stdout)
        return {
            "elapsed_seconds": round(elapsed, 3),
            "timings": payload.get("timings") or {},
            "actions": payload.get("actions") or [],
        }

    def launch(
        self,
        task: dict[str, Any],
        condition: str,
        checkout: Path,
        output: Path,
        profile: str,
        cbm_project: str | None,
    ) -> tuple[subprocess.Popen[str], float, object, object, Path]:
        usage = output / f"{condition}.usage.json"
        stdout_path = output / f"{condition}.stdout.txt"
        stderr_path = output / f"{condition}.stderr.txt"
        prompt = COMMON_PROMPT + "\nTask:\n" + task["prompt"] + "\n"
        if condition == "cbm" and cbm_project:
            prompt += f"\nThe isolated CBM project is `{cbm_project}`. Use no other indexed project.\n"
        command = [
            str(self.hermes), "-p", profile, "-m", self.model,
            "--provider", self.provider, "--usage-file", str(usage),
        ]
        if condition == "cbm":
            command.extend(["--skills", "codebase-memory"])
        elif condition == "grimoire":
            command.extend(["--skills", "grimoire"])
        command.extend(["-z", prompt])
        environment = os.environ.copy()
        environment["HERMES_ACCEPT_HOOKS"] = "1"
        blocked = [self.cbm_binary.parent, self.grimoire_build / "bin"]
        selected = self.cbm_binary.parent if condition == "cbm" else self.grimoire_build / "bin" if condition == "grimoire" else None
        environment["PATH"] = isolated_path(selected, blocked)
        stdout_file = stdout_path.open("w", encoding="utf-8", newline="")
        stderr_file = stderr_path.open("w", encoding="utf-8", newline="")
        process = subprocess.Popen(command, cwd=checkout, env=environment, stdout=stdout_file, stderr=stderr_file, text=True)
        return process, time.perf_counter(), stdout_file, stderr_file, usage


def collect_runs(
    active: dict[str, tuple],
    *,
    checkout_by_condition: dict[str, Path],
    output: Path,
    expected_sections: list[str],
    required_path_prefixes: list[str],
    timeout_seconds: int = 2400,
) -> dict[str, dict[str, Any]]:
    results: dict[str, dict[str, Any]] = {}
    pending = set(active)
    while pending:
        for condition in list(pending):
            process, started, stdout_file, stderr_file, usage = active[condition]
            exit_code = process.poll()
            if exit_code is None and time.perf_counter() - started <= timeout_seconds:
                continue
            if exit_code is None:
                process.kill()
                try:
                    process.wait(timeout=10)
                except subprocess.TimeoutExpired:
                    pass
                exit_code = -9
            finished = time.perf_counter()
            stdout_file.close()
            stderr_file.close()
            usage_data = json.loads(usage.read_text(encoding="utf-8")) if usage.is_file() else None
            answer_path = output / f"{condition}.stdout.txt"
            audit_path = output / "grimoire.mcp-audit.jsonl" if condition == "grimoire" else None
            grounding = validate_answer(
                checkout_by_condition[condition],
                answer_path.read_text(encoding="utf-8") if answer_path.is_file() else "",
                exit_code=exit_code,
                expected_sections=expected_sections,
                audit_log=audit_path,
                require_grimoire_handles=False,
                required_path_prefixes=required_path_prefixes,
            )
            grounding_path = output / f"{condition}.grounding.json"
            grounding_path.write_text(json.dumps(grounding.to_dict(), indent=2) + "\n", encoding="utf-8")
            results[condition] = {
                "exit_code": exit_code,
                "elapsed_seconds": round(finished - started, 3),
                "usage": usage_data,
                "answer_bytes": answer_path.stat().st_size if answer_path.is_file() else 0,
                "discovery_output": summarize_audit(audit_path),
                "grounding": grounding.to_dict(),
                "valid": grounding.valid,
                "completed_at": datetime.now(timezone.utc).isoformat(),
            }
            pending.remove(condition)
        if pending:
            time.sleep(1)
    return results
