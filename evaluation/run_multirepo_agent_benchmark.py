from __future__ import annotations

import json
import os
import shutil
import subprocess
import time
from datetime import datetime, timezone
from pathlib import Path

import yaml

ROOT = Path(r"C:\!bin\workspace")
GRIMOIRE_REPO = ROOT / "grimoire"
OUT = GRIMOIRE_REPO / "evaluation" / "results" / "multi-repo-agent-benchmark-2026-07-27-v1"
CHECKOUT_ROOT = ROOT / "benchmark-checkouts" / "multi-repo-agent-benchmark-2026-07-27-v1"
HERMES = shutil.which("hermes") or "hermes"
HERMES_PROFILES = Path.home() / "AppData" / "Local" / "hermes" / "profiles"
BASE_PROFILE = HERMES_PROFILES / "benchplain"
CBM_BIN = ROOT / "cbm-bin" / "codebase-memory-mcp.exe"
CBM_PROXY = GRIMOIRE_REPO / "evaluation" / "cbm_mcp_unpaged_proxy.py"
GRIMOIRE_BUILD = GRIMOIRE_REPO / "evaluation" / "build-current"
GRIMOIRE_BIN = GRIMOIRE_BUILD / "bin" / "grimoire.exe"
LEXICON_BIN = GRIMOIRE_BUILD / "bin" / "lexicon.exe"
ARCANA_BIN = GRIMOIRE_BUILD / "bin" / "arcana.exe"
CBM_SKILL = GRIMOIRE_REPO / "evaluation" / "results" / "cbm-0.9.0-SKILL.md"
GRIMOIRE_SKILL = GRIMOIRE_REPO / "skills" / "grimoire" / "SKILL.md"
MODEL = "gpt-5.6-sol"
PROVIDER = "openai-codex"

COMMON = r'''Read-only repository investigation. Do not modify repository files, run generators, or implement the change. Normal shell and file discovery are allowed. Use only the discovery tool available in this condition; do not attempt to access the other discovery product.

Cite every material conclusion as repository-relative `path:line` or `path:start-end`, naming concrete symbols. Distinguish verified current behavior from proposed design. Check exact source before relying on documentation or tool summaries. Do not claim completeness without checking the relevant scope.
'''

TASKS = [
    {
        "id": "hikaricp-connection-lifecycle",
        "repo": ROOT / "corpus" / "java-hikaricp",
        "commit": "a4d93f4f85517f90e632b795486d7102e933d7ff",
        "prompt": r'''
Task: Trace the complete HikariCP JDBC connection lifecycle at this revision.

Start at the public `HikariDataSource.getConnection` entry point and follow a successful connection through pool acquisition, timeout/deadline handling, `ConcurrentBag` coordination, liveness/validation checks, proxy creation, leak-task scheduling, application close, recycle, eviction/retirement, and physical close. Also explain the principal failure, interruption, shutdown, and maintenance paths, including how housekeeper/fill activity interacts with acquisition.

Deliver:
1. A concrete end-to-end call/data-flow map for the normal lifecycle.
2. Acquisition-loop timing and timeout semantics.
3. `ConcurrentBag` handoff/waiter behavior and pool-growth interaction.
4. Validation, dead-connection, bypass-window, and exception paths.
5. Proxy/close/recycle ownership and state-reset behavior.
6. Leak detection scheduling and cancellation.
7. Eviction, max-lifetime/idle retirement, shutdown, and physical close.
8. Configuration, metrics/JMX, and observability hooks affecting this path.
9. Relevant tests and a focused failure-mode test plan.

End with exactly one line beginning `BENCHMARK_EVIDENCE_JSON:` followed by compact JSON. It must contain arrays named `public_entry`, `acquisition_loop`, `bag_coordination`, `validation`, `proxy_lifecycle`, `leak_detection`, `recycle_eviction`, `maintenance_shutdown`, and `tests`. Every item must contain `path`, `symbol`, `lines`, and `claim`.
''',
    },
    {
        "id": "detekt-ruleset-plugin-flow",
        "repo": ROOT / "corpus" / "kotlin-detekt",
        "commit": "f9e1d5cc239ab740ce499b1edb36b872012648e2",
        "prompt": r'''
Task: Trace how a third-party detekt `RuleSetProvider` plugin JAR is accepted, discovered, configured, executed, suppressed/baselined, and reported at this revision.

Cover both the command-line application and the Gradle plugin. Begin at the user-facing plugin/classpath configuration surfaces, follow classloader and `ServiceLoader` ownership, provider and rule-factory creation, configuration/activation, analysis execution, finding transformation/suppression/baseline handling, report generation, output/error propagation, and Gradle task/worker wiring. Identify where the two entry paths converge and where they differ.

Deliver:
1. CLI plugin-input to loaded-provider flow.
2. Gradle configuration/task/worker flow for custom plugins.
3. Classloader lifetime and service-discovery boundaries.
4. Provider, rule factory, configuration, activation, and validation behavior.
5. Analyzer execution and concurrency/lifecycle behavior.
6. Suppression and baseline ordering relative to rule findings.
7. Finding aggregation, severity/failure decisions, reporters, and output files.
8. Important plugin incompatibility, duplicate, malformed-service, and configuration failure modes.
9. Existing tests and a focused cross-entry-point test plan.

End with exactly one line beginning `BENCHMARK_EVIDENCE_JSON:` followed by compact JSON. It must contain arrays named `plugin_input`, `classloader_discovery`, `provider_rules`, `configuration`, `analysis_execution`, `suppression_baseline`, `findings_reporting`, `gradle_wiring`, and `tests`. Every item must contain `path`, `symbol`, `lines`, and `claim`.
''',
    },
    {
        "id": "nowinandroid-topic-notification-muting",
        "repo": ROOT / "corpus" / "kotlin-nowinandroid",
        "commit": "7d45eae4f8720a0c77f507712ba2437ff974b6ed",
        "prompt": r'''
Task: Design a concrete implementation plan for per-topic notification muting in Now in Android at this revision.

A muted topic remains followed and remains eligible for the For You feed, but newly synchronized news associated only with muted topics must not produce user-facing news notifications. Users can toggle mute independently from follow on the topic screen and manage muted topics from settings. The choice must work offline and integrate with the repository's actual user-data persistence, backup, demo/prod, and synchronization behavior. Do not invent remote user-data synchronization if it does not exist.

Deliver:
1. Current user-data, followed-topic, feed, sync, and notification ownership/data-flow map.
2. Exact model and persistence changes, including defaults, schema evolution, migration, and backup implications.
3. Repository/domain API changes that preserve follow and mute as independent states.
4. Exact notification-filtering boundary and behavior for multi-topic news, empty topic sets, removed topics, repeat sync, and stale data.
5. Topic-screen state/events/UI changes and accessibility behavior.
6. Settings-screen state/events/UI for reviewing and unmuting topics.
7. Demo/prod, DI/module, analytics, and test-double changes.
8. Concrete unit, integration, migration, UI, and sync/notification tests.
9. Important failure modes and a minimal ordered implementation sequence.

End with exactly one line beginning `BENCHMARK_EVIDENCE_JSON:` followed by compact JSON. It must contain arrays named `user_data_model`, `persistence_migration`, `repository_domain`, `sync_behavior`, `notification_filtering`, `topic_ui`, `settings_ui`, `dependency_wiring`, `analytics_accessibility`, and `tests`. Every item must contain `path`, `symbol`, `lines`, and `claim`.
''',
    },
]


def run_checked(command: list[str], cwd: Path | None = None, env: dict[str, str] | None = None, timeout: int = 1200) -> subprocess.CompletedProcess[str]:
    result = subprocess.run(command, cwd=cwd, env=env, text=True, capture_output=True, timeout=timeout)
    if result.returncode != 0:
        raise RuntimeError(f"command failed ({result.returncode}): {command}\nstdout:\n{result.stdout}\nstderr:\n{result.stderr}")
    return result


def prepare_worktree(repo: Path, destination: Path, commit: str) -> None:
    destination.parent.mkdir(parents=True, exist_ok=True)
    if destination.exists():
        try:
            run_checked(["git", "worktree", "remove", "--force", str(destination)], cwd=repo, timeout=300)
        except Exception:
            shutil.rmtree(destination, ignore_errors=True)
    run_checked(["git", "worktree", "prune"], cwd=repo, timeout=300)
    run_checked(["git", "worktree", "add", "--detach", str(destination), commit], cwd=repo, timeout=600)


def base_config() -> dict:
    return yaml.safe_load((BASE_PROFILE / "config.yaml").read_text(encoding="utf-8"))


def prepare_profile(name: str, condition: str, checkout: Path, cbm_cache: Path | None = None) -> Path:
    profile = HERMES_PROFILES / name
    shutil.rmtree(profile, ignore_errors=True)
    profile.mkdir(parents=True, exist_ok=True)
    shutil.copy2(BASE_PROFILE / ".env", profile / ".env")
    config = base_config()
    config["model"] = {"provider": PROVIDER, "default": MODEL, "base_url": "https://chatgpt.com/backend-api/codex"}
    config.setdefault("memory", {})["memory_enabled"] = False
    config.setdefault("memory", {})["user_profile_enabled"] = False
    config["mcp_servers"] = {}
    if condition == "cbm":
        assert cbm_cache is not None
        config["mcp_servers"] = {
            "codebase-memory": {
                "command": str(Path(HERMES).parent / "python.exe"),
                "args": [str(CBM_PROXY), "--binary", str(CBM_BIN), "--cwd", str(checkout), "--cache-dir", str(cbm_cache)],
                "connect_timeout": 120.0,
                "enabled": True,
            }
        }
    elif condition == "grimoire":
        config["mcp_servers"] = {
            "grimoire": {
                "command": str(GRIMOIRE_BIN),
                "args": ["mcp", "--root", str(checkout), "--state-mode", "refresh-if-needed"],
                "connect_timeout": 300.0,
                "enabled": True,
            }
        }
    (profile / "config.yaml").write_text(yaml.safe_dump(config, sort_keys=False), encoding="utf-8")
    if condition == "cbm":
        target = profile / "skills" / "codebase-memory"
        target.mkdir(parents=True, exist_ok=True)
        shutil.copy2(CBM_SKILL, target / "SKILL.md")
    elif condition == "grimoire":
        target = profile / "skills" / "grimoire"
        target.mkdir(parents=True, exist_ok=True)
        shutil.copy2(GRIMOIRE_SKILL, target / "SKILL.md")
    return profile


def isolated_path(selected: Path | None) -> str:
    blocked = {str(CBM_BIN.parent).casefold(), str(GRIMOIRE_BUILD / "bin").casefold()}
    entries = [entry for entry in os.environ.get("PATH", "").split(os.pathsep) if entry]
    kept = [entry for entry in entries if str(Path(entry)).casefold() not in blocked]
    if selected is not None:
        kept.insert(0, str(selected))
    return os.pathsep.join(kept)


def prewarm_cbm(checkout: Path, cache: Path, log_dir: Path) -> tuple[str, float]:
    shutil.rmtree(cache, ignore_errors=True)
    cache.mkdir(parents=True, exist_ok=True)
    env = os.environ.copy()
    env["CBM_CACHE_DIR"] = str(cache)
    started = time.perf_counter()
    result = run_checked([str(CBM_BIN), "cli", "index_repository", "--repo-path", str(checkout)], env=env, timeout=1800)
    elapsed = time.perf_counter() - started
    (log_dir / "cbm-index.stdout.txt").write_text(result.stdout, encoding="utf-8")
    (log_dir / "cbm-index.stderr.txt").write_text(result.stderr, encoding="utf-8")
    payload = json.loads(result.stdout)
    project = payload.get("project") or payload.get("name")
    if not project:
        raise RuntimeError(f"CBM index result omitted project name: {result.stdout}")
    return str(project), elapsed


def prewarm_grimoire(checkout: Path, log_dir: Path) -> float:
    env = os.environ.copy()
    env["PATH"] = isolated_path(GRIMOIRE_BUILD / "bin")
    started = time.perf_counter()
    result = run_checked([
        str(GRIMOIRE_BIN), "status", "--root", str(checkout), "--force",
        "--lexicon-command", str(LEXICON_BIN), "--arcana-command", str(ARCANA_BIN),
    ], env=env, timeout=3600)
    elapsed = time.perf_counter() - started
    (log_dir / "grimoire-prewarm.stdout.txt").write_text(result.stdout, encoding="utf-8")
    (log_dir / "grimoire-prewarm.stderr.txt").write_text(result.stderr, encoding="utf-8")
    return elapsed


def launch_run(task: dict, condition: str, checkout: Path, task_out: Path, profile_name: str, cbm_project: str | None) -> tuple[subprocess.Popen[str], float, object, object, Path]:
    usage = task_out / f"{condition}.usage.json"
    stdout_path = task_out / f"{condition}.stdout.txt"
    stderr_path = task_out / f"{condition}.stderr.txt"
    prompt = COMMON + task["prompt"]
    if condition == "cbm" and cbm_project:
        prompt += f'\nThe CBM index project for this isolated checkout is `{cbm_project}`. Use that project and no other indexed project.\n'
    command = [HERMES, "-p", profile_name, "-m", MODEL, "--provider", PROVIDER, "--usage-file", str(usage)]
    if condition == "cbm":
        command.extend(["--skills", "codebase-memory"])
    elif condition == "grimoire":
        command.extend(["--skills", "grimoire"])
    command.extend(["-z", prompt])
    env = os.environ.copy()
    env["HERMES_ACCEPT_HOOKS"] = "1"
    if condition == "cbm":
        env["PATH"] = isolated_path(CBM_BIN.parent)
    elif condition == "grimoire":
        env["PATH"] = isolated_path(GRIMOIRE_BUILD / "bin")
    else:
        env["PATH"] = isolated_path(None)
    stdout_file = stdout_path.open("w", encoding="utf-8", newline="")
    stderr_file = stderr_path.open("w", encoding="utf-8", newline="")
    process = subprocess.Popen(command, cwd=checkout, env=env, stdout=stdout_file, stderr=stderr_file, text=True)
    return process, time.perf_counter(), stdout_file, stderr_file, usage


def collect_wave(active: dict[str, tuple]) -> dict[str, dict]:
    results: dict[str, dict] = {}
    pending = set(active)
    while pending:
        for condition in list(pending):
            process, started, stdout_file, stderr_file, usage = active[condition]
            exit_code = process.poll()
            if exit_code is None:
                if time.perf_counter() - started > 2400:
                    process.kill()
                    exit_code = -9
                else:
                    continue
            finished = time.perf_counter()
            stdout_file.close()
            stderr_file.close()
            usage_data = json.loads(usage.read_text(encoding="utf-8")) if usage.is_file() else None
            results[condition] = {
                "exit_code": exit_code,
                "elapsed_seconds": round(finished - started, 3),
                "usage": usage_data,
                "completed_at": datetime.now(timezone.utc).isoformat(),
            }
            pending.remove(condition)
        if pending:
            time.sleep(1)
    return results


def main() -> int:
    for required in [HERMES, CBM_BIN, CBM_PROXY, GRIMOIRE_BIN, LEXICON_BIN, ARCANA_BIN, CBM_SKILL, GRIMOIRE_SKILL]:
        if not Path(required).exists():
            raise RuntimeError(f"required benchmark dependency missing: {required}")
    OUT.mkdir(parents=True, exist_ok=True)
    CHECKOUT_ROOT.mkdir(parents=True, exist_ok=True)
    suite = {
        "model": MODEL,
        "provider": PROVIDER,
        "parallel_within_task": True,
        "sequential_tasks": True,
        "started_at": datetime.now(timezone.utc).isoformat(),
        "tasks": {},
    }
    for task in TASKS:
        task_id = task["id"]
        task_out = OUT / task_id
        task_out.mkdir(parents=True, exist_ok=True)
        checkouts: dict[str, Path] = {}
        for condition in ["plain", "cbm", "grimoire"]:
            checkout = CHECKOUT_ROOT / task_id / condition
            prepare_worktree(task["repo"], checkout, task["commit"])
            checkouts[condition] = checkout
        cbm_cache = task_out / "cbm-cache"
        cbm_project, cbm_index_seconds = prewarm_cbm(checkouts["cbm"], cbm_cache, task_out)
        grimoire_prewarm_seconds = prewarm_grimoire(checkouts["grimoire"], task_out)
        profiles: dict[str, str] = {}
        for condition in ["plain", "cbm", "grimoire"]:
            profile_name = f"bench-{task_id[:18]}-{condition}"
            prepare_profile(profile_name, condition, checkouts[condition], cbm_cache)
            profiles[condition] = profile_name
        active = {
            condition: launch_run(task, condition, checkouts[condition], task_out, profiles[condition], cbm_project)
            for condition in ["plain", "cbm", "grimoire"]
        }
        runs = collect_wave(active)
        suite["tasks"][task_id] = {
            "repository": str(task["repo"]),
            "commit": task["commit"],
            "cbm_project": cbm_project,
            "cbm_index_seconds": round(cbm_index_seconds, 3),
            "grimoire_prewarm_seconds": round(grimoire_prewarm_seconds, 3),
            "profiles": profiles,
            "runs": runs,
        }
        (OUT / "summary.partial.json").write_text(json.dumps(suite, indent=2) + "\n", encoding="utf-8")
        time.sleep(3)
    suite["completed_at"] = datetime.now(timezone.utc).isoformat()
    (OUT / "summary.json").write_text(json.dumps(suite, indent=2) + "\n", encoding="utf-8")
    print(json.dumps(suite, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
