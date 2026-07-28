from __future__ import annotations

from pathlib import Path
import unittest

from benchmark_tasks import SUPPORTED_CATEGORIES, load_task_suite


class BenchmarkTaskSuiteTests(unittest.TestCase):
    def test_version_two_suite_covers_intended_task_classes_without_prompt_checklists(self) -> None:
        workspace = Path(__file__).resolve().parents[2]
        suite = load_task_suite(
            Path(__file__).with_name("agent_benchmark_tasks.v2.json"),
            workspace,
        )
        tasks = suite["tasks"]
        self.assertEqual({task["category"] for task in tasks}, SUPPORTED_CATEGORIES)
        self.assertEqual(suite["evidence_sections"], ["evidence"])
        for task in tasks:
            prompt = task["prompt"]
            self.assertNotIn("Deliver:", prompt)
            self.assertNotIn("BENCHMARK_EVIDENCE_JSON", prompt)
            self.assertTrue(task["rubric"]["dimensions"])
            self.assertTrue(task["rubric"]["required_evidence"])
            self.assertTrue(task["repo"].is_dir())


if __name__ == "__main__":
    unittest.main()
