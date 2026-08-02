from __future__ import annotations

import argparse
import hashlib
import json
import os
import platform
import shutil
import subprocess
import sys
import time
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path
from typing import Any

from validation_summary import summarize


def executable(name: str, *fallbacks: Path) -> str:
    names = [name]
    if name.lower().endswith(".exe"):
        names.append(name[:-4])
    candidates = [
        *(shutil.which(candidate) for candidate in names),
        *(str(path) for path in fallbacks),
    ]
    for candidate in candidates:
        if candidate and Path(candidate).is_file():
            return candidate
    raise FileNotFoundError(f"cannot locate executable: {name}")


def npm_command(*args: str) -> list[str]:
    if os.name != "nt":
        return [executable("npm"), *args]
    node = Path(executable("node.exe", Path("C:/Program Files/nodejs/node.exe")))
    npm_cli = node.parent / "node_modules" / "npm" / "bin" / "npm-cli.js"
    if not npm_cli.is_file():
        raise FileNotFoundError(f"cannot locate npm CLI: {npm_cli}")
    return [str(node), str(npm_cli), *args]


def dotnet_command(*args: str) -> list[str]:
    configured = os.environ.get("LEXICON_DOTNET")
    if configured:
        if not Path(configured).is_file():
            raise FileNotFoundError(f"LEXICON_DOTNET does not exist: {configured}")
        return [configured, *args]
    return [executable("dotnet", Path("C:/Program Files/dotnet/dotnet.exe")), *args]


def dotnet_runtime_identifier() -> str:
    architecture = platform.machine().lower()
    if architecture in {"amd64", "x86_64"}:
        architecture = "x64"
    elif architecture in {"arm64", "aarch64"}:
        architecture = "arm64"
    else:
        raise RuntimeError(f"unsupported .NET architecture: {architecture}")

    if sys.platform == "win32":
        operating_system = "win"
    elif sys.platform == "darwin":
        operating_system = "osx"
    elif sys.platform.startswith("linux"):
        operating_system = "linux"
    else:
        raise RuntimeError(f"unsupported .NET platform: {sys.platform}")
    return f"{operating_system}-{architecture}"


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
        raise RuntimeError(
            f"command failed ({completed.returncode}): {rendered}\n{completed.stdout}"
        )
    return completed.stdout


def workspace_root(root: Path) -> Path:
    common_dir = Path(
        run(
            ["git", "rev-parse", "--path-format=absolute", "--git-common-dir"],
            root,
        ).strip()
    )
    return common_dir.parent.parent


def go_adapter_binary(root: Path, adapter: str) -> Path:
    name = f"lexicon-{adapter}.exe" if os.name == "nt" else f"lexicon-{adapter}"
    return root / "evaluation" / "bin" / name


def build_go_adapter(root: Path, adapter: str) -> None:
    go = executable("go.exe", Path("C:/Program Files/Go/bin/go.exe"))
    run(
        [go, "build", "-trimpath", "-buildvcs=false", "-o", str(go_adapter_binary(root, adapter)), "."],
        root / "adapters" / adapter,
    )


def build_adapters(root: Path, adapters: set[str]) -> None:
    bin_dir = root / "evaluation" / "bin"
    bin_dir.mkdir(parents=True, exist_ok=True)
    for adapter in sorted(adapters & {"c-family", "gdscript", "java", "kotlin", "lotusscript"}):
        build_go_adapter(root, adapter)
    if "typescript" in adapters:
        run(npm_command("run", "build"), root / "adapters" / "typescript")
    if "rust" in adapters:
        cargo = executable("cargo.exe", Path.home() / ".cargo" / "bin" / "cargo.exe")
        run(
            [cargo, "build", "--release", "--manifest-path", "adapters/rust/Cargo.toml"],
            root,
        )
    if "csharp" in adapters:
        output = bin_dir / "csharp"
        if output.exists():
            shutil.rmtree(output)
        run(
            dotnet_command(
                "publish",
                "adapters/csharp/Lexicon.CSharp.csproj",
                "--configuration",
                "Release",
                "--nologo",
                "--output",
                str(output),
                "--self-contained",
                "true",
                "--runtime",
                dotnet_runtime_identifier(),
                "-p:AssemblyName=lexicon-csharp",
                "-p:ContinuousIntegrationBuild=true",
                "-p:DebugSymbols=false",
                "-p:DebugType=None",
                "-p:Deterministic=true",
                "-p:UseAppHost=true",
            ),
            root,
        )


def adapter_command(
    root: Path,
    adapter: str,
    repository: Path,
    output: Path,
    extra_args: list[str] | None = None,
) -> tuple[list[str], dict[str, str]]:
    env = os.environ.copy()
    arguments = ["--repo", str(repository), "--output", str(output), *(extra_args or [])]
    if adapter in {"c-family", "gdscript", "java", "kotlin", "lotusscript"}:
        return [str(go_adapter_binary(root, adapter)), *arguments], env
    if adapter == "python":
        env["PYTHONPATH"] = str(root / "adapters" / "python")
        return [sys.executable, "-m", "lexicon_python", *arguments], env
    if adapter == "ruby":
        ruby = executable("ruby.exe", Path("C:/Ruby34-x64/bin/ruby.exe"))
        return [ruby, str(root / "adapters" / "ruby" / "lexicon_ruby.rb"), *arguments], env
    if adapter == "typescript":
        node = executable("node.exe", Path("C:/Program Files/nodejs/node.exe"))
        return [node, str(root / "adapters" / "typescript" / "dist" / "cli.js"), *arguments], env
    if adapter == "rust":
        cargo = Path(executable("cargo.exe", Path.home() / ".cargo" / "bin" / "cargo.exe"))
        env["PATH"] = str(cargo.parent) + os.pathsep + env.get("PATH", "")
        binary = root / "adapters" / "rust" / "target" / "release" / (
            "lexicon-rust-adapter.exe" if os.name == "nt" else "lexicon-rust-adapter"
        )
        return [str(binary), *arguments], env
    if adapter == "csharp":
        configured = env.get("LEXICON_DOTNET")
        if configured and Path(configured).is_file():
            dotnet_root = str(Path(configured).resolve().parent)
            env["DOTNET_ROOT"] = dotnet_root
            env["PATH"] = dotnet_root + os.pathsep + env.get("PATH", "")
        binary = root / "evaluation" / "bin" / "csharp" / (
            "lexicon-csharp.exe" if os.name == "nt" else "lexicon-csharp"
        )
        return [str(binary), *arguments], env
    raise ValueError(f"unsupported adapter: {adapter}")


def validate_case(
    root: Path,
    workspace: Path,
    output_root: Path,
    case: dict[str, Any],
) -> dict[str, Any]:
    case_dir = output_root / case["id"]
    if case_dir.exists():
        shutil.rmtree(case_dir)
    case_dir.mkdir(parents=True)
    repository = workspace / case["repository"]
    if not repository.exists():
        raise FileNotFoundError(f"missing corpus repository: {repository}")
    expected_revision = case.get("revision")
    if expected_revision:
        actual_revision = run(["git", "rev-parse", "HEAD"], repository).strip()
        if actual_revision != expected_revision:
            raise RuntimeError(
                f"corpus revision mismatch for {case['id']}: "
                f"expected {expected_revision}, found {actual_revision}"
            )

    durations: list[float] = []
    outputs: list[Path] = []
    for run_number in (1, 2):
        output = case_dir / f"run{run_number}.jsonl"
        command, env = adapter_command(
            root,
            case["adapter"],
            repository,
            output,
            case.get("adapter_args"),
        )
        started = time.perf_counter()
        run(command, root, env)
        durations.append(round(time.perf_counter() - started, 6))
        run([sys.executable, "tools/validate_jsonl.py", str(output)], root)
        outputs.append(output)

    first_bytes = outputs[0].read_bytes()
    deterministic = first_bytes == outputs[1].read_bytes()
    semantic = summarize(outputs[0], case)
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
    unexpected_unresolved_call_reasons = [
        reason
        for reason in case.get("expected_zero_unresolved_call_reasons", [])
        if semantic["unresolved_call_reasons"].get(reason, 0) != 0
    ]
    sample_path = case_dir / "audit_samples.json"
    sample_path.write_text(
        json.dumps(semantic["samples"], indent=2, sort_keys=True) + "\n",
        encoding="utf-8",
    )
    del semantic["samples"]
    result = {
        "adapter": case["adapter"],
        "deterministic": deterministic,
        "durations_seconds": durations,
        "id": case["id"],
        "missing_required_relations": missing,
        "unexpected_nonzero_relations": unexpected,
        "unexpected_nonzero_unresolved_call_reasons": unexpected_unresolved_call_reasons,
        "output_bytes": len(first_bytes),
        "output_sha256": hashlib.sha256(first_bytes).hexdigest(),
        "repository": case["repository"],
        "revision": case.get("revision"),
        "split": case["split"],
        **semantic,
    }
    (case_dir / "summary.json").write_text(
        json.dumps(result, indent=2, sort_keys=True) + "\n",
        encoding="utf-8",
    )
    return result


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--adapter",
        action="append",
        choices=(
            "c-family",
            "csharp",
            "gdscript",
            "java",
            "kotlin",
            "lotusscript",
            "python",
            "ruby",
            "rust",
            "typescript",
        ),
    )
    parser.add_argument("--case", action="append")
    parser.add_argument("--jobs", type=int, default=3)
    args = parser.parse_args()

    root = Path(__file__).resolve().parents[1]
    workspace = workspace_root(root)
    manifest = json.loads(
        (Path(__file__).with_name("corpus.json")).read_text(encoding="utf-8")
    )
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
    summary_path.write_text(
        json.dumps(summary, indent=2, sort_keys=True) + "\n",
        encoding="utf-8",
    )

    run1_paths = [str(output_root / result["id"] / "run1.jsonl") for result in results]
    if run1_paths:
        run(
            [
                sys.executable,
                "tools/semantic_report.py",
                *run1_paths,
                "--output",
                str(output_root / "semantic_report.json"),
            ],
            root,
        )

    failed_gates = failures or any(
        not result["deterministic"]
        or result["missing_required_edges"]
        or result["missing_required_nodes"]
        or result["missing_required_relations"]
        or result["unexpected_nonzero_relations"]
        or result["unexpected_nonzero_unresolved_call_reasons"]
        for result in results
    )
    complete_selection = (
        not args.adapter and not args.case and len(cases) == len(manifest["cases"])
    )
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
