from __future__ import annotations

import json
import pathlib
import subprocess
import sys
import unittest


ROOT = pathlib.Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / "src"))


class WebExportTests(unittest.TestCase):
    def test_export_script_writes_expected_cases(self) -> None:
        subprocess.run(
            [sys.executable, str(ROOT / "scripts" / "export_web_cases.py")],
            check=True,
            cwd=ROOT,
        )

        payload = json.loads((ROOT / "web" / "data" / "cases.json").read_text(encoding="utf-8"))
        case_map = {item["id"]: item for item in payload["cases"]}

        self.assertIn("path-p6-k4", case_map)
        self.assertIn("cycle-c7-k4", case_map)
        self.assertEqual(case_map["path-p2-k1"]["initialWinner"], "Bob")
        self.assertEqual(case_map["path-p6-k4"]["initialWinner"], "Alice")
        self.assertEqual(case_map["cycle-c6-k5"]["initialWinner"], "Alice")
        self.assertEqual(case_map["cycle-c5-k4"]["initialWinner"], "Bob")

        start_state = case_map["path-p3-k3"]["states"]["0.0.0|A"]
        self.assertTrue(start_state["aliceWins"])
        self.assertGreaterEqual(len(start_state["legalMoves"]), 1)


if __name__ == "__main__":
    unittest.main()
