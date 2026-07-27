#!/usr/bin/env python3
"""Deterministic smoke checks for the root release workflow."""

from __future__ import annotations

import hashlib
import tempfile
import unittest
import zipfile
from pathlib import Path

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

    def test_version_validation_rejects_path_values(self) -> None:
        with self.assertRaises(ValueError):
            workflow.validate_version("../outside")


if __name__ == "__main__":
    run_smoke()
