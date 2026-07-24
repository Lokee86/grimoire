from __future__ import annotations

import importlib.util
import json
import unittest
from pathlib import Path

MODULE_PATH = Path(__file__).with_name("reconcile_runtime.py")
SPEC = importlib.util.spec_from_file_location("reconcile_runtime", MODULE_PATH)
assert SPEC and SPEC.loader
MODULE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(MODULE)


def identity(char: str) -> str:
    return "sha256:" + char * 64


class RuntimeReconciliationTest(unittest.TestCase):
    def static_records(self):
        source = identity("a")
        definite = identity("b")
        possible = identity("c")
        unmodeled = identity("d")
        records = [
            {"record": "lexicon", "schema_version": 1, "adapter_version": "0.1.0", "language": "go", "repository": "example/repo"},
        ]
        for node_id in (source, definite, possible, unmodeled):
            records.append({"record": "node", "id": node_id, "kind": "function", "name": node_id[-1], "path": "main.go", "qualified_name": node_id[-1]})
        records.extend(
            [
                {"record": "edge", "source": source, "target": definite, "relation": "calls"},
                {"record": "edge", "source": source, "target": possible, "relation": "possible-calls"},
            ]
        )
        return records, source, definite, possible, unmodeled

    def test_classifies_static_and_external_targets(self):
        static, source, definite, possible, unmodeled = self.static_records()
        runtime = [
            {"record": "lexicon-runtime", "schema_version": 1, "repository": "example/repo", "run_id": "run-1"},
            {"record": "observation", "relation": "calls", "source": source, "external_target": "runtime:generated", "count": 7},
            {"record": "observation", "relation": "calls", "source": source, "target": definite, "count": 2},
            {"record": "observation", "relation": "calls", "source": source, "target": possible, "count": 3},
            {"record": "observation", "relation": "calls", "source": source, "target": unmodeled, "count": 5},
        ]
        report = MODULE.reconcile(static, runtime)
        self.assertEqual(
            report["observation_counts"],
            {
                "confirmed-definite": 2,
                "confirmed-possible": 3,
                "external-runtime-target": 7,
                "unmodeled-static-target": 5,
            },
        )

    def test_rejects_unsorted_observations(self):
        _, source, definite, possible, _ = self.static_records()
        runtime = [
            {"record": "lexicon-runtime", "schema_version": 1, "repository": "example/repo", "run_id": "run-1"},
            {"record": "observation", "relation": "writes", "source": source, "target": definite, "count": 1},
            {"record": "observation", "relation": "calls", "source": source, "target": possible, "count": 1},
        ]
        with self.assertRaisesRegex(ValueError, "canonically sorted"):
            MODULE.validate_runtime(runtime)

    def test_rejects_duplicate_aggregate(self):
        _, source, definite, _, _ = self.static_records()
        runtime = [
            {"record": "lexicon-runtime", "schema_version": 1, "repository": "example/repo", "run_id": "run-1"},
            {"record": "observation", "relation": "calls", "source": source, "target": definite, "count": 1},
            {"record": "observation", "relation": "calls", "source": source, "target": definite, "count": 2},
        ]
        with self.assertRaisesRegex(ValueError, "duplicate observation aggregate"):
            MODULE.validate_runtime(runtime)

    def test_rejects_mismatched_repository(self):
        static, source, definite, _, _ = self.static_records()
        runtime = [
            {"record": "lexicon-runtime", "schema_version": 1, "repository": "other/repo", "run_id": "run-1"},
            {"record": "observation", "relation": "calls", "source": source, "target": definite, "count": 1},
        ]
        with self.assertRaisesRegex(ValueError, "repository identities differ"):
            MODULE.reconcile(static, runtime)


if __name__ == "__main__":
    unittest.main()
