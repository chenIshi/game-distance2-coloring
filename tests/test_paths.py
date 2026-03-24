from __future__ import annotations

import pathlib
import sys
import unittest

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parents[1] / "src"))

from game_coloring import alice_wins, analyze_game, game_distance_2_chromatic_number, make_path_graph


class PathGraphTests(unittest.TestCase):
    def test_small_known_values(self) -> None:
        self.assertEqual(game_distance_2_chromatic_number(make_path_graph(1)), 1)
        self.assertEqual(game_distance_2_chromatic_number(make_path_graph(2)), 2)
        self.assertEqual(game_distance_2_chromatic_number(make_path_graph(3)), 3)

    def test_losing_below_known_thresholds(self) -> None:
        self.assertFalse(alice_wins(make_path_graph(1), 0))
        self.assertFalse(alice_wins(make_path_graph(2), 1))
        self.assertFalse(alice_wins(make_path_graph(3), 2))

    def test_bob_analysis_contains_dead_vertex(self) -> None:
        analysis = analyze_game(make_path_graph(2), 1)
        self.assertEqual(analysis.winner, "Bob")
        self.assertEqual(analysis.dead_vertices, (1,))
        self.assertEqual(len(analysis.example_line), 1)

    def test_alice_analysis_contains_opening_move(self) -> None:
        analysis = analyze_game(make_path_graph(4), 3)
        self.assertEqual(analysis.winner, "Alice")
        self.assertIsNotNone(analysis.winning_opening_move)
        self.assertGreaterEqual(len(analysis.example_line), 1)


if __name__ == "__main__":
    unittest.main()
