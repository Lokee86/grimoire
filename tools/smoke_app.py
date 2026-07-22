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
        (repository / "main.rb").write_text("def answer = 42\n", encoding="utf-8")
        (repository / "main.gd").write_text("extends Node\nfunc answer():\n    return 42\n", encoding="utf-8")
        (repository / "main.go").write_text("package smoke\n\nfunc Answer() int { return 42 }\n", encoding="utf-8")
        (repository / "go.mod").write_text("module example.com/smoke\n\ngo 1.26\n", encoding="utf-8")
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
        for language in ("gdscript", "go", "python", "ruby", "rust", "typescript"):
            library = library_root / f"{language}.jsonl"
            records = [json.loads(line) for line in library.read_text(encoding="utf-8").splitlines()]
            if not records or records[0].get("record") != "lexicon":
                raise RuntimeError(f"{language} library did not contain a Lexicon header")
        library = library_root / "python.jsonl"
        initial_id, initial_manifest = validate_snapshot(repository)

        source.write_text("def answer():\n    return 43\n", encoding="utf-8")
        updated = run(binary, "scan", "--repo", str(repository))
        if "updated 1 files: python" not in updated.stdout:
            raise RuntimeError(f"unexpected incremental scan output: {updated.stdout}")
        updated_id, updated_manifest = validate_snapshot(repository)
        if updated_id == initial_id:
            raise RuntimeError("source change did not publish a new snapshot")
        initial_objects = file_objects(initial_manifest)
        updated_objects = file_objects(updated_manifest)
        if initial_objects[("ruby", "main.rb")] != updated_objects[("ruby", "main.rb")]:
            raise RuntimeError("unchanged Ruby fact object was not reused")

        current_path = repository / ".lexicon" / "CURRENT"
        current_path.write_text(initial_id + "\n", encoding="utf-8")
        repaired = run(binary, "scan", "--repo", str(repository))
        if "Lexicon is current" not in repaired.stdout:
            raise RuntimeError(f"unexpected recovery scan output: {repaired.stdout}")
        repaired_id, _ = validate_snapshot(repository)
        if repaired_id != updated_id:
            raise RuntimeError("stale CURRENT pointer was not repaired")

        current_path.unlink()
        run(binary, "scan", "--repo", str(repository))
        restored_id, _ = validate_snapshot(repository)
        if restored_id != updated_id:
            raise RuntimeError("missing CURRENT pointer was not restored")

        before_daemon = library.read_bytes()
        before_snapshot = current_path.read_bytes()
        daemon = subprocess.Popen(
            [
                str(binary),
                "daemon",
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
            wait_for_change(library, before_daemon)
            wait_for_change(current_path, before_snapshot)
            validate_snapshot(repository)
        finally:
            daemon.terminate()
            try:
                daemon.wait(timeout=5)
            except subprocess.TimeoutExpired:
                daemon.kill()
                daemon.wait(timeout=5)

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
