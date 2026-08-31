from __future__ import annotations

import pathlib
import subprocess
import sys
import unittest


ROOT = pathlib.Path(__file__).resolve().parents[1]


def run_cli(*args: str) -> subprocess.CompletedProcess:
    return subprocess.run(
        [sys.executable, str(ROOT / "run_cli.py"), *args],
        capture_output=True,
        text=True,
        cwd=ROOT,
    )


class ChromaticNumberOutputTests(unittest.TestCase):
    def test_path_reports_value_and_every_k_below_it(self) -> None:
        result = run_cli("path", "6")
        self.assertEqual(result.returncode, 0)
        self.assertEqual(
            result.stdout.splitlines(),
            [
                "P_6: chi_g,2 = 4",
                "  k=1: Bob",
                "  k=2: Bob",
                "  k=3: Bob",
                "  k=4: Alice",
            ],
        )

    def test_cycle_label(self) -> None:
        result = run_cli("cycle", "7")
        self.assertEqual(result.returncode, 0)
        self.assertEqual(result.stdout.splitlines()[0], "C_7: chi_g,2 = 4")

    def test_star_label(self) -> None:
        result = run_cli("star", "4")
        self.assertEqual(result.returncode, 0)
        self.assertEqual(result.stdout.splitlines()[0], "K_1,4: chi_g,2 = 5")

    def test_only_the_final_k_is_an_alice_win(self) -> None:
        lines = run_cli("cycle", "6").stdout.splitlines()
        self.assertEqual(lines[0], "C_6: chi_g,2 = 5")
        self.assertTrue(all(line.endswith("Bob") for line in lines[1:-1]))
        self.assertTrue(lines[-1].endswith("Alice"))


class FixedColourCountTests(unittest.TestCase):
    def test_reports_bob(self) -> None:
        result = run_cli("path", "2", "--k", "1")
        self.assertEqual(result.returncode, 0)
        self.assertEqual(result.stdout.strip(), "P_2 with k=1: Bob")

    def test_reports_alice(self) -> None:
        result = run_cli("path", "4", "--k", "3")
        self.assertEqual(result.returncode, 0)
        self.assertEqual(result.stdout.strip(), "P_4 with k=3: Alice")


class ExplainTests(unittest.TestCase):
    def test_bob_win_shows_failed_line_and_dead_vertex(self) -> None:
        result = run_cli("path", "2", "--k", "1", "--explain")
        self.assertEqual(result.returncode, 0)
        self.assertEqual(
            result.stdout.splitlines(),
            [
                "P_2 with k=1: Bob",
                "one failed line:",
                "  Alice: v0 -> color 1",
                "dead vertex set: v1",
            ],
        )

    def test_alice_win_shows_opening_move_and_continuation(self) -> None:
        result = run_cli("path", "4", "--k", "3", "--explain")
        self.assertEqual(result.returncode, 0)
        lines = result.stdout.splitlines()
        self.assertEqual(lines[0], "P_4 with k=3: Alice")
        self.assertTrue(lines[1].startswith("winning opening move: Alice colors v"))
        self.assertEqual(lines[2], "sample continuation:")
        self.assertEqual(lines[-1], "final coloring completed successfully")
        moves = lines[3:-1]
        self.assertEqual(len(moves), 4)
        for index, line in enumerate(moves):
            self.assertTrue(line.startswith("  Alice: " if index % 2 == 0 else "  Bob: "))

    def test_explain_is_ignored_without_k(self) -> None:
        # --explain only applies to a fixed-k run; the sweep output is unchanged.
        with_flag = run_cli("path", "4", "--explain").stdout
        without_flag = run_cli("path", "4").stdout
        self.assertEqual(with_flag, without_flag)


class ArgumentHandlingTests(unittest.TestCase):
    def test_missing_subcommand_exits_with_usage_error(self) -> None:
        result = run_cli()
        self.assertEqual(result.returncode, 2)
        self.assertIn("usage:", result.stderr)

    def test_unknown_family_exits_with_usage_error(self) -> None:
        result = run_cli("wheel", "5")
        self.assertEqual(result.returncode, 2)

    def test_non_integer_size_exits_with_usage_error(self) -> None:
        result = run_cli("path", "abc")
        self.assertEqual(result.returncode, 2)


if __name__ == "__main__":
    unittest.main()
