from __future__ import annotations

import json
import subprocess
import tempfile
import time
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]

def run(binary: Path, *arguments: str) -> subprocess.CompletedProcess[str]:
    result = subprocess.run(
        [str(binary), *arguments],
        check=False,
        capture_output=True,
        text=True,
    )
    if result.returncode != 0:
        raise RuntimeError(result.stdout + result.stderr)
    return result


def build_binary(directory: Path) -> Path:
    binary = directory / "lexicon.exe"
    result = subprocess.run(
        ["go", "build", "-o", str(binary), "./cmd/lexicon"],
        cwd=ROOT,
        check=False,
        capture_output=True,
        text=True,
    )
    if result.returncode != 0:
        raise RuntimeError(result.stdout + result.stderr)
    return binary


def wait_for_change(path: Path, previous: bytes, timeout: float = 10.0) -> None:
    deadline = time.monotonic() + timeout
    while time.monotonic() < deadline:
        if path.exists() and path.read_bytes() != previous:
            return
        time.sleep(0.1)
    raise RuntimeError(f"file did not change: {path}")


def load_snapshot(repository: Path) -> tuple[str, dict[str, object]]:
    root = repository / ".lexicon"
    snapshot_id = (root / "CURRENT").read_text(encoding="utf-8").strip()
    manifest_path = root / "snapshots" / f"{snapshot_id.removeprefix('sha256:')}.json"
    manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
    return snapshot_id, manifest


def validate_snapshot(repository: Path) -> tuple[str, dict[str, object]]:
    snapshot_id, manifest = load_snapshot(repository)
    root = repository / ".lexicon"
    for language in manifest["languages"]:
        object_ids = [file["object_id"] for file in language["files"]]
        shared = language.get("shared_object_id")
        if shared:
            object_ids.append(shared)
        for object_id in object_ids:
            digest = object_id.removeprefix("sha256:")
            if not (root / "objects" / digest[:2] / digest[2:]).is_file():
                raise RuntimeError(f"missing fact object: {object_id}")
    return snapshot_id, manifest


def file_objects(manifest: dict[str, object]) -> dict[tuple[str, str], str]:
    result: dict[tuple[str, str], str] = {}
    for language in manifest["languages"]:
        for file in language["files"]:
            result[(language["language"], file["path"])] = file["object_id"]
    return result


def main() -> int:
    with tempfile.TemporaryDirectory(prefix="lexicon-smoke-") as directory:
        temporary = Path(directory)
        binary = build_binary(temporary)
        repository = temporary / "repository"
        repository.mkdir()
        source = repository / "main.py"
        source.write_text("def answer():\n    return 42\n", encoding="utf-8")
        (repository / "helper.py").write_text("def helper():\n    return 7\n", encoding="utf-8")
        (repository / "main.rb").write_text("def answer = 42\n", encoding="utf-8")
        (repository / "main.gd").write_text("extends Node\nfunc answer():\n    return 42\n", encoding="utf-8")
        (repository / "main.go").write_text("package smoke\n\nfunc Answer() int { return 42 }\n", encoding="utf-8")
        (repository / "go.mod").write_text("module example.com/smoke\n\ngo 1.26\n", encoding="utf-8")
        (repository / "main.c").write_text("int answer(void) { return 42; }\n", encoding="utf-8")
        (repository / "app.ts").write_text("export function answer(): number { return 42; }\n", encoding="utf-8")
        (repository / "package.json").write_text('{"name":"lexicon-smoke","private":true}\n', encoding="utf-8")
        (repository / "src").mkdir()
        (repository / "src" / "lib.rs").write_text("pub fn answer() -> i32 { 42 }\n", encoding="utf-8")
        (repository / "Cargo.toml").write_text(
            '[package]\nname = "lexicon-smoke"\nversion = "0.1.0"\nedition = "2024"\n',
            encoding="utf-8",
        )

        run(binary, "init", "--repo", str(repository), "--adapters", str(ROOT / "adapters"))
        library_root = repository / ".lexicon" / "repo" / "library"
        if library_root.exists():
            raise RuntimeError("normal initialization retained a materialized JSONL library")
        initial_id, initial_manifest = validate_snapshot(repository)
        languages = {entry["language"] for entry in initial_manifest["languages"]}
        expected = {"c-family", "gdscript", "go", "python", "ruby", "rust", "typescript"}
        if languages != expected:
            raise RuntimeError(f"unexpected snapshot languages: {sorted(languages)}")

        source.write_text("def answer():\n    return 43\n", encoding="utf-8")
        updated = run(binary, "scan", "--repo", str(repository))
        if "analyzing python files: 1" not in updated.stdout or "updated 1 files: python" not in updated.stdout or "expanding python to full analysis" in updated.stdout:
            raise RuntimeError(f"unexpected incremental scan output: {updated.stdout}")
        updated_id, updated_manifest = validate_snapshot(repository)
        if updated_id == initial_id:
            raise RuntimeError("source change did not publish a new snapshot")
        initial_objects = file_objects(initial_manifest)
        updated_objects = file_objects(updated_manifest)
        if initial_objects[("ruby", "main.rb")] != updated_objects[("ruby", "main.rb")]:
            raise RuntimeError("unchanged Ruby fact object was not reused")
        if initial_objects[("python", "helper.py")] != updated_objects[("python", "helper.py")]:
            raise RuntimeError("unchanged Python fact object was not reused")

        incremental_updates = (
            ("c-family", repository / "main.c", "int answer(void) { return 43; }\n"),
            ("ruby", repository / "main.rb", "def answer = 43\n"),
            ("gdscript", repository / "main.gd", "extends Node\nfunc answer():\n    return 43\n"),
            ("go", repository / "main.go", "package smoke\n\nfunc Answer() int { return 43 }\n"),
            ("typescript", repository / "app.ts", "export function answer(): number { return 43; }\n"),
            ("rust", repository / "src" / "lib.rs", "pub fn answer() -> i32 { 43 }\n"),
        )
        for language, path, content in incremental_updates:
            path.write_text(content, encoding="utf-8")
            result = run(binary, "scan", "--repo", str(repository))
            if f"analyzing {language} files: 1" not in result.stdout or f"expanding {language} to full analysis" in result.stdout:
                raise RuntimeError(f"{language} did not stay incremental: {result.stdout}")
            updated_id, _ = validate_snapshot(repository)

        current_path = repository / ".lexicon" / "CURRENT"
        before_snapshot = current_path.read_bytes()
        demon = subprocess.Popen(
            [
                str(binary),
                "demon",
                "--repo",
                str(repository),
                "--debounce",
                "50ms",
                "--reconcile",
                "5s",
            ],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
        )
        try:
            time.sleep(0.5)
            source.write_text("def answer():\n    return 44\n", encoding="utf-8")
            wait_for_change(current_path, before_snapshot)
            validate_snapshot(repository)
            if library_root.exists():
                raise RuntimeError("demon recreated a materialized JSONL library")
        finally:
            demon.terminate()
            try:
                demon.wait(timeout=5)
            except subprocess.TimeoutExpired:
                demon.kill()
                demon.wait(timeout=5)

        state = repository / ".lexicon" / "repo"
        count = subprocess.run(
            ["git", "rev-list", "--count", "HEAD"],
            cwd=state,
            check=True,
            capture_output=True,
            text=True,
        ).stdout.strip()
        if count != "1":
            raise RuntimeError(f"expected one reachable state commit, got {count}")

    print("Lexicon application smoke test passed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
