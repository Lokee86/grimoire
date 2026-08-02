from __future__ import annotations

import json
import os
import shutil
import subprocess
import tempfile
from pathlib import Path

ADAPTER = Path(__file__).resolve().parents[1]
SMOKE_FIXTURE = ADAPTER / "testdata" / "smoke"
PROJECT_FIXTURE = ADAPTER / "testdata" / "project-graph"
PROJECT = ADAPTER / "Lexicon.CSharp.csproj"
REQUIRED_RELATIONS = {"calls", "reads", "writes", "depends-on", "extends", "implements", "overrides"}


def dotnet() -> str:
    candidates = [
        os.environ.get("LEXICON_DOTNET"),
        shutil.which("dotnet"),
        r"C:\Program Files\dotnet\dotnet.exe",
        str(Path.home() / ".dotnet" / "dotnet.exe"),
    ]
    for candidate in candidates:
        if candidate and Path(candidate).is_file():
            return candidate
    raise FileNotFoundError("dotnet SDK not found; set LEXICON_DOTNET")


def adapter_env() -> dict[str, str]:
    env = os.environ.copy()
    executable = Path(dotnet()).resolve()
    env["LEXICON_DOTNET"] = str(executable)
    env["DOTNET_ROOT"] = str(executable.parent)
    env["PATH"] = str(executable.parent) + os.pathsep + env.get("PATH", "")
    return env


def prepare_adapter() -> None:
    completed = subprocess.run(
        [dotnet(), "build", str(PROJECT), "--nologo"],
        cwd=ADAPTER,
        env=adapter_env(),
        text=True,
        capture_output=True,
    )
    if completed.returncode != 0:
        raise RuntimeError(completed.stdout + completed.stderr)


def prepare_project_fixture() -> None:
    for project in (
        PROJECT_FIXTURE / "App" / "App.csproj",
        PROJECT_FIXTURE / "Other" / "Other.csproj",
    ):
        completed = subprocess.run(
            [dotnet(), "restore", str(project), "--nologo"],
            cwd=PROJECT_FIXTURE,
            env=adapter_env(),
            text=True,
            capture_output=True,
        )
        if completed.returncode != 0:
            raise RuntimeError(completed.stdout + completed.stderr)


def run_adapter(
    repository: Path,
    output: Path,
    *scope: str,
) -> list[dict[str, object]]:
    command = [
        dotnet(),
        "run",
        "--no-build",
        "--no-restore",
        "--project",
        str(PROJECT),
        "--",
        "--repo",
        str(repository),
        "--output",
        str(output),
        *scope,
    ]
    completed = subprocess.run(
        command,
        cwd=ADAPTER,
        env=adapter_env(),
        text=True,
        capture_output=True,
    )
    if completed.returncode != 0:
        raise RuntimeError(completed.stdout + completed.stderr)
    return [json.loads(line) for line in output.read_text(encoding="utf-8").splitlines() if line]


def assert_valid(records: list[dict[str, object]]) -> None:
    assert records and records[0]["record"] == "lexicon"
    nodes = {record["id"] for record in records if record.get("record") == "node"}
    for edge in (record for record in records if record.get("record") == "edge"):
        assert edge["source"] in nodes, edge
        assert edge["target"] in nodes, edge


def repository_node(records: list[dict[str, object]]) -> dict[str, object]:
    return next(
        record
        for record in records
        if record.get("record") == "node" and record.get("kind") == "repository"
    )


def assert_project_reference_resolution(records: list[dict[str, object]]) -> None:
    nodes = {
        record["id"]: record
        for record in records
        if record.get("record") == "node"
    }
    calls = [
        (nodes[edge["source"]], nodes[edge["target"]])
        for edge in records
        if edge.get("record") == "edge" and edge.get("relation") == "calls"
    ]
    matching = [
        (source, target)
        for source, target in calls
        if source.get("path") == "App/Program.cs"
        and target.get("name") == "Execute"
    ]
    assert matching, "App call to Execute was not resolved"
    assert {target["path"] for _, target in matching} == {"Lib/Worker.cs"}, matching


def main() -> int:
    prepare_adapter()
    prepare_project_fixture()
    with tempfile.TemporaryDirectory(prefix="lexicon-csharp-test-") as temporary:
        directory = Path(temporary)
        first_path = directory / "first.jsonl"
        second_path = directory / "second.jsonl"
        incremental_path = directory / "incremental.jsonl"
        project_path = directory / "project.jsonl"
        project_second_path = directory / "project-second.jsonl"
        fallback_path = directory / "fallback.jsonl"
        sharded_path = directory / "sharded.jsonl"

        first = run_adapter(SMOKE_FIXTURE, first_path)
        second = run_adapter(SMOKE_FIXTURE, second_path)
        assert first_path.read_bytes() == second_path.read_bytes()
        assert_valid(first)
        relations = {record["relation"] for record in first if record.get("record") == "edge"}
        assert REQUIRED_RELATIONS <= relations, sorted(REQUIRED_RELATIONS - relations)
        assert repository_node(first)["attributes"]["analysis_mode"] == "files"

        incremental = run_adapter(
            SMOKE_FIXTURE,
            incremental_path,
            "--changed-file",
            "Program.cs",
        )
        assert_valid(incremental)
        header = incremental[0]
        assert header["mode"] == "incremental"
        assert header["changed_files"] == ["Program.cs"]
        assert header["removed_files"] == []
        assert header["shared_complete"] is True

        project = run_adapter(
            PROJECT_FIXTURE,
            project_path,
            "--project-loading",
            "msbuild",
        )
        assert_valid(project)
        attributes = repository_node(project)["attributes"]
        assert attributes["analysis_mode"] == "msbuild", attributes
        assert attributes["project_count"] >= 3, attributes
        assert_project_reference_resolution(project)
        project_second = run_adapter(
            PROJECT_FIXTURE,
            project_second_path,
            "--project-loading",
            "msbuild",
        )
        assert project_path.read_bytes() == project_second_path.read_bytes()
        assert_project_reference_resolution(project_second)

        fallback = run_adapter(
            PROJECT_FIXTURE,
            fallback_path,
            "--project-loading",
            "files",
        )
        assert_valid(fallback)
        assert repository_node(fallback)["attributes"]["analysis_mode"] == "files"

        sharded_root = directory / "sharded"
        sharded_root.mkdir()
        for index in range(513):
            (sharded_root / f"Type{index:04d}.cs").write_text(
                f"internal static class Type{index:04d} {{ internal static int Value => {index}; }}\n",
                encoding="utf-8",
            )
        sharded = run_adapter(
            sharded_root,
            sharded_path,
            "--project-loading",
            "files",
        )
        assert_valid(sharded)
        sharded_attributes = repository_node(sharded)["attributes"]
        assert sharded_attributes["analysis_mode"] == "files"
        assert sharded_attributes["compilation_count"] == 2, sharded_attributes

    print("C# adapter smoke, MSBuild graph, sharding, determinism, relation, and incremental checks passed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
