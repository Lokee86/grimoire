from __future__ import annotations

import json
from pathlib import Path
import tempfile
import unittest

from benchmark_grounding import summarize_audit, validate_answer


class BenchmarkGroundingTests(unittest.TestCase):
    def test_validates_paths_ranges_and_grimoire_handles(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            (root / "target.go").write_text("package fixture\n\nfunc Target() {}\n", encoding="utf-8")
            audit = root / "audit.jsonl"
            audit.write_text(json.dumps({
                "response": {
                    "delta": {
                        "new_source_ranges": [{
                            "handle": "g1_range",
                            "evidence": {"path": "target.go", "start_line": 3, "end_line": 3},
                        }],
                    },
                },
            }) + "\n", encoding="utf-8")
            answer = (
                "Target is declared here at target.go:3 and repeated as `target.go:3`.\n"
                'BENCHMARK_EVIDENCE_JSON:{"ownership":[{"path":"target.go","symbol":"Target",'
                '"lines":"3","claim":"Target owns the behavior","handle":"g1_range"}]}\n'
            )
            report = validate_answer(
                root,
                answer,
                exit_code=0,
                expected_sections=["ownership"],
                audit_log=audit,
                require_grimoire_handles=True,
                required_path_prefixes=["target.go", "missing/"],
            )
            self.assertTrue(report.valid, report.to_dict())
            self.assertEqual(report.required_prefixes_found, ["target.go"])
            self.assertEqual(report.missing_required_prefixes, ["missing"])
            self.assertEqual(report.inline_citations, 2)
            self.assertEqual(report.canonical_handle_items, 1)
            summary = summarize_audit(audit)
            self.assertEqual(summary["calls"], 1)
            self.assertEqual(summary["new_source_ranges"], 1)
            self.assertGreater(summary["response_bytes"], 0)

    def test_rejects_missing_files_bad_ranges_and_mismatched_handles(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            (root / "target.go").write_text("one\ntwo\n", encoding="utf-8")
            audit = root / "audit.jsonl"
            audit.write_text(json.dumps({
                "response": {
                    "delta": {
                        "new_source_ranges": [{
                            "handle": "g1_range",
                            "evidence": {"path": "target.go", "start_line": 1, "end_line": 1},
                        }],
                    },
                },
            }) + "\n", encoding="utf-8")
            answer = (
                "Bad citations. `missing.go:1` `target.go:1-9`\n"
                'BENCHMARK_EVIDENCE_JSON:{"ownership":[{"path":"target.go","symbol":"Target",'
                '"lines":"2","claim":"wrong range","handle":"g1_range"}]}\n'
            )
            report = validate_answer(
                root,
                answer,
                exit_code=0,
                expected_sections=["ownership"],
                audit_log=audit,
                require_grimoire_handles=True,
            )
            self.assertFalse(report.valid)
            codes = {finding.code for finding in report.findings}
            self.assertIn("inline_citation_missing_path", codes)
            self.assertIn("inline_citation_out_of_range", codes)
            self.assertIn("handle_range_mismatch", codes)

    def test_accepts_non_session_inspection_handles(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            (root / "target.go").write_text("one\ntwo\nthree\n", encoding="utf-8")
            audit = root / "audit.jsonl"
            audit.write_text(json.dumps({
                "request": {"mode": "inspect"},
                "response": {"inspections": [{
                    "containing_span": {
                        "path": "target.go", "start_line": 2, "end_line": 3,
                        "handle": {"value": "grimoire:v1:range"},
                    },
                }]},
            }) + "\n", encoding="utf-8")
            answer = (
                'BENCHMARK_EVIDENCE_JSON:{"evidence":[{"path":"target.go","symbol":"Target",'
                '"lines":"2-3","claim":"canonical inspection","handle":"grimoire:v1:range"}]}\n'
            )
            report = validate_answer(
                root,
                answer,
                exit_code=0,
                expected_sections=["evidence"],
                audit_log=audit,
            )
            self.assertTrue(report.valid, report.to_dict())
            self.assertEqual(report.canonical_handle_items, 1)

    def test_reports_malformed_audit_logs(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            (root / "target.go").write_text("one\n", encoding="utf-8")
            audit = root / "audit.jsonl"
            audit.write_text("not-json\n", encoding="utf-8")
            answer = (
                'BENCHMARK_EVIDENCE_JSON:{"evidence":[{"path":"target.go","symbol":"Target",'
                '"lines":"1","claim":"source evidence"}]}\n'
            )
            report = validate_answer(
                root,
                answer,
                exit_code=0,
                expected_sections=["evidence"],
                audit_log=audit,
            )
            self.assertFalse(report.valid)
            self.assertIn("invalid_audit_log", {finding.code for finding in report.findings})
            summary = summarize_audit(audit)
            self.assertEqual(summary["invalid_records"], 1)

    def test_rejects_refusal_and_empty_sections(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            report = validate_answer(
                Path(temporary),
                "I cannot complete this investigation.\nBENCHMARK_EVIDENCE_JSON:{\"ownership\":[]}",
                exit_code=0,
                expected_sections=["ownership"],
            )
            self.assertFalse(report.valid)
            self.assertTrue(report.refusal_detected)
            self.assertIn("missing_evidence_section", {finding.code for finding in report.findings})


if __name__ == "__main__":
    unittest.main()
