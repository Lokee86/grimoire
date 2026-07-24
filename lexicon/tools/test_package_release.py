from __future__ import annotations

import importlib.util
import os
import stat
from pathlib import Path


MODULE_PATH = Path(__file__).with_name("package_release.py")
SPEC = importlib.util.spec_from_file_location("package_release", MODULE_PATH)
assert SPEC is not None and SPEC.loader is not None
package_release = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(package_release)


def test_copy_installers_uses_public_release_names(tmp_path: Path) -> None:
    repo = tmp_path / "repo"
    packaging = repo / "packaging"
    packaging.mkdir(parents=True)
    (packaging / "install.ps1").write_text("powershell installer\n", encoding="utf-8")
    (packaging / "install.sh").write_text("shell installer\n", encoding="utf-8")

    output = tmp_path / "release"
    output.mkdir()
    package_release.copy_installers(repo, output)

    if os.name == "nt":
        assert (output / "install.ps1").read_text(encoding="utf-8") == "powershell installer\n"
        assert not (output / "install.sh").exists()
    else:
        assert (output / "install.sh").read_text(encoding="utf-8") == "shell installer\n"
        assert not (output / "install.ps1").exists()
        assert (output / "install.sh").stat().st_mode & stat.S_IXUSR
