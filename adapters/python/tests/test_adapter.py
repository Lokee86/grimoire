from __future__ import annotations

import hashlib
import json
import os
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


ADAPTER_ROOT = Path(__file__).resolve().parents[1]
REPO_ROOT = ADAPTER_ROOT.parents[1]


class PythonAdapterTest(unittest.TestCase):
    def setUp(self) -> None:
        self.tempdir = tempfile.TemporaryDirectory()
        self.repo = Path(self.tempdir.name) / "fixture-repository"
        (self.repo / "pkg").mkdir(parents=True)
        (self.repo / "vendor").mkdir()
        (self.repo / "build").mkdir()
        (self.repo / "__pycache__").mkdir()
        self._write("pkg/__init__.py", "from .base import Base\n")
        self._write(
            "pkg/base.py",
            "class Base:\n"
            "    def run(self):\n"
            "        return 1\n",
        )
        self._write(
            "pkg/known.py",
            "class Worker:\n"
            "    def run(self):\n"
            "        return 1\n",
        )
        self._write(
            "pkg/other.py",
            "class Alternate:\n"
            "    def run(self):\n"
            "        return 1\n",
        )
        self._write(
            "pkg/child.py",
            "from .base import Base\n"
            "import pkg.base as base\n"
            "class Child(Base):\n"
            "    def run(self):\n"
            "        return base.Base()\n",
        )
        self._write(
            "main.py",
            "from pkg.child import Child\n"
            "class Root(Child):\n"
            "    pass\n\n"
            "def factory():\n"
            "    return Root\n\n"
            "def use():\n"
            "    return factory()\n",
        )
        self._write("dynamic.py", "import importlib\nmodule = importlib.import_module(name)\n")
        self._write("vendor/ignored.py", "class Ignored: pass\n")
        self._write("build/ignored.py", "def ignored(): pass\n")
        self._write("__pycache__/ignored.py", "def ignored(): pass\n")
        self._write("tools/data_sync/config.py", "class NestedConfig:\n    pass\n")
        self._write(
            "nested_consumer.py",
            "from data_sync.config import NestedConfig\n\n"
            "def build_nested():\n"
            "    return NestedConfig()\n",
        )

    def tearDown(self) -> None:
        self.tempdir.cleanup()

    def _write(self, relative: str, content: str) -> None:
        path = self.repo / relative
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(content, encoding="utf-8")

    def _run(self, output: Path) -> list[dict[str, object]]:
        environment = os.environ.copy()
        environment["PYTHONPATH"] = str(ADAPTER_ROOT)
        result = subprocess.run(
            [sys.executable, "-m", "lexicon_python", "--repo", str(self.repo), "--output", str(output)],
            cwd=REPO_ROOT,
            env=environment,
            check=True,
            capture_output=True,
            text=True,
        )
        self.assertEqual(result.stderr, "")
        return [json.loads(line) for line in output.read_text(encoding="utf-8").splitlines()]

    def test_declarations_imports_inheritance_and_exclusions(self) -> None:
        records = self._run(self.repo / "facts.jsonl")
        nodes = [record for record in records if record["record"] == "node"]
        edges = [record for record in records if record["record"] == "edge"]
        unresolved = [record for record in records if record["record"] == "unresolved"]

        kinds = {record["kind"] for record in nodes}
        self.assertTrue({"repository", "directory", "file", "module", "type", "function", "method", "import"} <= kinds)
        paths = {record["path"] for record in nodes}
        self.assertNotIn("vendor/ignored.py", paths)
        self.assertNotIn("build/ignored.py", paths)
        self.assertNotIn("__pycache__/ignored.py", paths)

        relations = {record["relation"] for record in edges}
        self.assertTrue({"contains", "defines", "imports", "extends"} <= relations)
        self.assertTrue(any(record["relation"] == "extends" for record in edges))
        self.assertTrue(any(record["relation"] == "imports" for record in edges))
        self.assertTrue(any(record["relation"] == "calls" for record in unresolved))
        self.assertTrue(any(record["reason"] == "dynamic-target" for record in unresolved))

    def test_nested_package_prefix_imports_resolve(self) -> None:
        records = self._run(self.repo / "facts.jsonl")
        nodes = [record for record in records if record["record"] == "node"]
        edges = [record for record in records if record["record"] == "edge"]
        consumer_module = next(
            record["id"]
            for record in nodes
            if record["kind"] == "module" and record["path"] == "nested_consumer.py"
        )
        nested_type = next(
            record["id"]
            for record in nodes
            if record["kind"] == "type" and record["qualified_name"] == "tools.data_sync.config.NestedConfig"
        )
        builder = next(
            record["id"]
            for record in nodes
            if record["kind"] == "function" and record["qualified_name"] == "nested_consumer.build_nested"
        )

        self.assertTrue(
            any(
                record["relation"] == "imports"
                and record["source"] == consumer_module
                and record["target"] == nested_type
                for record in edges
            )
        )
        self.assertTrue(
            any(
                record["relation"] == "calls"
                and record["source"] == builder
                and record["target"] == nested_type
                for record in edges
            )
        )

    def test_local_constructor_calls_are_precise_and_conservative(self) -> None:
        self._write("left/config.py", "class Worker:\n    def run(self): pass\n")
        self._write("right/config.py", "class Worker:\n    def run(self): pass\n")
        self._write(
            "local_precision.py",
            "from pkg.known import Worker\n"
            "from pkg.other import Alternate\n"
            "from config import Worker as AmbiguousWorker\n\n"
            "def precise():\n"
            "    worker = Worker()\n"
            "    return worker.run()\n\n"
            "def branch_dependent(flag):\n"
            "    worker = Worker()\n"
            "    if flag:\n"
            "        worker = Worker()\n"
            "    return worker.run()\n\n"
            "def conflicting():\n"
            "    worker = Worker()\n"
            "    worker = Alternate()\n"
            "    return worker.run()\n\n"
            "def attribute_based():\n"
            "    holder.worker = Worker()\n"
            "    return holder.worker.run()\n\n"
            "def unknown_method():\n"
            "    worker = Worker()\n"
            "    return worker.missing()\n\n"
            "def ambiguous_class():\n"
            "    worker = AmbiguousWorker()\n"
            "    return worker.run()\n",
        )
        records = self._run(self.repo / "facts.jsonl")
        nodes = [record for record in records if record["record"] == "node"]
        edges = [record for record in records if record["record"] == "edge"]
        unresolved = [record for record in records if record["record"] == "unresolved"]
        worker_run = next(record for record in nodes if record.get("qualified_name") == "pkg.known.Worker.run")
        alternate_run = next(record for record in nodes if record.get("qualified_name") == "pkg.other.Alternate.run")
        function_ids = {
            record["qualified_name"].rsplit(".", 1)[-1]: record["id"]
            for record in nodes
            if record["kind"] == "function" and record["path"] == "local_precision.py"
        }

        self.assertTrue(
            any(
                record["relation"] == "calls"
                and record["source"] == function_ids["precise"]
                and record["target"] == worker_run["id"]
                for record in edges
            )
        )
        for function_name in ("branch_dependent", "conflicting", "attribute_based", "unknown_method", "ambiguous_class"):
            self.assertFalse(
                any(
                    record["relation"] == "calls"
                    and record["source"] == function_ids[function_name]
                    and record["target"] in {worker_run["id"], alternate_run["id"]}
                    for record in edges
                )
            )
        self.assertTrue(any(record["expression"] == "worker.run()" and record["reason"] == "ambiguous-target" for record in unresolved))
        self.assertTrue(any(record["expression"] == "holder.worker.run()" for record in unresolved))
        self.assertTrue(any(record["expression"] == "worker.missing()" and record["reason"] == "missing-target" for record in unresolved))

    def test_runtime_builtin_namespace_is_classified(self) -> None:
        self._write(
            "builtins_usage.py",
            "def use(values):\n"
            "    sorted(values)\n"
            "    frozenset(values)\n"
            "    KeyError('missing')\n"
            "    Exception('failed')\n"
            "    SystemExit(1)\n",
        )
        records = self._run(self.repo / "facts.jsonl")
        unresolved = [
            record
            for record in records
            if record["record"] == "unresolved"
            and record["relation"] == "calls"
            and record.get("span", {}).get("path") == "builtins_usage.py"
        ]
        self.assertEqual(len(unresolved), 5)
        self.assertEqual({record["reason"] for record in unresolved}, {"builtin-target"})

    def test_ids_content_and_header(self) -> None:
        output = self.repo / "facts.jsonl"
        records = self._run(output)
        header = records[0]
        self.assertEqual(header, {
            "adapter_version": "0.2.0",
            "language": "python",
            "record": "lexicon",
            "repository": "fixture-repository",
            "schema_version": 1,
        })
        file_data = (self.repo / "main.py").read_bytes()
        expected_content = "sha256:" + hashlib.sha256(file_data).hexdigest()
        expected_node = "sha256:" + hashlib.sha256(
            b"lexicon:v1\0python\0file\0main.py"
        ).hexdigest()
        main_node = next(record for record in records if record.get("path") == "main.py" and record.get("kind") == "file")
        self.assertEqual(main_node["id"], expected_node)
        self.assertEqual(main_node["content_id"], expected_content)
        self.assertTrue(all(record["record"] == "lexicon" or "D:\\" not in json.dumps(record) for record in records))

    def test_deterministic_ordering_and_repeat_runs(self) -> None:
        first = self.repo / "first.jsonl"
        second = self.repo / "second.jsonl"
        first_records = self._run(first)
        second_records = self._run(second)
        self.assertEqual(first.read_bytes(), second.read_bytes())
        self.assertEqual(first_records[0]["record"], "lexicon")
        records = first_records[1:]
        phases = [0 if item["record"] == "node" else 1 if item["record"] == "edge" else 2 for item in records]
        self.assertEqual(phases, sorted(phases))
        self.assertEqual(len({item["id"] for item in records if item["record"] == "node"}), len([item for item in records if item["record"] == "node"]))

    def test_contract_validator_accepts_cli_output(self) -> None:
        output = self.repo / "facts.jsonl"
        self._run(output)
        result = subprocess.run(
            [sys.executable, str(REPO_ROOT / "tools" / "validate_jsonl.py"), str(output)],
            cwd=REPO_ROOT,
            capture_output=True,
            text=True,
        )
        self.assertEqual(result.returncode, 0, result.stderr)


if __name__ == "__main__":
    unittest.main()
