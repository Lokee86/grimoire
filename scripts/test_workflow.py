#!/usr/bin/env python3
"""Deterministic smoke checks for the root release workflow."""

from __future__ import annotations

import hashlib
import tempfile
import unittest
import zipfile
from pathlib import Path
from unittest import mock

import workflow


def run_smoke() -> None:
    suite = unittest.defaultTestLoader.loadTestsFromTestCase(WorkflowSmokeTests)
    result = unittest.TextTestRunner(verbosity=2).run(suite)
    if not result.wasSuccessful():
        raise RuntimeError("workflow smoke checks failed")


class WorkflowSmokeTests(unittest.TestCase):
    def test_windows_package_and_install_layout(self) -> None:
        with tempfile.TemporaryDirectory(prefix="grimoire-workflow-smoke-") as temporary:
            root = Path(temporary)
            build = root / "build"
            (build / "bin").mkdir(parents=True)
            (build / "native").mkdir()
            for name in ("grimoire.exe", "lexicon.exe", "arcana.exe"):
                (build / "bin" / name).write_bytes(name.encode())
            (build / "native" / "lodestone_ffi.dll").write_bytes(b"dll")

            release_root = workflow.package_artifacts(
                build, root / "dist", "1.2.3", platform_name="Windows", machine="AMD64"
            )
            self.assertTrue((release_root / "SHA256SUMS.txt").is_file())
            checksum_text = (release_root / "SHA256SUMS.txt").read_bytes()
            self.assertNotIn(b"\r", checksum_text)
            archives = sorted(release_root.glob("*.zip"))
            self.assertEqual(len(archives), 4)
            for line in checksum_text.decode().splitlines():
                digest, _, name = line.partition("  ")
                self.assertEqual(digest, hashlib.sha256((release_root / name).read_bytes()).hexdigest())
            combined = release_root / "grimoire-bundle-1.2.3-windows-x86_64.zip"
            with zipfile.ZipFile(combined) as archive:
                self.assertIn("bin/grimoire.exe", archive.namelist())
                self.assertIn("native/lodestone_ffi.dll", archive.namelist())
                self.assertIn("install.py", archive.namelist())
                self.assertEqual((archive.getinfo("bin/grimoire.exe").external_attr >> 16) & 0o777, 0o755)
                self.assertEqual((archive.getinfo("install.py").external_attr >> 16) & 0o777, 0o755)

            installed = root / "selected-bin"
            workflow.install(build, installed)
            self.assertEqual((installed / "grimoire.exe").read_bytes(), b"grimoire.exe")
            self.assertTrue((installed / "lodestone_ffi.dll").is_file())

            subset = root / "lexicon-only"
            workflow.install(build, subset, ("lexicon",))
            self.assertTrue((subset / "lexicon.exe").is_file())
            self.assertFalse((subset / "grimoire.exe").exists())
            self.assertFalse((subset / "lodestone_ffi.dll").exists())

    def test_component_tests_are_cpu_bounded_by_default(self) -> None:
        calls: list[tuple[list[str], Path, dict[str, str] | None]] = []

        def record(command: list[str], cwd: Path, env: dict[str, str] | None = None) -> None:
            calls.append((list(command), cwd, env))

        with mock.patch.object(workflow, "cargo_command", return_value="cargo"), \
                mock.patch.object(workflow, "run", side_effect=record):
            workflow.test()

        self.assertEqual(len(calls), 3)
        self.assertEqual(calls[0][0], ["go", "test", "-p", "1", "-parallel", "1", "./..."])
        self.assertEqual(calls[1][0], ["go", "test", "-p", "1", "-parallel", "1", "./..."])
        self.assertEqual(
            calls[2][0],
            [
                "cargo", "test", "--jobs", "1", "--all-targets", "--locked",
                "--manifest-path", str(workflow.ROOT / "arcana" / "Cargo.toml"),
                "--", "--test-threads", "1",
            ],
        )
        for _, _, environment in calls:
            self.assertIsNotNone(environment)
            self.assertEqual(environment["GOMAXPROCS"], "1")
            self.assertEqual(environment["CARGO_BUILD_JOBS"], "1")
            self.assertEqual(environment["RUST_TEST_THREADS"], "1")

    def test_release_jobs_default_and_override(self) -> None:
        default = workflow.parse_args(["release", "--version", "1.2.3"])
        self.assertEqual(default.jobs, 1)
        overridden = workflow.parse_args(["release", "--version", "1.2.3", "--jobs", "3"])
        self.assertEqual(overridden.jobs, 3)
        with self.assertRaises(ValueError):
            workflow.validate_jobs(0)

    def test_version_validation_rejects_path_values(self) -> None:
        with self.assertRaises(ValueError):
            workflow.validate_version("../outside")


if __name__ == "__main__":
    run_smoke()
